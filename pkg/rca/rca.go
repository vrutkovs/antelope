package rca

import (
	"io"
	"sync"
)

var (
	rules = []Rule{
		infraFailureIfMatchBuildLogs(
			"level=fatal msg=\"Bootstrap failed to complete",
			CauseBootstrapTimeout,
		),

		infraFailureIfMatchBuildLogs(
			"Throttling: Rate exceeded",
			CauseBootstrapTimeout,
		),

		failedTests,
	}
)

type Rule func(j job, failures chan<- Cause) error

type job interface {
	Result() (string, error)
	BuildLog() (io.Reader, error)
	Machines() (io.Reader, error)
	Nodes() (io.Reader, error)
	JUnit() (io.Reader, error)
}

func Find(j job) (<-chan Cause, <-chan error) {

	failures := make(chan Cause)
	errs := make(chan error, len(rules))

	res, _ := j.Result()
	if res == "SUCCESS" {
		close(failures)
		close(errs)
		return failures, errs
	}

	var wg sync.WaitGroup
	for _, rule := range rules {
		wg.Add(1)
		go func(rule Rule) {
			if err := rule(j, failures); err != nil {
				errs <- err
			}
			wg.Done()
		}(rule)
	}

	go func() {
		wg.Wait()
		close(failures)
		close(errs)
	}()

	return uniqueFilter(failures), errs
}

func uniqueFilter(inCh <-chan Cause) <-chan Cause {
	var (
		outCh = make(chan Cause)
		cache = make(map[Cause]struct{})
	)

	go func() {
		for cause := range inCh {
			if _, ok := cache[cause]; ok {
				continue
			}
			outCh <- cause
			cache[cause] = struct{}{}
		}
		close(outCh)
	}()

	return outCh
}
