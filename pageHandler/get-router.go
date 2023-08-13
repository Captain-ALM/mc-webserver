package pageHandler

import (
	"github.com/gorilla/mux"
	"golang.captainalm.com/mc-webserver/conf"
	"golang.captainalm.com/mc-webserver/pageHandler/utils"
	"net/http"
)

var theRouter *mux.Router
var thePageHandler *PageHandler

func GetRouter(config conf.ConfigYaml) http.Handler {
	if theRouter == nil {
		theRouter = mux.NewRouter()
		if thePageHandler == nil {
			thePageHandler = NewPageHandler(config.Serve)
		}
		if len(config.Serve.Domains) == 0 {
			theRouter.PathPrefix("/").HandlerFunc(thePageHandler.ServeHTTP)
		} else {
			for _, domain := range config.Serve.Domains {
				theRouter.Host(domain).HandlerFunc(thePageHandler.ServeHTTP)
			}
			theRouter.PathPrefix("/").HandlerFunc(domainNotAllowed)
		}
		if config.Listen.Identify {
			theRouter.Use(headerMiddleware)
		}
	}
	return theRouter
}

func domainNotAllowed(rw http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet || req.Method == http.MethodHead {
		utils.WriteResponseHeaderCanWriteBody(req.Method, rw, http.StatusNotFound, "Domain Not Allowed")
	} else {
		rw.Header().Set("Allow", http.MethodOptions+", "+http.MethodGet+", "+http.MethodHead)
		if req.Method == http.MethodOptions {
			utils.WriteResponseHeaderCanWriteBody(req.Method, rw, http.StatusOK, "")
		} else {
			utils.WriteResponseHeaderCanWriteBody(req.Method, rw, http.StatusMethodNotAllowed, "")
		}
	}
}

func headerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "Clerie Gilbert")
		w.Header().Set("X-Powered-By", "Love")
		w.Header().Set("X-Friendly", "True")
		next.ServeHTTP(w, r)
	})
}
