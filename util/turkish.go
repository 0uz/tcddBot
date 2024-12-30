package util

import "strings"

var turkishToLower = map[rune]rune{
	'İ': 'i',
	'I': 'ı',
	'Ğ': 'ğ',
	'Ü': 'ü',
	'Ş': 'ş',
	'Ö': 'ö',
	'Ç': 'ç',
}

var turkishToUpper = map[rune]rune{
	'i': 'İ',
	'ı': 'I',
	'ğ': 'Ğ',
	'ü': 'Ü',
	'ş': 'Ş',
	'ö': 'Ö',
	'ç': 'Ç',
}

var turkishToASCII = map[rune]string{
	'ı': "i",
	'İ': "I",
	'ğ': "g",
	'Ğ': "G",
	'ü': "u",
	'Ü': "U",
	'ş': "s",
	'Ş': "S",
	'ö': "o",
	'Ö': "O",
	'ç': "c",
	'Ç': "C",
}

// ToLowerTurkish converts a string to lowercase considering Turkish characters
func ToLowerTurkish(s string) string {
	result := make([]rune, 0, len(s))
	for _, r := range s {
		if lower, ok := turkishToLower[r]; ok {
			result = append(result, lower)
		} else {
			result = append(result, []rune(strings.ToLower(string(r)))...)
		}
	}
	return string(result)
}

// ToUpperTurkish converts a string to uppercase considering Turkish characters
func ToUpperTurkish(s string) string {
	result := make([]rune, 0, len(s))
	for _, r := range s {
		if upper, ok := turkishToUpper[r]; ok {
			result = append(result, upper)
		} else {
			result = append(result, []rune(strings.ToUpper(string(r)))...)
		}
	}
	return string(result)
}

// ToASCII converts Turkish characters to their ASCII equivalents
func ToASCII(s string) string {
	var result strings.Builder
	for _, r := range s {
		if ascii, ok := turkishToASCII[r]; ok {
			result.WriteString(ascii)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
