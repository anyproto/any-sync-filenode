package redisprovider

type configSource interface {
	GetRedis() Config
}

type Config struct {
	IsCluster bool   `yaml:"isCluster"`
	Url       string `yaml:"url"`
}
