package conf

type CacheSettingsYaml struct {
	EnableTemplateCaching                bool `yaml:"enableTemplateCaching"`
	EnableTemplateCachePurge             bool `yaml:"enableTemplateCachePurge"`
	EnableContentsCaching                bool `yaml:"enableContentsCaching"`
	EnableContentsCachePurge             bool `yaml:"enableContentsCachePurge"`
	MaxAge                               uint `yaml:"maxAge"`
	NotModifiedResponseUsingLastModified bool `yaml:"notModifiedUsingLastModified"`
	NotModifiedResponseUsingETags        bool `yaml:"notModifiedUsingETags"`
}
