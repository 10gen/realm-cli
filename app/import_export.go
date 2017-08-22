package app

import (
	"bytes"
	"encoding/json"

	"github.com/burntsushi/toml"
)

// TODO: decide on the import/export representation

func Import(payload []byte) (a App, err error) {
	return ImportTOML(payload)
}

func (a App) Export() []byte {
	return a.ExportTOML()
}

func ImportTOML(payload []byte) (a App, err error) {
	_, err = toml.Decode(string(payload), &a)
	return
}

func (a App) ExportTOML() []byte {
	buf := new(bytes.Buffer)
	enc := toml.NewEncoder(buf)
	enc.Encode(a)
	return buf.Bytes()
}

func ImportJSON(payload []byte) (a App, err error) {
	err = json.Unmarshal(payload, &a)
	return
}

func (a App) ExportJSON() []byte {
	raw, _ := json.Marshal(a)
	return raw
}
