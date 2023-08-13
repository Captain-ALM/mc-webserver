package main

import (
	"golang.captainalm.com/mc-webserver/conf"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func getListener(config conf.ConfigYaml, cwd string) net.Listener {
	split := strings.Split(strings.ToLower(config.Listen.WebNetwork), ":")
	if len(split) == 0 {
		log.Fatalln("Invalid Web Network")
		return nil
	} else {
		var theListener net.Listener
		var theError error
		log.Println("[Main] Socket Network Type: " + split[0])
		log.Printf("[Main] Starting up %s server on %s...\n", config.Listen.WebMethod, config.Listen.Web)
		switch split[0] {
		case "tcp", "tcp4", "tcp6":
			theListener, theError = net.Listen(strings.ToLower(config.Listen.WebNetwork), config.Listen.Web)
		case "unix", "unixgram", "unixpacket":
			socketPath := config.Listen.Web
			if !filepath.IsAbs(socketPath) {
				if !filepath.IsAbs(cwd) {
					log.Fatalln("Web Path Not Absolute And No Working Directory.")
					return nil
				}
				socketPath = path.Join(cwd, socketPath)
			}
			log.Println("[Main] Removing old socket.")
			if err := os.RemoveAll(socketPath); err != nil {
				log.Fatalln("Could Not Remove Old Socket.")
				return nil
			}
			theListener, theError = net.Listen(strings.ToLower(config.Listen.WebNetwork), config.Listen.Web)
		default:
			log.Fatalln("Unknown Web Network.")
			return nil
		}
		if theError != nil {
			log.Fatalln("Failed to listen due to:", theError)
			return nil
		}
		return theListener
	}
}

func runBackgroundHttp(s *http.Server, l net.Listener, tlsEnabled bool) {
	var err error
	if tlsEnabled {
		err = s.ServeTLS(l, "", "")
	} else {
		err = s.Serve(l)
	}
	if err != nil {
		if err == http.ErrServerClosed {
			log.Println("The http server shutdown successfully")
		} else {
			log.Fatalf("[Http] Error trying to host the http server: %s\n", err.Error())
		}
	}
}

func runBackgroundFCgi(h http.Handler, l net.Listener) {
	err := fcgi.Serve(l, h)
	if err != nil {
		if err == net.ErrClosed {
			log.Println("The fcgi server shutdown successfully")
		} else {
			log.Fatalf("[Http] Error trying to host the fcgi server: %s\n", err.Error())
		}
	}
}
