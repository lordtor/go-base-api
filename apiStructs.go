package go_base_api

import (
	"github.com/imdario/mergo"
)

type JSONResult struct {
	Code    int         `json:"code" `
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type HealthCheck struct {
	Alive bool `json:"alive"`
}

type ApiServer struct {
	ListenPort      int         `yaml:"listen_port"`
	WriteTimeout    int         `yaml:"write_timeout"`
	ReadTimeout     int         `yaml:"read_timeout"`
	GracefulTimeout int         `yaml:"graceful_timeout"`
	IdleTimeout     int         `yaml:"idle_timeout"`
	Swagger         bool        `yaml:"swagger"`
	LocalSwagger    bool        `yaml:"local_swagger"`
	Schema          string      `yaml:"schema"`
	App             string      `yaml:"app"`
	Host            string      `yaml:"host"`
	AppConfig       interface{} `json:"-"`
}

func (c *ApiServer) Update(conf ApiServer) {
	err := mergo.MapWithOverwrite(c, conf)
	if err != nil {
		Log.Error("Cannot Merge data: ", err)
	}
}

func (c *ApiServer) init() {
	err := mergo.Merge(c, ApiServer{
		ListenPort:      8765,
		WriteTimeout:    30,
		ReadTimeout:     30,
		GracefulTimeout: 15,
		IdleTimeout:     60,
		Swagger:         false,
		LocalSwagger:    false,
		Schema:          "https",
	})
	if err != nil {
		Log.Error("Cannot Merge data: ", err)
	}
}
