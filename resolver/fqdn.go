package resolver

import (
	"fmt"
	"strings"
)

func (s *Service) FQDN() string {
	return fmt.Sprintf("%s.%s", s.PackageName(), s.Name)
}

func (m *Method) FQDN() string {
	return fmt.Sprintf("%s/%s", m.Service.FQDN(), m.Name)
}

func (m *Message) FQDN() string {
	return strings.Join(
		append(append([]string{m.PackageName()}, m.ParentMessageNames()...), m.Name),
		".",
	)
}

func (f *Field) FQDN() string {
	if f.Message == nil {
		return f.Name
	}
	return fmt.Sprintf("%s.%s", f.Message.FQDN(), f.Name)
}

func (f *OneofField) FQDN() string {
	return fmt.Sprintf("%s.%s", f.Oneof.Message.FQDN(), f.Name)
}

func (e *Enum) FQDN() string {
	if e == nil {
		return ""
	}
	if e.Message != nil {
		return fmt.Sprintf("%s.%s", e.Message.FQDN(), e.Name)
	}
	return fmt.Sprintf("%s.%s", e.PackageName(), e.Name)
}

func (v *EnumValue) FQDN() string {
	return fmt.Sprintf("%s.%s", v.Enum.FQDN(), v.Value)
}

func (t *Type) FQDN() string {
	var repeated string
	if t.Repeated {
		repeated = "repeated "
	}
	if t.OneofField != nil {
		return repeated + t.OneofField.FQDN()
	}
	if t.Message != nil {
		if t.Message.IsMapEntry {
			return "map<" + t.Message.Fields[0].Type.FQDN() + ", " + t.Message.Fields[1].Type.FQDN() + ">"
		}
		return repeated + t.Message.FQDN()
	}
	if t.Enum != nil {
		return repeated + t.Enum.FQDN()
	}
	return repeated + t.Kind.ToString()
}

func (n *MessageDependencyGraphNode) FQDN() string {
	return fmt.Sprintf("%s_%s", n.BaseMessage.FQDN(), n.VariableDefinition.Name)
}
