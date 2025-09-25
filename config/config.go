package config

import (
	"os"

	hcl "github.com/hashicorp/hcl/v2/hclsimple"
)

type Config struct {
	PodName  string
	Host     string    `hcl:"host"`
	Port     int32     `hcl:"port"`
	Forwards []Forward `hcl:"forward,block"`
}

type Forward struct {
	Name      string     `hcl:"name,label"`
	Templates *Templates `hcl:"templates,block"`
}

type Templates struct {
	Rewrites []Rewrite `hcl:"rewrite,block"`
}

type Rewrite struct {
	Name string `hcl:"name,label"`
	To   string `hcl:"to,attr"`
}

func Load(filename string) (*Config, error) {
	appConfig := Config{}
	if err := hcl.DecodeFile(filename, nil, &appConfig); err != nil {
		return nil, err
	}
	appConfig.PodName = os.Getenv("POD_NAME")
	return &appConfig, nil
}
