package phone

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"runtime"
	"sync"
)

type country struct {
	MinLen int `json:"minLen"`
	MaxLen int `json:"maxLen"`
	Code   int `json:"code"`
	// # for the real number
	// everything else is format
	// ex: (0###) ##-####
	LocalFormat string `json:"local"`
	I18nFormat  string `json:"i18n"`
}

var once = sync.Once{}
var _countryDetails map[string]country = nil

func countryDetails() map[string]country {
	if _countryDetails == nil {
		_, filename, _, _ := runtime.Caller(1)
		once.Do(func() {
			_countryDetails = fromFile(path.Join(path.Dir(filename), "raw/country-phones.json"))
		})
	}
	return _countryDetails
}

func fromFile(filename string) map[string]country {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var cnts map[string]country
	err = json.Unmarshal(file, &cnts)
	if err != nil {
		panic(err)
	}
	return cnts
}
