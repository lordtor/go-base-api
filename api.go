package go_base_api

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"encoding/json"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	common_lib "github.com/lordtor/go-common-lib"
	logging "github.com/lordtor/go-logging"
	version "github.com/lordtor/go-version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	muxprom "gitlab.com/msvechla/mux-prometheus/pkg/middleware"
)

var (
	DefaultCT = []string{"Content-Type", "application/json"}
	Log       = logging.Log
	Con       = ApiServer{}
)

func HTTP(con ApiServer, config interface{}, r *mux.Router) {
	Con.Update(con)
	Con.AppConfig = config
	hostAPI := ""
	if Con.LocalSwagger {
		hostAPI = fmt.Sprintf("%s:%d", Con.Host, Con.ListenPort)
	} else {
		hostAPI = Con.Host
	}
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*time.Duration(Con.GracefulTimeout), "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	instrumentation := muxprom.NewDefaultInstrumentation()
	header := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "SERVICE-AGENT", "Access-Control-Allow-Methods", "Date", "X-FORWARDED-FOR", "Accept", "Content-Length", "Accept-Encoding", "Service-Agent"})
	credentials := handlers.AllowCredentials()
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "HEAD", "OPTIONS"})
	allowedOrigins := []string{"*"}
	allowedOrigins = common_lib.UpdateStructList(allowedOrigins, Con.AllowedOrigins)
	allowedOrigins = common_lib.UpdateList(allowedOrigins, hostAPI)
	origins := handlers.AllowedOrigins(allowedOrigins)
	r.Use(instrumentation.Middleware)
	if Con.Swagger {
		if Con.LocalSwagger {
			r.PathPrefix("/swagger/").Handler(
				httpSwagger.Handler(
					httpSwagger.URL(fmt.Sprintf("%s://%s:%d/swagger/doc.json", Con.Schema, Con.Host, Con.ListenPort)),
					httpSwagger.DeepLinking(true),
					httpSwagger.DocExpansion("none"),
					httpSwagger.DomID("#swagger-ui"),
				),
			)
		} else {
			r.PathPrefix("/swagger/").Handler(
				httpSwagger.Handler(
					httpSwagger.URL(fmt.Sprintf("%s://%s/direct-container-url/%s/swagger/doc.json", Con.Schema, Con.Host, Con.App)),
					httpSwagger.DeepLinking(true),
					httpSwagger.DocExpansion("none"),
					httpSwagger.DomID("#swagger-ui"),
				),
			)
		}
	}
	r.Use(mux.CORSMethodMiddleware(r))
	srv := &http.Server{
		Handler:      handlers.CORS(header, credentials, methods, origins)(r),
		Addr:         fmt.Sprint(":", Con.ListenPort),
		WriteTimeout: time.Duration(Con.WriteTimeout) * time.Second,
		ReadTimeout:  time.Duration(Con.ReadTimeout) * time.Second,
		IdleTimeout:  time.Second * time.Duration(Con.IdleTimeout),
	}

	// -- Internal api
	Mount(r, "/", RouterApi())
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logging.Log.Println(err)
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)
}

func Mount(r *mux.Router, path string, handler http.Handler) {
	r.PathPrefix(path).Handler(
		http.StripPrefix(
			strings.TrimSuffix(path, "/"),
			handler,
		),
	)
}

func RouterApi() *mux.Router {
	router := mux.NewRouter()
	router.Path("/prometheus").Handler(promhttp.Handler()).Methods("GET")
	router.Path("/env").HandlerFunc((ShowConfig)).Methods("GET")
	router.Path("/health").HandlerFunc(Health).Methods("GET")
	router.Path("/info").HandlerFunc(ShowInfo).Methods("GET")
	return router
}

func Resp(data *JSONResult, w http.ResponseWriter) {
	w.Header().Set(DefaultCT[0], DefaultCT[1])
	resp, err := json.Marshal(data)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(data.Code)
	intE, err := w.Write(resp)
	if err != nil {
		http.Error(w, fmt.Sprint(err), intE)
		return
	}
}

// ShowInfo godoc
// @Summary Info a service
// @Tags internal
// @Description get information about service
// @Accept  json
// @Produce  json
// @Success 200 {object} JSONResult "desc"
// @Failure 400,404 {object} JSONResult
// @Router /info [get]
// @BasePath /
func ShowInfo(w http.ResponseWriter, r *http.Request) {
	Resp(&JSONResult{
		Code:    http.StatusOK,
		Message: "",
		Data:    version.GetVersion(),
	}, w)
}

// ShowConfig godoc
// @Summary Show config for service
// @Tags internal
// @Description Internal method
// @Accept  json
// @Produce  json
// @Success 200 {object}  JSONResult{data=ApiServer} "desc"
// @Failure 400,404,405 {object} JSONResult
// @Failure 500 {object} JSONResult
// _Security ApiKeyAuth
// @Router /conf [get]
func ShowConfig(w http.ResponseWriter, r *http.Request) {
	Resp(&JSONResult{
		Code: http.StatusOK,
		Data: Con.AppConfig}, w)
}

// Health godoc
// @Summary Health check
// @Tags internal
// @Description Internal method
// @Accept  json
// @Produce  json
// @Success 200 {object}  JSONResult{data=HealthCheck} "desc"
// @Failure 400,404 {object} JSONResult
// @Failure 500 {object} JSONResult
// @Router /health [get]
func Health(w http.ResponseWriter, r *http.Request) {
	respData := JSONResult{Code: http.StatusOK, Data: HealthCheck{Alive: true}, Message: ""}
	Resp(&respData, w)
}
