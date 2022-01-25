// swagger:meta
// @termsOfService https://sample.url

// @contact.name API Support
// @contact.url https://jira.url
// @contact.email admin@mail.ru

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// _securityDefinitions.apikey ApiKeyAuth
// _in header
// _name Authorization

package main

import (
	"context"
	"fmt"

	"github.com/lordtor/go-base-api/cmd/docs"

	api "github.com/lordtor/go-base-api"
	ex "github.com/lordtor/go-base-api/internal/pkg/extend_config"
	trace "github.com/lordtor/go-trace-lib"

	logging "github.com/lordtor/go-logging"
	version "github.com/lordtor/go-version"
)

var (
	Log  = logging.Log
	Conf = ex.C{}
	// Bootstrap local & bin version.
	binVersion      = "0.1.1"
	aBuildNumber    = "00000"
	aBuildTimeStamp = ""
	aGitBranch      = "master"
	aGitHash        = ""
)

func init() {
	// Bootstrap configs and logging.
	logging.InitLog("")
	version.InitVersion(binVersion, aBuildNumber, aBuildTimeStamp, aGitBranch, aGitHash)
	Conf.ReloadConfig()
	Conf.API.App = Conf.AppName
	Conf.Trace.Environment = Conf.ProfileName
	Conf.Trace.ServiceName = Conf.AppName
	Conf.Trace.ServiceVersion = version.AppVersion.Version
	Conf.API.InitializeApiServerConfig(Conf.API, Conf)
	logging.ChangeLogLevel(Conf.LogLevel)
	logging.Log.Error(Conf)
}
func main() {
	ctx := context.Background()
	// Bootstrap swagger
	if Conf.API.Swagger {
		docs.SwaggerInfo.Title = fmt.Sprintf("Swagger  %s", Conf.AppName)
		docs.SwaggerInfo.Version = version.GetVersion().Version
		docs.SwaggerInfo.BasePath = fmt.Sprintf("/%s", Conf.AppName)
		docs.SwaggerInfo.Description = "Basic only internal methods!"
		docs.SwaggerInfo.Schemes = []string{Conf.API.Schema}
		docs.SwaggerInfo.Host = Conf.API.ApiHost
	}
	// Bootstrap tracer.
	prv, err := trace.NewProvider(Conf.Trace)
	if err != nil {
		Log.Fatalln(err)
	}
	defer prv.Close(ctx)
	// Bootstrap api.
	a := api.API{}
	a.Initialize(Conf.API, Conf)
	// Mount anothe routes
	// a.Mount(fmt.Sprintf("/%s/anothe api/", Conf.AppName), anothe.Routes(Conf.Anothe))
	a.Run()
}
