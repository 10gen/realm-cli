const babel = require("@babel/standalone");
const transformObjectRestSpread = require("@babel/plugin-proposal-object-rest-spread");

babel.registerPlugin("transform-object-rest-spread", transformObjectRestSpread);

function processData(input) {
  const parsedInput = JSON.parse(input);
  const output = {
    results: [],
    errors: [],
    warnings: [],
  };
  const opts = {
    presets: [
      ["env", { exclude: ["babel-plugin-transform-async-to-generator"] }],
    ],
    plugins: ["transform-object-rest-spread"],
    parserOpts: {
      allowReturnOutsideFunction: true,
    },
    sourceMap: true,
  };
  for (let i = 0; i < parsedInput.length; i++) {
    const code = parsedInput[i];
    try {
      const result = babel.transform(code, opts);
      if (output.errors.length > 0) {
        continue;
      }
      output.results.push({
        // This has been introduced to replace the use of `eval` in @protobufjs/inquire
        // which is used by any service module from the GCP family.
        // TODO(REALMC-5102): Remove this temporary workaround to have the inquire
        // module work (no more .replace should be needed with `eval` support). Alternatively
        // if we decide to stay away from `eval` then remove this TODO.
        code: result.code.replace(
          `eval("quire".replace(/^/, "re"))`,
          `require`
        ),
        map: result.map,
      });
    } catch (e) {
      const error = {
        index: i,
        // error message includes the snippet of code, which we don't want in the message
        message: (e.message || "").split("\n")[0],
      };
      if (e.loc) {
        error.line = e.loc.line;
        error.column = e.loc.column;
      }
      /*
        this was introduced trying to transpile this file:
        https://github.com/protobufjs/protobuf.js/blob/master/cli/wrappers/es6.js
        the file contains unrecoverable errors for the parser:
              import * as $protobuf from $DEPENDENCY;
              $OUTPUT;
              export { $root as default };
        this error was added as a warning to still have visibility without causing termination
        ---
        ETA(REALMC-6584): the original check was failing TestES6FunctionsLoggedInWithApp/Creating_functions_with_invalid_js_syntax_in_the_source_code_should_fail/(0)_name_"badSyntax3"
        Changing the check to "$DEPENDENCY" satisfies the source of this original issue (protobuf),
        but still allows for other arbitrary errors containing "$" in the output to be recognized as errors
      */
      if (e.message.includes("$DEPENDENCY")) {
        output.warnings.push(error);
        continue;
      }

      output.errors.push(error);
    }
  }

  if (output.errors.length == 0) {
    delete output.errors;
  } else {
    delete output.results;
  }
  if (output.warnings.length == 0) {
    delete output.warnings;
  }

  const outData = JSON.stringify(output, null, 2);
  const cb = () => process.exit(0);

  // For large inputs the entire stdout buffer might not be written after the first call to write(),
  // so explicitly wait for it to drain before exiting the process.
  // See https://nodejs.org/api/stream.html#stream_writable_write_chunk_encoding_callback
  if (!process.stdout.write(outData, cb)) {
    process.stdout.once("drain", cb);
  } else {
    process.nextTick(cb);
  }
}

(function handleStream() {
  let data = "";
  process.stdin.on("readable", () => {
    while (true) {
      const chunk = process.stdin.read();
      if (!chunk) {
        break;
      }
      data += chunk;
    }
  });

  process.stdin.on("end", () => {
    processData(data);
  });

  process.stdin.on("error", () => {
    process.exit(2);
  });
})();
