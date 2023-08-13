package conf

type ListenYaml struct {
	Web        string `yaml:"web"`
	WebMethod  string `yaml:"webMethod"`
	WebNetwork string `yaml:"webNetwork"`
	Identify   bool   `yaml:"identify"`
}
