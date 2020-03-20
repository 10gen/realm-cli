package transpiler

import (
	"bytes"
	"context"
	"testing"

	"github.com/robertkrimen/otto"
)

func TestTranspiler(t *testing.T) {
	type testCase struct {
		description         string
		codes               []string
		expectedTranspileds []string
		expectedError       error
		expectedStacktraces []string
	}
	for _, tc := range []testCase{
		{
			"plain JS source",
			[]string{`() => {console.log('hi')}`},
			[]string{"'use strict';\n\n(function () {\n  console.log('hi');\n});"},
			nil,
			[]string{""},
		},
		{
			"JS source that uses classes, with sourcemap",
			[]string{"class Duck{\nquack(){\nthrow new Error();return 'quack'\n}\n}\nvar x = new Duck();\n var y = x.quack()"},
			[]string{"'use strict';\n\nvar _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if (\"value\" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();\n\nfunction _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError(\"Cannot call a class as a function\"); } }\n\nvar Duck = function () {\n  function Duck() {\n    _classCallCheck(this, Duck);\n  }\n\n  _createClass(Duck, [{\n    key: 'quack',\n    value: function quack() {\n      throw new Error();return 'quack';\n    }\n  }]);\n\n  return Duck;\n}();\n\nvar x = new Duck();\nvar y = x.quack();"},
			nil,
			[]string{"Error\n    at quack (unknown:3:10)\n    at unknown:7:9\n"},
		},
		{
			"plain JS that throws an error, with sourcemap",
			[]string{"(()=> { (()=> {var foo = 'aaaaaaaaaa';throw new Error('boo!', this)})()})()"},
			[]string{"'use strict';\n\n(function () {\n  (function () {\n    var foo = 'aaaaaaaaaa';throw new Error('boo!', undefined);\n  })();\n})();"},
			nil,
			[]string{""},
		},
		{
			"Destructuring",
			[]string{`() => { var foo = {a:"a!", b:"b!", c: "c!"}; var {a,b,c} = foo;}`},
			[]string{"\"use strict\";\n\n(function () {\n  var foo = { a: \"a!\", b: \"b!\", c: \"c!\" };var a = foo.a,\n      b = foo.b,\n      c = foo.c;\n});"},
			nil,
			[]string{""},
		},
		{
			"Spread",
			[]string{`() => { var foo = {a:"a!", b:"b!", c: {x:1,y:2,z:3}}; var {a,b,...otherObj} = foo;}`},
			[]string{"\"use strict\";\n\nfunction _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }\n\n(function () {\n  var foo = { a: \"a!\", b: \"b!\", c: { x: 1, y: 2, z: 3 } };\n  var a = foo.a,\n      b = foo.b,\n      otherObj = _objectWithoutProperties(foo, [\"a\", \"b\"]);\n});"},
			nil,
			[]string{""},
		},
		{
			"invalid source code",
			[]string{")"},
			[]string{""},
			TranspileErrors{&TranspileError{
				Message: "unknown: Unexpected token (1:0)",
				Line:    1,
				Column:  0,
			}},
			[]string{""},
		},
		{
			"good bad good bad good",
			[]string{`() => {console.log('hi')}`, ")", `() => {console.log('hello')}`, "<", `() => {console.log('bye')}`},
			[]string{""},
			TranspileErrors{
				&TranspileError{
					Index:   1,
					Message: "unknown: Unexpected token (1:0)",
					Line:    1,
					Column:  0,
				},
				&TranspileError{
					Index:   3,
					Message: "unknown: Unexpected token (1:0)",
					Line:    1,
					Column:  0,
				},
			},
			[]string{""},
		},
		{
			"good good",
			[]string{`() => {console.log('hi')}`, "class Duck{\nquack(){\nthrow new Error();return 'quack'\n}\n}\nvar x = new Duck();\n var y = x.quack()"},
			[]string{"'use strict';\n\n(function () {\n  console.log('hi');\n});", "'use strict';\n\nvar _createClass = function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if (\"value\" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; }();\n\nfunction _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError(\"Cannot call a class as a function\"); } }\n\nvar Duck = function () {\n  function Duck() {\n    _classCallCheck(this, Duck);\n  }\n\n  _createClass(Duck, [{\n    key: 'quack',\n    value: function quack() {\n      throw new Error();return 'quack';\n    }\n  }]);\n\n  return Duck;\n}();\n\nvar x = new Duck();\nvar y = x.quack();"},
			nil,
			[]string{"", "Error\n    at quack (unknown:3:10)\n    at unknown:7:9\n"},
		},
	} {
		t.Run(tc.description, func(t *testing.T) {
			transpiler := externalTranspiler{DefaultTranspilerCommand}
			results, err := transpiler.Transpile(context.Background(), tc.codes...)
			if err != nil {
				if tc.expectedError == nil {
					t.Fatal(err)
				}
				if err.Error() != tc.expectedError.Error() {
					t.Fatalf("Error: %s does not match expected: %s", err.Error(), tc.expectedError.Error())
				}
				return
			}

			// error is nil but we are expecting one
			if tc.expectedError != nil {
				t.Fatal("expected error ", tc.expectedError)
			}

			if len(results) != len(tc.expectedTranspileds) {
				t.Fatalf("results length %d did not match expected length %d", len(results), len(tc.expectedTranspileds))
			}
			// sort.Sort(results)
			for i, expected := range tc.expectedTranspileds {
				if results[i].Code != expected {
					t.Fatalf("Results does not match expected \nexpected: %s\nactual:%s", expected, results[i].Code)
				}
			}

			for i, expected := range tc.expectedStacktraces {
				if expected == "" {
					continue
				}
				vm := otto.New()

				script, err := vm.CompileWithSourceMap("foo", tc.expectedTranspileds[i], bytes.NewBuffer(results[i].SourceMap))
				if err != nil {
					t.Fatal(err)
				}

				_, err = vm.Run(script)

				ottoErr, ok := err.(*otto.Error)
				if !ok {
					t.Fatal("expected to get an error with a stack trace")

				}
				if ottoErr.String() != expected {
					t.Fatalf("expected %s to equal %s", ottoErr.String(), expected)
				}
			}
		})
	}

}
