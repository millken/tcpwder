package api

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/millken/tcpwder/config"
)

/* gin app */
var app *gin.Engine

/**
 * Initialize module
 */
func init() {
	gin.SetMode(gin.ReleaseMode)
}

/**
 * Starts REST API server
 */
func Start(cfg config.ApiConfig) {

	if !cfg.Enabled {
		log.Printf("[INFO] API disabled")
		return
	}

	log.Printf("[INFO] Starting up API")

	app = gin.New()

	if cfg.Cors {
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowCredentials = true
		corsConfig.AllowMethods = []string{"PUT", "POST", "DELETE", "GET", "OPTIONS"}
		corsConfig.AllowHeaders = []string{"Origin", "Authorization"}

		app.Use(cors.New(corsConfig))
		log.Printf("[INFO] API CORS enabled")
	}

	r := app.Group("/")

	if cfg.BasicAuth != nil {
		log.Printf("[INFO] Using HTTP Basic Auth")
		r.Use(gin.BasicAuth(gin.Accounts{
			cfg.BasicAuth.Login: cfg.BasicAuth.Password,
		}))
	}

	/* attach endpoints */
	attachRoot(r)
	attachServers(r)

	var err error
	/* start rest api server */
	if cfg.Tls != nil {
		log.Printf("[INFO] Starting HTTPS server %s", cfg.Bind)
		err = app.RunTLS(cfg.Bind, cfg.Tls.CertPath, cfg.Tls.KeyPath)
	} else {
		log.Printf("[INFO] Starting HTTP server %s", cfg.Bind)
		err = app.Run(cfg.Bind)
	}

	if err != nil {
		log.Fatal(err)
	}

}
