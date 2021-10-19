#!/usr/bin/env node

const fs = require('fs');
const childProcess = require('child_process');

function run() {
  const [, , BUILD_ID, VERSION, PATH = 'manifest.json'] = process.argv;

  if (!BUILD_ID) {
    throw new Error('error: must specify a valid build id');
  }

  if (!VERSION) {
    throw new Error('error: must specify a valid version');
  }

  let data = '{}'
  try {
    data = fs.readFileSync(PATH)
  } catch (err) { /* create a new manifest.json file */ }

  const { version, info, past_releases: pastReleases = [] } = JSON.parse(data)

  if (version && info) {
    pastReleases.unshift({ version, info })
  }


  const manifest = {
    version: VERSION,
    info: {
      'linux-amd64': {
        url: `https://s3.amazonaws.com/realm-clis/${BUILD_ID}/linux-amd64/realm-cli`,
      },
      'macos-amd64': {
        url: `https://s3.amazonaws.com/realm-clis/${BUILD_ID}/macos-amd64/realm-cli`,
      },
      'windows-amd64': {
        url: `https://s3.amazonaws.com/realm-clis/${BUILD_ID}/windows-amd64/realm-cli.exe`,
      },
    },
    past_releases: pastReleases,
  };

  fs.writeFileSync(PATH, JSON.stringify(manifest, null, '  '))
}

try {
  run()
} catch (err) {
  console.error(err);
} finally {
  process.exit();
}
