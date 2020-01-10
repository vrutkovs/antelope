package web

import (
	"github.com/gin-gonic/gin"
	"cloud.google.com/go/storage"

	"github.com/vrutkovs/antelope/pkg/gcs"
)

type Settings struct {
	GcsBucket *storage.BucketHandle
}

func SetGinRoutes(r *gin.Engine) {
	// Prepare server settings
	s := &Settings{
		GcsBucket: gcs.GetGCSBucket(),
	}
	// Add job route
	r.GET("/job/:name", s.job)
}
