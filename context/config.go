package goCMS_context

import (
	"github.com/menklab/goCMS/models"
	"log"
)

var Config *GoCMSConfig

type GoCMSConfig struct {
	// DB (GET FROM ENV)
	DbName     string
	DbUser     string
	DbPassword string
	DbServer   string

	// Elastic Search
	ElasticSearchConnectionUrl string
	ElasticSearchUseAwsSignedClient bool
	ElasticSearchAwsUser string
	ElasticSearchAwsSecret string

	// Debug
	Debug         bool
	DebugSecurity bool

	// App Config
	Port                string
	PublicApiUrl        string
	RedirectRootUrl     string
	CorsHost            string
	OpenRegistration    bool
	SettingsRefreshRate int64

	// Authentication
	AuthKey                string
	UserAuthTimeout        int64
	PasswordResetTimeout   int64
	EmailActivationTimeout int64
	DeviceAuthTimeout      int64
	TwoFactorCodeTimeout   int64
	UseTwoFactor           bool
	PasswordComplexity     int64

	// SMTP
	SMTPServer      string
	SMTPPort        int64
	SMTPUser        string
	SMTPPassword    string
	SMTPFromAddress string
	SMTPSimulate    bool
}

func (c *GoCMSConfig) ApplySettingsToConfig(settings map[string]goCMS_models.Setting) {

	log.Println("Refreshed GoCMS Settings")

	// Elastic Search
	c.ElasticSearchConnectionUrl = GetStringOrFail(settings["ES_CONNECTION_URL"].Value)
	c.ElasticSearchUseAwsSignedClient = GetBoolOrFail(settings["ES_USE_AWS_SIGNED_CLIENT"].Value)
	c.ElasticSearchAwsUser = GetStringOrFail(settings["ES_AWS_USER"].Value)
	c.ElasticSearchAwsSecret = GetStringOrFail(settings["ES_AWS_SECRET"].Value)

	// Debug
	c.Debug = GetBoolOrFail(settings["DEBUG"].Value)
	c.DebugSecurity = GetBoolOrFail(settings["DEBUG_SECURITY"].Value)

	// App Config
	c.Port = GetStringOrFail(settings["PORT"].Value)
	c.PublicApiUrl = GetStringOrFail(settings["PUBLIC_API_URL"].Value)
	c.RedirectRootUrl = GetStringOrFail(settings["REDIRECT_ROOT_URL"].Value)
	c.CorsHost = GetStringOrFail(settings["CORS_HOST"].Value)
	c.SettingsRefreshRate = GetIntOrFail(settings["SETTINGS_REFRESH_RATE"].Value)

	// Authentication
	c.AuthKey = GetStringOrFail(settings["AUTHENTICATION_KEY"].Value)
	c.UserAuthTimeout = GetIntOrFail(settings["USER_AUTHENTICATION_TIMEOUT"].Value)
	c.PasswordResetTimeout = GetIntOrFail(settings["PASSWORD_RESET_TIMEOUT"].Value)
	c.DeviceAuthTimeout = GetIntOrFail(settings["DEVICE_AUTHENTICATION_TIMEOUT"].Value)
	c.TwoFactorCodeTimeout = GetIntOrFail(settings["TWO_FACTOR_CODE_TIMEOUT"].Value)
	c.EmailActivationTimeout = GetIntOrFail(settings["EMAIL_ACTIVATION_TIMEOUT"].Value)
	c.UseTwoFactor = GetBoolOrFail(settings["USE_TWO_FACTOR"].Value)
	c.PasswordComplexity = GetIntOrFail(settings["PASSWORD_COMPLEXITY"].Value)
	c.OpenRegistration = GetBoolOrFail(settings["OPEN_REGISTRATION"].Value)

	// SMTP
	c.SMTPServer = GetStringOrFail(settings["SMTP_SERVER"].Value)
	c.SMTPPort = GetIntOrFail(settings["SMTP_PORT"].Value)
	c.SMTPUser = GetStringOrFail(settings["SMTP_USER"].Value)
	c.SMTPPassword = GetStringOrFail(settings["SMTP_PASSWORD"].Value)
	c.SMTPFromAddress = GetStringOrFail(settings["SMTP_FROM_ADDRESS"].Value)
	c.SMTPSimulate = GetBoolOrFail(settings["SMTP_SIMULATE"].Value)
}
