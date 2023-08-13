package conf

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ServeYaml struct {
	DataStorage      string            `yaml:"dataStorage"`
	Domains          []string          `yaml:"domains"`
	RangeSupported   bool              `yaml:"rangeSupported"`
	EnableGoInfoPage bool              `yaml:"enableGoInfoPage"`
	CacheSettings    CacheSettingsYaml `yaml:"cacheSettings"`
}

func (sy ServeYaml) GetDomainString() string {
	if len(sy.Domains) == 0 {
		return "all"
	} else {
		return strings.Join(sy.Domains, " ")
	}
}

func (sy ServeYaml) GetDataStoragePath() string {
	if sy.DataStorage == "" || !filepath.IsAbs(sy.DataStorage) {
		wd, err := os.Getwd()
		if err != nil {
			return ""
		} else {
			if sy.DataStorage == "" {
				return wd
			} else {
				return path.Join(wd, sy.DataStorage)
			}
		}
	} else {
		return sy.DataStorage
	}
}
