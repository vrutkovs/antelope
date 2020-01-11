package main

import (
	"github.com/gin-gonic/contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/vrutkovs/antelope/pkg/web"
)

// health is k8s endpoint for liveness check
func healthz(c *gin.Context) {
	c.String(http.StatusOK, "")
}

func main() {
	r := gin.New()

	// Allow all
	r.Use(cors.Default())

	// Don't log k8s health endpoint
	r.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/healthz"),
		gin.Recovery(),
	)
	r.GET("/healthz", healthz)

	web.SetGinRoutes(r)

	r.Run(":3000")
}
