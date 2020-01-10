package main

import (
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"net/http"
)

// health is k8s endpoint for liveness check
func health(c *gin.Context) {
	c.String(http.StatusOK, "")
}

func job(c *gin.Context) {
	jobName := c.Params.ByName("name")
	if len(jobName) == 0 {
		return
	}

	// TODO: fetch latest job ID
	// TODO: paginate
	// TODO: fetch job status
	// TODO: send job results via webhook
}

func main() {
	r := gin.New()

	// Server static HTML
	r.Use(static.Serve("/", static.LocalFile("./html", true)))

	// Don't log k8s health endpoint
	r.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/health"),
		gin.Recovery(),
	)
	r.GET("/health", health)

	r.GET("/job/:name", job)

	r.Run(":8080")
}
