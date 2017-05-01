package content_ctrl

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gocms-io/goCMS/routes"
	"github.com/gocms-io/goCMS/services"
	"net/http"
)

type DocumentationController struct {
	serviceGroup *services.ServicesGroup
	routes       *routes.Routes
}

func DefaultDocumentationController(routes *routes.Routes, sg *services.ServicesGroup) *DocumentationController {
	dc := &DocumentationController{
		serviceGroup: sg,
		routes:       routes,
	}

	dc.Default()
	return dc
}

func (dc *DocumentationController) Default() {
	// register goCMS Docs Route
	dc.routes.Root.Static("/docs/goCMS", "./content/docs")
	docsMap := make(map[string]string)

	// register plugins route
	for _, plugin := range dc.serviceGroup.PluginsService.GetActivePlugins() {
		name := plugin.Manifest.Name
		link := fmt.Sprintf("docs/%s", plugin.Manifest.Bin)
		docsMap[name] = link
		dc.routes.Root.Static(fmt.Sprintf("docs/%s", plugin.Manifest.Bin), fmt.Sprintf("./plugins/%s/docs", plugin.Manifest.Bin))
	}

	docsMap["GoCMS"] = "/docs/goCMS"

	dc.routes.Root.GET("/docs", func(c *gin.Context) {
		c.HTML(http.StatusOK, "docs.tmpl", docsMap)
	})

}
