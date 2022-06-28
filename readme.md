
### org-scrounger

A highly opinionated CLI to aid me in my day-to-day tasks engineer managing a large github, k8s estate.

[![ci](https://github.com/mdevilliers/org-scrounger/actions/workflows/ci.yaml/badge.svg)](https://github.com/mdevilliers/org-scrounger/actions/workflows/ci.yaml)
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


### List all of the docker images used in a kustomize configuration and map to repositories

```
export GITHUB_TOKEN=xxxxxxxxxxx

./team-reporter images kustomize --root {some-path} --root {some-other-path } --mapping {some-file-path}
```

An example mapping file 

```
# this is a comment

# default github owner
owner = "org-1"

# know container repositories
container_repositories = [
  "foo-container-repo",
  "bar-container-repo"
]

# a container that doesn't map to a repo 
_ > "please/ignore"

# static is a repo we can't discover from the image name 
static > _

# the container 'bar' maps to repo 'foo' at the owner above
foo > "bar"
# the container 'other-org' maps to another to github repo org-2/foo
org-2/foo > "other-org"
# the container 'no', 'yes' and 'maybe' maps to repo 'needle' at the owner above
needle > ["no", "yes", "maybe"]

```

Example output

```
[
 {
    "name": "foo-container-repo/bar",
    "version": "0.3.2",
    "count": 1,
    "repo": {
      "name": "foo",
      "url": "https://github.com/org-1/foo",
      "is_archived": false,
      "topics": [
        "one",
        "two"
      ],
      "languages": {
        "Dockerfile": 1572,
        "Go": 92022,
        "Makefile": 1609,
        "Shell": 1332
      }
    }
  }
]


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
