package app

import (
	"encoding/json"
)

func Import(payload []byte) (a App, err error) {
	err = json.Unmarshal(payload, &a)
	return
}

func (a App) Export() []byte {
	raw, _ := json.Marshal(a)
	return raw
}
