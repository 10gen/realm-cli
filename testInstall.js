#!/usr/bin/env node

const os = require('os');
const fs = require('fs');
const path = require('path');
const spawnSync = require('child_process').spawnSync;

function directoryExists(file) {
  try {
    const stat = fs.lstatSync(file);
    return stat.isDirectory();
  } catch (err) {
    return false;
  }
}

function fileExists(file) {
  try {
    const stat = fs.lstatSync(file);
    return stat.isFile();
  } catch (err) {
    return false;
  }
}

function removeFolder(dir) {
  if (!fs.existsSync(dir)) return;
  fs.readdirSync(dir).forEach(function(file) {
    const curPath = dir + path.sep + file;
    if (fs.lstatSync(curPath).isDirectory()) {
      removeFolder(curPath);
    } else {
      fs.unlinkSync(curPath);
    }
  });
  fs.rmdirSync(dir);
}

function checkSpawn(spawnInfo) {
  if (spawnInfo.stdout) {
    if (typeof spawnInfo.stdout !== 'string') {
      console.log(spawnInfo.stdout.toString('utf8'));
    } else {
      console.log(spawnInfo.stdout);
    }
  }
  if (spawnInfo.stderr) {
    if (typeof spawnInfo.error !== 'string') {
      console.error(spawnInfo.stderr.toString('utf8'));
    } else {
      console.error(spawnInfo.stderr);
    }
  }
  if (spawnInfo.status !== 0 || spawnInfo.error) {
    console.error('Failed when spawning.');
    process.exit(1);
  }
  if (typeof spawnInfo.stdout !== 'string') {
    return spawnInfo.stdout.toString('utf8');
  }

  return spawnInfo.stdout;
}

function sleep(milliseconds) {
  const inAFewMilliseconds = new Date(new Date().getTime() + milliseconds);
  while (inAFewMilliseconds > new Date()) {
    // wait for completion
  }
}

const tempInstallPath = path.resolve(os.tmpdir(), 'realm-cli-test');
if (directoryExists(tempInstallPath)) {
  console.log(`Deleting directory '${tempInstallPath}'.`);
  removeFolder(tempInstallPath);
}

fs.mkdirSync(tempInstallPath);

if (process.platform === 'win32') {
  sleep(2000); // wait 2 seconds until everything is in place
  checkSpawn(spawnSync('cmd.exe', ['/c', `npm i ${__dirname}`], { cwd: tempInstallPath }));
} else {
  checkSpawn(spawnSync('npm', ['i', `${__dirname}`], { cwd: tempInstallPath }));
}

const executable = path.resolve(
  tempInstallPath,
  'node_modules',
  'mongodb-realm-cli',
  `realm-cli${os.platform() === 'win32' ? '.exe' : ''}`
);
if (fileExists(executable)) {
  console.log(`Realm CLI installed fine.`);
} else {
  console.error(`REALM CLI did not install correctly, file '${executable}' was not found.`);
  process.exit(2);
}

try {
  removeFolder(tempInstallPath);
} catch (err) {
  console.error(`Could not delete folder '${tempInstallPath}'.`);
}
