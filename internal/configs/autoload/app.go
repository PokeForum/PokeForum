package autoload

type APP struct {
	Host         string `mapstructure:"host" json:"host" yaml:"host"`
	Port         string `mapstructure:"port" json:"port" yaml:"port"`
	Debug        bool   `mapstructure:"debug" json:"debug" yaml:"debug"`
	Prometheus   bool   `mapstructure:"prometheus" json:"prometheus" yaml:"prometheus"`
	CookieDomain string `mapstructure:"cookie-domain" json:"cookie-domain" yaml:"cookie-domain"`
}
