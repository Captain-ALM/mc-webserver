package pageHandler

import (
	"html/template"
	"net/url"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"golang.captainalm.com/mc-webserver/conf"
	"golang.captainalm.com/mc-webserver/utils/info"
	"golang.captainalm.com/mc-webserver/utils/io"
)

const templateName = "goinfo.go.html"

func newGoInfoPage(handlerIn *PageHandler, dataStore string, cacheTemplates bool) *goInfoPage {
	var ptm *sync.Mutex
	if cacheTemplates {
		ptm = &sync.Mutex{}
	}
	pageToReturn := &goInfoPage{
		Handler:           handlerIn,
		DataStore:         dataStore,
		PageTemplateMutex: ptm,
	}
	return pageToReturn
}

type goInfoPage struct {
	Handler           *PageHandler
	DataStore         string
	PageTemplateMutex *sync.Mutex
	PageTemplate      *template.Template
}

func (gipg *goInfoPage) GetCacheIDExtension(urlParameters url.Values) string {
	if urlParameters.Has("full") {
		return "full"
	} else {
		return ""
	}
}

type goInfoTemplateMarshal struct {
	FullOutput         bool
	CurrentTime        time.Time
	RegisteredPages    []string
	CachedPages        []string
	ProcessID          int
	ParentProcessID    int
	ProductLocation    string
	ProductName        string
	ProductDescription string
	BuildVersion       string
	BuildDate          string
	WorkingDirectory   string
	Hostname           string
	PageSize           int
	GoVersion          string
	GoRoutineNum       int
	GoCGoCallNum       int64
	NumCPU             int
	GoRoot             string
	GoMaxProcs         int
	Compiler           string
	GoArch             string
	GoOS               string
	ListenSettings     conf.ListenYaml
	ServeSettings      conf.ServeYaml
	Environment        []string
}

func (gipg *goInfoPage) GetPath() string {
	return "/goinfo.go"
}

func (gipg *goInfoPage) GetLastModified() time.Time {
	return time.Now()
}

func (gipg *goInfoPage) GetContents(urlParameters url.Values) (contentType string, contents []byte, canCache bool) {
	theTemplate, err := gipg.getPageTemplate()
	if err != nil {
		return "text/plain", []byte("Cannot Get Info.\r\n" + err.Error()), false
	}
	var regPages []string
	var cacPages []string
	env := make([]string, 0)
	if urlParameters.Has("full") {
		regPages = gipg.Handler.GetRegisteredPages()
		cacPages = gipg.Handler.GetCachedPages()
		env = os.Environ()
	} else {
		regPages = make([]string, len(gipg.Handler.PageProviders))
		cacPages = make([]string, gipg.Handler.GetNumberOfCachedPages())
	}
	theBuffer := &io.BufferedWriter{}
	err = theTemplate.ExecuteTemplate(theBuffer, templateName, &goInfoTemplateMarshal{
		FullOutput:         urlParameters.Has("full"),
		CurrentTime:        time.Now(),
		RegisteredPages:    regPages,
		CachedPages:        cacPages,
		ProcessID:          os.Getpid(),
		ParentProcessID:    os.Getppid(),
		ProductLocation:    getStringOrError(os.Executable),
		ProductName:        info.BuildName,
		ProductDescription: info.BuildDescription,
		BuildVersion:       info.BuildVersion,
		BuildDate:          info.BuildDate,
		WorkingDirectory:   getStringOrError(os.Getwd),
		Hostname:           getStringOrError(os.Hostname),
		PageSize:           os.Getpagesize(),
		GoVersion:          runtime.Version(),
		GoRoutineNum:       runtime.NumGoroutine(),
		GoCGoCallNum:       runtime.NumCgoCall(),
		NumCPU:             runtime.NumCPU(),
		GoRoot:             runtime.GOROOT(),
		GoMaxProcs:         runtime.GOMAXPROCS(0),
		Compiler:           runtime.Compiler,
		GoArch:             runtime.GOARCH,
		GoOS:               runtime.GOOS,
		ListenSettings:     info.ListenSettings,
		ServeSettings:      info.ServeSettings,
		Environment:        env,
	})
	if err != nil {
		return "text/plain", []byte("Cannot Get Info.\r\n" + err.Error()), false
	}
	return "text/html", theBuffer.Data, false
}

func (gipg *goInfoPage) PurgeTemplate() {
	if gipg.PageTemplateMutex != nil {
		gipg.PageTemplateMutex.Lock()
		gipg.PageTemplate = nil
		gipg.PageTemplateMutex.Unlock()
	}
}

func (gipg *goInfoPage) getPageTemplate() (*template.Template, error) {
	if gipg.PageTemplateMutex != nil {
		gipg.PageTemplateMutex.Lock()
		defer gipg.PageTemplateMutex.Unlock()
	}
	if gipg.PageTemplate == nil {
		thePath := templateName
		if gipg.DataStore != "" {
			thePath = path.Join(gipg.DataStore, thePath)
		}
		loadedData, err := os.ReadFile(thePath)
		if err != nil {
			return nil, err
		}
		tmpl, err := template.New(templateName).Parse(string(loadedData))
		if err != nil {
			return nil, err
		}
		if gipg.PageTemplateMutex != nil {
			gipg.PageTemplate = tmpl
		}
		return tmpl, nil
	} else {
		return gipg.PageTemplate, nil
	}
}

func getStringOrError(funcIn func() (string, error)) string {
	toReturn, err := funcIn()
	if err == nil {
		return toReturn
	} else {
		return "Error: " + err.Error()
	}
}
