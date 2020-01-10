package gcs

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func ListBucket(bucket *storage.BucketHandle, ctx context.Context, job string, start int, max int) ([]int, error) {
	var jobIDs []int
	prefix := basePrefix + job + "/"
	it := bucket.Objects(ctx, &storage.Query{
		Prefix:    prefix,
		Delimiter: "/",
	})
	fmt.Printf("Fetching objects in bucket for prefix %s\n", prefix)
	for {
		attrs, err := it.Next()

		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		strJobID := strings.Replace(strings.Replace(attrs.Prefix, prefix, "", 1), "/", "", 1)
		jobID, err := strconv.Atoi(strJobID)
		if err != nil {
			fmt.Printf("Failed to convert %s to integer\n", strJobID)
			continue
		}
		jobIDs = append(jobIDs, jobID)
	}
	// Sort jobIDs
	sort.Sort(sort.Reverse(sort.IntSlice(jobIDs)))
	return jobIDs[start:max], nil
}
