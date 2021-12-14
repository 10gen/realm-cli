const fs = require('fs');
const path = require('path');
const request = require('request');

const packageMetadata = require('./package');

const manifestURL =
  'https://s3.amazonaws.com/realm-clis/versions/cloud-prod/CURRENT';

function fetchManifest() {
  return new Promise((resolve, reject) => {
    request.get(manifestURL, (err, _, body) => {
      if (err) {
        reject(err);
        return;
      }
      resolve(JSON.parse(body));
    });
  });
}

function getDownloadURL({ past_releases: pastReleases = [], ...manifest }) {
  const { arch, platform } = process;

  let tag = '';

  if (platform === 'linux') {
    if (arch !== 'x64') {
      throw new Error('Only Linux 64 bits supported.');
    }
    tag = 'linux-amd64';
  } else if (platform === 'darwin' || platform === 'freebsd') {
    if (arch !== 'x64' && arch !== 'arm64') {
      throw new Error('Only Mac 64 bits supported.');
    }
    tag = 'macos-amd64';
  } else if (platform === 'win32') {
    if (arch !== 'x64') {
      throw new Error('Only Windows 64 bits supported.');
    }
    tag = 'windows-amd64';
  }

  if (tag === '') {
    throw new Error(`Unexpected platform or architecture: ${platform} ${arch}`);
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

function requstBinary(downloadURL, baseName = 'realm-cli') {
  console.log(`downloading "${baseName}" from "${downloadURL}"`);

  return new Promise((resolve, reject) => {
    let count = 0;
    let notifiedCount = 0;

    const binaryName =
      process.platform === 'win32' ? baseName + '.exe' : baseName;
    const filePath = path.join(process.cwd(), binaryName);
    const outFile = fs.openSync(filePath, 'w');

    const requestOptions = {
      uri: downloadURL,
      method: 'GET',
    };
    const client = request(requestOptions);

    client.on('error', (err) => {
      reject(new Error(`Error with http(s) request: ${err}`));
    });

    client.on('data', (data) => {
      fs.writeSync(outFile, data, 0, data.length, null);
      count += data.length;
      if (count - notifiedCount > 800000) {
        process.stdout.write(`Received ${Math.floor(count / 1024)} K...\r`);
        notifiedCount = count;
      }
    });

    client.on('end', () => {
      console.log(`Received ${Math.floor(count / 1024)} K total.`);
      fs.closeSync(outFile);
      fixFilePermissions(filePath);
      resolve(true);
    });
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
  .catch((err) => {
    console.error('failed to download Realm CLI:', err);
    process.exit(1);
  });
