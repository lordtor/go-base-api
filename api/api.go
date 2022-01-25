package api

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/imdario/mergo"
	common_lib "github.com/lordtor/go-common-lib"
	logging "github.com/lordtor/go-logging"
	trace "github.com/lordtor/go-trace-lib"
	version "github.com/lordtor/go-version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swagger "github.com/swaggo/http-swagger"
	muxprom "gitlab.com/msvechla/mux-prometheus/pkg/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
)

var (
	DefaultCT = []string{"Content-Type", "application/json"}
	Log       = logging.Log
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
		ListenPort:      8080,
		WriteTimeout:    30,
		ReadTimeout:     30,
		GracefulTimeout: 15,
		IdleTimeout:     60,
		Swagger:         false,
		Prometheus:      false,
		LocalSwagger:    false,
		Schema:          "http",
		AllowedOrigins:  allowedOrigins,
		AllowedHeaders:  allowedHeaders,
		AllowedMethods:  allowedMethods,
	})
	if err != nil {
		Log.Error("Cannot Merge data: ", err)
	}
	con.ApiServerConfigUpdate(conf, config)
}

type API struct {
	Router *mux.Router
	Config ApiServerConfig
}

func (a *API) Initialize(conf ApiServerConfig, config interface{}) {
	a.Router = mux.NewRouter()
	a.Config.InitializeApiServerConfig(conf, config)
	a.InitializeSwagger()
	a.InitializePrometheus()
	a.Router.Use(Logging)
	a.Router.Use(PanicRecovery)
	a.Router.Use(otelmux.Middleware(a.Config.App))
	a.initializeBaseRoutes()

}

func (a *API) InitializeSwagger() {
	if a.Config.Swagger {
		if a.Config.LocalSwagger {
			a.Router.PathPrefix("/swagger/").Handler(
				swagger.Handler(
					swagger.URL(fmt.Sprintf("%s://%s:%d/swagger/doc.json", a.Config.Schema, a.Config.Host, a.Config.ListenPort)),
					swagger.DeepLinking(true),
					swagger.DocExpansion("none"),
					swagger.DomID("#swagger-ui"),
				),
			)
		} else {
			a.Router.PathPrefix("/swagger/").Handler(
				swagger.Handler(
					swagger.URL(fmt.Sprintf("%s://%s/direct-container-url/%s/swagger/doc.json", a.Config.Schema, a.Config.Host, a.Config.App)),
					swagger.DeepLinking(true),
					swagger.DocExpansion("none"),
					swagger.DomID("#swagger-ui"),
				),
			)
		}
	}
}
func (a *API) InitializePrometheus() {
	if a.Config.Prometheus {
		instrumentation := muxprom.NewDefaultInstrumentation()
		a.Router.Use(instrumentation.Middleware)
		a.Router.HandleFunc("/prometheus", otelhttp.NewHandler(promhttp.Handler(), "Prometheus").ServeHTTP).Methods(http.MethodGet)
	}
}
func (a *API) InitializeCORS() (header handlers.CORSOption, credentials handlers.CORSOption,
	methods handlers.CORSOption, origins handlers.CORSOption) {
	a.Router.Use(mux.CORSMethodMiddleware(a.Router))
	header = handlers.AllowedHeaders(a.Config.AllowedHeaders)
	credentials = handlers.AllowCredentials()
	methods = handlers.AllowedMethods(a.Config.AllowedMethods)
	allowedOrigins := []string{}
	allowedOrigins = common_lib.UpdateStructList(allowedOrigins, a.Config.AllowedOrigins)
	allowedOrigins = common_lib.UpdateList(allowedOrigins, a.Config.ApiHost)
	origins = handlers.AllowedOrigins(allowedOrigins)
	return header, credentials, methods, origins
}

func (a *API) Run() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*time.Duration(a.Config.GracefulTimeout), "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()
	srv := &http.Server{
		Handler:      handlers.CORS(a.InitializeCORS())(a.Router),
		Addr:         fmt.Sprint(":", a.Config.ListenPort),
		WriteTimeout: time.Duration(a.Config.WriteTimeout) * time.Second,
		ReadTimeout:  time.Duration(a.Config.ReadTimeout) * time.Second,
		IdleTimeout:  time.Second * time.Duration(a.Config.IdleTimeout),
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			Log.Error(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)
	// Block until we receive our signal.
	<-c
	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	err := srv.Shutdown(ctx)
	if err != nil {
		Log.Error(err.Error())
	}
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	Log.Info("shutting down")
	os.Exit(0)
}

func (a *API) Mount(path string, handler http.Handler) {
	a.Router.PathPrefix(path).Handler(
		http.StripPrefix(strings.TrimSuffix(path, "/"), handler),
	)
}

func (a *API) initializeBaseRoutes() {
	a.Router.HandleFunc("/prometheus", promhttp.Handler().ServeHTTP).Methods(http.MethodGet)
	a.Router.HandleFunc("/env", a.ShowConfig()).Methods(http.MethodGet)
	a.Router.HandleFunc("/health", a.Health()).Methods(http.MethodGet)
	a.Router.HandleFunc("/info", a.ShowInfo()).Methods(http.MethodGet)
}

// ShowInfo godoc
// @Summary Show config for service
// @Tags internal
// @Description Internal method
// @Accept  json
// @Produce  json
// @Success 200 {object}  JSONResult "desc"
// @Failure 400,404,405 {object} JSONResult
// @Failure 500 {object} JSONResult
// _Security ApiKeyAuth
// @Router /info [get]
func (a *API) ShowInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ver := getVersion(r.Context())
		a.Resp(&JSONResult{
			Code:    http.StatusOK,
			Message: "",
			Data:    ver,
		}, w, r.Context())
	}
}
func getVersion(ctx context.Context) version.ApplicationVersion {
	v := version.GetVersion()
	_, span := trace.NewSpan(ctx, "ShowInfo.getVersion", nil)
	defer span.End()
	span.SetAttributes(attribute.String("BuildTimeStamp", v.BuildTimeStamp),
		attribute.String("GitBranch", v.GitBranch),
		attribute.String("GitHash", v.GitHash),
		attribute.String("Version", v.Version))
	span.SetStatus(2, "")
	return v
}

// ShowConfig godoc
// @Summary Show config for service
// @Tags internal
// @Description Internal method
// @Accept  json
// @Produce  json
// @Success 200 {object}  JSONResult "desc"
// @Failure 400,404,405 {object} JSONResult
// @Failure 500 {object} JSONResult
// _Security ApiKeyAuth
// @Router /env [get]
func (a *API) ShowConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.Resp(&JSONResult{
			Code: http.StatusOK,
			Data: getConfig(r.Context(), a.Config.AppConfig)}, w, r.Context())
	}
}
func getConfig(ctx context.Context, con interface{}) interface{} {

	_, span := trace.NewSpan(ctx, "ShowInfo.getVersion", nil)
	defer span.End()
	ravC, err := json.Marshal(con)
	if err != nil {
		Log.Error(err)
		span.SetStatus(1, "json.Marshal")
		span.RecordError(err)
	}
	c := string(ravC)
	span.SetAttributes(attribute.Key("Env").String(string(c)))
	span.SetStatus(2, "")
	return con
}

// Health godoc
// @Summary Health check
// @Tags internal
// @Description Internal method
// @Accept  json
// @Produce  json
// @Success 200 {object}  JSONResult "desc"
// @Failure 400,404 {object} JSONResult
// @Failure 500 {object} JSONResult
// @Router /health [get]
func (a *API) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respData := JSONResult{Code: http.StatusOK, Data: map[string]bool{"Alive": true}, Message: ""}
		a.Resp(&respData, w, r.Context())
	}
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, req)
		Log.Debugf("%s %s %s", req.Method, req.RequestURI, time.Since(start))
	})
}

func PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				Log.Error(err)
			}
		}()
		next.ServeHTTP(w, req)
	})
}

func (a *API) Resp(data *JSONResult, w http.ResponseWriter, ctx context.Context) {
	_, span := trace.NewSpan(ctx, "Resp", nil)
	defer span.End()
	w.Header().Set(DefaultCT[0], DefaultCT[1])
	resp, err := json.Marshal(data)
	span.SetStatus(2, data.Message)
	span.SetAttributes(attribute.Key("Data").String(string(resp)))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(1, data.Message)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(data.Code)
	intE, err := w.Write(resp)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(1, data.Message)
		http.Error(w, err.Error(), intE)
		return
	}
}
