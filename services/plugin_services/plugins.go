package plugin_services

import (
	"github.com/menklab/goCMS/models"
	"path/filepath"
	"os"
	"log"
	"io/ioutil"
	"encoding/json"
	"github.com/menklab/goCMS/utility/errors"
	"github.com/menklab/goCMS/routes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/menklab/goCMS/controllers/middleware/plugins/proxy"
	"os/exec"
	"bufio"
)

type Plugin struct {
	BinaryPath string
	Host       string
	Port       int64
	Schema     string
	Manifest   *models.PluginManifest
	Proxy      *plugin_proxy_mdl.PluginProxyMiddleware
	cmd        *exec.Cmd
}

type ProxyRoute struct {
	Schema string
	Host   string
	Port   string
}

type IPluginsService interface {
	FindPlugins() error
	StartPlugins() error
	RegisterPluginRoutes(routes *routes.Routes) error
}

type PluginsService struct {
	Plugins []*Plugin
}

func DefaultPluginsService() *PluginsService {

	pluginsService := &PluginsService{}

	return pluginsService

}

func (ps *PluginsService) StartPlugins() error {

	var port int64 = 30002
	for _, plugin := range ps.Plugins {
		cmd := exec.Command(plugin.BinaryPath, "-port", "9091")


		cmdReader, err := cmd.StdoutPipe()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
			return err
		}

		scanner := bufio.NewScanner(cmdReader)
		go func() {
			for scanner.Scan() {
				fmt.Printf("\t > %s - %s\n", plugin.Manifest.Name, scanner.Text())
			}
		}()

		log.Printf("Starting microservice: %s\n", plugin.Manifest.Name)
		err = cmd.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
			return err
		}

		//err = cmd.Wait()
		//if err != nil {
		//	fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
		//	return err
		//}

		plugin.cmd = cmd
		port++

		// if success create proxy
		plugin.Proxy = &plugin_proxy_mdl.PluginProxyMiddleware{
			Port: port,
			Schema: "http",
			Host: "localhost",
		}
	}

	return nil

}

func (ps *PluginsService) RegisterPluginRoutes(routes *routes.Routes) error {
	for _, plugin := range ps.Plugins {

		// loop through each manifest and apply each route to the middleware proxy
		for _, routeManifest := range plugin.Manifest.Routes {
			routerGroup, err := ps.getRouteGroup(routeManifest.Route, routes)
			if err != nil {
				es := fmt.Sprintf("Plugin %s -> Route %s -> Method %s, Url %s, Error: %s\n", plugin.Manifest.Name, routeManifest.Route, routeManifest.Method, routeManifest.Url, err.Error())
				log.Print(es)
				return err
			} else {
				ps.registerPluginProxyOnRoute(routerGroup, routeManifest.Method, routeManifest.Url, plugin.Proxy)
			}
		}
	}
	return nil
}

func (ps *PluginsService) registerPluginProxyOnRoute(route *gin.RouterGroup, method string, url string, pluginProxy *plugin_proxy_mdl.PluginProxyMiddleware) {
	route.Handle(method, url, pluginProxy.ReverseProxy())
}

func (ps *PluginsService) getRouteGroup(pluginRoute string, routes *routes.Routes) (*gin.RouterGroup, error) {
	switch pluginRoute {
	case "Public":
		return routes.Public, nil
	case "PreTwofactor":
		return routes.PreTwofactor, nil
	case "Auth":
		return routes.Auth, nil
	case "Admin":
		return routes.Admin, nil
	case "Root":
		return routes.Root, nil
	default:
		return nil, errors.New(fmt.Sprintf("Route %s doesn't exist.\n", pluginRoute))
	}
}

func (ps *PluginsService) FindPlugins() error {

	// find all plugins
	err := filepath.Walk("./plugins/", ps.visitPlugin)
	if err != nil {
		log.Printf("Error finding plugins while traversing plugin directory: %s\n", err.Error())
		return err
	}

	return err
}

func (ps *PluginsService) visitPlugin(path string, f os.FileInfo, err error) (error) {
	if err != nil {
		log.Printf("Error traversing %s, %s\n", path, err.Error())
	}

	// parse manifests as they are found
	if f.Mode().IsRegular() && f.Name() == "manifest.json" {
		manifest, err := ps.parseManifest(path)
		if err != nil {
			return err
		}

		// verify that there is a main.go file
		mainPath, _ := filepath.Split(path)
		mainFilePath := filepath.Join(mainPath, "main.go")
		mainFile, err := os.Stat(mainFilePath)
		if err != nil {
			log.Printf("No main file for plugin %s: %s\n", manifest.Name, err.Error())
			return err
		}

		if !mainFile.Mode().IsRegular() {
			log.Printf("Main file for plugin %s, apprears to be corrupted: %s\n", manifest.Name, err.Error())
			return err
		}

		plugin := Plugin{
			BinaryPath: mainFilePath,
			Manifest: manifest,
		}

		ps.Plugins = append(ps.Plugins, &plugin)
	}

	return nil
}

func (ps *PluginsService) parseManifest(fileUri string) (*models.PluginManifest, error) {
	var manifest models.PluginManifest

	// read file in
	raw, err := ioutil.ReadFile(fileUri)
	if err != nil {
		log.Printf("Error reading raw plugin manifest file %s: %s\n", fileUri, err.Error())
		return nil, err
	}

	err = json.Unmarshal(raw, &manifest)
	if err != nil {
		log.Printf("Error parsing manifest file %s: %s\n", fileUri, err.Error())
		return nil, err
	}

	return &manifest, nil
}