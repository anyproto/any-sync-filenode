package config

import (
	"os"

	commonaccount "github.com/anyproto/any-sync/accountservice"
	"github.com/anyproto/any-sync/app"
	"github.com/anyproto/any-sync/metric"
	"github.com/anyproto/any-sync/net/rpc"
	"github.com/anyproto/any-sync/net/transport/quic"
	"github.com/anyproto/any-sync/net/transport/yamux"
	"github.com/anyproto/any-sync/nodeconf"
	"gopkg.in/yaml.v3"

	"github.com/anyproto/any-sync-filenode/redisprovider"
	"github.com/anyproto/any-sync-filenode/store/s3store"
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
	Account                  commonaccount.Config   `yaml:"account"`
	Drpc                     rpc.Config             `yaml:"drpc"`
	Yamux                    yamux.Config           `yaml:"yamux"`
	Quic                     quic.Config            `yaml:"quic"`
	Metric                   metric.Config          `yaml:"metric"`
	S3Store                  s3store.Config         `yaml:"s3Store"`
	FileDevStore             FileDevStore           `yaml:"fileDevStore"`
	Redis                    redisprovider.Config   `yaml:"redis"`
	Network                  nodeconf.Configuration `yaml:"network"`
	NetworkStorePath         string                 `yaml:"networkStorePath"`
	NetworkUpdateIntervalSec int                    `yaml:"networkUpdateIntervalSec"`
	CafeMigrateKey           string                 `yaml:"cafeMigrateKey"`
	DefaultLimit             uint64                 `yaml:"defaultLimit"`
	PersistTtl               uint                   `yaml:"persistTtl"`
}

func (c *Config) Init(a *app.App) (err error) {
	return
}

func (c *Config) Name() (name string) {
	return CName
}

func (c *Config) GetAccount() commonaccount.Config {
	return c.Account
}

func (c *Config) GetS3Store() s3store.Config {
	return c.S3Store
}

func (c *Config) GetDevStore() FileDevStore {
	return c.FileDevStore
}

func (c *Config) GetDrpc() rpc.Config {
	return c.Drpc
}

func (c *Config) GetMetric() metric.Config {
	return c.Metric
}

func (c *Config) GetRedis() redisprovider.Config {
	return c.Redis
}

func (c *Config) GetNodeConf() nodeconf.Configuration {
	return c.Network
}

func (c *Config) GetNodeConfStorePath() string {
	return c.NetworkStorePath
}

func (c *Config) GetNodeConfUpdateInterval() int {
	return c.NetworkUpdateIntervalSec
}

func (c *Config) GetYamux() yamux.Config {
	return c.Yamux
}

func (c *Config) GetQuic() quic.Config {
	return c.Quic
}
