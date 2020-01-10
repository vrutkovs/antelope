package main

import (
	"context"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"net/http"

	"cloud.google.com/go/storage"

	"github.com/vrutkovs/antelope/pkg/gcs"
)

type ServerSettings struct {
	gcsBucket *storage.BucketHandle
}

// health is k8s endpoint for liveness check
func healthz(c *gin.Context) {
	c.String(http.StatusOK, "")
}

func (s *ServerSettings) job(c *gin.Context) {
	jobName := c.Params.ByName("name")
	if len(jobName) == 0 {
		return
	}

	ctx := context.Background()

	// TODO: don't filter output in ListBucket, use LRU
	jobIDs, err := gcs.ListBucket(s.gcsBucket, ctx, jobName, 0, 40)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "{'message': 'internal error'}")
		return
	}

	c.JSON(http.StatusOK, jobIDs)
	// TODO: fetch latest job ID
	// TODO: paginate
	// TODO: fetch job status
	// TODO: send job results via websocket
}

func main() {
	r := gin.New()

	// Server static HTML
	r.Use(static.Serve("/", static.LocalFile("./html", true)))

	// Don't log k8s health endpoint
	r.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/healthz"),
		gin.Recovery(),
	)
	r.GET("/healthz", healthz)

	// Prepare server settings
	s := &ServerSettings{
		gcsBucket: gcs.GetGCSBucket(),
	}
	// Add job route
	r.GET("/job/:name", s.job)

	r.Run(":8080")
}
