package web

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/vrutkovs/antelope/pkg/gcs"
	"github.com/vrutkovs/antelope/pkg/job"
)

func (s *Settings) listJobIDs(c *gin.Context) {
	jobName := c.Params.ByName("name")
	if len(jobName) == 0 {
		c.JSON(http.StatusNotFound, nil)
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

	result := jobIDs[:0]

	for _, id := range jobIDs {
		j := &job.Job{
			Name:   jobName,
			ID:     id,
			Bucket: s.GcsBucket,
			Cache:  s.Cache,
		}
		if err := j.GetBasicInfo(); err != nil {
			// Skip the element, its running or broken
			continue
		}
		fmt.Printf("Created job %s with ID %d\n", jobName, id)
		result = append(result, id)
	}
	c.JSON(http.StatusOK, result)
}

func (s *Settings) getJobInfo(c *gin.Context) {
	jobName := c.Params.ByName("name")
	if len(jobName) == 0 {
		c.JSON(http.StatusNotFound, nil)
		return
	}
	strJobId := c.Params.ByName("id")
	if len(strJobId) == 0 {
		c.JSON(http.StatusNotFound, nil)
		return
	}
	jobId, err := strconv.Atoi(strJobId)
	if err != nil {
		c.JSON(http.StatusNotFound, nil)
		return
	}

	j := &job.Job{
		Name:   jobName,
		ID:     jobId,
		Bucket: s.GcsBucket,
		Cache:  s.Cache,
	}
	if err := j.GetBasicInfo(); err != nil {
		c.JSON(http.StatusNotFound, nil)
		return
	}

	// Skip errors here as these are being checked in GetBasicInfo
	clusterType, _ := j.GetClusterType()
	artifactsSubdir, _ := j.GetArtifactsSubdir()
	buildLogUrl := j.GetBuildLogUrl()

	c.JSON(http.StatusOK, gin.H{
		"type":          clusterType,
		"artifacts":     artifactsSubdir,
		"build_log_url": buildLogUrl,
	})
}
