package job

import (
	"bufio"
	"cloud.google.com/go/storage"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/vrutkovs/antelope/pkg/cache"
)

type Job struct {
	Name string
	ID   int

	Bucket *storage.BucketHandle
	Cache  *cache.Cache

	artifactsSubdir string
	clusterType     string
}

func (j *Job) subPath(path string) string {
	return j.Name + "/" + strconv.Itoa(j.ID) + "/" + path
}

func (j *Job) fetch(file string) (io.Reader, error) {
	fmt.Printf("job: Fetching %s\n", file)
	return j.Cache.Get(file)
}

func (j *Job) StartTime() (time.Time, error) {
	f, err := j.fetch(j.subPath("started.json"))
	if err != nil {
		return time.Time{}, err
	}

	var started metadata
	if err := json.NewDecoder(f).Decode(&started); err != nil {
		return time.Time{}, err
	}

	return started.time, nil
}

func (j *Job) FinishTime() (time.Time, error) {
	f, err := j.fetch(j.subPath("finished.json"))
	if err != nil {
		return time.Time{}, err
	}

	var finished metadata
	if err := json.NewDecoder(f).Decode(&finished); err != nil {
		return time.Time{}, err
	}

	return finished.time, nil
}

func (j *Job) GetClusterType() (string, error) {
	if j.clusterType != "" {
		return j.clusterType, nil
	}

	f, err := j.fetch(j.subPath("artifacts/build-resources/templateinstances.json"))
	if err != nil {
		return "", err
	}

	var tParams templateParams
	if err := json.NewDecoder(f).Decode(&tParams); err != nil {
		return "", err
	}
	for _, params := range tParams.Parameters {
		if params.Name == "CLUSTER_TYPE" {
			j.clusterType = params.Value
			return params.Value, nil
		}
	}
	return "", fmt.Errorf("Failed to find cluster type: %s", "no CLUSTER_TYPE param found")
}

func (j *Job) GetBasicInfo() error {
	var err error
	fmt.Printf("Fetching basic info for %s #%d\n", j.Name, j.ID)
	if _, err = j.FinishTime(); err != nil {
		return errors.New("Test is not yet complete, skipping")
	}
	if _, err = j.GetClusterType(); err != nil {
		return fmt.Errorf("Failed to fetch cluster type: %s", err)
	}
	if _, err = j.GetArtifactsSubdir(); err != nil {
		return fmt.Errorf("Failed to find artifacts subdir: %s", err)
	}
	return nil
}

func (j *Job) GetArtifactsSubdir() (string, error) {
	if j.artifactsSubdir != "" {
		return j.artifactsSubdir, nil
	}

	f, err := j.fetch(j.subPath("artifacts/build-resources/templateinstances.json"))
	if err != nil {
		return "", err
	}
	var tParams templateParams
	if err := json.NewDecoder(f).Decode(&tParams); err != nil {
		return "", err
	}
	for _, params := range tParams.Parameters {
		if params.Name == "JOB_NAME_SAFE" {
			j.artifactsSubdir = params.Value
			return params.Value, nil
		}
	}
	return "", fmt.Errorf("Failed to find cluster type: %s", "no CLUSTER_TYPE param found")
}

func (j *Job) Result() (string, error) {
	f, err := j.fetch(j.subPath("finished.json"))
	if err != nil {
		return "", err
	}

	var finished metadata
	if err := json.NewDecoder(f).Decode(&finished); err != nil {
		return "", err
	}

	return finished.result, nil
}

func (j *Job) BuildLog() (io.Reader, error) {
	return j.fetch(j.subPath("build-log.txt"))
}

func (j *Job) Machines() (io.Reader, error) {
	return j.fetch(j.subPath("artifacts" + j.artifactsSubdir + "/machines.json"))
}

func (j *Job) Nodes() (io.Reader, error) {
	return j.fetch(j.subPath("artifacts" + j.artifactsSubdir + "/openstack_nodes.log"))
}

func (j *Job) JUnitURL() (string, error) {
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

func (j *Job) JUnit() (io.Reader, error) {
	u, err := j.JUnitURL()
	if err != nil {
		return nil, err
	}
	return j.fetch(u)
}
