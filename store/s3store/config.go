package s3store

type configSource interface {
	GetS3Store() Config
}

type Credentials struct {
	AccessKey string `yaml:"accessKey"`
	SecretKey string `yaml:"secretKey"`
}

type Config struct {
	Profile        string      `yaml:"profile"`
	Region         string      `yaml:"region"`
	Bucket         string      `yaml:"bucket"`
	IndexBucket    string      `yaml:"indexBucket"`
	Endpoint       string      `yaml:"endpoint"`
	MaxThreads     int         `yaml:"maxThreads"`
	Credentials    Credentials `yaml:"credentials"`
	ForcePathStyle bool        `yaml:"forcePathStyle"`
}
