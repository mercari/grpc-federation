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
	return loc, nil
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
	return loc
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
		// Constant variables
		cel.Constant(
			createTimeName("LAYOUT"),
			types.StringType,
			types.String(time.Layout),
		),
		cel.Constant(
			createTimeName("ANSIC"),
			types.StringType,
			types.String(time.ANSIC),
		),
		cel.Constant(
			createTimeName("UNIX_DATE"),
			types.StringType,
			types.String(time.UnixDate),
		),
		cel.Constant(
			createTimeName("RUBY_DATE"),
			types.StringType,
			types.String(time.RubyDate),
		),
		cel.Constant(
			createTimeName("RFC822"),
			types.StringType,
			types.String(time.RFC822),
		),
		cel.Constant(
			createTimeName("RFC822Z"),
			types.StringType,
			types.String(time.RFC822Z),
		),
		cel.Constant(
			createTimeName("RFC850"),
			types.StringType,
			types.String(time.RFC850),
		),
		cel.Constant(
			createTimeName("RFC1123"),
			types.StringType,
			types.String(time.RFC1123),
		),
		cel.Constant(
			createTimeName("RFC1123Z"),
			types.StringType,
			types.String(time.RFC1123Z),
		),
		cel.Constant(
			createTimeName("RFC3339"),
			types.StringType,
			types.String(time.RFC3339),
		),
		cel.Constant(
			createTimeName("RFC3339NANO"),
			types.StringType,
			types.String(time.RFC3339Nano),
		),
		cel.Constant(
			createTimeName("KITCHEN"),
			types.StringType,
			types.String(time.Kitchen),
		),
		cel.Constant(
			createTimeName("STAMP"),
			types.StringType,
			types.String(time.Stamp),
		),
		cel.Constant(
			createTimeName("STAMP_MILLI"),
			types.StringType,
			types.String(time.StampMilli),
		),
		cel.Constant(
			createTimeName("STAMP_MICRO"),
			types.StringType,
			types.String(time.StampMicro),
		),
		cel.Constant(
			createTimeName("STAMP_NANO"),
			types.StringType,
			types.String(time.StampNano),
		),
		cel.Constant(
			createTimeName("DATETIME"),
			types.StringType,
			types.String(time.DateTime),
		),
		cel.Constant(
			createTimeName("DATE_ONLY"),
			types.StringType,
			types.String(time.DateOnly),
		),
		cel.Constant(
			createTimeName("TIME_ONLY"),
			types.StringType,
			types.String(time.TimeOnly),
		),
		cel.Constant(
			createTimeName("NANOSECOND"),
			types.IntType,
			types.Int(time.Nanosecond),
		),
		cel.Constant(
			createTimeName("MICROSECOND"),
			types.IntType,
			types.Int(time.Microsecond),
		),
		cel.Constant(
			createTimeName("MILLISECOND"),
			types.IntType,
			types.Int(time.Millisecond),
		),
		cel.Constant(
			createTimeName("SECOND"),
			types.IntType,
			types.Int(time.Second),
		),
		cel.Constant(
			createTimeName("MINUTE"),
			types.IntType,
			types.Int(time.Minute),
		),
		cel.Constant(
			createTimeName("HOUR"),
			types.IntType,
			types.Int(time.Hour),
		),
		cel.Function(
			createTimeName("LOCAL"),
			cel.Overload(createTimeID("local_location"), []*cel.Type{}, LocationType,
				cel.FunctionBinding(func(_ ...ref.Val) ref.Val {
					return &Location{Location: time.Local}
				}),
			),
		),
		cel.Function(
			createTimeName("UTC"),
			cel.Overload(createTimeID("utc_location"), []*cel.Type{}, LocationType,
				cel.FunctionBinding(func(_ ...ref.Val) ref.Val {
					return &Location{Location: time.UTC}
				}),
			),
		),

		// Duration functions
		cel.Function(
			createTimeName("toDuration"),
			cel.Overload(createTimeID("to_duration_int_duration"), []*cel.Type{cel.IntType}, cel.DurationType,
				cel.UnaryBinding(func(v ref.Val) ref.Val {
					return types.Duration{Duration: time.Duration(v.(types.Int))}
				}),
			),
		),
		cel.Function(
			createTimeName("parseDuration"),
			cel.Overload(createTimeID("parse_duration_string_duration"), []*cel.Type{cel.StringType}, cel.DurationType,
				cel.UnaryBinding(func(s ref.Val) ref.Val {
					d, err := time.ParseDuration(string(s.(types.String)))
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Duration{Duration: d}
				}),
			),
		),
		cel.Function(
			createTimeName("since"),
			cel.Overload(createTimeID("since_timestamp_duration"), []*cel.Type{cel.TimestampType}, cel.DurationType,
				cel.UnaryBinding(func(t ref.Val) ref.Val {
					return types.Duration{
						Duration: time.Since(t.(types.Timestamp).Time),
					}
				}),
			),
		),
		cel.Function(
			createTimeName("until"),
			cel.Overload(createTimeID("until_timestamp_duration"), []*cel.Type{cel.TimestampType}, cel.DurationType,
				cel.UnaryBinding(func(t ref.Val) ref.Val {
					return types.Duration{
						Duration: time.Until(t.(types.Timestamp).Time),
					}
				}),
			),
		),
		cel.Function(
			"abs",
			cel.MemberOverload(createTimeID("abs_duration_duration"), []*cel.Type{cel.DurationType}, cel.DurationType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Duration{
						Duration: self.(types.Duration).Duration.Abs(),
					}
				}),
			),
		),
		cel.Function(
			"hours",
			cel.MemberOverload(createTimeID("hours_duration_double"), []*cel.Type{cel.DurationType}, cel.DoubleType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Double(self.(types.Duration).Duration.Hours())
				}),
			),
		),
		cel.Function(
			"microseconds",
			cel.MemberOverload(createTimeID("microseconds_duration_int"), []*cel.Type{cel.DurationType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(types.Duration).Duration.Microseconds())
				}),
			),
		),
		cel.Function(
			"milliseconds",
			cel.MemberOverload(createTimeID("milliseconds_duration_int"), []*cel.Type{cel.DurationType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(types.Duration).Duration.Milliseconds())
				}),
			),
		),
		cel.Function(
			"minutes",
			cel.MemberOverload(createTimeID("minutes_duration_int"), []*cel.Type{cel.DurationType}, cel.DoubleType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Double(self.(types.Duration).Duration.Minutes())
				}),
			),
		),
		cel.Function(
			"nanoseconds",
			cel.MemberOverload(createTimeID("nanoseconds_duration_int"), []*cel.Type{cel.DurationType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(types.Duration).Duration.Nanoseconds())
				}),
			),
		),
		cel.Function(
			"round",
			cel.MemberOverload(createTimeID("round_duration_int_duration"), []*cel.Type{cel.DurationType, cel.IntType}, cel.DurationType,
				cel.BinaryBinding(func(self, m ref.Val) ref.Val {
					return types.Duration{
						Duration: self.(types.Duration).Duration.Round(time.Duration(m.(types.Int))),
					}
				}),
			),
			cel.MemberOverload(createTimeID("round_duration_duration_duration"), []*cel.Type{cel.DurationType, cel.DurationType}, cel.DurationType,
				cel.BinaryBinding(func(self, m ref.Val) ref.Val {
					return types.Duration{
						Duration: self.(types.Duration).Duration.Round(m.(types.Duration).Duration),
					}
				}),
			),
		),
		cel.Function(
			"seconds",
			cel.MemberOverload(createTimeID("seconds_duration_int"), []*cel.Type{cel.DurationType}, cel.DoubleType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Double(self.(types.Duration).Duration.Seconds())
				}),
			),
		),
		cel.Function(
			"string",
			cel.MemberOverload(createTimeID("string_duration_string"), []*cel.Type{cel.DurationType}, cel.StringType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.String(self.(types.Duration).Duration.String())
				}),
			),
		),
		cel.Function(
			"truncate",
			cel.MemberOverload(createTimeID("truncate_duration_int_duration"), []*cel.Type{cel.DurationType, cel.IntType}, cel.DurationType,
				cel.BinaryBinding(func(self, m ref.Val) ref.Val {
					return types.Duration{
						Duration: self.(types.Duration).Duration.Truncate(time.Duration(m.(types.Int))),
					}
				}),
			),
			cel.MemberOverload(createTimeID("truncate_duration_duration_duration"), []*cel.Type{cel.DurationType, cel.DurationType}, cel.DurationType,
				cel.BinaryBinding(func(self, m ref.Val) ref.Val {
					return types.Duration{
						Duration: self.(types.Duration).Duration.Truncate(m.(types.Duration).Duration),
					}
				}),
			),
		),

		// Location functions
		cel.Function(
			createTimeName("fixedZone"),
			cel.Overload(createTimeID("fixed_zone_string_int_location"), []*cel.Type{cel.StringType, cel.IntType}, LocationType,
				cel.BinaryBinding(func(name, offset ref.Val) ref.Val {
					loc := time.FixedZone(string(name.(types.String)), int(offset.(types.Int)))
					return &Location{loc}
				}),
			),
			cel.Overload(createTimeID("fixed_zone_string_double_location"), []*cel.Type{cel.StringType, cel.DoubleType}, LocationType,
				cel.BinaryBinding(func(name, offset ref.Val) ref.Val {
					loc := time.FixedZone(string(name.(types.String)), int(offset.(types.Double)))
					return &Location{loc}
				}),
			),
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
			createTimeName("loadLocationFromTZData"),
			cel.Overload(createTimeID("load_location_from_tz_data_string_bytes_location"), []*cel.Type{cel.StringType, cel.BytesType}, LocationType,
				cel.BinaryBinding(func(name, data ref.Val) ref.Val {
					loc, err := time.LoadLocationFromTZData(string(name.(types.String)), []byte(data.(types.Bytes)))
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

		// Timestamp functions
		cel.Function(
			createTimeName("date"),
			cel.Overload(
				createTimeID("date_int_int_int_int_int_int_int_location_timestamp"),
				[]*cel.Type{cel.IntType, cel.IntType, cel.IntType, cel.IntType, cel.IntType, cel.IntType, cel.IntType, LocationType},
				cel.TimestampType,
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
			createTimeName("parseInLocation"),
			cel.Overload(createTimeID("parse_in_location_string_string_location_timestamp"), []*cel.Type{cel.StringType, cel.StringType, LocationType}, cel.TimestampType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					t, err := time.ParseInLocation(string(values[0].(types.String)), string(values[1].(types.String)), values[2].(*Location).Location)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Timestamp{Time: t}
				}),
			),
		),
		cel.Function(
			createTimeName("unix"),
			cel.Overload(createTimeID("unix_int_int_timestamp"), []*cel.Type{cel.IntType, cel.IntType}, cel.TimestampType,
				cel.BinaryBinding(func(sec, nsec ref.Val) ref.Val {
					return types.Timestamp{
						Time: time.Unix(int64(sec.(types.Int)), int64(nsec.(types.Int))),
					}
				}),
			),
		),
		cel.Function(
			createTimeName("unixMicro"),
			cel.Overload(createTimeID("unix_micro_int_timestamp"), []*cel.Type{cel.IntType}, cel.TimestampType,
				cel.UnaryBinding(func(usec ref.Val) ref.Val {
					return types.Timestamp{
						Time: time.UnixMicro(int64(usec.(types.Int))),
					}
				}),
			),
		),
		cel.Function(
			createTimeName("unixMilli"),
			cel.Overload(createTimeID("unix_milli_int_timestamp"), []*cel.Type{cel.IntType}, cel.TimestampType,
				cel.UnaryBinding(func(msec ref.Val) ref.Val {
					return types.Timestamp{
						Time: time.UnixMilli(int64(msec.(types.Int))),
					}
				}),
			),
		),
		cel.Function(
			"add",
			cel.MemberOverload(createTimeID("add_timestamp_int_timestamp"), []*cel.Type{cel.TimestampType, cel.IntType}, cel.TimestampType,
				cel.BinaryBinding(func(self, d ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.Add(time.Duration(d.(types.Int))),
					}
				}),
			),
			cel.MemberOverload(createTimeID("add_timestamp_duration_timestamp"), []*cel.Type{cel.TimestampType, cel.DurationType}, cel.TimestampType,
				cel.BinaryBinding(func(self, d ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.Add(d.(types.Duration).Duration),
					}
				}),
			),
		),
		cel.Function(
			"addDate",
			cel.MemberOverload(createTimeID("add_date_timestamp_int_int_int_timestamp"), []*cel.Type{cel.TimestampType, cel.IntType, cel.IntType, cel.IntType}, cel.TimestampType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					self := values[0].(types.Timestamp)
					return types.Timestamp{
						Time: self.AddDate(int(values[1].(types.Int)), int(values[2].(types.Int)), int(values[3].(types.Int))),
					}
				}),
			),
		),
		cel.Function(
			"after",
			cel.MemberOverload(createTimeID("after_timestamp_timestamp_bool"), []*cel.Type{cel.TimestampType, cel.TimestampType}, cel.BoolType,
				cel.BinaryBinding(func(self, u ref.Val) ref.Val {
					return types.Bool(self.(types.Timestamp).Time.After(u.(types.Timestamp).Time))
				}),
			),
		),
		cel.Function(
			"appendFormat",
			cel.MemberOverload(createTimeID("append_format_timestamp_string_string_bytes"), []*cel.Type{cel.TimestampType, cel.StringType, cel.StringType}, cel.BytesType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Bytes(
						values[0].(types.Timestamp).Time.AppendFormat(
							[]byte(values[1].(types.String)),
							string(values[2].(types.String)),
						),
					)
				}),
			),
			cel.MemberOverload(createTimeID("append_format_timestamp_bytes_string_bytes"), []*cel.Type{cel.TimestampType, cel.BytesType, cel.StringType}, cel.BytesType,
				cel.FunctionBinding(func(values ...ref.Val) ref.Val {
					return types.Bytes(
						values[0].(types.Timestamp).Time.AppendFormat(
							[]byte(values[1].(types.Bytes)),
							string(values[2].(types.String)),
						),
					)
				}),
			),
		),
		cel.Function(
			"before",
			cel.MemberOverload(createTimeID("before_timestamp_timestamp_bool"), []*cel.Type{cel.TimestampType, cel.TimestampType}, cel.BoolType,
				cel.BinaryBinding(func(self, u ref.Val) ref.Val {
					return types.Bool(self.(types.Timestamp).Time.Before(u.(types.Timestamp).Time))
				}),
			),
		),
		cel.Function(
			"compare",
			cel.MemberOverload(createTimeID("compare_timestamp_timestamp_int"), []*cel.Type{cel.TimestampType, cel.TimestampType}, cel.IntType,
				cel.BinaryBinding(func(self, u ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Compare(u.(types.Timestamp).Time))
				}),
			),
		),
		cel.Function(
			"day",
			cel.MemberOverload(createTimeID("day_timestamp_int"), []*cel.Type{cel.TimestampType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Day())
				}),
			),
		),
		cel.Function(
			"equal",
			cel.MemberOverload(createTimeID("equal_timestamp_timestamp_bool"), []*cel.Type{cel.TimestampType, cel.TimestampType}, cel.BoolType,
				cel.BinaryBinding(func(self, u ref.Val) ref.Val {
					return types.Bool(self.(types.Timestamp).Time.Equal(u.(types.Timestamp).Time))
				}),
			),
		),
		cel.Function(
			"format",
			cel.MemberOverload(createTimeID("format_timestamp_string_string"), []*cel.Type{cel.TimestampType, cel.StringType}, cel.StringType,
				cel.BinaryBinding(func(self, layout ref.Val) ref.Val {
					return types.String(self.(types.Timestamp).Time.Format(string(layout.(types.String))))
				}),
			),
		),
		cel.Function(
			"hour",
			cel.MemberOverload(createTimeID("hour_timestamp_int"), []*cel.Type{cel.TimestampType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Hour())
				}),
			),
		),
		cel.Function(
			"withLocation",
			cel.MemberOverload(createTimeID("with_location_timestamp_location_timestamp"), []*cel.Type{cel.TimestampType, LocationType}, cel.TimestampType,
				cel.BinaryBinding(func(self, loc ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.In(loc.(*Location).Location),
					}
				}),
			),
		),
		cel.Function(
			"isDST",
			cel.MemberOverload(createTimeID("is_dst_timestamp_bool"), []*cel.Type{cel.TimestampType}, cel.BoolType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Bool(self.(types.Timestamp).Time.IsDST())
				}),
			),
		),
		cel.Function(
			"isZero",
			cel.MemberOverload(createTimeID("is_zero_timestamp_bool"), []*cel.Type{cel.TimestampType}, cel.BoolType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Bool(self.(types.Timestamp).Time.IsZero())
				}),
			),
		),
		cel.Function(
			"local",
			cel.MemberOverload(createTimeID("local_timestamp_timestamp"), []*cel.Type{cel.TimestampType}, cel.TimestampType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.Local(),
					}
				}),
			),
		),
		cel.Function(
			"location",
			cel.MemberOverload(createTimeID("location_timestamp_location"), []*cel.Type{cel.TimestampType}, LocationType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return &Location{Location: self.(types.Timestamp).Time.Location()}
				}),
			),
		),
		cel.Function(
			"minute",
			cel.MemberOverload(createTimeID("minute_timestamp_int"), []*cel.Type{cel.TimestampType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Minute())
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
		cel.Function(
			"nanosecond",
			cel.MemberOverload(createTimeID("nanosecond_timestamp_int"), []*cel.Type{cel.TimestampType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Nanosecond())
				}),
			),
		),
		cel.Function(
			"round",
			cel.MemberOverload(createTimeID("round_timestamp_int_timestamp"), []*cel.Type{cel.TimestampType, cel.IntType}, cel.TimestampType,
				cel.BinaryBinding(func(self, d ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.Round(time.Duration(d.(types.Int))),
					}
				}),
			),
			cel.MemberOverload(createTimeID("round_timestamp_duration_timestamp"), []*cel.Type{cel.TimestampType, cel.DurationType}, cel.TimestampType,
				cel.BinaryBinding(func(self, d ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.Round(d.(types.Duration).Duration),
					}
				}),
			),
		),
		cel.Function(
			"second",
			cel.MemberOverload(createTimeID("second_timestamp_int"), []*cel.Type{cel.TimestampType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Second())
				}),
			),
		),
		cel.Function(
			"string",
			cel.MemberOverload(createTimeID("string_timestamp_string"), []*cel.Type{cel.TimestampType}, cel.StringType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.String(self.(types.Timestamp).Time.String())
				}),
			),
		),
		cel.Function(
			"sub",
			cel.MemberOverload(createTimeID("sub_timestamp_timestamp_duration"), []*cel.Type{cel.TimestampType, cel.TimestampType}, cel.DurationType,
				cel.BinaryBinding(func(self, u ref.Val) ref.Val {
					return types.Duration{
						Duration: self.(types.Timestamp).Time.Sub(u.(types.Timestamp).Time),
					}
				}),
			),
		),
		cel.Function(
			"truncate",
			cel.MemberOverload(createTimeID("truncate_timestamp_int_timestamp"), []*cel.Type{cel.TimestampType, cel.IntType}, cel.TimestampType,
				cel.BinaryBinding(func(self, d ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.Truncate(time.Duration(d.(types.Int))),
					}
				}),
			),
			cel.MemberOverload(createTimeID("truncate_timestamp_duration_timestamp"), []*cel.Type{cel.TimestampType, cel.DurationType}, cel.TimestampType,
				cel.BinaryBinding(func(self, d ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.Truncate(d.(types.Duration).Duration),
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
			"unix",
			cel.MemberOverload(createTimeID("unix_timestamp_int"), []*cel.Type{cel.TimestampType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Unix())
				}),
			),
		),
		cel.Function(
			"unixMicro",
			cel.MemberOverload(createTimeID("unix_micro_timestamp_int"), []*cel.Type{cel.TimestampType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.UnixMicro())
				}),
			),
		),
		cel.Function(
			"unixMilli",
			cel.MemberOverload(createTimeID("unix_milli_timestamp_int"), []*cel.Type{cel.TimestampType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.UnixMilli())
				}),
			),
		),
		cel.Function(
			"unixNano",
			cel.MemberOverload(createTimeID("unix_nano_timestamp_int"), []*cel.Type{cel.TimestampType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.UnixNano())
				}),
			),
		),
		cel.Function(
			"weekday",
			cel.MemberOverload(createTimeID("weekday_timestamp_int"), []*cel.Type{cel.TimestampType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Weekday())
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
			"yearDay",
			cel.MemberOverload(createTimeID("year_day_timestamp_int"), []*cel.Type{cel.TimestampType}, cel.IntType,
				cel.UnaryBinding(func(self ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.YearDay())
				}),
			),
		),
	}
	return opts
}

func (lib *TimeLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
