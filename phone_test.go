package phone

import (
	"testing"
)

var testMap = map[string]Phone{
	"+14154403838":    Phone{1, 4154403838, false, "US"},
	"14154403838":     Phone{1, 4154403838, false, "US"},
	"011919839022222": Phone{91, 9839022222, false, "IN"},
	"00919839022222":  Phone{91, 9839022222, false, "IN"},
	"+64021234567":    Phone{64, 21234567, true, "NZ"},
	"+640212345678":   Phone{64, 212345678, true, "NZ"},
	"+6402123456789":  Phone{64, 2123456789, true, "NZ"},
	"+61423 652 915":  Phone{61, 423652915, false, "AU"},
}

var testMapDefault = map[string]map[string]Phone{
	"US": {"4154403838": Phone{1, 4154403838, false, "US"}},
	"IN": {
		"9839022222":  Phone{91, 9839022222, false, "IN"},
		"09839022222": Phone{91, 9839022222, false, "IN"},
	},
	"NZ": {"021 234 5678": Phone{64, 212345678, true, "NZ"}},
}

func TestNumberParsing(t *testing.T) {
	for k, v := range testMap {
		phone, err := ParseNumber(k)
		if err != nil {
			t.Errorf(err.Error())
		} else if phone.CountryCode != v.CountryCode || phone.NationalNumber != v.NationalNumber || phone.HasLeadingZero != v.HasLeadingZero {
			t.Errorf("Unmatched phone. Expecting: %+v \n Got: %+v", v, phone)
		}
	}

	for def, tst := range testMapDefault {
		for k, v := range tst {
			phone, err := ParseNumberForCountry(k, def)
			if err != nil {
				t.Errorf(err.Error())
			} else if phone.CountryCode != v.CountryCode || phone.NationalNumber != v.NationalNumber || phone.HasLeadingZero != v.HasLeadingZero {
				t.Errorf("Unmatched phone. Expecting: %+v \n Got: %+v", v, phone)
			}
		}
	}
}

var testOptMap = map[string]Phone{
	"+91123456789004": Phone{91, 123456789004, false, "IN"},
	"+1607546":        Phone{1, 607546, false, "US"},
}

func TestParsingWithOpts(t *testing.T) {
	for k, v := range testOptMap {
		if phone, err := ParseNumberWithOpts(k, ParsingOpts{MinLengthToUse: 4, DontUseMaxLength: true}); err != nil {
			t.Errorf("Should have parsed fine, but didn't")
		} else if *phone != v {
			t.Errorf("Expected: %+v , found: %+v, for: %s", v, phone, k)
		}
	}
}
