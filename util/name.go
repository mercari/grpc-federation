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
		snakeV      []byte
	)
	for _, c := range v {
		if c == '_' {
			isLargeChar = true
			continue
		}
		if isLargeChar {
			snakeV = append(snakeV, bytes.ToUpper([]byte{byte(c)})...)
			isLargeChar = false
		} else {
			snakeV = append(snakeV, byte(c))
		}
	}
	return string(snakeV)
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
