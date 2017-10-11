package auth_ctrl

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gocms-io/gocms/routes"
	"github.com/gocms-io/gocms/services"
	"time"

	"github.com/gocms-io/gocms/context"
	"github.com/gocms-io/gocms/utility"
)

const (
	REDIRECT_LOGIN         = "login"
	REDIRECT_VERIFY_DEVICE = "verifyDevice"
)

type AuthController struct {
	routes        *routes.Routes
	ServicesGroup *services.ServicesGroup
}

func DefaultAuthController(routes *routes.Routes, sg *services.ServicesGroup) *AuthController {

	// create controller
	authController := &AuthController{
		routes:        routes,
		ServicesGroup: sg,
	}

	authController.Default()

	return authController
}

/**
* @apiDefine Authenticated Authenticated User
* User must be logged in and authenticated.
 */

func (ac *AuthController) Default() {
	ac.routes.Public.POST("/register", ac.register)
	ac.routes.Public.POST("/login", ac.login)
	ac.routes.Public.POST("/login/facebook", ac.loginFacebook)
	ac.routes.Public.POST("/login/google", ac.loginGoogle)
	ac.routes.Public.POST("/reset-password", ac.resetPassword)
	ac.routes.Public.PUT("/reset-password", ac.setPassword)
	ac.routes.Auth.GET("/verify", ac.verifyUser)

	if context.Config.UseTwoFactor {
		ac.routes.PreTwofactor.GET("/verify-device", ac.getDeviceCode)
		ac.routes.PreTwofactor.POST("/verify-device", ac.verifyDevice)
	}
}

func (ac *AuthController) createToken(userId int) (string, error) {
	expire := time.Now().Add(time.Minute * utility.GetTimeout(context.Config.UserAuthTimeout))
	userToken := jwt.New(jwt.SigningMethodHS256)
	userToken.Claims["userId"] = userId
	userToken.Claims["exp"] = expire.Unix() * 1000 // get milliseconds
	return userToken.SignedString([]byte(context.Config.AuthKey))
}
