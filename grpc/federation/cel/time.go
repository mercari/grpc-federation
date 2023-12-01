package cel

import (
	"reflect"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

var (
	LocationType = types.NewObjectType(createTimeName("Location"))
)

type Location struct {
	*time.Location
}

func (loc *Location) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return loc.Location, nil
}

func (loc *Location) ConvertToType(typeValue ref.Type) ref.Val {
	return types.NewErr("grpc.federation.time: location type conversion does not support")
}

func (loc *Location) Equal(other ref.Val) ref.Val {
	if o, ok := other.(*Location); ok {
		return types.Bool(loc.String() == o.String())
	}
	return types.False
}

func (loc *Location) Type() ref.Type {
	return LocationType
}

func (loc *Location) Value() any {
	return loc.Location
}

const TimePackageName = "time"

type TimeLibrary struct {
}

func (lib *TimeLibrary) LibraryName() string {
	return packageName(TimePackageName)
}

func createTimeName(name string) string {
	return createName(TimePackageName, name)
}

func createTimeID(name string) string {
	return createID(TimePackageName, name)
}

func (lib *TimeLibrary) CompileOptions() []cel.EnvOption {
	opts := []cel.EnvOption{
		cel.Constant(
			createTimeName("RFC3339"),
			types.StringType,
			types.String(time.RFC3339),
		),
		cel.Constant(
			createTimeName("UTC"),
			LocationType,
			&Location{Location: time.UTC},
		),
		cel.Function(
			createTimeName("loadLocation"),
			cel.Overload(createTimeID("load_location_string_location"), []*cel.Type{cel.StringType}, LocationType,
				cel.UnaryBinding(func(name ref.Val) ref.Val {
					loc, err := time.LoadLocation(string(name.(types.String)))
					if err != nil {
						return types.NewErr(err.Error())
					}
					return &Location{loc}
				}),
			),
		),
		cel.Function(
			"string",
			cel.MemberOverload(createTimeID("string_location_string"), []*cel.Type{LocationType}, cel.StringType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.String(self.(*Location).String())
				}),
			),
		),
		cel.Function(
			createTimeName("date"),
			cel.Overload(createTimeID("date_int_int_int_int_int_int_int_location_timestamp"), []*cel.Type{cel.IntType, cel.IntType, cel.IntType, cel.IntType, cel.IntType, cel.IntType, cel.IntType, LocationType}, cel.TimestampType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Timestamp{
						Time: time.Date(
							int(values[0].(types.Int)),
							time.Month(values[1].(types.Int)),
							int(values[2].(types.Int)),
							int(values[3].(types.Int)),
							int(values[4].(types.Int)),
							int(values[5].(types.Int)),
							int(values[6].(types.Int)),
							values[7].(*Location).Location,
						),
					}
				}),
			),
		),
		cel.Function(
			createTimeName("now"),
			cel.Overload(createTimeID("now_timestamp"), nil, cel.TimestampType,
				cel.FunctionBinding(func(_ ...ref.Val) ref.Val {
					return types.Timestamp{Time: time.Now()}
				}),
			),
		),
		cel.Function(
			createTimeName("parse"),
			cel.Overload(createTimeID("parse_string_string_timestamp"), []*cel.Type{cel.StringType, cel.StringType}, cel.TimestampType,
				cel.BinaryBinding(func(layout, value ref.Val) ref.Val {
					t, err := time.Parse(string(layout.(types.String)), string(value.(types.String)))
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Timestamp{Time: t}
				}),
			),
		),
		cel.Function(
			createTimeName("fixedZone"),
			cel.Overload(createTimeID("fixed_zone_string_int_location"), []*cel.Type{cel.StringType, cel.IntType}, LocationType,
				cel.BinaryBinding(func(name, offset ref.Val) ref.Val {
					loc := time.FixedZone(string(name.(types.String)), int(offset.(types.Int)))
					return &Location{loc}
				}),
			),
		),
		cel.Function(
			"addDate",
			cel.MemberOverload(createTimeID("add_timestamp_int_int_int_timestamp"), []*cel.Type{cel.TimestampType, cel.IntType, cel.IntType, cel.IntType}, cel.TimestampType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					self := values[0].(types.Timestamp)
					return types.Timestamp{
						Time: self.AddDate(int(values[1].(types.Int)), int(values[2].(types.Int)), int(values[3].(types.Int))),
					}
				}),
			),
		),
		cel.Function(
			"utc",
			cel.MemberOverload(createTimeID("utc_timestamp_timestamp"), []*cel.Type{cel.TimestampType}, cel.TimestampType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Timestamp{Time: self.(types.Timestamp).Time.UTC()}
				}),
			),
		),
		cel.Function(
			"year",
			cel.MemberOverload(createTimeID("year_timestamp_int"), []*cel.Type{cel.TimestampType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Year())
				}),
			),
		),
		cel.Function(
			"month",
			cel.MemberOverload(createTimeID("month_timestamp_int"), []*cel.Type{cel.TimestampType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Month())
				}),
			),
		),
	}
	return opts
}

func (lib *TimeLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
