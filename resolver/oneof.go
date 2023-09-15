package resolver

import (
	"fmt"

	"github.com/mercari/grpc-federation/util"
)

func (f *OneofField) IsConflict() bool {
	msg := f.Oneof.Message
	fieldName := util.ToPublicGoVariable(f.Name)
	for _, m := range msg.NestedMessages {
		if util.ToPublicGoVariable(m.Name) == fieldName {
			return true
		}
	}
	for _, e := range msg.Enums {
		if util.ToPublicGoVariable(e.Name) == fieldName {
			return true
		}
	}
	return false
}

func (f *OneofField) FQDN() string {
	return fmt.Sprintf("%s.%s", f.Oneof.Message.FQDN(), f.Name)
}
