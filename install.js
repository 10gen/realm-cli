"use strict";

const fs = require('fs');
const path = require('path');
const axios = require('axios');

const packageMetadata = require('./package');

const manifestURL =
  'https://s3.amazonaws.com/realm-clis/versions/cloud-prod/CURRENT';

function fetchManifest() {
  return axios(manifestURL)
    .then(function (res) { return res.data; });
}

function getDownloadURL(manifest) {
  const pastReleases = manifest.past_releases || [];

  let tag = '';

  if (process.platform === 'linux') {
    if (process.arch !== 'x64') {
      throw new Error('Only Linux 64 bits supported.');
    }
    tag = 'linux-amd64';
  } else if (process.platform === 'darwin' || process.platform === 'freebsd') {
    if (process.arch !== 'x64' && process.arch !== 'arm64') {
      throw new Error('Only Mac 64 bits supported.');
    }
    tag = 'macos-amd64';
  } else if (process.platform === 'win32') {
    if (process.arch !== 'x64') {
      throw new Error('Only Windows 64 bits supported.');
    }
    tag = 'windows-amd64';
  }

  if (tag === '') {
    throw new Error(`Unexpected platform or architecture: ${process.platform} ${process.arch}`);
  }

  if (manifest.version !== packageMetadata.version) {
    for (let i = 0; i < pastReleases.length; i++) {
      const pastRelease = pastReleases[i];
      if (pastRelease.version === packageMetadata.version) {
        return pastRelease.info[tag].url;
      }
    }
  }

  return manifest.info[tag].url;
}

function requstBinary(downloadURL, baseName) {
  baseName = baseName || 'realm-cli';

  console.log(`downloading "${baseName}" from "${downloadURL}"`);

  return new Promise(function (resolve, reject) {
    let count = 0;
    let notifiedCount = 0;

    const binaryName =
      process.platform === 'win32' ? baseName + '.exe' : baseName;
    const filePath = path.join(process.cwd(), binaryName);
    const outFile = fs.openSync(filePath, 'w');

    axios.get(downloadURL, { responseType: 'stream' })
      .then(function (res) { return res.data; })
      .then(function (stream) {
        stream.on('error', function (err) {
          reject(new Error(`Error with http(s) request: ${err}`));
        });

        stream.on('data', function (data) {
          fs.writeSync(outFile, data, 0, data.length, null);
          count += data.length;
          if (count - notifiedCount > 800000) {
            process.stdout.write(`Received ${Math.floor(count / 1024)} K...\r`);
            notifiedCount = count;
          }
        });

        stream.on('end', function () {
          console.log(`Received ${Math.floor(count / 1024)} K total.`);
          fs.closeSync(outFile);
          fixFilePermissions(filePath);
          resolve(true);
        });
      })
      .catch(reject);
  });
}

function fixFilePermissions(filePath) {
  // Check that the binary is user-executable and fix it if it isn't
  if (process.platform !== 'win32') {
    const stat = fs.statSync(filePath);
    // 64 == 0100 (no octal literal in strict mode)
    // eslint-disable-next-line no-bitwise
    if (!(stat.mode & 64)) {
      fs.chmodSync(filePath, '755');
    }
  }
}

fetchManifest()
  .then(getDownloadURL)
  .then(requstBinary)
  .catch(function (err) {
    console.error('failed to download Realm CLI:', err);
    process.exit(1);
  });
