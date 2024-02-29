#!/bin/bash
# if there is a dependabot.yml file 
if test -f .github/dependabot.yml ; then 
  # if package-ecosystem == Go, set schedule to weekly
  yq -i '(.updates.[] | select(.package-ecosystem == "gomod") | .schedule.interval) |= "weekly"' .github/dependabot.yml
  # if package-ecosystem == Go, add grouping for otel
  yq -i '(.updates.[] | select(.package-ecosystem == "gomod")) += { "groups": { "production-dependencies": { "dependency-type": "production"}}}' .github/dependabot.yml
fi
