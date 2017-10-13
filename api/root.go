package api

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/millken/tcpwder/config"
	"github.com/millken/tcpwder/manager"
)

/**
 * Attaches / handlers
 */
func attachRoot(app *gin.RouterGroup) {

	/**
	 * Global stats
	 */
	app.GET("/", func(c *gin.Context) {

		c.IndentedJSON(http.StatusOK, gin.H{
			"pid":           os.Getpid(),
			"time":          time.Now(),
			"startTime":     config.StartTime,
			"uptime":        time.Now().Sub(config.StartTime).String(),
			"version":       config.Version,
			"configuration": config.Configuration,
		})
	})

	/**
	 * Dump current config as TOML
	 */
	app.GET("/dump", func(c *gin.Context) {
		format := c.DefaultQuery("format", "toml")

		data, err := manager.DumpConfig(format)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err.Error())
			return
		}

		c.String(http.StatusOK, data)
	})
}
