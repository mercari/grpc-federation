package resolver

import "fmt"

func (v *EnumValue) FQDN() string {
	if v.Enum.Message != nil {
		return fmt.Sprintf("%s.%s", v.Enum.Message.FQDN(), v.Value)
	}
	return fmt.Sprintf("%s.%s", v.Enum.PackageName(), v.Value)
}
