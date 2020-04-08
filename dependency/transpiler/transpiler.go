package transpiler

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
)

// DefaultTranspilerCommand is the binary used for executing the transpiler
const DefaultTranspilerCommand = "transpiler"

// TranspileError contains the error message from a failed transpile attempt, and
// the line/column if available
type TranspileError struct {
	Index   int    `json:"index"` // index into code input
	Message string `json:"message"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
}

// TranspileResult contains the transpiled code and the source map obtained from a transpiler run
type TranspileResult struct {
	Code      string          `json:"code"`
	SourceMap json.RawMessage `json:"map"`
}

// Error returns the error messag
func (te TranspileError) Error() string {
	return te.Message
}

// TranspileErrors is a slice of TranspileError
type TranspileErrors []*TranspileError

func (tes TranspileErrors) Error() string {
	switch len(tes) {
	case 0:
		return ""
	case 1:
		return tes[0].Error()
	default:
		return "multiple errors in []*TranspileError"
	}
}

// Transpiler allows building transpiled source code and a source map from a given ES6 source string
type Transpiler interface {
	Transpile(ctx context.Context, code ...string) ([]TranspileResult, error)
}

type externalTranspiler struct {
	execCmd string
}

// NewExternalTranspiler returns an instance of Transpiler that works by invoking the binary
// at the given path
func NewExternalTranspiler(command string) Transpiler {
	return externalTranspiler{
		execCmd: command,
	}
}

// Transpile performs a transpile step by running an external binary
func (et externalTranspiler) Transpile(ctx context.Context, codes ...string) ([]TranspileResult, error) {
	if len(codes) == 0 {
		return nil, nil
	}

	cmd := exec.CommandContext(ctx, et.execCmd)

	codesInput, err := json.Marshal(codes)
	if err != nil {
		return nil, err
	}

	cmd.Stdin = bytes.NewReader(codesInput)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err = cmd.Run(); err != nil {
		return nil, err
	}
	type transpileOutResult struct {
		Results []TranspileResult `json:"results,omitempty"`
		Errors  TranspileErrors   `json:"errors,omitempty"`
	}
	dec := json.NewDecoder(bytes.NewReader(out.Bytes()))

	res := transpileOutResult{}
	if err = dec.Decode(&res); err != nil {
		return nil, err
	}
	if len(res.Errors) != 0 {
		return nil, res.Errors
	}

	return res.Results, nil
}
