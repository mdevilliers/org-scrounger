#!/bin/bash
# if there is a dependabot.yml file AND .package-ecosystem == "gomod change the interval to weekly
if test -f .github/dependabot.yml ; then 
  yq -i '(.updates.[] | select(.package-ecosystem == "gomod") | .schedule.interval) |= "weekly"' .github/dependabot.yml
fi
