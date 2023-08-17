package resolver

import "fmt"

func (m *Method) FQDN() string {
	return fmt.Sprintf("%s/%s", m.Service.FQDN(), m.Name)
}
