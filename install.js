const fs = require('fs');
const path = require('path');
const request = require('request');

const baseDownloadURL =
  'https://s3.amazonaws.com/stitch-clis/stitch_cli_linux_64_f21637afd3f02fde319df8bf05cf173e5892f88f_18_12_06_00_55_54';
const linuxDownloadURL = `${baseDownloadURL}/linux-amd64/stitch-cli`;
const macDownloadURL = `${baseDownloadURL}/macos-amd64/stitch-cli`;
const windowsDownloadURL = `${baseDownloadURL}/windows-amd64/stitch-cli.exe`;

function getDownloadURL() {
  const platform = process.platform;
  let downloadUrl;

  if (platform === 'linux') {
    if (process.arch === 'x64') {
      downloadUrl = linuxDownloadURL;
    } else {
      throw new Error('Only Linux 64 bits supported.');
    }
  } else if (platform === 'darwin' || platform === 'freebsd') {
    if (process.arch === 'x64') {
      downloadUrl = macDownloadURL;
    } else {
      throw new Error('Only Mac 64 bits supported.');
    }
  } else if (platform === 'win32') {
    if (process.arch === 'x64') {
      downloadUrl = windowsDownloadURL;
    } else {
      throw new Error('Only Windows 64 bits supported.');
    }
  } else {
    throw new Error(`Unexpected platform or architecture: ${process.platform} ${process.arch}`);
  }

  return downloadUrl;
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

function requstBinary(downloadUrl) {
  console.log(`downloading stitch-cli from "${downloadUrl}"`);

  return new Promise((resolve, reject) => {
    let count = 0;
    let notifiedCount = 0;

    const binaryName = process.platform === 'win32' ? 'stitch-cli.exe' : 'stitch-cli';
    const filePath = path.join(process.cwd(), binaryName);
    const outFile = fs.openSync(filePath, 'w');

    const requestOptions = {
      uri: downloadUrl,
      method: 'GET',
    };
    const client = request(requestOptions);

    client.on('error', err => {
      reject(new Error(`Error with http(s) request: ${err}`));
    });

    client.on('data', data => {
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

// Install script starts here
let downloadUrl;
try {
  downloadUrl = getDownloadURL();
} catch (err) {
  console.error('Stitch CLI installation failed:', err);
  process.exit(1);
}

requstBinary(downloadUrl).catch(err => {
  console.error('Stitch CLI installation failed while downloading:', err);
  process.exit(1);
});
