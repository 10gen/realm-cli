const babel = require('babel-standalone');

const babelES2015 = require('babel-preset-es2015');

babel.registerPreset('es2015', babelES2015);

function processData(input) {
  const parsedInput = JSON.parse(input);
  const output = {
    results: [],
    errors: [],
  };
  const opts = {
      presets: ['es2015'],
    plugins: [
      'transform-object-rest-spread',
      'transform-regenerator',
      // ,
      // 'transform-async-to-generator'
    ],
    sourceMap: true,
  };
  for (let i = 0; i < parsedInput.length; i++) {
    const code = parsedInput[i];
    try {
      const result = babel.transform(code, opts);
      if (output.errors.length > 0) {
        continue
      }
      output.results.push({code: result.code, map: result.map});
    } catch (e) {
      let error = {
        index: i,
        // error message includes the snippet of code, which we don't want in the message
        message: (e.message || '').split('\n')[0],        
      };
      if (e.loc) {
        error.line = e.loc.line;
        error.column = e.loc.column;
      }
      output.errors.push(error);
    }
  }

  if (output.errors.length == 0) {
    delete(output.errors);
  } else {
    delete(output.results);
  }

  const outData = JSON.stringify(output, null, 2);
  const cb = () => process.exit(0);

  // For large inputs the entire stdout buffer might not be written after the first call to write(),
  // so explicitly wait for it to drain before exiting the process.
  // See https://nodejs.org/api/stream.html#stream_writable_write_chunk_encoding_callback
  if (!process.stdout.write(outData, cb)) {
    process.stdout.once('drain', cb);
  } else {
    process.nextTick(cb);
  }
}

(function handleStream() {
  let data = '';
  process.stdin.on('readable', () => {
    while (true) {
      const chunk = process.stdin.read();
      if (!chunk) {
        break;
      }
      data += chunk;
    }
  });

  process.stdin.on('end', () => {
    processData(data);
  });

  process.stdin.on('error', () => {
    process.exit(2);
  });
})();
