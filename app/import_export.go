package app

import (
	"encoding/json"
)

// Import unmarshals an exported stitch app into the App representation defined
// in this package.
func Import(payload []byte) (a App, err error) {
	err = json.Unmarshal(payload, &a)
	return
}

// Export marshals the App such that it may be imported as a stitch application
// configuration.
func (a App) Export() []byte {
	raw, _ := json.Marshal(a)
	return raw
}
