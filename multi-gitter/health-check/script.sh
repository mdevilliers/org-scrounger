#!/bin/bash

# example usage
# OUTPUT_FILE="${PWD}/output" ./scrng mg --owner foo --topic bar --omit-archived --dry-run --script-path ./multi-gitter/health-check/script.sh --branch na --commit-message na

echo "${REPOSITORY} - [${LANGUAGE_PRIMARY}] ${TOPICS} " >> ${OUTPUT_FILE}

# is there is a dependabot.yml file
if  [ ! -e .github/dependabot.yml ] ; then
  echo -e "\tmissing dependabot file" >> ${OUTPUT_FILE}
fi

# is there is a auto-merge.yml file
if  [ ! -e .github/auto-merge.yml ] ; then
  echo -e "\tmissing auto-merge file" >> ${OUTPUT_FILE}
fi

# is there a CODEOWNERS file
if  [ ! -e .github/CODEOWNERS ] ; then
  echo -e "\tmissing CODEOWNERS file" >> ${OUTPUT_FILE}
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

  # does it publish docs
  if  [ ! -e .github/workflows/gh-pages.yml ] ; then
    echo -e "\tmissing publishing docs workflow" >> ${OUTPUT_FILE}
  fi

  # is the Docker file in the correct place
  if [ ! -e ./docker/production/Dockerfile ] ; then 
    echo -e "\tno production Dockerfile" >> ${OUTPUT_FILE}
  else
    grep "FROM gcr.io/distroless/static-debian11" ./docker/production/Dockerfile || echo -e "\tdockerfile not using distroless base image" >> ${OUTPUT_FILE}
    grep "USER 65532" ./docker/production/Dockerfile || echo -e "\tdockerfile not running as well known user" >> ${OUTPUT_FILE}
  fi
fi

if [ -e .gitmodules ] ; then
    echo -e "\tprobably using build-tools" >> ${OUTPUT_FILE}
fi
