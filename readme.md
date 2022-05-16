
### org-scrounger

A highly opinionated CLI to aid me in my day-to-day tasks engineer managing a large github, k8s estate.

[![CircleCI](https://circleci.com/gh/mdevilliers/org-scrounger.svg?style=svg)](https://circleci.com/gh/mdevilliers/org-scrounger)
[![ReportCard](https://goreportcard.com/badge/github.com/mdevilliers/org-scrounger)](https://goreportcard.com/report/github.com/mdevilliers/org-scrounger)


## Examples

Get some help.

```
# get some help
./team-reporter -h
```

### Run reports outputting either to JSON or format using a template file.

Ensure you have a github token in your env

```
export GITHUB_TOKEN=xxxxxxxxxxx

# get some help
./team-reporter -h

./team-reporter report --output template --topic foo --owner some-owner > team-foo.html # outputs html for all repos with tag
./team-reporter report --topic foo --owner some-owner  # outputs json
./team-reporter report --output template --repo some-repo --owner some-owner # outputs html for one repo
```

### List all of the docker images used in a kustomize configuration.

```
export GITHUB_TOKEN=xxxxxxxxxxx

./team-reporter images kustomize --root {some-path} --root {some-other-path } # list all images
```

### List all of the repos with some basic information for a team.

```
export GITHUB_TOKEN=xxxxxxxxxxx

./team-reporter list --topic foo --owner some-owner | jq
```

### List all repos without specific topics (in the example one, two or three)

```
export GITHUB_TOKEN=xxxxxxxxxxx

./team-reporter list -owner some-owner --omit-archived | jq -c '.[] |. as $parent | .topics |  select( all( test("one|two|three") == false )) | $parent' | jq -r  '.name' | sort | uniq
```

### List all non-archived repos

```
export GITHUB_TOKEN=xxxxxxxxxxx

./team-reporter list -owner some-owner --omit-archived
```
