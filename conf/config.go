package conf

type ConfigYaml struct {
	Listen ListenYaml `yaml:"listen"`
	Serve  ServeYaml  `yaml:"serve"`
}
