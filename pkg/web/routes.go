package web

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"

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
	if err != nil {
		c.JSON(http.StatusInternalServerError, "{'message': 'internal error'}")
		return
	}

	for id := range jobIDs {
		j := job.Job{
			Name:   jobName,
			ID:     string(id),
			Bucket: s.GcsBucket,
		}

		jobClusterType, err := j.GetClusterType()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(jobClusterType)
	}
	c.JSON(http.StatusOK, jobIDs)
	// TODO: fetch job status
	// TODO: send job results via websocket
}
