package config

import (
	"github.com/anytypeio/any-sync-filenode/redisprovider"
	"github.com/anytypeio/any-sync-filenode/s3store"
	commonaccount "github.com/anytypeio/any-sync/accountservice"
	"github.com/anytypeio/any-sync/app"
	"github.com/anytypeio/any-sync/metric"
	"github.com/anytypeio/any-sync/net"
	"gopkg.in/yaml.v3"
	"os"
)

const CName = "config"

func NewFromFile(path string) (c *Config, err error) {
	c = &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(data, c); err != nil {
		return nil, err
	}
	return
}

type Config struct {
	Account      commonaccount.Config `yaml:"account"`
	GrpcServer   net.Config           `yaml:"grpcServer"`
	Metric       metric.Config        `yaml:"metric"`
	S3Store      s3store.Config       `yaml:"s3Store"`
	FileDevStore FileDevStore         `yaml:"fileDevStore"`
	Redis        redisprovider.Config `yaml:"redis"`
}

func (c *Config) Init(a *app.App) (err error) {
	return
}

func (c Config) Name() (name string) {
	return CName
}

func (c Config) GetAccount() commonaccount.Config {
	return c.Account
}

func (c Config) GetS3Store() s3store.Config {
	return c.S3Store
}

func (c Config) GetDevStore() FileDevStore {
	return c.FileDevStore
}

func (c Config) GetNet() net.Config {
	return c.GrpcServer
}

func (c Config) GetMetric() metric.Config {
	return c.Metric
}

func (c Config) GetRedis() redisprovider.Config {
	return c.Redis
}
