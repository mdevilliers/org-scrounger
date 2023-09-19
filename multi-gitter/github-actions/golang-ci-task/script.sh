#!/bin/bash
# if there is a golangci-lint file 
if test -f .github/workflows/golangci-lint.yml ; then 
  yq -i '(.jobs | .golangci.steps[] | select(.name == "golangci-lint") | .with.version ) |= "v1.53"' .github/workflows/golangci-lint.yml
  yq -i '(.jobs | .golangci.steps[] | select(.name == "golangci-lint") | .with.only-new-issues ) |= false' .github/workflows/golangci-lint.yml 
  yq -i '(.jobs | .golangci.steps[] | select(.name == "golangci-lint") | .with.args ) |= "--timeout=4m"' .github/workflows/golangci-lint.yml
fi
