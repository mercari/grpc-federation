package resolver

import (
	"fmt"
)

func (f *OneofField) FQDN() string {
	return fmt.Sprintf("%s.%s", f.Oneof.Message.FQDN(), f.Name)
}
