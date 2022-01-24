package api

import (
	"fmt"

	"github.com/imdario/mergo"
)

type JSONResult struct {
	Code    int         `json:"code" `
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
type ApiServerConfig struct {
	ListenPort      int         `json:"listen_port" yaml:"listen_port"`
	WriteTimeout    int         `json:"write_timeout" yaml:"write_timeout"`
	ReadTimeout     int         `json:"read_timeout" yaml:"read_timeout"`
	GracefulTimeout int         `json:"graceful_timeout" yaml:"graceful_timeout"`
	IdleTimeout     int         `json:"idle_timeout" yaml:"idle_timeout"`
	Swagger         bool        `json:"swagger" yaml:"swagger"`
	Prometheus      bool        `json:"prometheus" yaml:"prometheus"`
	LocalSwagger    bool        `json:"local_swagger" yaml:"local_swagger"`
	Schema          string      `json:"schema" yaml:"schema"`
	App             string      `json:"app" yaml:"app"`
	Host            string      `json:"host" yaml:"host"`
	ApiHost         string      `json:"api_host" yaml:"api_host"`
	AllowedOrigins  []string    `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedHeaders  []string    `json:"allowed_header" yaml:"allowed_header"`
	AllowedMethods  []string    `json:"allowed_methods" yaml:"allowed_methods"`
	AppConfig       interface{} `json:"-"`
}

func (con *ApiServerConfig) ApiServerConfigUpdate(conf ApiServerConfig, config interface{}) {
	err := mergo.MapWithOverwrite(con, conf)
	if err != nil {
		Log.Error("Cannot Merge data: ", err)
	}
	con.AppConfig = config
	if con.LocalSwagger {
		con.ApiHost = fmt.Sprintf("%s:%d", con.Host, con.ListenPort)
	} else {
		con.ApiHost = con.Host
	}
}

func (con *ApiServerConfig) InitializeApiServerConfig(conf ApiServerConfig, config interface{}) {
	allowedOrigins := []string{"*"}
	allowedHeaders := []string{"X-Requested-With", "Content-Type", "Authorization",
		"SERVICE-AGENT", "Access-Control-Allow-Methods", "Date", "X-FORWARDED-FOR", "Accept",
		"Content-Length", "Accept-Encoding", "Service-Agent"}
	allowedMethods := []string{"GET", "POST", "PUT", "PATCH", "HEAD", "OPTIONS"}
	err := mergo.Merge(con, ApiServerConfig{
		ListenPort:      8765,
		WriteTimeout:    30,
		ReadTimeout:     30,
		GracefulTimeout: 15,
		IdleTimeout:     60,
		Swagger:         false,
		Prometheus:      true,
		LocalSwagger:    false,
		Schema:          "https",
		AllowedOrigins:  allowedOrigins,
		AllowedHeaders:  allowedHeaders,
		AllowedMethods:  allowedMethods,
	})
	if err != nil {
		Log.Error("Cannot Merge data: ", err)
	}
	con.ApiServerConfigUpdate(conf, config)
}
