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
}

// MessageOption represents grpc.federation.message option location.
type MessageOption struct {
	Resolver *ResolverOption
	Messages *MessageDependencyOption
	Alias    bool
}

// ResolverOption represents resolver location of grpc.federation.message option.
type ResolverOption struct {
	Method   bool
	Request  *RequestOption
	Response *ResponseOption
	Timeout  bool
	Retry    *RetryOption
}

// RequestOption represents resolver.request location of grpc.federation.message option.
type RequestOption struct {
	Idx   int
	Field bool
	By    bool
}

// ResponseOption represents resolver.response location of grpc.federation.message option.
type ResponseOption struct {
	Idx      int
	Name     bool
	Field    bool
	AutoBind bool
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

// MessageDependencyOption represents messages location of grpc.federation.message option.
type MessageDependencyOption struct {
	Idx     int
	Name    bool
	Message bool
	Args    *ArgumentOption
}

// ArgumentOption represents message argument location of grpc.federation.message option.
type ArgumentOption struct {
	Idx    int
	Name   bool
	By     bool
	Inline bool
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

// MessageAliasOptionLocation creates location for alias in grpc.federaiton.message option.
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

// MethodLocation creates location for resolver.method in grpc.federation.message.
func MethodLocation(fileName, msgName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				Resolver: &ResolverOption{
					Method: true,
				},
			},
		},
	}
}

// MethodTimeoutLocation creates location for resolver.timeout in grpc.federation.message.
func MethodTimeoutLocation(fileName, msgName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				Resolver: &ResolverOption{
					Timeout: true,
				},
			},
		},
	}
}

// MethodRetryConstantIntervalLocation creates location for resolver.retry.constant.interval in grpc.federation.message.
func MethodRetryConstantIntervalLocation(fileName, msgName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				Resolver: &ResolverOption{
					Retry: &RetryOption{
						Constant: &RetryConstantOption{
							Interval: true,
						},
					},
				},
			},
		},
	}
}

// MethodRetryExponentialInitialIntervalLocation creates location for resolver.retry.exponential.initial_interval in grpc.federation.message.
func MethodRetryExponentialInitialIntervalLocation(fileName, msgName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				Resolver: &ResolverOption{
					Retry: &RetryOption{
						Exponential: &RetryExponentialOption{
							InitialInterval: true,
						},
					},
				},
			},
		},
	}
}

// MethodRetryExponentialMaxIntervalLocation creates location for resolver.retry.exponential.max_interval in grpc.federation.message.
func MethodRetryExponentialMaxIntervalLocation(fileName, msgName string) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				Resolver: &ResolverOption{
					Retry: &RetryOption{
						Exponential: &RetryExponentialOption{
							MaxInterval: true,
						},
					},
				},
			},
		},
	}
}

// RequestFieldLocation creates location for resolver.request[*].field in grpc.federation.message.
func RequestFieldLocation(fileName, msgName string, idx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				Resolver: &ResolverOption{
					Request: &RequestOption{
						Idx:   idx,
						Field: true,
					},
				},
			},
		},
	}
}

// RequestByLocation creates location for resolver.request[*].by in grpc.federation.message.
func RequestByLocation(fileName, msgName string, idx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				Resolver: &ResolverOption{
					Request: &RequestOption{
						Idx: idx,
						By:  true,
					},
				},
			},
		},
	}
}

// ResponseFieldLocation creates location for resolver.response[*].field in grpc.federation.message.
func ResponseFieldLocation(fileName, msgName string, idx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				Resolver: &ResolverOption{
					Response: &ResponseOption{
						Idx:   idx,
						Field: true,
					},
				},
			},
		},
	}
}

// MessageDependencyMessageLocation creates location for messages[*].message in grpc.federation.message.
func MessageDependencyMessageLocation(fileName, msgName string, idx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				Messages: &MessageDependencyOption{
					Idx:     idx,
					Message: true,
				},
			},
		},
	}
}

// MessageDependencyArgumentNameLocation creates location for messages[*].args[*].name in grpc.federation.message.
func MessageDependencyArgumentNameLocation(fileName, msgName string, idx, argIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				Messages: &MessageDependencyOption{
					Idx: idx,
					Args: &ArgumentOption{
						Idx:  argIdx,
						Name: true,
					},
				},
			},
		},
	}
}

// MessageDependencyArgumentByLocation creates location for messages[*].args[*].by in grpc.federation.message.
func MessageDependencyArgumentByLocation(fileName, msgName string, idx, argIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				Messages: &MessageDependencyOption{
					Idx: idx,
					Args: &ArgumentOption{
						Idx: argIdx,
						By:  true,
					},
				},
			},
		},
	}
}

// MessageDependencyArgumentInlineLocation creates location for messages[*].args[*].inline in grpc.federation.message.
func MessageDependencyArgumentInlineLocation(fileName, msgName string, idx, argIdx int) *Location {
	return &Location{
		FileName: fileName,
		Message: &Message{
			Name: msgName,
			Option: &MessageOption{
				Messages: &MessageDependencyOption{
					Idx: idx,
					Args: &ArgumentOption{
						Idx:    argIdx,
						Inline: true,
					},
				},
			},
		},
	}
}
