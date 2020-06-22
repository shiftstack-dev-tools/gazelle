## What

Update the CI spreadsheet with information from latest CI runs.

## Root cause analysis

The program looks for errored machines in `machines.json` and in `openstack_nodes.log`. If it finds any, it directly prints `Provisioned VM in ERROR state`.

The rules are coded in `pkg/rca/rule.go`. When new rules are coded, and add them to the batch in `pkg/rca/rca.go`.

## How

### Build gazelle

```shell
go build ./cmd/cireport
```

### Setup

Get your `credentials.json` file from https://console.developers.google.com/apis/credentials and save it at the root of your git checkout.

On the first run of gazelle, it will prompt you for granting access. Visit the URL and paste the result in your terminal.

### Usage

```
$ ./cireport --help
Usage of ./cireport:
  -id string
        Job ID. If unset, it consists of all new runs since last time the spreadsheet was updated.
  -job string
        Full name of the test job (e.g. release-openshift-ocp-installer-e2e-openstack-serial-4.4). All known jobs if unset.
  -user string
        Username to use for CI Cop
```

Update the spreadsheet with all latest results for all jobs, with CI Cop Axel Foley:
```shell
./cireport -user "Axel Foley"
```

Update the spreadsheet with latest results for OCP Parallel 4.4:
```shell
./cireport -job release-openshift-ocp-installer-e2e-openstack-4.4
```

Add results for job 345 for OKD Serial 4.2:
```shell
./cireport -job release-openshift-origin-installer-e2e-openstack-serial-4.2 -id 345
```
