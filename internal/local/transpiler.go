package local

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
)

const (
	defaultTranspilerCommand = "transpiler"
)

// Transpiler is a transpiler
type Transpiler interface {
	Transpile(ctx context.Context, sources ...string) ([]string, error)
}

func newDefaultTranspiler() (Transpiler, error) {
	return newTranspiler(defaultTranspilerCommand)
}

func newTranspiler(cmd string) (Transpiler, error) {
	if _, err := exec.LookPath(cmd); err != nil {
		// TODO(REALMC-8127): should return an error here capable of instructing the user how to download/build the transpiler
		return nil, err
	}
	return &externalTranspiler{cmd}, nil
}

type externalTranspiler struct {
	cmd string
}

func (t *externalTranspiler) Transpile(ctx context.Context, sources ...string) ([]string, error) {
	if len(sources) == 0 {
		return nil, nil
	}

	cmd := exec.CommandContext(ctx, t.cmd)

	in, inErr := json.Marshal(sources)
	if inErr != nil {
		return nil, inErr
	}
	cmd.Stdin = bytes.NewReader(in)

	out := new(bytes.Buffer)
	cmd.Stdout = out

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var res transpilationResult
	if err := json.NewDecoder(bytes.NewReader(out.Bytes())).Decode(&res); err != nil {
		return nil, err
	}

	if len(res.Errors) > 0 {
		return nil, res.Errors
	}

	codes := make([]string, len(res.Sources))
	for i, source := range res.Sources {
		codes[i] = source.Code
	}
	return codes, nil
}

type transpilationResult struct {
	Sources []transpiledSource  `json:"results,omitempty"`
	Errors  transpilationErrors `json:"errors,omitempty"`
}

type transpiledSource struct {
	Code string          `json:"code"`
	Map  json.RawMessage `json:"map"`
}

type transpilationErrors []transpilationError

func (err transpilationErrors) Error() string {
	switch len(err) {
	case 0:
		return ""
	case 1:
		return err[0].Error()
	default:
		return "multiple errors occurred during transpilation"
	}
}

type transpilationError struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
}

func (err transpilationError) Error() string {
	return err.Message
}
