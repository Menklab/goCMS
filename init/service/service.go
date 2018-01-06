package service

import (
	"github.com/gocms-io/gocms/context"
	"github.com/gocms-io/gocms/domain/acl/access_control/access_control_service"
	"github.com/gocms-io/gocms/domain/acl/authentication/authentication_service"
	"github.com/gocms-io/gocms/domain/acl/permissions/permissions_service"
	"github.com/gocms-io/gocms/domain/email/email_service"
	"github.com/gocms-io/gocms/domain/health/health_service"
	"github.com/gocms-io/gocms/domain/mail/mail_service"
	"github.com/gocms-io/gocms/domain/plugin/plugin_services"
	"github.com/gocms-io/gocms/domain/setting/setting_service"
	"github.com/gocms-io/gocms/domain/user/user_service"
	"github.com/gocms-io/gocms/init/database"
	"github.com/gocms-io/gocms/init/repository"
	"github.com/gocms-io/gocms/utility/log"
	"time"
)

type ServicesGroup struct {
	SettingsService   setting_service.ISettingsService
	MailService       mail_service.IMailService
	AuthService       authentication_service.IAuthService
	PermissionService permission_service.IPermissionService
	UserService       user_service.IUserService
	AclService        access_control_service.IAclService
	EmailService      email_service.IEmailService
	PluginsService    plugin_services.IPluginsService
	HealthService     health_service.IHealthService
}

func DefaultServicesGroup(repositoriesGroup *repository.RepositoriesGroup, db *database.Database) *ServicesGroup {
	var pluginRelatedErr error

	// setup settings
	settingsService := setting_service.DefaultSettingsService(repositoriesGroup)
	settingsService.RegisterRefreshCallback(context.Config.DbVars.LoadDbVars)

	// refresh settings every x minutes
	refreshSettings := time.Duration(context.Config.DbVars.SettingsRefreshRate) * time.Minute
	context.Schedule.AddTicker(refreshSettings, func() {
		settingsService.RefreshSettingsCache()
	})

	// mail service
	mailService := mail_service.DefaultMailService()

	// start permissions cache
	aclService := access_control_service.DefaultAclService(repositoriesGroup)
	aclService.RefreshPermissionsCache()

	permissionService := permission_service.DefaultPermissionService(repositoriesGroup)

	authService := authentication_service.DefaultAuthService(repositoriesGroup, mailService)
	userService := user_service.DefaultUserService(repositoriesGroup, authService, mailService)

	// email service
	emailService := email_service.DefaultEmailService(repositoriesGroup, mailService, authService)

	// plugins service
	pluginsService := plugin_services.DefaultPluginsService(repositoriesGroup, aclService)
	pluginRelatedErr = pluginsService.RefreshInstalledPlugins()
	if pluginRelatedErr != nil {
		log.Errorf("Error finding plugins. Can't start plugin microservice: %s\n", pluginRelatedErr.Error())
	} else {
		pluginRelatedErr = pluginsService.StartPluginsService()
	}

	// heath service
	healthService := health_service.DefaultHealthService(db, pluginsService)

	sg := &ServicesGroup{
		SettingsService:   settingsService,
		MailService:       mailService,
		AuthService:       authService,
		PermissionService: permissionService,
		UserService:       userService,
		AclService:        aclService,
		EmailService:      emailService,
		PluginsService:    pluginsService,
		HealthService:     healthService,
	}

	return sg
}
