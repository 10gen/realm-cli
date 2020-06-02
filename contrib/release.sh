#!/bin/sh

set -e

VERSION=$1
GITHASH=$2

if [ "$GITHASH" == "" ]; then
  GITHASH=`git rev-parse HEAD`
fi


DIR=`aws s3 ls realm-clis --recursive | grep $GITHASH | sort | tail -n 1 | awk '{print $4}' | cut -f 1-1 -d "/"`

if [ "$DIR" == "" ]; then
  echo "error: failed to find release for git hash $GITHASH"
  exit 1
fi

if [ "$VERSION" == "" ]; then
  echo "error: must specify a valid version"
  exit 1
fi

CURRENT=$(cat <<EOF
{
  "version": "$VERSION",
  "info": {
    "linux-amd64": {
      "url": "https://s3.amazonaws.com/realm-clis/$DIR/linux-amd64/realm-cli"
    },
    "macos-amd64": {
      "url": "https://s3.amazonaws.com/realm-clis/$DIR/macos-amd64/realm-cli"
    },
    "windows-amd64": {
      "url": "https://s3.amazonaws.com/realm-clis/$DIR/windows-amd64/realm-cli.exe"
    }
  }
}
EOF
)

echo "uploading the following JSON file to S3 bucket 'realm-clis/versions/cloud-prod/CURRENT':\n"
echo "$CURRENT\n"

echo "proceed? [y/n]:"
read CONFIRM

if [ "$CONFIRM" != "y" ]; then
  echo "update canceled\n"
  exit 0
fi

echo "$CURRENT" | aws s3 cp - 's3://realm-clis/versions/cloud-prod/CURRENT' --content-type 'application/json' --acl 'public-read'
echo "successfully uploaded to S3"

echo "updating 'version.json'..."
cat <<EOF > version.json
{
  "version": "$VERSION",
  "baseDirectory": "$DIR"
}
EOF

echo "updating npm version and creating tag..."
npm version --no-git-tag-version $VERSION
git add ./version.json ./package*
git commit -m "$VERSION"
git tag -m "$VERSION" -a "v$VERSION"

echo "Success!\n"
