package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/menklab/goCMS/utility"
	"github.com/menklab/goCMS/routes"
)


// @SubApi User  [/users]
// @SubApi Allows you access to different features of the users , login , get status etc [/users]

type HealthyController struct {
	routes *routes.ApiRoutes
}

func DefaultHealthyController(routes *routes.ApiRoutes) *HealthyController{
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

// @Title Get Users Information
// @Description Get Users Information
// @Accept json
// @Param userId path int true "User ID"
// @Success 200 {object} string "Success"
// @Failure 401 {object} string "Access denied"
// @Failure 404 {object} string "Not Found"
// @Resource /users
// @Router /v1/users/:userId.json [get]
func (hc *HealthyController) healthy(c *gin.Context) {
	c.Status(http.StatusOK)
}

func (hc *HealthyController) user(c *gin.Context) {
	// get logged in user
	authUser, _ := utility.GetUserFromContext(c)
	c.JSON(http.StatusOK, authUser)
}
