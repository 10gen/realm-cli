const fs = require('fs');
const path = require('path');
const request = require('request');

const versionMetadata = require('./version');

const basedownloadURL = `https://s3.amazonaws.com/realm-clis/${versionMetadata.baseDirectory}`;
const linuxdownloadURL = `${basedownloadURL}/linux-amd64/realm-cli`;
const macdownloadURL = `${basedownloadURL}/macos-amd64/realm-cli`;
const windowsdownloadURL = `${basedownloadURL}/windows-amd64/realm-cli.exe`;
// transpiler downloads
const linuxTranspilerdownloadURL = `${basedownloadURL}/linux-amd64/transpiler`;
const macTranspilerdownloadURL = `${basedownloadURL}/macos-amd64/transpiler`;
const windowsTranspilerdownloadURL = `${basedownloadURL}/windows-amd64/transpiler.exe`;

function getdownloadURL(cli) {
  const platform = process.platform;
  let downloadURL;

  if (platform === 'linux') {
    if (process.arch === 'x64') {
      downloadURL = cli ? linuxdownloadURL : linuxTranspilerdownloadURL;
    } else {
      throw new Error('Only Linux 64 bits supported.');
    }
  } else if (platform === 'darwin' || platform === 'freebsd') {
    if (process.arch === 'x64') {
      downloadURL = cli ? macdownloadURL : macTranspilerdownloadURL;
    } else {
      throw new Error('Only Mac 64 bits supported.');
    }
  } else if (platform === 'win32') {
    if (process.arch === 'x64') {
      downloadURL = cli ? windowsdownloadURL : windowsTranspilerdownloadURL;
    } else {
      throw new Error('Only Windows 64 bits supported.');
    }
  } else {
    throw new Error(`Unexpected platform or architecture: ${process.platform} ${process.arch}`);
  }

  return downloadURL;
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

function requstBinary(downloadURL, baseName) {
  console.log(`downloading "${baseName}" from "${downloadURL}"`);

  return new Promise((resolve, reject) => {
    let count = 0;
    let notifiedCount = 0;

    const binaryName = process.platform === 'win32' ? baseName + '.exe' : baseName;
    const filePath = path.join(process.cwd(), binaryName);
    const outFile = fs.openSync(filePath, 'w');

    const requestOptions = {
      uri: downloadURL,
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
let downloadURL;
try {
  downloadURL = getdownloadURL(true);
} catch (err) {
  console.error('Realm CLI installation failed:', err);
  process.exit(1);
}

requstBinary(downloadURL, 'realm-cli').catch(err => {
  console.error('failed to download Realm CLI:', err);
  process.exit(1);
});

let transpilerdownloadURL;
try {
  transpilerdownloadURL = getdownloadURL(false);
} catch (err) {
  console.error('Realm CLI installation failed:', err);
  process.exit(1);
}

requstBinary(transpilerdownloadURL, 'transpiler').catch(err => {
  console.error('failed to download Realm CLI:', err);
  process.exit(1);
});
