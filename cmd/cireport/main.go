package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"sync"

	"github.com/pierreprinetti/go-sequence"
	"github.com/shiftstack/gazelle/pkg/gsheets"
	"github.com/shiftstack/gazelle/pkg/job"
	"github.com/shiftstack/gazelle/pkg/prow"
	"github.com/shiftstack/gazelle/pkg/rca"
)

var (
	fullJobName string
	jobIDs      string
	username    string
)

var valid_jobs = []string{
	"release-openshift-ocp-installer-e2e-openstack-4.4",
	"release-openshift-ocp-installer-e2e-openstack-serial-4.4",
	"release-openshift-ocp-installer-e2e-openstack-4.3",
	"release-openshift-ocp-installer-e2e-openstack-serial-4.3",
	"release-openshift-ocp-installer-e2e-openstack-4.2",
	"release-openshift-ocp-installer-e2e-openstack-serial-4.2",
}

func runJob(jobName, jobIDs string, client *gsheets.Client) {
	sheet := gsheets.Sheet{
		JobName: jobName,
		Client:  client,
	}

	if jobIDs == "" {
		lowerBound := sheet.GetLatestId() + 1
		upperBound := prow.GetLatestId(jobName)
		if lowerBound < upperBound {
			jobIDs = fmt.Sprintf("%d-%d", lowerBound, upperBound)
		} else if lowerBound == upperBound {
			jobIDs = fmt.Sprintf("%d", upperBound)
		} else {
			return
		}
	}
	fmt.Printf("Updating %s with results from jobs %s\n", jobName, jobIDs)

	ids, err := sequence.Int(jobIDs)
	if err != nil {
		panic(err)
	}

	for _, i := range ids {
		fmt.Printf("%s %v\n", jobName, i)
		j := job.Job{
			FullName: jobName,
			ID:       strconv.Itoa(i),
		}

		result, err := j.Result()
		if err == nil {
			j.ComputedResult = result
		} else {
			j.ComputedResult = "Pending"
		}

		var (
			testFailures  []string
			infraFailures []string
		)
		for failure := range rca.Find(j) {
			if failure.IsInfra() {
				infraFailures = append(infraFailures, failure.String())
			}
			testFailures = append(testFailures, failure.String())
		}

		j.RootCause = testFailures
		if len(infraFailures) > 0 {
			j.RootCause = infraFailures
			j.ComputedResult = "INFRA FAILURE"
		}

		sheet.AddRow(j, username)
	}
}

func main() {
	client := gsheets.NewClient()

	var jobs []string
	if fullJobName == "" {
		jobs = valid_jobs
	} else {
		jobs = append(jobs, fullJobName)
	}

	var wg sync.WaitGroup
	for _, jobName := range jobs {
		wg.Add(1)
		go func(jobName string) {
			runJob(jobName, jobIDs, &client)
			wg.Done()
		}(jobName)
	}
	wg.Wait()
}

func init() {
	flag.StringVar(&fullJobName, "job", "", "Full name of the test job (e.g. release-openshift-ocp-installer-e2e-openstack-serial-4.4). All known jobs if unset.")
	flag.StringVar(&jobIDs, "id", "", "Job IDs. If unset, it consists of all new runs since last time the spreadsheet was updated.")

	flag.StringVar(&username, "user", os.Getenv("CIREPORT_USER"), "Username to use for CI Cop")
	if username == "" {
		if u, err := user.Current(); err == nil {
			username = u.Username
		} else {
			username = "cireport"
		}
	}

	flag.Parse()
}
