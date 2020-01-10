package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pierreprinetti/go-sequence"
	"github.com/shiftstack/gazelle/pkg/job"
	"github.com/shiftstack/gazelle/pkg/rca"
)

var (
	jobName string
	target  string
	jobIDs  string
	output  string
)

type Report struct {
	startedAt  time.Time
	finishedAt time.Time
	result     string
	rootCause  []string
}

func main() {
	ids, err := sequence.Int(jobIDs)
	if err != nil {
		panic(err)
	}

	for _, i := range ids {
		j := job.Job{
			Name:   jobName,
			Target: target,
			ID:     strconv.Itoa(i),
		}

		startedAt, err := j.StartTime()
		if err != nil {
			panic(err)
		}

		finishedAt, err := j.FinishTime()
		if err != nil {
			panic(err)
		}

		result, err := j.Result()
		if err != nil {
			panic(err)
		}

		failures, errs := rca.Find(j)

		var wg sync.WaitGroup
		wg.Add(1)
		// panic at the first error
		go func() {
			for err := range errs {
				panic(err)
			}
			wg.Done()
		}()

		var (
			testFailures  []string
			infraFailures []string
		)
		for failure := range failures {
			if failure.IsInfra() {
				infraFailures = append(infraFailures, failure.String())
			}
			testFailures = append(testFailures, failure.String())
		}

		// Wait for the error handling to occur
		wg.Wait()

		rootCause := testFailures
		if len(infraFailures) > 0 {
			rootCause = infraFailures
			result = "INFRA FAILURE"
		}

		report := Report{
			startedAt:  startedAt,
			finishedAt: finishedAt,
			result:     result,
			rootCause:  rootCause,
		}

		if output == "html" {
			printHTML(j, report)
		} else {
			if report.result == "SUCCESS" {
				// Show only failures
				continue
			}
			printText(j, report)
		}

	}
}

func printHTML(j job.Job, report Report) {
	var s strings.Builder
	{
		s.WriteString(`<meta http-equiv="content-type" content="text/html; charset=utf-8"><meta name="generator" content="cireport"/><table xmlns="http://www.w3.org/1999/xhtml"><tbody><tr><td>`)
		s.WriteString(strings.Join([]string{
			`<a href="` + j.JobURL() + `">` + j.ID + `</a>`,
			report.startedAt.String(),
			report.finishedAt.Sub(report.startedAt).String(),
			report.result,
			"",
			`<a href="` + j.BuildLogURL() + `">` + j.BuildLogURL() + `</a>`,
			`<a href="` + j.MachinesURL() + `">` + j.MachinesURL() + `</a>`,
			`<a href="` + j.NodesURL() + `">` + j.NodesURL() + `</a>`,
			"cireport",
			strings.Join(report.rootCause, "<br />"),
		}, "</td><td>"))
		s.WriteString(`</td></tr></tbody></table>`)
	}
	fmt.Println(s.String())
}

func printText(j job.Job, report Report) {
	var s strings.Builder
	{
		s.WriteString(strings.Join([]string{
			`* ` + j.JobURL(),
			"\t" + report.startedAt.String() + ` ` + report.finishedAt.Sub(report.startedAt).String(),
			"\t" + report.result,
			"\t" + j.BuildLogURL(),
			"\t" + strings.Join(report.rootCause, "\n\t"),
			"----",
		}, "\n"))
	}
	fmt.Println(s.String())
}

func init() {
	flag.StringVar(&jobName, "job", "", "Name of the test job")
	flag.StringVar(&target, "target", "", "Target OpenShift version")
	flag.StringVar(&jobIDs, "id", "", "Job IDs")
	flag.StringVar(&output, "output", "html", "Output type")

	flag.Parse()
}
