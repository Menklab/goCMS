package goCMS_healthy_ctrl

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/menklab/goCMS/utility"
	"github.com/menklab/goCMS/routes"
)


type HealthyController struct {
	routes *goCMS_routes.ApiRoutes
}

func DefaultHealthyController(routes *goCMS_routes.ApiRoutes) *HealthyController{
	hc := &HealthyController{
		routes: routes,
	}

	hc.Default()
	return hc
}

func (hc *HealthyController) Default() {
	hc.routes.Public.GET("/healthy", hc.healthy)
	hc.routes.Auth.GET("/verify", hc.user)
}

/**
* @api {get} /healthy Service Health Status
* @apiDescription Used to verify that the services are up and running.
* @apiName GetHealthy
* @apiGroup Utility
*/
func (hc *HealthyController) healthy(c *gin.Context) {
	c.Status(http.StatusOK)
}

/**
* @api {get} /verify Verify User
* @apiDescription Used to verify that the user is authenticated.
* @apiName VerifyUser
* @apiGroup Authentication
*
* @apiUse UserAuthHeader
* @apiUse UserDisplay
* @apiPermission Authenticated
*/
func (hc *HealthyController) user(c *gin.Context) {
	// get logged in user
	authUser, _ := goCMS_utility.GetUserFromContext(c)
	c.JSON(http.StatusOK, authUser.GetUserDisplay())
}
