package server

import (
	"encoding/json"
	"io/ioutil"
	"network/connection"

	"gopkg.in/yaml.v2"
)

type Config struct {
	PConfig connection.PacketConfig `json:"packet_config" yaml:"packet_config"`
	Host    string                  `json:"host" yaml:"host"`
	Port    int                     `json:"port" yaml:"port"`
}

type HostInfo struct {
	Host string `json:"host" yaml:"host"`
	Port int    `json:"port" yaml:"port"`
}

func (c *Config) Load(path string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(content, c)
}

func (c *Config) Yaml(path string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(content, c)
}
