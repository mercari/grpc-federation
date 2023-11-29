package source

// Location represents semantic location information for grpc federation option.
type Location struct {
	FileName  string
	GoPackage bool
	Service   *Service
	Message   *Message
	Enum      *Enum
}

// Service represents service location.
type Service struct {
	Name   string
	Method *Method
	Option *ServiceOption
}

// Method represents service's method location.
type Method struct {
	Name     string
	Request  bool
	Response bool
	Option   *MethodOption
}

// ServiceOption represents grpc.federation.service option location.
type ServiceOption struct {
	Dependencies *ServiceDependencyOption
}

// ServiceDependencyOption represents dependencies option of service option.
type ServiceDependencyOption struct {
	Idx     int
	Name    bool
	Service bool
}

// MethodOption represents grpc.federation.method option location.
type MethodOption struct {
	Timeout bool
}

// Enum represents enum location.
type Enum struct {
	Name   string
	Option *EnumOption
	Value  *EnumValue
}

// EnumOption represents grpc.federation.enum option location.
type EnumOption struct {
	Alias bool
}

// EnumValue represents enum value location.
type EnumValue struct {
	Value  string
	Option *EnumValueOption
}

// EnumValueOption represents grpc.federation.enum_value option location.
type EnumValueOption struct {
	Alias   bool
	Default bool
}

// Message represents message location.
type Message struct {
	Name   string
	Option *MessageOption
	Field  *Field
	Enum   *Enum
	Oneof  *Oneof
}

// Field represents message field location.
type Field struct {
	Name   string
	Option *FieldOption
}

// FieldOption represents grpc.federation.field option location.
type FieldOption struct {
	By    bool
	Alias bool
	Oneof *FieldOneof
}

// FieldOneof represents grpc.federation.field.oneof location.
type FieldOneof struct {
	If                  bool
	Default             bool
	VariableDefinitions *VariableDefinitionOption
	By                  bool
}

type Oneof struct {
	Name   string
	Option *OneofOption
}

type OneofOption struct {
}

// MessageOption represents grpc.federation.message option location.
type MessageOption struct {
	VariableDefinitions *VariableDefinitionOption
	Alias               bool
}

// VariableDefinitionOption represents def location of grpc.federation.message option.
type VariableDefinitionOption struct {
	Idx        int
	Name       bool
	By         bool
	Map        *MapExprOption
	Call       *CallExprOption
	Message    *MessageExprOption
	Validation *ValidationExprOption
}

// MapExprOption represents def.map location of grpc.federation.message option.
type MapExprOption struct {
	Iterator *IteratorOption
	By       bool
	Message  *MessageExprOption
}

// IteratorOption represents def.map.iterator location of grpc.federation.message option.
type IteratorOption struct {
	Name   bool
	Source bool
}

// CallExprOption represents def.call location of grpc.federation.message option.
type CallExprOption struct {
	Method  bool
	Request *RequestOption
	Timeout bool
	Retry   *RetryOption
}

// MessageExprOption represents def.message location of grpc.federation.message option.
type MessageExprOption struct {
	Name bool
	Args *ArgumentOption
}

// RequestOption represents resolver.request location of grpc.federation.message option.
type RequestOption struct {
	Idx   int
	Field bool
	By    bool
}

// RetryOption represents resolver.retry location of grpc.federation.message option.
type RetryOption struct {
	Constant    *RetryConstantOption
	Exponential *RetryExponentialOption
}

// RetryConstantOption represents resolver.retry.constant location of grpc.federation.message option.
type RetryConstantOption struct {
	Interval   bool
	MaxRetries bool
}

// RetryExponentialOption represents resolver.retry.exponential location of grpc.federation.message option.
type RetryExponentialOption struct {
	InitialInterval     bool
	RandomizationFactor bool
	Multiplier          bool
	MaxInterval         bool
	MaxRetries          bool
}

// ArgumentOption represents message argument location of grpc.federation.message option.
type ArgumentOption struct {
	Idx    int
	Name   bool
	By     bool
	Inline bool
}

type ValidationExprOption struct {
	Name   bool
	Rule   bool
	Detail *ValidationDetailOption
}

type ValidationDetailOption struct {
	Idx                 int
	Rule                bool
	Message             *ValidationDetailMessageOption
	PreconditionFailure *ValidationDetailPreconditionFailureOption
	BadRequest          *ValidationDetailBadRequestOption
	LocalizedMessage    *ValidationDetailLocalizedMessageOption
}

type ValidationDetailMessageOption struct {
	Idx     int
	Message *MessageExprOption
}

type ValidationDetailPreconditionFailureOption struct {
	Idx       int
	Violation ValidationDetailPreconditionFailureViolationOption
}

type ValidationDetailPreconditionFailureViolationOption struct {
	Idx       int
	FieldName string
}

type ValidationDetailBadRequestOption struct {
	Idx            int
	FieldViolation ValidationDetailBadRequestFieldViolationOption
}

type ValidationDetailBadRequestFieldViolationOption struct {
	Idx       int
	FieldName string
}

type ValidationDetailLocalizedMessageOption struct {
	Idx       int
	FieldName string
}

// Position represents source position in proto file.
type Position struct {
	Line int
	Col  int
}

// FileLocation creates location for file.
func FileLocation(fileName string) *Location {
	return &Location{
		FileName: fileName,
	}
}

// GoPackageLocation creates location for go_package option.
func GoPackageLocation(fileName string) *Location {
	return &Location{
		FileName:  fileName,
		GoPackage: true,
	}
}

// ServiceLocation creates location for service name.
func ServiceLocation(fileName, svcName string) *Location {
	return &Location{
		FileName: fileName,
		Service: &Service{
			Name: svcName,
		},
	}
}

// ServiceOptionLocation creates location for grpc.federation.service option.
func ServiceOptionLocation(fileName, svcName string) *Location {
	return &Location{
		FileName: fileName,
		Service: &Service{
			Name:   svcName,
			Option: &ServiceOption{},
		},
	}
}

// ServiceDependencyLocation creates location for service dependencies.
func ServiceDependencyLocation(fileName, svcName string, idx int) *Location {
	return &Location{
		FileName: fileName,
		Service: &Service{
			Name: svcName,
			Option: &ServiceOption{
				Dependencies: &ServiceDependencyOption{Idx: idx},
			},
		},
	}
}

// ServiceDependencyNameLocation creates location for name of service dependencies.
func ServiceDependencyNameLocation(fileName, svcName string, idx int) *Location {
	return &Location{
		FileName: fileName,
		Service: &Service{
			Name: svcName,
			Option: &ServiceOption{
				Dependencies: &ServiceDependencyOption{Idx: idx, Name: true},
			},
		},
	}
}

// ServiceDependencyServiceLocation creates location for service of service dependencies.
func ServiceDependencyServiceLocation(fileName, svcName string, idx int) *Location {
	return &Location{
		FileName: fileName,
		Service: &Service{
			Name: svcName,
			Option: &ServiceOption{
				Dependencies: &ServiceDependencyOption{Idx: idx, Service: true},
			},
		},
	}
}

// ServiceMethodLocation creates location for method of service.
func ServiceMethodLocation(fileName, svcName, methodName string) *Location {
	return &Location{
		FileName: fileName,
		Service: &Service{
			Name: svcName,
			Method: &Method{
				Name: methodName,
			},
		},
	}
}

// ServiceMethodRequestLocation creates location for method request type of service.
func ServiceMethodRequestLocation(fileName, svcName, methodName string) *Location {
	return &Location{
		FileName: fileName,
		Service: &Service{
			Name: svcName,
			Method: &Method{
				Name:    methodName,
				Request: true,
			},
		},
	}
}

// ServiceMethodResponseLocation creates location for method response type of service.
func ServiceMethodResponseLocation(fileName, svcName, methodName string) *Location {
	return &Location{
		FileName: fileName,
		Service: &Service{
			Name: svcName,
			Method: &Method{
				Name:     methodName,
				Response: true,
			},
		},
	}
}

// ServiceMethodOptionLocation creates location for grpc.federation.method option.
func ServiceMethodOptionLocation(fileName, svcName, methodName string) *Location {
	return &Location{
		FileName: fileName,
		Service: &Service{
			Name: svcName,
			Method: &Method{
				Name:   methodName,
				Option: &MethodOption{},
			},
		},
	}
}

// ServiceMethodTimeoutLocation creates location for timeout of grpc.federation.method option.
func ServiceMethodTimeoutLocation(fileName, svcName, methodName string) *Location {
	return &Location{
		FileName: fileName,
		Service: &Service{
			Name: svcName,
			Method: &Method{
				Name: methodName,
				Option: &MethodOption{
					Timeout: true,
				},
			},
		},
	}
}

// MessageLocation creates location for message name.
func MessageLocation(fileName, msgName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
		},
	}
}

// EnumLocation creates location for enum.
func EnumLocation(fileName, msgName, enumName string) *Location {
	return enumLocation(fileName, msgName, &Enum{
		Name: enumName,
	})
}

// EnumOptionLocation creates location for grpc.federation.enum option.
func EnumOptionLocation(fileName, msgName, enumName string) *Location {
	return enumLocation(fileName, msgName, &Enum{
		Name:   enumName,
		Option: &EnumOption{},
	})
}

// EnumAliasLocation creates location for alias in grpc.federation.enum option.
func EnumAliasLocation(fileName, msgName, enumName string) *Location {
	return enumLocation(fileName, msgName, &Enum{
		Name: enumName,
		Option: &EnumOption{
			Alias: true,
		},
	})
}

// EnumValueLocation creates location for enum value.
func EnumValueLocation(fileName, msgName, enumName, enumValueName string) *Location {
	return enumLocation(fileName, msgName, &Enum{
		Name: enumName,
		Value: &EnumValue{
			Value: enumValueName,
		},
	})
}

// EnumValueOptionLocation creates location for grpc.federation.enum_value option.
func EnumValueOptionLocation(fileName, msgName, enumName, enumValueName string) *Location {
	return enumLocation(fileName, msgName, &Enum{
		Name: enumName,
		Value: &EnumValue{
			Value:  enumValueName,
			Option: &EnumValueOption{},
		},
	})
}

// EnumValueAliasLocation creates location for alias in grpc.federation.enum_value option.
func EnumValueAliasLocation(fileName, msgName, enumName, enumValueName string) *Location {
	return enumLocation(fileName, msgName, &Enum{
		Name: enumName,
		Value: &EnumValue{
			Value: enumValueName,
			Option: &EnumValueOption{
				Alias: true,
			},
		},
	})
}

func enumLocation(fileName, msgName string, enum *Enum) *Location {
	if msgName == "" {
		return &Location{FileName: fileName, Enum: enum}
	}
	return &Location{FileName: fileName, Message: &Message{Name: msgName, Enum: enum}}
}

// MessageFieldLocation creates location for message field.
func MessageFieldLocation(fileName, msgName, fieldName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Field: &Field{
				Name: fieldName,
			},
		},
	}
}

// MessageFieldOptionLocation creates location for grpc.federation.field option.
func MessageFieldOptionLocation(fileName, msgName, fieldName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Field: &Field{
				Name:   fieldName,
				Option: &FieldOption{},
			},
		},
	}
}

// MessageFieldByLocation creates location for by in grpc.federation.field option.
func MessageFieldByLocation(fileName, msgName, fieldName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Field: &Field{
				Name:   fieldName,
				Option: &FieldOption{By: true},
			},
		},
	}
}

// MessageFieldAliasLocation creates location for alias in grpc.federation.field option.
func MessageFieldAliasLocation(fileName, msgName, fieldName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Field: &Field{
				Name: fieldName,
				Option: &FieldOption{
					Alias: true,
				},
			},
		},
	}
}

// MessageFieldOneofLocation creates location for oneof in grpc.federation.field option.
func MessageFieldOneofLocation(fileName, msgName, fieldName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Field: &Field{
				Name:   fieldName,
				Option: &FieldOption{Oneof: &FieldOneof{}},
			},
		},
	}
}

// MessageFieldOneofIfLocation creates location for if in grpc.federation.field.oneof option.
func MessageFieldOneofIfLocation(fileName, msgName, fieldName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Field: &Field{
				Name: fieldName,
				Option: &FieldOption{
					Oneof: &FieldOneof{
						If: true,
					},
				},
			},
		},
	}
}

// MessageFieldOneofDefaultLocation creates location for default in grpc.federation.field.oneof option.
func MessageFieldOneofDefaultLocation(fileName, msgName, fieldName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Field: &Field{
				Name: fieldName,
				Option: &FieldOption{
					Oneof: &FieldOneof{
						Default: true,
					},
				},
			},
		},
	}
}

// MessageFieldOneofByLocation creates location for by in grpc.federation.field.oneof option.
func MessageFieldOneofByLocation(fileName, msgName, fieldName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Field: &Field{
				Name: fieldName,
				Option: &FieldOption{
					Oneof: &FieldOneof{
						By: true,
					},
				},
			},
		},
	}
}

// MessageFieldOneofDefLocation creates location for def in grpc.federation.field.oneof option.
func MessageFieldOneofDefLocation(fileName, msgName, fieldName string, idx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Field: &Field{
				Name: fieldName,
				Option: &FieldOption{
					Oneof: &FieldOneof{
						VariableDefinitions: &VariableDefinitionOption{
							Idx: idx,
						},
					},
				},
			},
		},
	}
}

// MessageFieldOneofDefMessageLocation creates location for def[*].message in grpc.federation.field.oneof.
func MessageFieldOneofDefMessageLocation(fileName, msgName, fieldName string, idx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Field: &Field{
				Name: fieldName,
				Option: &FieldOption{
					Oneof: &FieldOneof{
						VariableDefinitions: &VariableDefinitionOption{
							Idx:     idx,
							Message: &MessageExprOption{},
						},
					},
				},
			},
		},
	}
}

// MessageFieldOneofDefMessageArgumentLocation creates location for def[*].message.args[*] in grpc.federation.field.oneof.
func MessageFieldOneofDefMessageArgumentLocation(fileName, msgName, fieldName string, idx, argIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Field: &Field{
				Name: fieldName,
				Option: &FieldOption{
					Oneof: &FieldOneof{
						VariableDefinitions: &VariableDefinitionOption{
							Idx: idx,
							Message: &MessageExprOption{
								Args: &ArgumentOption{
									Idx: argIdx,
								},
							},
						},
					},
				},
			},
		},
	}
}

// MessageFieldOneofDefMessageArgumentNameLocation creates location for def[*].message.args[].name in grpc.federation.field.oneof.
func MessageFieldOneofDefMessageArgumentNameLocation(fileName, msgName, fieldName string, idx, argIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Field: &Field{
				Option: &FieldOption{
					Oneof: &FieldOneof{
						VariableDefinitions: &VariableDefinitionOption{
							Idx: idx,
							Message: &MessageExprOption{
								Args: &ArgumentOption{
									Idx:  argIdx,
									Name: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// MessageFieldOneofDefMessageArgumentByLocation creates location for def[*].message.args[].by in grpc.federation.field.oneof.
func MessageFieldOneofDefMessageArgumentByLocation(fileName, msgName, fieldName string, idx, argIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Field: &Field{
				Name: fieldName,
				Option: &FieldOption{
					Oneof: &FieldOneof{
						VariableDefinitions: &VariableDefinitionOption{
							Idx: idx,
							Message: &MessageExprOption{
								Args: &ArgumentOption{
									Idx: argIdx,
									By:  true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// MessageFieldOneofDefMessageArgumentInlineLocation creates location for def[*].message.args[].inline in grpc.federation.field.oneof.
func MessageFieldOneofDefMessageArgumentInlineLocation(fileName, msgName, fieldName string, idx, argIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Field: &Field{
				Name: fieldName,
				Option: &FieldOption{
					Oneof: &FieldOneof{
						VariableDefinitions: &VariableDefinitionOption{
							Idx: idx,
							Message: &MessageExprOption{
								Args: &ArgumentOption{
									Idx:    argIdx,
									Inline: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// MessageOptionLocation creates location for grpc.federaiton.message option.
func MessageOptionLocation(fileName, msgName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name:   msgName,
			Option: &MessageOption{},
		},
	}
}

// OneofOptionLocation creates location for grpc.federaiton.oneof option.
func OneofOptionLocation(fileName, msgName, oneofName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Oneof: &Oneof{
				Name: oneofName,
			},
		},
	}
}

// MessageAliasLocation creates location for alias in grpc.federaiton.message option.
func MessageAliasLocation(fileName, msgName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				Alias: true,
			},
		},
	}
}

// MethodLocation creates location for def[*].call.method in grpc.federation.message.
func MethodLocation(fileName, msgName string, idx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: idx,
					Call: &CallExprOption{
						Method: true,
					},
				},
			},
		},
	}
}

// MethodTimeoutLocation creates location for def[*].call.timeout in grpc.federation.message.
func MethodTimeoutLocation(fileName, msgName string, idx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: idx,
					Call: &CallExprOption{
						Timeout: true,
					},
				},
			},
		},
	}
}

// MethodRetryConstantIntervalLocation creates location for def[*].call.retry.constant.interval in grpc.federation.message.
func MethodRetryConstantIntervalLocation(fileName, msgName string, idx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: idx,
					Call: &CallExprOption{
						Retry: &RetryOption{
							Constant: &RetryConstantOption{
								Interval: true,
							},
						},
					},
				},
			},
		},
	}
}

// MethodRetryExponentialInitialIntervalLocation creates location for def[*].call.retry.exponential.initial_interval in grpc.federation.message.
func MethodRetryExponentialInitialIntervalLocation(fileName, msgName string, idx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: idx,
					Call: &CallExprOption{
						Retry: &RetryOption{
							Exponential: &RetryExponentialOption{
								InitialInterval: true,
							},
						},
					},
				},
			},
		},
	}
}

// MethodRetryExponentialMaxIntervalLocation creates location for def[*].call.retry.exponential.max_interval in grpc.federation.message.
func MethodRetryExponentialMaxIntervalLocation(fileName, msgName string, idx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: idx,
					Call: &CallExprOption{
						Retry: &RetryOption{
							Exponential: &RetryExponentialOption{
								MaxInterval: true,
							},
						},
					},
				},
			},
		},
	}
}

// RequestFieldLocation creates location for def[*].call.request[*].field in grpc.federation.message.
func RequestFieldLocation(fileName, msgName string, defIdx, reqIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: defIdx,
					Call: &CallExprOption{
						Request: &RequestOption{
							Idx:   reqIdx,
							Field: true,
						},
					},
				},
			},
		},
	}
}

// RequestByLocation creates location for def[*].call.request[*].by in grpc.federation.message.
func RequestByLocation(fileName, msgName string, defIdx, reqIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: defIdx,
					Call: &CallExprOption{
						Request: &RequestOption{
							Idx: reqIdx,
							By:  true,
						},
					},
				},
			},
		},
	}
}

// MessageExprLocation creates location for def[*].message in grpc.federation.message.
func MessageExprLocation(fileName, msgName string, defIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx:     defIdx,
					Message: &MessageExprOption{},
				},
			},
		},
	}
}

// MessageExprNameLocation creates location for def[*].message.name in grpc.federation.message.
func MessageExprNameLocation(fileName, msgName string, idx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: idx,
					Message: &MessageExprOption{
						Name: true,
					},
				},
			},
		},
	}
}

// MessageExprArgumentLocation creates location for def[*].message.args[*] in grpc.federation.message.
func MessageExprArgumentLocation(fileName, msgName string, idx, argIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: idx,
					Message: &MessageExprOption{
						Args: &ArgumentOption{
							Idx: argIdx,
						},
					},
				},
			},
		},
	}
}

// MessageExprArgumentNameLocation creates location for def[*].message.args[*].name in grpc.federation.message.
func MessageExprArgumentNameLocation(fileName, msgName string, idx, argIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: idx,
					Message: &MessageExprOption{
						Args: &ArgumentOption{
							Idx:  argIdx,
							Name: true,
						},
					},
				},
			},
		},
	}
}

// MessageExprArgumentByLocation creates location for messages[*].args[*].by in grpc.federation.message.
func MessageExprArgumentByLocation(fileName, msgName string, idx, argIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: idx,
					Message: &MessageExprOption{
						Args: &ArgumentOption{
							Idx: argIdx,
							By:  true,
						},
					},
				},
			},
		},
	}
}

// MessageExprArgumentInlineLocation creates location for def[*].message.args[*].inline in grpc.federation.message.
func MessageExprArgumentInlineLocation(fileName, msgName string, idx, argIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: idx,
					Message: &MessageExprOption{
						Args: &ArgumentOption{
							Idx:    argIdx,
							Inline: true,
						},
					},
				},
			},
		},
	}
}

// ValidationLocation creates location for def[*].validation in grpc.federation.message.
func ValidationLocation(fileName, msgName string, idx int, rule bool) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: idx,
					Validation: &ValidationExprOption{
						Rule: rule,
					},
				},
			},
		},
	}
}

// ValidationDetailLocation creates location for def[*].validation.error.details[*] in grpc.federation.message.
func ValidationDetailLocation(fileName, msgName string, vIdx, dIdx int, rule bool) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: vIdx,
					Validation: &ValidationExprOption{
						Detail: &ValidationDetailOption{
							Idx:  dIdx,
							Rule: rule,
						},
					},
				},
			},
		},
	}
}

// ValidationDetailPreconditionFailureLocation creates location for def[*].validation.error.details[*].precondition_failure[*] in grpc.federation.message.
func ValidationDetailPreconditionFailureLocation(fileName, msgName string, vIdx, dIdx, fIdx, fvIdx int, fieldName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: vIdx,
					Validation: &ValidationExprOption{
						Detail: &ValidationDetailOption{
							Idx: dIdx,
							PreconditionFailure: &ValidationDetailPreconditionFailureOption{
								Idx: fIdx,
								Violation: ValidationDetailPreconditionFailureViolationOption{
									Idx:       fvIdx,
									FieldName: fieldName,
								},
							},
						},
					},
				},
			},
		},
	}
}

// ValidationDetailBadRequestLocation creates location for def[*].validation.error.details[*].bad_request[*] in grpc.federation.message.
func ValidationDetailBadRequestLocation(fileName, msgName string, vIdx, dIdx, bIdx, fvIdx int, fieldName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: vIdx,
					Validation: &ValidationExprOption{
						Detail: &ValidationDetailOption{
							Idx: dIdx,
							BadRequest: &ValidationDetailBadRequestOption{
								Idx: bIdx,
								FieldViolation: ValidationDetailBadRequestFieldViolationOption{
									Idx:       fvIdx,
									FieldName: fieldName,
								},
							},
						},
					},
				},
			},
		},
	}
}

// ValidationDetailLocalizedMessageLocation creates location for def[*].validation.error.details[*].localized_message[*] in grpc.federation.message.
func ValidationDetailLocalizedMessageLocation(fileName, msgName string, vIdx, dIdx, lIdx int, fieldName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: vIdx,
					Validation: &ValidationExprOption{
						Detail: &ValidationDetailOption{
							Idx: dIdx,
							LocalizedMessage: &ValidationDetailLocalizedMessageOption{
								Idx:       lIdx,
								FieldName: fieldName,
							},
						},
					},
				},
			},
		},
	}
}

// MapIteratorNameLocation creates location for def.map.iterator.name in grpc.federation.message.
func MapIteratorNameLocation(fileName, msgName string, defIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: defIdx,
					Map: &MapExprOption{
						Iterator: &IteratorOption{
							Name: true,
						},
					},
				},
			},
		},
	}
}

// MapIteratorSourceLocation creates location for def.map.iterator.src in grpc.federation.message.
func MapIteratorSourceLocation(fileName, msgName string, defIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: defIdx,
					Map: &MapExprOption{
						Iterator: &IteratorOption{
							Source: true,
						},
					},
				},
			},
		},
	}
}

// MapExprByLocation creates location for def.map.by in grpc.federation.message.
func MapExprByLocation(fileName, msgName string, defIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: defIdx,
					Map: &MapExprOption{
						By: true,
					},
				},
			},
		},
	}
}

// MapMessageExprArgumentInlineLocation creates location for def.map.message.args[*].inline in grpc.federation.message.
func MapMessageExprArgumentInlineLocation(fileName, msgName string, defIdx, argIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: defIdx,
					Map: &MapExprOption{
						Message: &MessageExprOption{
							Args: &ArgumentOption{
								Idx:    argIdx,
								Inline: true,
							},
						},
					},
				},
			},
		},
	}
}

// MapMessageExprArgumentByLocation creates location for def.map.message.args[*].by in grpc.federation.message.
func MapMessageExprArgumentByLocation(fileName, msgName string, defIdx, argIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: defIdx,
					Map: &MapExprOption{
						Message: &MessageExprOption{
							Args: &ArgumentOption{
								Idx: argIdx,
								By:  true,
							},
						},
					},
				},
			},
		},
	}
}

// CallExprMethodLocation creates location for def.call.method in grpc.federation.message.
func CallExprMethodLocation(fileName, msgName string, defIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: defIdx,
					Call: &CallExprOption{
						Method: true,
					},
				},
			},
		},
	}
}

// CallExprRequestByLocation creates location for def.call.request.by in grpc.federation.message.
func CallExprRequestByLocation(fileName, msgName string, defIdx, reqIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: defIdx,
					Call: &CallExprOption{
						Request: &RequestOption{
							Idx: reqIdx,
							By:  true,
						},
					},
				},
			},
		},
	}
}

// VariableDefinitionLocation creates location for def in grpc.federation.message.
func VariableDefinitionLocation(fileName, msgName string, defIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: defIdx,
				},
			},
		},
	}
}

// VariableDefinitionByLocation creates location for def.by in grpc.federation.message.
func VariableDefinitionByLocation(fileName, msgName string, defIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: defIdx,
					By:  true,
				},
			},
		},
	}
}

// VariableDefinitionValidationDetailMessageLocation creates location for def.validation[*].error.details[*].message[*] in grpc.federation.message.
func VariableDefinitionValidationDetailMessageLocation(fileName, msgName string, defIdx, detIdx, msgIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: defIdx,
					Validation: &ValidationExprOption{
						Detail: &ValidationDetailOption{
							Idx: detIdx,
							Message: &ValidationDetailMessageOption{
								Idx: msgIdx,
							},
						},
					},
				},
			},
		},
	}
}

// VariableDefinitionValidationDetailMessageNameLocation creates location for def.validation[*].error.details[*].message[*] in grpc.federation.message.
func VariableDefinitionValidationDetailMessageNameLocation(fileName, msgName string, defIdx, detIdx, msgIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: defIdx,
					Validation: &ValidationExprOption{
						Detail: &ValidationDetailOption{
							Idx: detIdx,
							Message: &ValidationDetailMessageOption{
								Idx: msgIdx,
								Message: &MessageExprOption{
									Name: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// VariableDefinitionValidationDetailMessageArgumentLocation creates location for def.validation[*].error.details[*].message[*].args[*] in grpc.federation.message.
func VariableDefinitionValidationDetailMessageArgumentLocation(fileName, msgName string, defIdx, detIdx, msgIdx, argIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: defIdx,
					Validation: &ValidationExprOption{
						Detail: &ValidationDetailOption{
							Idx: detIdx,
							Message: &ValidationDetailMessageOption{
								Idx: msgIdx,
								Message: &MessageExprOption{
									Args: &ArgumentOption{
										Idx: argIdx,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// VariableDefinitionValidationDetailMessageArgumentByLocation creates location for def.validation[*].error.details[*].message[*].args[*].by in grpc.federation.message.
func VariableDefinitionValidationDetailMessageArgumentByLocation(fileName, msgName string, defIdx, detIdx, msgIdx, argIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: defIdx,
					Validation: &ValidationExprOption{
						Detail: &ValidationDetailOption{
							Idx: detIdx,
							Message: &ValidationDetailMessageOption{
								Idx: msgIdx,
								Message: &MessageExprOption{
									Args: &ArgumentOption{
										Idx: argIdx,
										By:  true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// VariableDefinitionValidationDetailMessageArgumentInlineLocation creates location for def.validation[*].error.details[*].message[*].args[*].inline in grpc.federation.message.
func VariableDefinitionValidationDetailMessageArgumentInlineLocation(fileName, msgName string, defIdx, detIdx, msgIdx, argIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				VariableDefinitions: &VariableDefinitionOption{
					Idx: defIdx,
					Validation: &ValidationExprOption{
						Detail: &ValidationDetailOption{
							Idx: detIdx,
							Message: &ValidationDetailMessageOption{
								Idx: msgIdx,
								Message: &MessageExprOption{
									Args: &ArgumentOption{
										Idx:    argIdx,
										Inline: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
