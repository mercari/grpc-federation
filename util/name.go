package util

import (
	"strings"
	_ "unsafe"
)

//go:linkname GoCamelCase google.golang.org/protobuf/internal/strs.GoCamelCase
func GoCamelCase(string) string

func ToPublicGoVariable(name string) string {
	return ToPublicVariable(ToGoVariable(name))
}

func ToPrivateGoVariable(name string) string {
	return ToPrivateVariable(ToGoVariable(name))
}

func ToGoVariable(v string) string {
	return GoCamelCase(v)
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
