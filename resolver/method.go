package resolver

func (m *Method) FederationResponse() *Message {
	if m.Rule != nil && m.Rule.Response != nil {
		return m.Rule.Response
	}
	return m.Response
}
