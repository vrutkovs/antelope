package web

import (
	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"

	"github.com/vrutkovs/antelope/pkg/cache"
	"github.com/vrutkovs/antelope/pkg/gcs"
)

type Settings struct {
	GcsBucket *storage.BucketHandle
	Cache     *cache.Cache
}

func SetGinRoutes(r *gin.Engine) {
	// Prepare server settings
	s := &Settings{
		GcsBucket: gcs.GetGCSBucket(),
	}
	s.Cache = &cache.Cache{
		Bucket: s.GcsBucket,
	}

	// Add job route
	r.GET("/api/job/:name", s.listJobIDs)
	r.GET("/api/job/:name/:id", s.getJobInfo)
}
