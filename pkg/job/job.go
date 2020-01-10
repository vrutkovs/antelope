package job

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shiftstack/gazelle/pkg/cache"
)

type Job struct {
	Name   string
	Target string
	ID     string

	client http.Client

	cache *cache.Cache
}

func (j Job) baseURL() string {
	prefix := "release-openshift-ocp-installer"
	if j.Target == "4.1" {
		// Pre-4.2 builds have "-origin-" prefix
		prefix = "release-openshift-origin-installer"
	}
	return "https://storage.googleapis.com/origin-ci-test/logs/" + prefix + "-" + j.Name + "-" + j.Target + "/" + j.ID
}

func (j *Job) fetch(file string) (io.Reader, error) {
	if j.cache == nil {
		j.cache = new(cache.Cache)
	}
	return j.cache.Get(file)
}

func (j Job) StartTime() (time.Time, error) {
	f, err := j.fetch(j.baseURL() + "/started.json")
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
	f, err := j.fetch(j.baseURL() + "/finished.json")
	if err != nil {
		return time.Time{}, err
	}

	var finished metadata
	if err := json.NewDecoder(f).Decode(&finished); err != nil {
		return time.Time{}, err
	}

	return finished.time, nil
}

func (j Job) Result() (string, error) {
	f, err := j.fetch(j.baseURL() + "/finished.json")
	if err != nil {
		return "", err
	}

	var finished metadata
	if err := json.NewDecoder(f).Decode(&finished); err != nil {
		return "", err
	}

	return finished.result, nil
}

func (j Job) BuildLogURL() string {
	return j.baseURL() + "/build-log.txt"
}

func (j Job) BuildLog() (io.Reader, error) {
	return j.fetch(j.BuildLogURL())
}

func (j Job) MachinesURL() string {
	return j.baseURL() + "/artifacts/" + j.Name + "/machines.json"
}

func (j Job) Machines() (io.Reader, error) {
	return j.fetch(j.MachinesURL())
}

func (j Job) NodesURL() string {
	return j.baseURL() + "/artifacts/" + j.Name + "/openstack_nodes.log"
}

func (j Job) Nodes() (io.Reader, error) {
	return j.fetch(j.NodesURL())
}

func (j Job) JobURL() string {
	return strings.Replace(j.baseURL(), "storage.googleapis.com", "prow.svc.ci.openshift.org/view/gcs", 1)
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
			return j.baseURL() + "/artifacts/" + j.Name + "/junit/" + filename, scanner.Err()
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
