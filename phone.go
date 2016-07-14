// Package phone is a package to parse phone numbers from strings. It figures out the country and the other details for a given phone number and supports 156 countries as of now.
// It also supports checking for test and toll free phone numbers within the US.
package phone

import "fmt"

import "strconv"

import "strings"

import "errors"

import (
	"math"
	"github.com/davegardnerisme/phonegeocode"
)

type ErrUnsupportedCountry struct {
	Country string
}

func (e ErrUnsupportedCountry) Error() string {
	return fmt.Sprintf("Unsupported country: %s", e.Country)
}

type Phone struct {
	CountryCode    int    `json:"country_code"`
	NationalNumber int64  `json:"national_number"`
	HasLeadingZero bool   `json:"has_leading_zero"`
	CountryName    string `json:"country_name"`
}

type ParsingOpts struct {
	DefaultCountry   string
	MinLengthToUse   int
	DontUseMaxLength bool
}

func (p *Phone) IsEmpty() bool {
	return p.CountryCode == 0 && p.NationalNumber == 0
}

func (p *Phone) GetId() string {
	return fmt.Sprintf("%d:%d", p.CountryCode, p.NationalNumber)
}

func (p *Phone) Equals(other *Phone) bool {
	return other != nil &&
		other.CountryCode == p.CountryCode &&
		other.NationalNumber == p.NationalNumber
}

func PhoneStringFromId(id string) string {
	splits := strings.Split(id, ":")
	if len(splits) != 2 {
		return ""
	}
	return fmt.Sprintf("+%s%s", splits[0], splits[1])
}

func PhoneFromId(id string) (*Phone, error) {
	splits := strings.Split(id, ":")
	if len(splits) != 2 {
		return nil, errors.New("Invalid phone number id")
	}
	return ParseNumber(fmt.Sprintf("%s%s", splits[0], splits[1]))
}

func (p *Phone) String() string {
	return fmt.Sprintf("+%d%d", p.CountryCode, p.NationalNumber)
}

var phoneCoder = phonegeocode.New()
var numberCleaner = strings.NewReplacer(".", "", "-", "", " ", "", "(", "", ")", "", "\u00a0", "")



// Parse phone numbers.
var defaultOpts = ParsingOpts{}

// Parses a phone number string as a Phone object and returns a descriptive error if the number is invalid or unsupported
func ParseNumber(number string) (*Phone, error) {
	return ParseNumberWithOpts(number, defaultOpts)
}

// Same as ParseNumber. The given country is used as the default if the country cannot be found
func ParseNumberForCountry(number string, defaultCountry string) (*Phone, error) {
	return ParseNumberWithOpts(number, ParsingOpts{DefaultCountry: defaultCountry})
}

// Used for parsing phone numbers defined as constants
// Panics if the number is invalid
func MustParse(number string) Phone {
	phone, err := ParseNumber(number)
	if err != nil {
		panic(fmt.Sprintf("Trying to MustParse an invalid number: %s . Error: %s", number, err.Error()))
	}
	return *phone
}

func GetCountryFromPhone(number string) string {
	if cnt, err := phoneCoder.Country(number); err != nil {
		return "UNKNOWN"
	} else {
		return cnt
	}
}

func ParseNumberWithOpts(number string, opts ParsingOpts) (*Phone, error) {
	origNumber := number
	number = strings.TrimPrefix(number, "011")
	number = strings.TrimPrefix(number, "00")
	number = strings.TrimPrefix(number, "+")
	hasCountryCode := false
	if len(number) != len(origNumber) {
		hasCountryCode = true
	}
	number = numberCleaner.Replace(number)

	var cntStr string
	var err error
	var cntDeterminedByPrefix = false
	if hasCountryCode || opts.DefaultCountry == "" {
		cntStr, err = phoneCoder.Country(number)
		if err != nil {
			return nil, err
		}
		cntDeterminedByPrefix = true
	} else {
		cntStr = opts.DefaultCountry
	}
	cnt, ok := countryDetails()[cntStr]
	if !ok {
		return nil, ErrUnsupportedCountry{cntStr}
	}

	if cntDeterminedByPrefix {
		number = strings.TrimPrefix(number, fmt.Sprintf("%d", cnt.Code))
	}

	minLen := cnt.MinLen
	if opts.MinLengthToUse > 0 {
		minLen = opts.MinLengthToUse
	}

	if len(number) < minLen {
		return nil, errors.New(fmt.Sprintf("Number is too short. Expected length at least %d, your national number is %s", minLen, number))
	}

	phone := new(Phone)
	phone.CountryCode = cnt.Code
	phone.CountryName = cntStr

	lenToFetch := int(math.Min(float64(len(number)), float64(cnt.MaxLen)))
	if opts.DontUseMaxLength {
		lenToFetch = len(number)
	}

	numToSave := number[len(number)-lenToFetch:]
	if strings.HasPrefix(numToSave, "0") {
		phone.HasLeadingZero = true
	}
	phone.NationalNumber, err = strconv.ParseInt(numToSave, 10, 64)
	return phone, err
}

func defaultLocalFormat() string {
	return "###############"
}

func defaultI18nFormat(countryCode int) string {
	return fmt.Sprintf("+%d#############", countryCode)
}

// in-place
func reverse(a []int) []int {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
	return a
}

func digits(num int64) []int {
	output := make([]int, 0, 11)
	for num > 0 {
		output = append(output, int(num%10))
		num /= 10
	}
	return reverse(output)
}

func phoneFormat(format string, num int64) string {
	nums := digits(num)
	currNum := 0
	toReturn := make([]rune, 0, len(format))
	for _, char := range []rune(format) {
		if char == '#' {
			if currNum < len(nums) {
				toReturn = append(toReturn, rune(nums[currNum])+'0')
				currNum += 1
			} else {
				// otherwise, just skip this character
			}
		} else {
			toReturn = append(toReturn, char)
		}
	}
	return string(toReturn)
}

func (p *Phone) Format(relativeTo *Phone) string {
	if p.CountryName == relativeTo.CountryName {
		return p.FormatLocal()
	} else {
		return p.FormatI18n()
	}
}

func (p *Phone) FormatLocal() string {
	if country, ok := countryDetails()[p.CountryName]; ok && country.LocalFormat != "" {
		return phoneFormat(country.LocalFormat, p.NationalNumber)
	} else {
		return phoneFormat(defaultLocalFormat(), p.NationalNumber)
	}
}

func (p *Phone) FormatI18n() string {
	if country, ok := countryDetails()[p.CountryName]; ok && country.I18nFormat != "" {
		return phoneFormat(country.I18nFormat, p.NationalNumber)
	} else {
		return phoneFormat(defaultI18nFormat(p.CountryCode), p.NationalNumber)
	}
}

func (p *Phone) IsFromCountry(country string) bool {
	if p.CountryName != "" {
		return p.CountryName == country
	}
	if country, ok := countryDetails()[country]; ok {
		return country.Code == p.CountryCode
	}
	return false
}

// from: http://en.wikipedia.org/wiki/Toll-free_telephone_number
func (p *Phone) IsInvitable() bool {
	if p.CountryCode == 1 {
		areaCode := p.NationalNumber / (100 * 100 * 1000) /* strip local XXX-XXXX */
		switch areaCode {
		case 800:
			fallthrough
		case 822:
			fallthrough
		case 833:
			fallthrough
		case 844:
			fallthrough
		case 855:
			fallthrough
		case 866:
			fallthrough
		case 877:
			fallthrough
		case 880:
			fallthrough
		case 881:
			fallthrough
		case 882:
			fallthrough
		case 883:
			fallthrough
		case 884:
			fallthrough
		case 885:
			fallthrough
		case 886:
			fallthrough
		case 887:
			fallthrough
		case 888:
			fallthrough
		case 889:
			fallthrough
		case 900:
			return false
		default:
			return true
		}
	} else {
		return true
	}
}

func (p *Phone) IsTest() bool {
	if p.CountryCode == 1 {
		// Check if the middle three digits are 555
		if int64(math.Mod(float64(p.NationalNumber), 10000000))/10000 == 555 {
			return true
		}
	}
	return false
}
