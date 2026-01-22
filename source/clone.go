package source

func (loc *Location) Clone() *Location {
	if loc == nil {
		return nil
	}
	return &Location{
		FileName:  loc.FileName,
		Export:    loc.Export.Clone(),
		GoPackage: loc.GoPackage,
		Service:   loc.Service.Clone(),
		Message:   loc.Message.Clone(),
		Enum:      loc.Enum.Clone(),
	}
}

func (e *Export) Clone() *Export {
	if e == nil {
		return nil
	}
	return &Export{
		Name:      e.Name,
		Wasm:      e.Wasm.Clone(),
		Types:     e.Types.Clone(),
		Functions: e.Functions.Clone(),
	}
}

func (w *Wasm) Clone() *Wasm {
	if w == nil {
		return nil
	}
	return &Wasm{
		URL:    w.URL,
		Sha256: w.Sha256,
	}
}

func (t *PluginType) Clone() *PluginType {
	if t == nil {
		return nil
	}
	return &PluginType{
		Idx:     t.Idx,
		Name:    t.Name,
		Methods: t.Methods.Clone(),
	}
}

func (f *PluginFunction) Clone() *PluginFunction {
	if f == nil {
		return nil
	}
	return &PluginFunction{
		Idx:        f.Idx,
		Name:       f.Name,
		Args:       f.Args.Clone(),
		ReturnType: f.ReturnType,
	}
}

func (a *PluginFunctionArgument) Clone() *PluginFunctionArgument {
	if a == nil {
		return nil
	}
	return &PluginFunctionArgument{
		Idx:  a.Idx,
		Type: a.Type,
	}
}

func (s *Service) Clone() *Service {
	if s == nil {
		return nil
	}
	return &Service{
		Name:   s.Name,
		Method: s.Method.Clone(),
		Option: s.Option.Clone(),
	}
}

func (m *Method) Clone() *Method {
	if m == nil {
		return nil
	}
	return &Method{
		Name:     m.Name,
		Request:  m.Request,
		Response: m.Response,
		Option:   m.Option.Clone(),
	}
}

func (o *ServiceOption) Clone() *ServiceOption {
	if o == nil {
		return nil
	}
	return &ServiceOption{
		Env: o.Env.Clone(),
		Var: o.Var.Clone(),
	}
}

func (e *Env) Clone() *Env {
	if e == nil {
		return nil
	}
	return &Env{
		Message: e.Message,
		Var:     e.Var.Clone(),
	}
}

func (v *EnvVar) Clone() *EnvVar {
	if v == nil {
		return nil
	}
	return &EnvVar{
		Idx:    v.Idx,
		Name:   v.Name,
		Type:   v.Type,
		Option: v.Option.Clone(),
	}
}

func (o *EnvVarOption) Clone() *EnvVarOption {
	if o == nil {
		return nil
	}
	return &EnvVarOption{
		Alternate: o.Alternate,
		Default:   o.Default,
		Required:  o.Required,
		Ignored:   o.Ignored,
	}
}

func (sv *ServiceVariable) Clone() *ServiceVariable {
	if sv == nil {
		return nil
	}
	return &ServiceVariable{
		Idx:        sv.Idx,
		Name:       sv.Name,
		If:         sv.If,
		By:         sv.By,
		Map:        sv.Map.Clone(),
		Message:    sv.Message.Clone(),
		Enum:       sv.Enum.Clone(),
		Validation: sv.Validation.Clone(),
	}
}

func (v *ServiceVariableValidationExpr) Clone() *ServiceVariableValidationExpr {
	if v == nil {
		return nil
	}
	return &ServiceVariableValidationExpr{
		If:      v.If,
		Message: v.Message,
	}
}

func (o *MethodOption) Clone() *MethodOption {
	if o == nil {
		return nil
	}
	return &MethodOption{
		Timeout:  o.Timeout,
		Response: o.Response,
	}
}

func (e *Enum) Clone() *Enum {
	if e == nil {
		return nil
	}
	return &Enum{
		Name:   e.Name,
		Option: e.Option.Clone(),
		Value:  e.Value.Clone(),
	}
}

func (o *EnumOption) Clone() *EnumOption {
	if o == nil {
		return nil
	}
	return &EnumOption{
		Alias: o.Alias,
	}
}

func (v *EnumValue) Clone() *EnumValue {
	if v == nil {
		return nil
	}
	return &EnumValue{
		Value:  v.Value,
		Option: v.Option.Clone(),
	}
}

func (o *EnumValueOption) Clone() *EnumValueOption {
	if o == nil {
		return nil
	}
	return &EnumValueOption{
		Alias:   o.Alias,
		Default: o.Default,
		Attr:    o.Attr.Clone(),
	}
}

func (a *EnumValueAttribute) Clone() *EnumValueAttribute {
	if a == nil {
		return nil
	}
	return &EnumValueAttribute{
		Idx:   a.Idx,
		Name:  a.Name,
		Value: a.Value,
	}
}

func (m *Message) Clone() *Message {
	if m == nil {
		return nil
	}
	return &Message{
		Name:          m.Name,
		Option:        m.Option.Clone(),
		Field:         m.Field.Clone(),
		Enum:          m.Enum.Clone(),
		Oneof:         m.Oneof.Clone(),
		NestedMessage: m.NestedMessage.Clone(),
	}
}

func (f *Field) Clone() *Field {
	if f == nil {
		return nil
	}
	return &Field{
		Name:   f.Name,
		Option: f.Option.Clone(),
	}
}

func (o *FieldOption) Clone() *FieldOption {
	if o == nil {
		return nil
	}
	return &FieldOption{
		By:    o.By,
		Alias: o.Alias,
		Oneof: o.Oneof.Clone(),
	}
}

func (o *FieldOneof) Clone() *FieldOneof {
	if o == nil {
		return nil
	}
	return &FieldOneof{
		If:      o.If,
		Default: o.Default,
		Def:     o.Def.Clone(),
		By:      o.By,
	}
}

func (o *Oneof) Clone() *Oneof {
	if o == nil {
		return nil
	}
	return &Oneof{
		Name:   o.Name,
		Option: o.Option,
		Field:  o.Field.Clone(),
	}
}

func (o *OneofOption) Clone() *OneofOption {
	if o == nil {
		return nil
	}
	return &OneofOption{}
}

func (o *MessageOption) Clone() *MessageOption {
	if o == nil {
		return nil
	}
	return &MessageOption{
		Def:   o.Def.Clone(),
		Alias: o.Alias,
	}
}

func (o *VariableDefinitionOption) Clone() *VariableDefinitionOption {
	if o == nil {
		return nil
	}
	return &VariableDefinitionOption{
		Idx:        o.Idx,
		Name:       o.Name,
		If:         o.If,
		By:         o.By,
		Map:        o.Map.Clone(),
		Call:       o.Call.Clone(),
		Message:    o.Message.Clone(),
		Enum:       o.Enum.Clone(),
		Validation: o.Validation.Clone(),
		Switch:     o.Switch.Clone(),
	}
}

func (o *MapExprOption) Clone() *MapExprOption {
	if o == nil {
		return nil
	}
	return &MapExprOption{
		Iterator: o.Iterator.Clone(),
		By:       o.By,
		Message:  o.Message.Clone(),
		Enum:     o.Enum.Clone(),
	}
}

func (o *IteratorOption) Clone() *IteratorOption {
	if o == nil {
		return nil
	}
	return &IteratorOption{
		Name:   o.Name,
		Source: o.Source,
	}
}

func (o *CallExprOption) Clone() *CallExprOption {
	if o == nil {
		return nil
	}
	return &CallExprOption{
		Method:   o.Method,
		Request:  o.Request.Clone(),
		Timeout:  o.Timeout,
		Retry:    o.Retry.Clone(),
		Error:    o.Error.Clone(),
		Option:   o.Option.Clone(),
		Metadata: o.Metadata,
	}
}

func (o *GRPCCallOption) Clone() *GRPCCallOption {
	if o == nil {
		return nil
	}
	return &GRPCCallOption{
		ContentSubtype:     o.ContentSubtype,
		Header:             o.Header,
		Trailer:            o.Trailer,
		MaxCallRecvMsgSize: o.MaxCallRecvMsgSize,
		MaxCallSendMsgSize: o.MaxCallSendMsgSize,
		StaticMethod:       o.StaticMethod,
		WaitForReady:       o.WaitForReady,
	}
}

func (o *MessageExprOption) Clone() *MessageExprOption {
	if o == nil {
		return nil
	}
	return &MessageExprOption{
		Name: o.Name,
		Args: o.Args.Clone(),
	}
}

func (o *EnumExprOption) Clone() *EnumExprOption {
	if o == nil {
		return nil
	}
	return &EnumExprOption{
		Name: o.Name,
		By:   o.By,
	}
}

func (o *RequestOption) Clone() *RequestOption {
	if o == nil {
		return nil
	}
	return &RequestOption{
		Idx:   o.Idx,
		Field: o.Field,
		By:    o.By,
	}
}

func (o *RetryOption) Clone() *RetryOption {
	if o == nil {
		return nil
	}
	return &RetryOption{
		Constant:    o.Constant.Clone(),
		Exponential: o.Exponential.Clone(),
	}
}

func (o *RetryConstantOption) Clone() *RetryConstantOption {
	if o == nil {
		return nil
	}
	return &RetryConstantOption{
		Interval:   o.Interval,
		MaxRetries: o.MaxRetries,
	}
}

func (o *RetryExponentialOption) Clone() *RetryExponentialOption {
	if o == nil {
		return nil
	}
	return &RetryExponentialOption{
		InitialInterval:     o.InitialInterval,
		RandomizationFactor: o.RandomizationFactor,
		Multiplier:          o.Multiplier,
		MaxInterval:         o.MaxInterval,
		MaxRetries:          o.MaxRetries,
	}
}

func (o *ArgumentOption) Clone() *ArgumentOption {
	if o == nil {
		return nil
	}
	return &ArgumentOption{
		Idx:    o.Idx,
		Name:   o.Name,
		By:     o.By,
		Inline: o.Inline,
	}
}

func (o *ValidationExprOption) Clone() *ValidationExprOption {
	if o == nil {
		return nil
	}
	return &ValidationExprOption{
		Name:  o.Name,
		Error: o.Error.Clone(),
	}
}

func (o *GRPCErrorOption) Clone() *GRPCErrorOption {
	if o == nil {
		return nil
	}
	return &GRPCErrorOption{
		Idx:     o.Idx,
		Def:     o.Def.Clone(),
		If:      o.If,
		Code:    o.Code,
		Message: o.Message,
		Ignore:  o.Ignore,
		Detail:  o.Detail.Clone(),
	}
}

func (o *GRPCErrorDetailOption) Clone() *GRPCErrorDetailOption {
	if o == nil {
		return nil
	}
	return &GRPCErrorDetailOption{
		Idx:                 o.Idx,
		Def:                 o.Def.Clone(),
		If:                  o.If,
		Message:             o.Message.Clone(),
		PreconditionFailure: o.PreconditionFailure.Clone(),
		BadRequest:          o.BadRequest.Clone(),
		LocalizedMessage:    o.LocalizedMessage.Clone(),
	}
}

func (o *GRPCErrorDetailPreconditionFailureOption) Clone() *GRPCErrorDetailPreconditionFailureOption {
	if o == nil {
		return nil
	}
	return &GRPCErrorDetailPreconditionFailureOption{
		Idx:       o.Idx,
		Violation: o.Violation,
	}
}

func (o *GRPCErrorDetailBadRequestOption) Clone() *GRPCErrorDetailBadRequestOption {
	if o == nil {
		return nil
	}
	return &GRPCErrorDetailBadRequestOption{
		Idx:            o.Idx,
		FieldViolation: o.FieldViolation,
	}
}

func (o *GRPCErrorDetailLocalizedMessageOption) Clone() *GRPCErrorDetailLocalizedMessageOption {
	if o == nil {
		return nil
	}
	return &GRPCErrorDetailLocalizedMessageOption{
		Idx:       o.Idx,
		FieldName: o.FieldName,
	}
}

func (o *SwitchExprOption) Clone() *SwitchExprOption {
	if o == nil {
		return nil
	}
	return &SwitchExprOption{
		Case:    o.Case.Clone(),
		Default: o.Default.Clone(),
	}
}

func (o *SwitchCaseExprOption) Clone() *SwitchCaseExprOption {
	if o == nil {
		return nil
	}
	return &SwitchCaseExprOption{
		Idx: o.Idx,
		If:  o.If,
		By:  o.By,
	}
}

func (o *SwitchDefaultExprOption) Clone() *SwitchDefaultExprOption {
	if o == nil {
		return nil
	}
	return &SwitchDefaultExprOption{
		By: o.By,
	}
}
