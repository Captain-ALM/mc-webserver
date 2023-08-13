package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"golang.captainalm.com/mc-webserver/conf"
	"golang.captainalm.com/mc-webserver/pageHandler"
	"golang.captainalm.com/mc-webserver/utils/info"
	"gopkg.in/yaml.v3"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	buildName        = ""
	buildDescription = "Minecraft Status Web APP"
	buildVersion     = "develop"
	buildDate        = ""
)

func main() {
	log.Printf("[Main] Starting up %s (%s) #%s (%s)\n", buildDescription, buildName, buildVersion, buildDate)
	y := time.Now()
	info.SetupProductInfo(buildName, buildDescription, buildVersion, buildDate)

	//Hold main thread till safe shutdown exit:
	wg := &sync.WaitGroup{}
	wg.Add(1)

	//Get working directory:
	cwdDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	//Load environment file:
	err = godotenv.Load()
	if err != nil {
		log.Fatalln("Error loading .env file")
	}

	//Data directory processing:
	dataDir := os.Getenv("DIR_DATA")
	if dataDir == "" {
		dataDir = path.Join(cwdDir, ".data")
	}

	check(os.MkdirAll(dataDir, 0777))

	//Config file processing:
	configLocation := os.Getenv("CONFIG_FILE")
	if configLocation == "" {
		configLocation = path.Join(dataDir, "config.yml")
	} else {
		if !filepath.IsAbs(configLocation) {
			configLocation = path.Join(dataDir, configLocation)
		}
	}

	//Config loading:
	configFile, err := os.Open(configLocation)
	if err != nil {
		log.Fatalln("Failed to open config.yml")
	}

	var configYml conf.ConfigYaml
	groupsDecoder := yaml.NewDecoder(configFile)
	err = groupsDecoder.Decode(&configYml)
	if err != nil {
		log.Fatalln("Failed to parse config.yml:", err)
	}
	err = configFile.Close()
	if err != nil {
		log.Println("Failed to close config file.")
	}

	//Server definitions:
	var webServer *http.Server
	var fcgiListen net.Listener
	info.ListenSettings = configYml.Listen
	info.ServeSettings = configYml.Serve
	switch strings.ToLower(configYml.Listen.WebMethod) {
	case "http":
		webServer = &http.Server{Handler: pageHandler.GetRouter(configYml)}
		go runBackgroundHttp(webServer, getListener(configYml, cwdDir), false)
	case "fcgi":
		fcgiListen = getListener(configYml, cwdDir)
		if fcgiListen == nil {
			log.Fatalln("Listener Nil")
		} else {
			go runBackgroundFCgi(pageHandler.GetRouter(configYml), fcgiListen)
		}
	default:
		log.Fatalln("Unknown Web Method.")
	}

	//=====================
	// Safe shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	//Startup complete:
	z := time.Now().Sub(y)
	log.Printf("[Main] Took '%s' to fully initialize modules\n", z.String())

	go func() {
		<-sigs
		fmt.Printf("\n")

		log.Printf("[Main] Attempting safe shutdown\n")
		a := time.Now()

		if webServer != nil {
			log.Printf("[Main] Shutting down HTTP server...\n")
			err := webServer.Close()
			if err != nil {
				log.Println(err)
			}
		}

		if fcgiListen != nil {
			log.Printf("[Main] Shutting down FCGI server...\n")
			err := fcgiListen.Close()
			if err != nil {
				log.Println(err)
			}
		}

		log.Printf("[Main] Signalling program exit...\n")
		b := time.Now().Sub(a)
		log.Printf("[Main] Took '%s' to fully shutdown modules\n", b.String())
		wg.Done()
	}()
	//
	//=====================
	wg.Wait()
	log.Println("[Main] Goodbye")
	//os.Exit(0)
}
