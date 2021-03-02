#!/usr/bin/env node

const fs = require('fs');
const child_process = require('child_process');
const readline = require('readline').createInterface({
  input: process.stdin,
  output: process.stdout,
});

function prompt(question) {
  return new Promise(resolve => {
    readline.question(`${question}: `, response => {
      readline.close();
      resolve(response.trim());
    });
  });
}

async function run() {
  const [, , VERSION] = process.argv;

  if (!VERSION) {
    throw new Error('error: must specify a valid version');
  }

  const githash = child_process
    .execSync('git rev-parse HEAD')
    .toString()
    .trim();

  const dir = child_process
    .execSync(
      `aws s3 ls realm-clis --recursive | grep ${githash} | sort | tail -n 1 | awk '{print $4}' | cut -f 1-1 -d "/"`
    )
    .toString()
    .trim();

  if (!dir) {
    throw new Error(`error: failed to find release for git hash ${githash}`);
  }

  const currentData = JSON.parse(
    child_process
      .execSync(`aws s3 cp 's3://realm-clis/versions/cloud-prod/CURRENT' -`)
      .toString()
      .trim()
  );

  const updatedData = JSON.stringify(
    {
      version: VERSION,
      info: {
        'linux-amd64': {
          url: `https://s3.amazonaws.com/realm-clis/${dir}/linux-amd64/realm-cli`,
        },
        'macos-amd64': {
          url: `https://s3.amazonaws.com/realm-clis/${dir}/macos-amd64/realm-cli`,
        },
        'windows-amd64': {
          url: `https://s3.amazonaws.com/realm-clis/${dir}/windows-amd64/realm-cli.exe`,
        },
      },
      past_releases: [
        {
          version: currentData.version,
          info: currentData.info,
        },
        ...(currentData.pastReleases || []),
      ],
    },
    null,
    '  '
  );

  console.info(
    "uploading the following JSON file to S3 bucket 'realm-clis/versions/cloud-prod/CURRENT':\n",
    updatedData
  );

  const response = await prompt('proceed? [y/n]');
  if (response !== 'y') {
    console.info('update canceled');
    return;
  }

  const uploadResult = child_process
    .execSync(
      `echo '${updatedData}' | aws s3 cp - 's3://realm-clis/versions/cloud-prod/CURRENT' --content-type 'application/json' --acl 'public-read'`
    )
    .toString()
    .trim();
  if (uploadResult !== '') {
    throw new Error(`error: failed to upload manifest data to S3: ${uploadResult}`);
  }
  console.info('successfully uploaded to S3');
  console.info('updating version.json...');

  fs.writeFileSync(
    'version.json',
    JSON.stringify(
      {
        version: VERSION,
        baseDirectory: dir,
      },
      null,
      '  '
    )
  );

  console.info('updating npm version and creating tag...');
  child_process.execSync(`npm version --no-git-tag-version ${VERSION}`);
  child_process.execSync(`git add ./version.json ./package*`);
  child_process.execSync(`git commit -m "${VERSION}"`);
  child_process.execSync(`git tag -m "${VERSION}" -a "v${VERSION}"`);

  console.info('Success!');
}

run().catch(err => {
  console.error(err);
  process.exit();
});
