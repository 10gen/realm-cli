#!/bin/sh

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $DIR
cd ..

NEXT_VERSION=`cat .evg.yml | grep cli_version: | awk '{ print $2 }' | tail -n 1`

LAST_VERSION=`node -e 'console.log(require("./package.json").version)'`

echo "Bumping CLI v$LAST_VERSION to v$NEXT_VERSION"

npm version $NEXT_VERSION --no-git-tag-version

JIRA_TICKET=`git branch --show-current`

git add .evg.yml package* && git commit -m "$JIRA_TICKET: Bump version to $NEXT_VERSION"

git log -p -1
