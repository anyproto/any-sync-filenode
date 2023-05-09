package s3store

type configSource interface {
	GetS3Store() Config
}

type Config struct {
	Profile    string `yaml:"profile"`
	Region     string `yaml:"region"`
	Bucket     string `yaml:"bucket"`
	Endpoint   string `yaml:"endpoint"`
	MaxThreads int    `yaml:"maxThreads"`
}
