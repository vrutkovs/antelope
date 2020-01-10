package gcs

import (
	"context"
	"fmt"
	"io/ioutil"
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

func FetchFile(bucket *storage.BucketHandle, ctx context.Context, gcsPath string) ([]byte, error) {

	if !hasFile(bucket, ctx, gcsPath) {
		return nil, fmt.Errorf("No such path: %s", gcsPath)
	}

	rc, err := bucket.Object(gcsPath).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func hasFile(bucket *storage.BucketHandle, ctx context.Context, path string) bool {
	it := bucket.Objects(ctx, &storage.Query{
		Prefix:    path,
		Delimiter: "/",
	})
	for {
		_, err := it.Next()

		if err == iterator.Done {
			break
		}
		if err != nil {
			return false
		}
		return true
	}
	return false
}
