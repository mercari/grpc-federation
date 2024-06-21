package resolver

var (
	AnyType              *Type
	TimestampType        *Type
	DurationType         *Type
	DurationRepeatedType *Type
	EmptyType            *Type
	Int64ValueType       *Type
	Int32ValueType       *Type
	Uint64ValueType      *Type
	Uint32ValueType      *Type
	DoubleValueType      *Type
	FloatValueType       *Type
	BytesValueType       *Type
	BoolValueType        *Type
	StringValueType      *Type
	WrapperNumberTypeMap map[string]struct{}
)

func init() {
	files, err := New(nil).ResolveWellknownFiles()
	if err != nil {
		panic(err)
	}
	for _, file := range files.FindByPackageName("google.protobuf") {
		if msg := file.Message("Any"); msg != nil {
			AnyType = NewMessageType(msg, false)
		}
		if msg := file.Message("Timestamp"); msg != nil {
			TimestampType = NewMessageType(msg, false)
		}
		if msg := file.Message("Duration"); msg != nil {
			DurationType = NewMessageType(msg, false)
			DurationRepeatedType = NewMessageType(msg, true)
		}
		if msg := file.Message("Empty"); msg != nil {
			EmptyType = NewMessageType(msg, false)
		}
		if msg := file.Message("Int64Value"); msg != nil {
			Int64ValueType = NewMessageType(msg, false)
		}
		if msg := file.Message("UInt64Value"); msg != nil {
			Uint64ValueType = NewMessageType(msg, false)
		}
		if msg := file.Message("Int32Value"); msg != nil {
			Int32ValueType = NewMessageType(msg, false)
		}
		if msg := file.Message("UInt32Value"); msg != nil {
			Uint32ValueType = NewMessageType(msg, false)
		}
		if msg := file.Message("BoolValue"); msg != nil {
			BoolValueType = NewMessageType(msg, false)
		}
		if msg := file.Message("BytesValue"); msg != nil {
			BytesValueType = NewMessageType(msg, false)
		}
		if msg := file.Message("FloatValue"); msg != nil {
			FloatValueType = NewMessageType(msg, false)
		}
		if msg := file.Message("DoubleValue"); msg != nil {
			DoubleValueType = NewMessageType(msg, false)
		}
		if msg := file.Message("StringValue"); msg != nil {
			StringValueType = NewMessageType(msg, false)
		}
	}
	WrapperNumberTypeMap = map[string]struct{}{
		Int64ValueType.FQDN():  {},
		Uint64ValueType.FQDN(): {},
		Int32ValueType.FQDN():  {},
		Uint32ValueType.FQDN(): {},
		DoubleValueType.FQDN(): {},
		FloatValueType.FQDN():  {},
	}
}
