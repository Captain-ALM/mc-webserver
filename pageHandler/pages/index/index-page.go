package index

import (
	"html/template"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"golang.captainalm.com/mc-webserver/utils/io"
	"gopkg.in/yaml.v3"
)

const templateName = "index.go.html"
const yamlName = "index.go.yml"

func NewPage(dataStore string, cacheTemplates bool) *Page {
	var ptm *sync.Mutex
	var sdm *sync.Mutex
	if cacheTemplates {
		ptm = &sync.Mutex{}
		sdm = &sync.Mutex{}
	}
	pageToReturn := &Page{
		DataStore:         dataStore,
		StoredDataMutex:   sdm,
		PageTemplateMutex: ptm,
		CacheMCMutex:      &sync.Mutex{},
	}
	return pageToReturn
}

type Page struct {
	DataStore            string
	StoredDataMutex      *sync.Mutex
	StoredData           *DataYaml
	LastModifiedData     time.Time
	PageTemplateMutex    *sync.Mutex
	PageTemplate         *template.Template
	LastModifiedTemplate time.Time
	CachedMC             *MC
	CollectedMCExpiry    time.Time
	LastModifiedMC       time.Time
	CacheMCMutex         *sync.Mutex
}

func (p *Page) GetPath() string {
	return "/index.go"
}

func (p *Page) GetLastModified() time.Time {
	var toTest time.Time
	if p.LastModifiedData.After(p.LastModifiedTemplate) {
		toTest = p.LastModifiedData
	} else {
		toTest = p.LastModifiedTemplate
	}
	if p.LastModifiedMC.After(toTest) {
		return p.LastModifiedMC
	} else {
		return toTest
	}
}

func (p *Page) GetCacheIDExtension(urlParameters url.Values) string {
	toReturn := ""
	if urlParameters.Has("players") {
		toReturn += "players&"
	}
	if urlParameters.Has("mods") {
		toReturn += "mods&"
	}
	if urlParameters.Has("extended") {
		toReturn += "extended&"
	}
	if urlParameters.Has("dark") {
		toReturn += "dark"
	}
	return strings.TrimRight(toReturn, "&")
}

func (p *Page) GetContents(urlParameters url.Values) (contentType string, contents []byte, canCache bool) {
	theTemplate, err := p.getPageTemplate()
	if err != nil {
		return "text/plain", []byte("Cannot Get Index.\r\n" + err.Error()), false
	}
	theData, err := p.getPageData()
	if err != nil {
		return "text/plain", []byte("Cannot Get Data.\r\n" + err.Error()), false
	}
	theMarshal := &Marshal{
		Data:          *theData,
		Dark:          urlParameters.Has("dark"),
		PlayersShown:  urlParameters.Has("players"),
		ModsShown:     urlParameters.Has("mods"),
		ExtendedShown: urlParameters.Has("extended"),
		Online:        true,
	}

	theMarshal.Queried = p.GetMC(theData, theMarshal)
	theBuffer := &io.BufferedWriter{}
	err = theTemplate.ExecuteTemplate(theBuffer, templateName, theMarshal)
	if err != nil {
		return "text/plain", []byte("Cannot Get Page.\r\n" + err.Error()), false
	}
	return "text/html", theBuffer.Data, true
}

func (p *Page) GetMC(theData *DataYaml, theMarshal *Marshal) MC {
	var theMC MC
	var err error
	defer p.CacheMCMutex.Unlock()
	p.CacheMCMutex.Lock()
	if time.Now().After(p.CollectedMCExpiry) || time.Now().Equal(p.CollectedMCExpiry) {
		theMC, err = theMarshal.NewMC()
		if err == nil {
			p.CachedMC = &theMC
		} else {
			theMarshal.Online = false
			p.CachedMC = nil
		}
		p.LastModifiedMC = time.Now()
		p.CollectedMCExpiry = p.LastModifiedMC.Add(theData.MCQueryInterval)
	} else {
		if p.CachedMC == nil {
			theMC = MC{}
			theMarshal.Online = false
		} else {
			theMC = *p.CachedMC
		}
	}
	return theMC
}

func (p *Page) PurgeTemplate() {
	if p.PageTemplateMutex != nil {
		p.PageTemplateMutex.Lock()
		p.PageTemplate = nil
		p.PageTemplateMutex.Unlock()
	}
	if p.StoredDataMutex != nil {
		p.StoredDataMutex.Lock()
		p.StoredData = nil
		p.StoredDataMutex.Unlock()
	}
}

func (p *Page) getPageTemplate() (*template.Template, error) {
	if p.PageTemplateMutex != nil {
		p.PageTemplateMutex.Lock()
		defer p.PageTemplateMutex.Unlock()
	}
	if p.PageTemplate == nil {
		thePath := templateName
		if p.DataStore != "" {
			thePath = path.Join(p.DataStore, thePath)
		}
		stat, err := os.Stat(thePath)
		if err != nil {
			return nil, err
		}
		p.LastModifiedTemplate = stat.ModTime()
		loadedData, err := os.ReadFile(thePath)
		if err != nil {
			return nil, err
		}
		tmpl, err := template.New(templateName).Parse(string(loadedData))
		if err != nil {
			return nil, err
		}
		if p.PageTemplateMutex != nil {
			p.PageTemplate = tmpl
		}
		return tmpl, nil
	} else {
		return p.PageTemplate, nil
	}
}

func (p *Page) getPageData() (*DataYaml, error) {
	if p.StoredDataMutex != nil {
		p.StoredDataMutex.Lock()
		defer p.StoredDataMutex.Unlock()
	}
	if p.StoredData == nil {
		thePath := yamlName
		if p.DataStore != "" {
			thePath = path.Join(p.DataStore, thePath)
		}
		stat, err := os.Stat(thePath)
		if err != nil {
			return nil, err
		}
		p.LastModifiedData = stat.ModTime()
		fileHandle, err := os.Open(thePath)
		if err != nil {
			return nil, err
		}
		dataYaml := &DataYaml{}
		decoder := yaml.NewDecoder(fileHandle)
		err = decoder.Decode(dataYaml)
		if err != nil {
			_ = fileHandle.Close()
			return nil, err
		}
		err = fileHandle.Close()
		if err != nil {
			return nil, err
		}
		if p.StoredDataMutex != nil {
			p.StoredData = dataYaml
		}
		return dataYaml, nil
	} else {
		return p.StoredData, nil
	}
}
