package resolver

var (
	AnyType       *Type
	TimestampType *Type
	DurationType  *Type
	EmptyType     *Type
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
		}
		if msg := file.Message("Empty"); msg != nil {
			EmptyType = NewMessageType(msg, false)
		}
	}
}
