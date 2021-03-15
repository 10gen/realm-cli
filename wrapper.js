#!/usr/bin/env node

const path = require('path');
const { spawn } = require('child_process');

function onExit(childProcess) {
  return new Promise((resolve, reject) => {
    childProcess.on('exit', code => {
      if (code === 0) {
        resolve();
      } else {
        reject(code);
      }
    });
  });
}

const binaryPath =
  process.platform === 'win32' ? path.join(__dirname, 'realm-cli.exe') : path.join(__dirname, 'realm-cli');
const args = process.argv.slice(2);

const childProcess = spawn(binaryPath, args, { stdio: [process.stdin, process.stdout, process.stderr] });
onExit(childProcess).catch(code => process.exit(code));
