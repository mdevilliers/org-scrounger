#!/bin/bash

# example usage
# OUTPUT_FILE="${PWD}/output" ./scrng mg --owner foo --topic bar --omit-archived --dry-run --script-path ./multi-gitter/health-check/script.sh --branch na --commit-message na

echo "${REPOSITORY}" >> ${OUTPUT_FILE}

# is there is a dependabot.yml file
if  [ ! -e .github/dependabot.yml ] ; then 
  echo -e "\tmissing dependabot file" >> ${OUTPUT_FILE}
fi

# is it a golang project
if [ -e go.mod ] ; then
  # does it have a ci file
  if [ ! -e .circleci/config.yml ] ; then 
    echo -e "\tno circleci config file" >> ${OUTPUT_FILE}
  else
    # is the repo upgraded to using Dagger
    grep "Run Dagger pipeline" .circleci/config.yml || echo -e "\tdagger not implemented" >> ${OUTPUT_FILE}
  fi
fi
