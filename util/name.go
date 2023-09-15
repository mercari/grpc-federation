package util

import (
	"bytes"
	"strings"
)

func ToPublicGoVariable(name string) string {
	return ToGoVariable(ToPublicVariable(name))
}

func ToGoVariable(v string) string {
	var (
		isLargeChar bool
		camelV      []byte
	)
	for _, c := range v {
		if c == '_' || c == '-' {
			isLargeChar = true
			continue
		}
		if isLargeChar {
			camelV = append(camelV, bytes.ToUpper([]byte{byte(c)})...)
			isLargeChar = false
		} else {
			camelV = append(camelV, byte(c))
		}
	}
	return string(camelV)
}

func ToPublicVariable(name string) string {
	if len(name) <= 1 {
		return strings.ToUpper(name)
	}
	return strings.ToUpper(string(name[0])) + name[1:]
}

func ToPrivateVariable(name string) string {
	if len(name) <= 1 {
		return strings.ToLower(name)
	}
	return strings.ToLower(string(name[0])) + name[1:]
}
