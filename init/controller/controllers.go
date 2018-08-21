package controller

import (
	"net/http"
	"fmt"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/myanrichal/gocms/context"
	"github.com/myanrichal/gocms/domain/acl/authentication/authentication_controller"
	"github.com/myanrichal/gocms/domain/acl/authentication/authentication_middleware"
	"github.com/myanrichal/gocms/domain/acl/cors"
	"github.com/myanrichal/gocms/domain/content/documentation"
	"github.com/myanrichal/gocms/domain/content/react"
	"github.com/myanrichal/gocms/domain/content/template"
	"github.com/myanrichal/gocms/domain/content/theme"
	"github.com/myanrichal/gocms/domain/email/email_controller"
	"github.com/myanrichal/gocms/domain/health/health_controller"
	"github.com/myanrichal/gocms/domain/health/health_middleware"
	"github.com/myanrichal/gocms/domain/plugin/plugin_services"
	"github.com/myanrichal/gocms/domain/user/user_admin_controller"
	"github.com/myanrichal/gocms/domain/user/user_controller"
	"github.com/myanrichal/gocms/domain/user/user_middleware"
	"github.com/myanrichal/gocms/init/service"
	"github.com/myanrichal/gocms/routes"
)

type ControllersGroup struct {
	Routes             *routes.Routes
	ApiControllers     *ApiControllers
	ContentControllers *ContentControllers
}

type ContentControllers struct {
	DocumentationController *documentation_controller.DocumentationController
	TemplateControllers     *template_controller.TemplatesController
	ThemesControllers       *theme_controller.ThemesController
	ReactControllers        *react_controller.ReactController
}

type ApiControllers struct {
	AuthController      *authentication_controller.AuthController
	HealthyController   *health_controller.HealthController
	AdminUserController *user_admin_controller.UserAdminController
	UserController      *user_controller.UserController
	EmailController     *email_controller.EmailController
}

var (
	defaultRoutePrefix = "/api"
)

func DefaultControllerGroup(r *gin.Engine, sg *service.ServicesGroup) *ControllersGroup {

	// create plugin middleware handle
	pluginMiddlewareProxy := sg.PluginsService.NewPluginMiddlewareProxyByRank()
	// apply plugin middleware rank 1
	r.Use(pluginMiddlewareProxy.ApplyForRank(plugin_services.MIDDLEWARE_RANK_1)...)

	// top level middleware
	r.Use(user_middleware.UUID())
	r.Use(cors.CORS())
	r.Use(user_middleware.Timezone())
	am := authentication_middleware.DefaultAuthMiddleware(sg)
	hm := health_middleware.DefaultHealthMiddleware(sg)
	r.Use(am.AddUserToContextIfValidToken())

	// apply plugin middleware rank 1000
	r.Use(pluginMiddlewareProxy.ApplyForRank(plugin_services.MIDDLEWARE_RANK_1000)...)

	//r.LoadHTMLGlob("./content/templates/*.tmpl")
	r.HTMLRender = createMyRender()
	// setup route groups
	routes := &routes.Routes{
		Root:    r.Group("/"),
		Public:  r.Group(defaultRoutePrefix),
		Auth:    r.Group(defaultRoutePrefix),
		NoRoute: r.NoRoute,
	}

	// apply auth middleware
	am.ApplyAuthToRoutes(routes)
	hm.ApplyHealthToRoutes(routes)

	// apply plugin middleware rank 2000
	r.Use(pluginMiddlewareProxy.ApplyForRank(plugin_services.MIDDLEWARE_RANK_2000)...)

	// define routes and apply middleware
	apiControllers := &ApiControllers{
		AuthController:      authentication_controller.DefaultAuthController(routes, sg),
		AdminUserController: user_admin_controller.DefaultUserAdminController(routes, sg),
		HealthyController:   health_controller.DefaultHealthController(routes, sg),
		UserController:      user_controller.DefaultUserController(routes, sg),
		EmailController:     email_controller.DefaultEmailController(routes, sg),
	}

	// define after for 404 catcher
	contentControllers := &ContentControllers{
		DocumentationController: documentation_controller.DefaultDocumentationController(routes, sg),
		ThemesControllers:       theme_controller.DefaultThemesController(r, routes),
		TemplateControllers:     template_controller.DefaultTemplatesController(routes),
		ReactControllers:        react_controller.DefaultReactController(routes, sg),
	}

	// apply plugin middleware rank 3000
	r.Use(pluginMiddlewareProxy.ApplyForRank(plugin_services.MIDDLEWARE_RANK_3000)...)

	// register plugin routes
	sg.PluginsService.RegisterActivePluginRoutes(routes)

	// apply plugin middleware rank 4000
	r.Use(pluginMiddlewareProxy.ApplyForRank(plugin_services.MIDDLEWARE_RANK_4000)...)

	// add no route controller
	routes.NoRoute(func(c *gin.Context) {
		
		c.Redirect(http.StatusMovedPermanently, "/");
		
		return
		//present default 404 error
		
		
		// previously if the page was not found it was handed to react to look. This is unneccesary. Go knows what react routes work.  
		// React was continue looking for non existing pages and "load" forever. 
		// returning here insures we keep a uniform 404 page instead of having to write seperate 404 handlers on back/front end. 


		/* OLD CODE 
			paths := strings.Split(c.Request.RequestURI, "/")
			if paths[1] == "api" {
				return // handle default not route
			}
			contentControllers.ReactControllers.ServeReact(c)
		*/
	})

	controllersGroup := &ControllersGroup{
		ApiControllers:     apiControllers,
		ContentControllers: contentControllers,
		Routes:             routes,
	}

	return controllersGroup
}

func createMyRender() multitemplate.Render {
	r := multitemplate.New()
	r.AddFromGlob("docs.tmpl", "./content/templates/docs.tmpl")
	r.AddFromFiles("react.tmpl", "./content/templates/react.tmpl",
		fmt.Sprintf("./content/themes/%v/theme_header.tmpl", context.Config.DbVars.ActiveTheme),
		fmt.Sprintf("./content/themes/%v/theme_body.tmpl", context.Config.DbVars.ActiveTheme),
		fmt.Sprintf("./content/themes/%v/theme_footer.tmpl", context.Config.DbVars.ActiveTheme),
	)
	return r
}
