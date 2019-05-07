package config

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Upstream server config
type Upstream struct {
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	Weight int    `yaml:"weight"`
	Scheme string `yaml:"scheme"`
}

// VHost config for per service(host)
type VHost struct {
	MaxFails    int        `yaml:"max_fails"`
	MaxTimeout  int64      `yaml:"max_timeout"`
	EnableHTTPS bool       `yaml:"enable_https"`
	Regex       string     `yaml:"regex"`
	LoadBlance  string     `yaml:"load_balance"`
	ServerName  []string   `yaml:"server_name"`
	Upstreams   []Upstream `yaml:"upstreams"`
}

// GlobalConfig echoes server global config
type GlobalConfig struct {
	Worker int     `yaml:"worker"`
	LogDir string  `yaml:"log_dir"`
	VHosts []VHost `yaml:"vhosts"`
}

// ReadConfig initialize config from specify path
func ReadConfig(file string) (*GlobalConfig, error) {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	conf := new(GlobalConfig)
	err = yaml.Unmarshal(yamlFile, conf)
	if err != nil {
		return nil, err
	}

	// check config
	err = checkConfig(conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func checkConfig(c *GlobalConfig) error {
	// TODO : complete config check
	for _, vhost := range c.VHosts {
		if len(vhost.Upstreams) == 0 {
			return errors.New("Miss upstreams config")
		}
	}
	return nil
}
