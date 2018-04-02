package utils

import (
	"bytes"
	"crypto/rand"
	"math/big"
)

var (
	stringAlpha               = []rune("abcdefghijklmnopqrstuvwxyz")
	stringAlphaUpper          = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	stringNumeric             = []rune("1234567890")
	stringSpecial             = []rune("!@#$%^&*()_-")
	stringAlphaNumeric        = append(append(append([]rune{}, stringAlpha...), stringAlphaUpper...), stringNumeric...)
	stringAlphaNumericSpecial = append(stringAlphaNumeric, stringSpecial...)

	stringAlphaMapping      = map[rune]struct{}{}
	stringAlphaUpperMapping = map[rune]struct{}{}
	stringNumericMapping    = map[rune]struct{}{}
	stringSpecialMapping    = map[rune]struct{}{}
)

func init() {
	for _, r := range stringAlpha {
		stringAlphaMapping[r] = struct{}{}
	}
	for _, r := range stringAlphaUpper {
		stringAlphaUpperMapping[r] = struct{}{}
	}
	for _, r := range stringNumeric {
		stringNumericMapping[r] = struct{}{}
	}
	for _, r := range stringSpecial {
		stringSpecialMapping[r] = struct{}{}
	}
}

func randomStringFromAlphabet(alpha []rune, length int) string {
	var buf bytes.Buffer

	for i := 0; i < length; i++ {
		val, err := rand.Int(rand.Reader, big.NewInt(int64(len(alpha))))
		if err != nil {
			panic(err)
		}
		buf.WriteRune(alpha[val.Int64()])
	}

	return buf.String()
}

// RandomAlphaString generates a new random alphabetic string of given length
func RandomAlphaString(length int) string {
	return randomStringFromAlphabet(stringAlpha, length)
}

// RandomAlphaNumericString generates a new random alphanumeric key
func RandomAlphaNumericString(length int) string {
	return randomStringFromAlphabet(stringAlphaNumeric, length)
}

// RandomAlphaNumericSpecialString generates a new random alphanumeric key
// with special characters
func RandomAlphaNumericSpecialString(length int) string {
	return randomAlphaNumericSpecialString(length, false)
}

// RandomAlphaNumericSpecialStringStrict generates a new random alphanumeric key
// with special characters
func RandomAlphaNumericSpecialStringStrict(length int) string {
	return randomAlphaNumericSpecialString(length, true)
}

func randomAlphaNumericSpecialString(length int, strict bool) string {
	for {
		candidate := randomStringFromAlphabet(stringAlphaNumericSpecial, length)

		if !strict {
			return candidate
		}

		lower, upper, number, special := false, false, false, false
		for _, sym := range candidate {

			if !lower {
				_, lower = stringAlphaMapping[sym]
			}
			if !upper {
				_, upper = stringAlphaUpperMapping[sym]
			}
			if !number {
				_, number = stringNumericMapping[sym]
			}
			if !special {
				_, special = stringSpecialMapping[sym]
			}

			if lower && upper && special && number {
				return candidate
			}
		}
	}
}
