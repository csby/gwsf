package gtype

import "regexp"

func IsEmailFormat(value string) bool {
	pattern := `^[0-9a-zA-Z][_.0-9a-zA-Z-]{0,31}@([0-9a-zA-Z][0-9a-zA-Z-]{0,30}[0-9a-zA-Z]\.){1,4}[a-zA-Z]{2,4}$`

	reg := regexp.MustCompile(pattern)
	return reg.MatchString(value)
}

func IsMobileFormat(value string) bool {
	pattern := "^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,3,5-8])|(18[0-9])|166|198|199|(147))\\d{8}$"

	reg := regexp.MustCompile(pattern)
	return reg.MatchString(value)
}

func IsAccountFormat(value string) bool {
	if IsMobileFormat(value) {
		return false
	}
	if IsEmailFormat(value) {
		return false
	}

	pattern := `^[0-9a-zA-Z]{1,31}[_.0-9a-zA-Z-]{0,31}`
	ok, err := regexp.MatchString(pattern, value)
	if err != nil {
		return false
	}

	return ok
}
