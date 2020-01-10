package web

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/vrutkovs/antelope/pkg/cache"
	"github.com/vrutkovs/antelope/pkg/gcs"
	"github.com/vrutkovs/antelope/pkg/job"
)

func (s *Settings) job(c *gin.Context) {
	jobName := c.Params.ByName("name")
	if len(jobName) == 0 {
		return
	}

	ctx := context.Background()

	// TODO: don't filter output in ListBucket, use LRU
	jobIDs, err := gcs.ListBucket(s.GcsBucket, ctx, jobName, 0, 40)
	fmt.Printf("Found JobIDs %d -> %d\n", jobIDs[len(jobIDs)-1], jobIDs[0])
	if err != nil {
		c.JSON(http.StatusInternalServerError, "{'message': 'internal error'}")
		return
	}

	fmt.Printf("Initialised cache\n")
	cache := &cache.Cache{
		Bucket: s.GcsBucket,
	}

	for _, id := range jobIDs {
		j := &job.Job{
			Name:   jobName,
			ID:     id,
			Bucket: s.GcsBucket,
			Cache:  cache,
		}
		if err := j.GetBasicInfo(); err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("Created job %s with ID %d\n", jobName, id)

		jobClusterType, err := j.GetClusterType()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("type: %s\n", jobClusterType)
	}
	c.JSON(http.StatusOK, jobIDs)
	// TODO: fetch job status
	// TODO: send job results via websocket
}
