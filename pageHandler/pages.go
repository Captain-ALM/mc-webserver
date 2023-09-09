package pageHandler

import (
	"golang.captainalm.com/mc-webserver/conf"
	"golang.captainalm.com/mc-webserver/pageHandler/pages/index"
	"strings"
)

var providers map[string]PageProvider

func GetProviders(cacheTemplates bool, dataStorage string, pageHandler *PageHandler, templateStorage string, pageSettings []conf.PageYaml, ymlDataFallback bool) map[string]PageProvider {
	if providers == nil {
		providers = make(map[string]PageProvider)
		if pageHandler != nil {
			infoPage := newGoInfoPage(pageHandler, dataStorage, cacheTemplates)
			providers[infoPage.GetPath()] = infoPage //Go Information Page
		}
		for _, cpg := range pageSettings { //Register pages
			if strings.EqualFold(cpg.PageName, index.PageName) {
				indexPage := index.NewPage(dataStorage, cacheTemplates, templateStorage, cpg.GetPagePath(), ymlDataFallback)
				providers[indexPage.GetPath()] = indexPage
			}
		}
	}
	return providers
}
