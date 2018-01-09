package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/gocms-io/gocms/context"
	"github.com/gocms-io/gocms/init/controller"
	"github.com/gocms-io/gocms/init/database"
	"github.com/gocms-io/gocms/init/repository"
	"github.com/gocms-io/gocms/init/service"
	"github.com/gocms-io/gocms/utility/log"
	"net/http"
	"os"
	"github.com/gocms-io/gocms/utility/security"
	"golang.org/x/sync/errgroup"
)

var (
	egocms *Engine
	igocms *InternalEngine
	g      errgroup.Group
)

type Engine struct {
	Gin               *gin.Engine
	ControllersGroup  *controller.ControllersGroup
	ServicesGroup     *service.ServicesGroup
	RepositoriesGroup *repository.RepositoriesGroup
	Database          *database.Database
}

type InternalEngine struct {
	Gin               *gin.Engine
	InternalControllersGroup  *controller.InternalControllersGroup
	ServicesGroup     *service.ServicesGroup
	RepositoriesGroup *repository.RepositoriesGroup
	Database          *database.Database
}

type gocmsRuntimeSettings struct {
	port               string
	msPort             string
	noExtneralServices bool
	noInternalServices bool
}

// todo write an optimizer for requirejs

//go:generate apidoc -c ./ -i ./models -i ./controllers/ -o ./content/docs/ -f ".*\\.go$" -f ".*\\.js$"
func Default() (e *Engine, ie *InternalEngine) {

	// setup database
	db := database.DefaultSQL()

	// migrate cms db
	db.SQL.MigrateSql()

	// check for rsa keys
	security.CheckOrGenRSAKeysAndSecrets(db.SQL.Dbx)

	// setup log level
	switch context.Config.EnvVars.LogLevel {
	case log.LOG_LEVEL_CRITICAL:
		fallthrough
	case log.LOG_LEVEL_ERROR:
		gin.SetMode(gin.ReleaseMode)
	case log.LOG_LEVEL_WARNING:
		gin.SetMode(gin.TestMode)
	case log.LOG_LEVEL_DEBUG:
		gin.SetMode(gin.DebugMode)
	}
	r := gin.Default()
	ir := gin.Default()

	// setup repositories
	rg := repository.DefaultRepositoriesGroup(db.SQL.Dbx)

	// setup services
	sg := service.DefaultServicesGroup(rg, db)

	// setup controllers
	cg := controller.DefaultControllerGroup(r, sg)
	icg := controller.DefaultInternalControllerGroup(ir, sg)

	// create engine
	e = &Engine{
		Gin:               r,
		ControllersGroup:  cg,
		ServicesGroup:     sg,
		RepositoriesGroup: rg,
		Database:          db,
	}

	// create engine
	ie = &InternalEngine{
		Gin:               ir,
		InternalControllersGroup: icg,
		ServicesGroup:     sg,
		RepositoriesGroup: rg,
		Database:          db,
	}

	return e, ie
}

func (engine *Engine) Listen(uri string) error {

	log.Infof("Listening on: %v\n", uri)
	err := http.ListenAndServe(uri, engine.Gin)
	return err

}

func (engine *InternalEngine) Listen(uri string) error {

	log.Infof("Listening on: %v\n", uri)
	err := http.ListenAndServe(uri, engine.Gin)
	return err

}

func main() {

	// startup defaults
	egocms, igocms = Default()

	// get ports
	rs := getRuntimeSettings()

	// no external if needed
	if !rs.noExtneralServices {
		g.Go(func() error {
			return egocms.Listen(":" + rs.port)
		})
	}

	// no internal if needed
	if !rs.noInternalServices {
		g.Go(func() error {
			return igocms.Listen(":" + rs.msPort)
		})
	}

	if err := g.Wait(); err != nil {
		log.Criticalf("Error launching services: %v\n", err.Error())
	}
}

func getRuntimeSettings() *gocmsRuntimeSettings {
	// define flags
	portFlag := flag.String("port", "", "port to run on. Overrides all.")
	msPortFlag := flag.String("msPort", "", "msPort to run on. Overrides all.")
	noExternalServiceFlag := flag.Bool("noExternal", false, "noExternal when this flag is set gocms will not run external services.")
	noInternalServiceFlag := flag.Bool("noInternal", false, "noInternal when this flag is set gocms will not run internal services.")
	flag.Parse()

	noExternalService := *noExternalServiceFlag
	noInternalService := *noInternalServiceFlag

	///////// PORT ///////////
	// get server port in order of importance
	// 3. db
	port := context.Config.DbVars.Port
	// 2. env
	portEnv := os.Getenv("PORT")
	if portEnv != "" {
		port = portEnv
	}
	// 1. flag
	if *portFlag != "" {
		port = *portFlag
	}
	// 0. if still unset
	if port == "" {
		port = "8080"
	}

	///////// MS PORT ///////////
	// get server port in order of importance
	// 3. db
	msPort := context.Config.DbVars.MsPort
	// 2. env
	msPortEnv := os.Getenv("msPort")
	if msPortEnv != "" {
		msPort = msPortEnv
	}
	// 1. flag
	// check for msPort flag and override all
	if *msPortFlag != "" {
		msPort = *msPortFlag
	}
	// 0. if still unset
	if msPort == "" {
		msPort = "8081"
	}

	return &gocmsRuntimeSettings{
		port:   port,
		msPort: msPort,
		noExtneralServices: noExternalService,
		noInternalServices: noInternalService,
	}
}
