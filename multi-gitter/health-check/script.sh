#!/bin/bash

# example usage
# OUTPUT_FILE="${PWD}/output" ./scrng mg --owner foo --topic bar --omit-archived --dry-run --script-path ./multi-gitter/health-check/script.sh --branch na --commit-message na

echo "${REPOSITORY}" >> ${OUTPUT_FILE}

# is there is a dependabot.yml file
if  [ ! -e .github/dependabot.yml ] ; then 
  echo -e "\tmissing dependabot file" >> ${OUTPUT_FILE}
fi

# is it a golang project and does the build upgraded to dagger
if [ -e go.mod ] ; then
  if [ ! -e .circleci/config.yml ] ; then 
    echo -e "no circleci config file" >> ${OUTPUT_FILE}
  else
    grep "Run Dagger pipeline" .circleci/config.yml || echo -e "\tdagger not implemented" >> ${OUTPUT_FILE}
  fi
fi
