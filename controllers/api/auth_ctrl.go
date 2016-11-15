package api

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"bitbucket.org/menklab/grnow-services/services"
	"bitbucket.org/menklab/grnow-services/config"
	"bitbucket.org/menklab/grnow-services/utility/errors"
	"bitbucket.org/menklab/grnow-services/utility"
	"bitbucket.org/menklab/grnow-services/controllers/routes"
	"bitbucket.org/menklab/grnow-services/controllers/api/middleware"
)

// Login form structure.
type LoginDisplay struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Verify device form structure
type VerifyDeviceDisplay struct {
	DeviceCode string `json:"deviceCode" binding:"required"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

type ResetPassword struct {
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	ResetCode string `json:"resetCode" binding:"required"`
}

type AuthController struct {
	userService      services.IUserService
	authServices     services.IAuthService
}
var authController *AuthController

func init() {
	authController = &AuthController{
		userService: new(services.UserService),
		authServices: new(services.AuthService),
	}
}


func (ac *AuthController) Apply() {
	routes := routes.Routes()
	routes.Public.POST("/login", ac.login)
	routes.Public.POST("/reset-password", ac.resetPassword)
	routes.Public.PUT("/reset-password", ac.setPassword)
}

// login controller
/**
 * @api {post} /api/login Login
 * @apiName login
 * @apiGroup Authentication
 *
 * @apiParam {json} email User's email address.
 * @apiParam {json} password User's password.
 *
 * @apiSuccess {Header} X-Auth-Token JWT token used for subsequent authenticated requests.
 *
 * @apiSuccessExample Success-Response:
 *     	HTTP/1.1 200 OK
 */
func (ac *AuthController) login(c *gin.Context) {

	var loginDisplay LoginDisplay

	// get login values
	if c.BindJSON(&loginDisplay) != nil {
		errors.Response(c, http.StatusUnauthorized, "Missing Email or Password", middleware.REDIRECT_LOGIN)
		return
	}

	// auth user
	authUser := services.AuthUser{
		Email:    loginDisplay.Email,
		Password: loginDisplay.Password,
	}
	user, authed := authController.authServices.AuthUser(&authUser)
	if !authed {
		errors.ResponseWithSoftRedirect(c, http.StatusUnauthorized, "Incorrect Email / Password", middleware.REDIRECT_LOGIN)
		return
	}


	// create token
	expire := time.Now().Add(time.Minute * utility.GetTimeout(config.UserAuthTimeout))
	userToken := jwt.New(jwt.SigningMethodHS256)
	userToken.Claims["userId"] = user.Id
	userToken.Claims["exp"] = expire.Unix()
	tokenString, err := userToken.SignedString([]byte(config.AuthKey))

	if err != nil {
		errors.ResponseWithSoftRedirect(c, http.StatusUnauthorized, "Error generating token.", middleware.REDIRECT_LOGIN)
		return
	}

	c.Header("X-AUTH-TOKEN", tokenString)

	c.JSON(http.StatusOK, user)
	return
}


func (ac *AuthController) resetPassword(c *gin.Context) {

	// get email for reset
	var resetRequest ResetPasswordRequest
	err := c.BindJSON(&resetRequest) // update any changes from request
	if err != nil {
		errors.Response(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	// respond as everything after this doesn't matter to the requester
	c.String(http.StatusOK, "Email will be sent if account exists.")

	// send password reset link
	err = authController.authServices.SendPasswordResetCode(resetRequest.Email)
	if err != nil {
		return
	}
}

func (ac *AuthController) setPassword(c *gin.Context) {
	// get password and code for reset
	var resetPassword ResetPassword
	err := c.BindJSON(&resetPassword) // update any changes from request
	if err != nil {
		errors.Response(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	// get user
	user, err := authController.userService.GetByEmail(resetPassword.Email)
	if err != nil {
		errors.Response(c, http.StatusBadRequest, "Couldn't reset password.", err)
		return
	}

	// verify code
	if ok := authController.authServices.VerifyPasswordResetCode(user.Id, resetPassword.ResetCode); !ok {
		errors.ResponseWithSoftRedirect(c, http.StatusUnauthorized, "Error resetting password.", middleware.REDIRECT_LOGIN)
		return
	}

	// reset password
	user.NewPassword = resetPassword.Password
	err = authController.userService.Update(user.Id, user)
	if err != nil {
		errors.Response(c, http.StatusBadRequest, "Couldn't reset password.", err)
		return
	}

	c.Status(http.StatusOK)
}
