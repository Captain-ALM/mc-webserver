package conf

import "strings"

type PageYaml struct {
	PageName string `yaml:"pageName"`
	PagePath string `yaml:"pagePath"`
}

func (py PageYaml) GetPagePath() string {
	toReturn := py.PagePath
	if !strings.HasSuffix(toReturn, ".go") {
		toReturn += ".go"
	}
	if !strings.HasPrefix(toReturn, "/") {
		toReturn = "/" + toReturn
	}
	return toReturn
}
