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
	ListenPort      int         `json:"listen_port" yaml:"listen_port"`
	WriteTimeout    int         `json:"write_timeout" yaml:"write_timeout"`
	ReadTimeout     int         `json:"read_timeout" yaml:"read_timeout"`
	GracefulTimeout int         `json:"graceful_timeout" yaml:"graceful_timeout"`
	IdleTimeout     int         `json:"idle_timeout" yaml:"idle_timeout"`
	Swagger         bool        `json:"swagger" yaml:"swagger"`
	LocalSwagger    bool        `json:"local_swagger" yaml:"local_swagger"`
	Schema          string      `json:"schema" yaml:"schema"`
	App             string      `json:"app" yaml:"app"`
	Host            string      `json:"host" yaml:"host"`
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
