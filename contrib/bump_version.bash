#!/bin/sh

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $DIR
cd ..

BUMP_TYPE=$1
if [ "$BUMP_TYPE" != "patch" ] && [ "$BUMP_TYPE" != "minor" ] && [ "$BUMP_TYPE" != "major" ]; then
	echo $"Usage: $0 <patch|minor|major>"
	exit 1
fi

LAST_VERSION=`node -e 'console.log(require("./package.json").version)'`

echo "Bumping CLI v$LAST_VERSION to"

npm version $BUMP_TYPE --no-git-tag-version

NEXT_VERSION=`node -e 'console.log(require("./package.json").version)'`

JIRA_TICKET=`git branch --show-current`

git add .evg.yml package* && git commit -m "$JIRA_TICKET: Bump version to $NEXT_VERSION"

git log -p -1
