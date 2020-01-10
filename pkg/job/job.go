package job

import (
	"bufio"
	"cloud.google.com/go/storage"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/vrutkovs/antelope/pkg/cache"
)

type Job struct {
	Name string
	ID   string

	Bucket *storage.BucketHandle

	cache    *cache.Cache
	testType string
}

func (j *Job) subPath(path string) string {
	return j.Name + "/" + j.ID + "/" + path
}

func (j *Job) fetch(file string) (io.Reader, error) {
	if j.cache == nil {
		j.cache = &cache.Cache{
			Bucket: j.Bucket,
		}
	}
	return j.cache.Get(file)
}

func (j Job) StartTime() (time.Time, error) {
	f, err := j.fetch(j.subPath("/started.json"))
	if err != nil {
		return time.Time{}, err
	}

	var started metadata
	if err := json.NewDecoder(f).Decode(&started); err != nil {
		return time.Time{}, err
	}

	return started.time, nil
}

func (j Job) FinishTime() (time.Time, error) {
	f, err := j.fetch(j.subPath("/finished.json"))
	if err != nil {
		return time.Time{}, err
	}

	var finished metadata
	if err := json.NewDecoder(f).Decode(&finished); err != nil {
		return time.Time{}, err
	}

	return finished.time, nil
}

func (j Job) GetClusterType() (string, error) {
	f, err := j.fetch(j.subPath("artifacts/build-resources/templateinstances.json"))
	if err != nil {
		return "", err
	}
	var inst templateinstance
	if err := json.NewDecoder(f).Decode(&inst); err != nil {
		return "", err
	}
	for _, params := range inst.Items[0].Spec.Template.Parameters {
		if params.Name == "CLUSTER_TYPE" {
			return params.Value, nil
		}
	}
	return "", fmt.Errorf("Failed to find cluster type: %s", "no CLUSTER_TYPE param found")
}

func (j Job) Result() (string, error) {
	f, err := j.fetch(j.subPath("/finished.json"))
	if err != nil {
		return "", err
	}

	var finished metadata
	if err := json.NewDecoder(f).Decode(&finished); err != nil {
		return "", err
	}

	return finished.result, nil
}

func (j Job) BuildLog() (io.Reader, error) {
	return j.fetch(j.subPath("/build-log.txt"))
}

func (j Job) Machines() (io.Reader, error) {
	return j.fetch(j.subPath("/artifacts" + j.testType + "/machines.json"))
}

func (j Job) Nodes() (io.Reader, error) {
	return j.fetch(j.subPath("/artifacts" + j.testType + "/openstack_nodes.log"))
}

func (j Job) JUnitURL() (string, error) {
	const (
		target       = "Writing JUnit report to /tmp/artifacts/junit/"
		targetLength = len(target)
	)

	buildLog, err := j.BuildLog()
	if err != nil {
		return "", err
	}

	// The default scanner buffer (60*1024 bytes) is too short for some
	// build logs. This sets the initial buffer capacity to the package
	// default, but a higher maximum value.
	scanner := bufio.NewScanner(buildLog)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) >= targetLength && line[:len(target)] == target {
			filename := line[len(target):]
			return "/artifacts/" + j.Name + "/junit/" + filename, scanner.Err()
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", io.EOF
}

func (j Job) JUnit() (io.Reader, error) {
	u, err := j.JUnitURL()
	if err != nil {
		return nil, err
	}
	return j.fetch(u)
}
