
### org-scrounger

A highly opinionated CLI to aid me in my day-to-day tasks engineer managing a large github, k8s estate.

[![CircleCI](https://circleci.com/gh/mdevilliers/org-scrounger.svg?style=svg)](https://circleci.com/gh/mdevilliers/org-scrounger)
[![ReportCard](https://goreportcard.com/badge/github.com/mdevilliers/org-scrounger)](https://goreportcard.com/report/github.com/mdevilliers/org-scrounger)


## run

Ensure you have a github token in your env

```
export GITHUB_TOKEN=xxxxxxxxxxx

cd ./cmd/team-reporter/
go build

# get some help
./team-reporter -h

# run some reports
./team-reporter report --output template --label foo --owner some-owner > team-foo.html # outputs html for all repos with tag
./team-reporter report --label foo --owner some-owner  # outputs json
./team-reporter report --output template --repo some-repo --owner some-owner # outputs html for one repo
```

