package source

func (loc *Location) IsDefinedTypeName() bool {
	return loc.Message.IsDefinedTypeName()
}

func (loc *Location) IsDefinedFieldType() bool {
	return loc.Message.IsDefinedFieldType()
}

func (loc *Location) IsDefinedTypeAlias() bool {
	switch {
	case loc.Message != nil:
		return loc.Message.IsDefinedTypeAlias()
	case loc.Enum != nil:
		return loc.Enum.IsDefinedTypeAlias()
	}
	return false
}

func (m *Message) IsDefinedTypeName() bool {
	if m == nil {
		return false
	}
	switch {
	case m.NestedMessage != nil:
		return m.NestedMessage.IsDefinedTypeName()
	case m.Option != nil:
		return m.Option.IsDefinedTypeName()
	case m.Oneof != nil:
		return m.Oneof.IsDefinedTypeName()
	case m.Field != nil:
		return m.Field.IsDefinedTypeName()
	}
	return false
}

func (o *Oneof) IsDefinedTypeName() bool {
	if o == nil {
		return false
	}
	return o.Field.IsDefinedTypeName()
}

func (f *Field) IsDefinedTypeName() bool {
	if f == nil {
		return false
	}
	return f.Option.IsDefinedTypeName()
}

func (o *FieldOption) IsDefinedTypeName() bool {
	if o == nil {
		return false
	}
	return o.Oneof.IsDefinedTypeName()
}

func (o *FieldOneof) IsDefinedTypeName() bool {
	if o == nil {
		return false
	}
	return o.Def.IsDefinedTypeName()
}

func (o *MessageOption) IsDefinedTypeName() bool {
	if o == nil {
		return false
	}
	return o.Def.IsDefinedTypeName()
}

func (o *VariableDefinitionOption) IsDefinedTypeName() bool {
	if o == nil {
		return false
	}
	switch {
	case o.Map != nil:
		return o.Map.IsDefinedTypeName()
	case o.Call != nil:
		return o.Call.IsDefinedTypeName()
	case o.Message != nil:
		return o.Message.IsDefinedTypeName()
	case o.Enum != nil:
		return o.Enum.IsDefinedTypeName()
	case o.Validation != nil:
		return o.Validation.IsDefinedTypeName()
	}
	return false
}

func (o *MapExprOption) IsDefinedTypeName() bool {
	if o == nil {
		return false
	}
	switch {
	case o.Message != nil:
		return o.Message.IsDefinedTypeName()
	case o.Enum != nil:
		return o.Enum.IsDefinedTypeName()
	}
	return false
}

func (o *CallExprOption) IsDefinedTypeName() bool {
	if o == nil {
		return false
	}
	return o.Error.IsDefinedTypeName()
}

func (o *MessageExprOption) IsDefinedTypeName() bool {
	if o == nil {
		return false
	}
	return o.Name
}

func (o *EnumExprOption) IsDefinedTypeName() bool {
	if o == nil {
		return false
	}
	return o.Name
}

func (o *ValidationExprOption) IsDefinedTypeName() bool {
	if o == nil {
		return false
	}
	return o.Error.IsDefinedTypeName()
}

func (o *GRPCErrorOption) IsDefinedTypeName() bool {
	if o == nil {
		return false
	}
	switch {
	case o.Def != nil:
		return o.Def.IsDefinedTypeName()
	case o.Detail != nil:
		return o.Detail.IsDefinedTypeName()
	}
	return false
}

func (o *GRPCErrorDetailOption) IsDefinedTypeName() bool {
	if o == nil {
		return false
	}
	switch {
	case o.Def != nil:
		return o.Def.IsDefinedTypeName()
	case o.Message != nil:
		return o.Message.IsDefinedTypeName()
	}
	return false
}

func (m *Message) IsDefinedFieldType() bool {
	if m == nil {
		return false
	}
	switch {
	case m.NestedMessage != nil:
		return m.NestedMessage.IsDefinedFieldType()
	case m.Field != nil:
		return m.Field.IsDefinedFieldType()
	case m.Oneof != nil:
		return m.Oneof.IsDefinedFieldType()
	}
	return false
}

func (f *Field) IsDefinedFieldType() bool {
	if f == nil {
		return false
	}
	return f.Type
}

func (o *Oneof) IsDefinedFieldType() bool {
	if o == nil {
		return false
	}
	return o.Field.IsDefinedFieldType()
}

func (m *Message) IsDefinedTypeAlias() bool {
	if m == nil {
		return false
	}
	switch {
	case m.Enum != nil:
		return m.Enum.IsDefinedTypeAlias()
	case m.NestedMessage != nil:
		return m.NestedMessage.IsDefinedTypeAlias()
	case m.Option != nil:
		return m.Option.IsDefinedTypeAlias()
	case m.Oneof != nil:
		return m.Oneof.IsDefinedTypeAlias()
	case m.Field != nil:
		return m.Field.IsDefinedTypeAlias()
	}
	return false
}

func (o *MessageOption) IsDefinedTypeAlias() bool {
	if o == nil {
		return false
	}
	return o.Alias
}

func (o *Oneof) IsDefinedTypeAlias() bool {
	if o == nil {
		return false
	}
	return o.Field.IsDefinedTypeAlias()
}

func (f *Field) IsDefinedTypeAlias() bool {
	if f == nil {
		return false
	}
	return f.Option.IsDefinedTypeAlias()
}

func (o *FieldOption) IsDefinedTypeAlias() bool {
	if o == nil {
		return false
	}
	return o.Alias
}

func (e *Enum) IsDefinedTypeAlias() bool {
	if e == nil {
		return false
	}
	switch {
	case e.Option != nil:
		return e.Option.IsDefinedTypeAlias()
	case e.Value != nil:
		return e.Value.IsDefinedTypeAlias()
	}
	return false
}

func (o *EnumOption) IsDefinedTypeAlias() bool {
	if o == nil {
		return false
	}
	return o.Alias
}

func (v *EnumValue) IsDefinedTypeAlias() bool {
	if v == nil {
		return false
	}
	return v.Option.IsDefinedTypeAlias()
}

func (o *EnumValueOption) IsDefinedTypeAlias() bool {
	if o == nil {
		return false
	}
	return o.Alias
}
