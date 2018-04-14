package botnames

import "encoding/json"

//go:generate go-bindata -pkg $GOPACKAGE bots.json

//New returns a list of Wikipedia bots
func New() (bots []string, err error) {
	bytes, err := Asset("bots.json")
	if err != nil {
		return
	}

	bots = make([]string, 0, len(bytes)/10)
	if err = json.Unmarshal(bytes, &bots); err != nil {
		bots = nil
	}

	return
}
