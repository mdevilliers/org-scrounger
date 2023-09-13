#!/bin/bash
# if there is a circle-ci yaml file AND .sonarcloud is defined, update version
if test -f .circleci/config.yml ; then 
  yq -i '(.orbs | select(has("sonarcloud")) .sonarcloud ) |= "sonarsource/sonarcloud@2.0.0"' .circleci/config.yml
fi
