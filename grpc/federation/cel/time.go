package cel

import (
	"context"
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
	}
	for _, funcOpts := range [][]cel.EnvOption{
		BindFunction(
			createTimeName("LOCAL"),
			OverloadFunc(createTimeID("local_location"), []*cel.Type{}, LocationType,
				func(_ context.Context, _ ...ref.Val) ref.Val {
					return &Location{Location: time.Local}
				},
			),
		),
		BindFunction(
			createTimeName("UTC"),
			OverloadFunc(createTimeID("utc_location"), []*cel.Type{}, LocationType,
				func(_ context.Context, _ ...ref.Val) ref.Val {
					return &Location{Location: time.UTC}
				},
			),
		),

		// Duration functions
		BindFunction(
			createTimeName("toDuration"),
			OverloadFunc(createTimeID("to_duration_int_duration"), []*cel.Type{cel.IntType}, cel.DurationType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Duration{Duration: time.Duration(args[0].(types.Int))}
				},
			),
		),
		BindFunction(
			createTimeName("parseDuration"),
			OverloadFunc(createTimeID("parse_duration_string_duration"), []*cel.Type{cel.StringType}, cel.DurationType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					d, err := time.ParseDuration(string(args[0].(types.String)))
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Duration{Duration: d}
				},
			),
		),
		BindFunction(
			createTimeName("since"),
			OverloadFunc(createTimeID("since_timestamp_duration"), []*cel.Type{cel.TimestampType}, cel.DurationType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Duration{
						Duration: time.Since(args[0].(types.Timestamp).Time),
					}
				},
			),
		),
		BindFunction(
			createTimeName("until"),
			OverloadFunc(createTimeID("until_timestamp_duration"), []*cel.Type{cel.TimestampType}, cel.DurationType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Duration{
						Duration: time.Until(args[0].(types.Timestamp).Time),
					}
				},
			),
		),
		BindMemberFunction(
			"abs",
			MemberOverloadFunc(createTimeID("abs_duration_duration"), cel.DurationType, []*cel.Type{}, cel.DurationType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Duration{
						Duration: self.(types.Duration).Duration.Abs(),
					}
				},
			),
		),
		BindMemberFunction(
			"hours",
			MemberOverloadFunc(createTimeID("hours_duration_double"), cel.DurationType, []*cel.Type{}, cel.DoubleType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Double(self.(types.Duration).Duration.Hours())
				},
			),
		),
		BindMemberFunction(
			"microseconds",
			MemberOverloadFunc(createTimeID("microseconds_duration_int"), cel.DurationType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Duration).Duration.Microseconds())
				},
			),
		),
		BindMemberFunction(
			"milliseconds",
			MemberOverloadFunc(createTimeID("milliseconds_duration_int"), cel.DurationType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Duration).Duration.Milliseconds())
				},
			),
		),
		BindMemberFunction(
			"minutes",
			MemberOverloadFunc(createTimeID("minutes_duration_int"), cel.DurationType, []*cel.Type{}, cel.DoubleType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Double(self.(types.Duration).Duration.Minutes())
				},
			),
		),
		BindMemberFunction(
			"nanoseconds",
			MemberOverloadFunc(createTimeID("nanoseconds_duration_int"), cel.DurationType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Duration).Duration.Nanoseconds())
				},
			),
		),
		BindMemberFunction(
			"round",
			MemberOverloadFunc(createTimeID("round_duration_int_duration"), cel.DurationType, []*cel.Type{cel.IntType}, cel.DurationType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Duration{
						Duration: self.(types.Duration).Duration.Round(time.Duration(args[0].(types.Int))),
					}
				},
			),
			MemberOverloadFunc(createTimeID("round_duration_duration_duration"), cel.DurationType, []*cel.Type{cel.DurationType}, cel.DurationType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Duration{
						Duration: self.(types.Duration).Duration.Round(args[0].(types.Duration).Duration),
					}
				},
			),
		),
		BindMemberFunction(
			"seconds",
			MemberOverloadFunc(createTimeID("seconds_duration_int"), cel.DurationType, []*cel.Type{}, cel.DoubleType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Double(self.(types.Duration).Duration.Seconds())
				},
			),
		),
		BindMemberFunction(
			"string",
			MemberOverloadFunc(createTimeID("string_duration_string"), cel.DurationType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.String(self.(types.Duration).Duration.String())
				},
			),
		),
		BindMemberFunction(
			"truncate",
			MemberOverloadFunc(createTimeID("truncate_duration_int_duration"), cel.DurationType, []*cel.Type{cel.IntType}, cel.DurationType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Duration{
						Duration: self.(types.Duration).Duration.Truncate(time.Duration(args[0].(types.Int))),
					}
				},
			),
			MemberOverloadFunc(createTimeID("truncate_duration_duration_duration"), cel.DurationType, []*cel.Type{cel.DurationType}, cel.DurationType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Duration{
						Duration: self.(types.Duration).Duration.Truncate(args[0].(types.Duration).Duration),
					}
				},
			),
		),

		// Location functions
		BindFunction(
			createTimeName("fixedZone"),
			OverloadFunc(createTimeID("fixed_zone_string_int_location"), []*cel.Type{cel.StringType, cel.IntType}, LocationType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					loc := time.FixedZone(string(args[0].(types.String)), int(args[1].(types.Int)))
					return &Location{loc}
				},
			),
			OverloadFunc(createTimeID("fixed_zone_string_double_location"), []*cel.Type{cel.StringType, cel.DoubleType}, LocationType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					loc := time.FixedZone(string(args[0].(types.String)), int(args[1].(types.Double)))
					return &Location{loc}
				},
			),
		),
		BindFunction(
			createTimeName("loadLocation"),
			OverloadFunc(createTimeID("load_location_string_location"), []*cel.Type{cel.StringType}, LocationType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					loc, err := time.LoadLocation(string(args[0].(types.String)))
					if err != nil {
						return types.NewErr(err.Error())
					}
					return &Location{loc}
				},
			),
		),
		BindFunction(
			createTimeName("loadLocationFromTZData"),
			OverloadFunc(createTimeID("load_location_from_tz_data_string_bytes_location"), []*cel.Type{cel.StringType, cel.BytesType}, LocationType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					loc, err := time.LoadLocationFromTZData(string(args[0].(types.String)), []byte(args[1].(types.Bytes)))
					if err != nil {
						return types.NewErr(err.Error())
					}
					return &Location{loc}
				},
			),
		),
		BindMemberFunction(
			"string",
			MemberOverloadFunc(createTimeID("string_location_string"), LocationType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.String(self.(*Location).String())
				},
			),
		),

		// Timestamp functions
		BindFunction(
			createTimeName("date"),
			OverloadFunc(
				createTimeID("date_int_int_int_int_int_int_int_location_timestamp"),
				[]*cel.Type{cel.IntType, cel.IntType, cel.IntType, cel.IntType, cel.IntType, cel.IntType, cel.IntType, LocationType},
				cel.TimestampType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Timestamp{
						Time: time.Date(
							int(args[0].(types.Int)),
							time.Month(args[1].(types.Int)),
							int(args[2].(types.Int)),
							int(args[3].(types.Int)),
							int(args[4].(types.Int)),
							int(args[5].(types.Int)),
							int(args[6].(types.Int)),
							args[7].(*Location).Location,
						),
					}
				},
			),
		),
		BindFunction(
			createTimeName("now"),
			OverloadFunc(createTimeID("now_timestamp"), nil, cel.TimestampType,
				func(_ context.Context, _ ...ref.Val) ref.Val {
					return types.Timestamp{Time: time.Now()}
				},
			),
		),
		BindFunction(
			createTimeName("parse"),
			OverloadFunc(createTimeID("parse_string_string_timestamp"), []*cel.Type{cel.StringType, cel.StringType}, cel.TimestampType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					t, err := time.Parse(string(args[0].(types.String)), string(args[1].(types.String)))
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Timestamp{Time: t}
				},
			),
		),
		BindFunction(
			createTimeName("parseInLocation"),
			OverloadFunc(createTimeID("parse_in_location_string_string_location_timestamp"), []*cel.Type{cel.StringType, cel.StringType, LocationType}, cel.TimestampType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					t, err := time.ParseInLocation(string(args[0].(types.String)), string(args[1].(types.String)), args[2].(*Location).Location)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Timestamp{Time: t}
				},
			),
		),
		BindFunction(
			createTimeName("unix"),
			OverloadFunc(createTimeID("unix_int_int_timestamp"), []*cel.Type{cel.IntType, cel.IntType}, cel.TimestampType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Timestamp{
						Time: time.Unix(int64(args[0].(types.Int)), int64(args[1].(types.Int))),
					}
				},
			),
		),
		BindFunction(
			createTimeName("unixMicro"),
			OverloadFunc(createTimeID("unix_micro_int_timestamp"), []*cel.Type{cel.IntType}, cel.TimestampType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Timestamp{
						Time: time.UnixMicro(int64(args[0].(types.Int))),
					}
				},
			),
		),
		BindFunction(
			createTimeName("unixMilli"),
			OverloadFunc(createTimeID("unix_milli_int_timestamp"), []*cel.Type{cel.IntType}, cel.TimestampType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return types.Timestamp{
						Time: time.UnixMilli(int64(args[0].(types.Int))),
					}
				},
			),
		),
		BindMemberFunction(
			"add",
			MemberOverloadFunc(createTimeID("add_timestamp_int_timestamp"), cel.TimestampType, []*cel.Type{cel.IntType}, cel.TimestampType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.Add(time.Duration(args[0].(types.Int))),
					}
				},
			),
			MemberOverloadFunc(createTimeID("add_timestamp_duration_timestamp"), cel.TimestampType, []*cel.Type{cel.DurationType}, cel.TimestampType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.Add(args[0].(types.Duration).Duration),
					}
				},
			),
		),
		BindMemberFunction(
			"addDate",
			MemberOverloadFunc(createTimeID("add_date_timestamp_int_int_int_timestamp"), cel.TimestampType, []*cel.Type{cel.IntType, cel.IntType, cel.IntType}, cel.TimestampType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.AddDate(int(args[0].(types.Int)), int(args[1].(types.Int)), int(args[2].(types.Int))),
					}
				},
			),
		),
		BindMemberFunction(
			"after",
			MemberOverloadFunc(createTimeID("after_timestamp_timestamp_bool"), cel.TimestampType, []*cel.Type{cel.TimestampType}, cel.BoolType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Bool(self.(types.Timestamp).Time.After(args[0].(types.Timestamp).Time))
				},
			),
		),
		BindMemberFunction(
			"appendFormat",
			MemberOverloadFunc(createTimeID("append_format_timestamp_string_string_bytes"), cel.TimestampType, []*cel.Type{cel.StringType, cel.StringType}, cel.BytesType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Bytes(
						self.(types.Timestamp).Time.AppendFormat(
							[]byte(args[0].(types.String)),
							string(args[1].(types.String)),
						),
					)
				},
			),
			MemberOverloadFunc(createTimeID("append_format_timestamp_bytes_string_bytes"), cel.TimestampType, []*cel.Type{cel.BytesType, cel.StringType}, cel.BytesType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Bytes(
						self.(types.Timestamp).Time.AppendFormat(
							[]byte(args[0].(types.Bytes)),
							string(args[1].(types.String)),
						),
					)
				},
			),
		),
		BindMemberFunction(
			"before",
			MemberOverloadFunc(createTimeID("before_timestamp_timestamp_bool"), cel.TimestampType, []*cel.Type{cel.TimestampType}, cel.BoolType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Bool(self.(types.Timestamp).Time.Before(args[0].(types.Timestamp).Time))
				},
			),
		),
		BindMemberFunction(
			"compare",
			MemberOverloadFunc(createTimeID("compare_timestamp_timestamp_int"), cel.TimestampType, []*cel.Type{cel.TimestampType}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Compare(args[0].(types.Timestamp).Time))
				},
			),
		),
		BindMemberFunction(
			"day",
			MemberOverloadFunc(createTimeID("day_timestamp_int"), cel.TimestampType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Day())
				},
			),
		),
		BindMemberFunction(
			"equal",
			MemberOverloadFunc(createTimeID("equal_timestamp_timestamp_bool"), cel.TimestampType, []*cel.Type{cel.TimestampType}, cel.BoolType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Bool(self.(types.Timestamp).Time.Equal(args[0].(types.Timestamp).Time))
				},
			),
		),
		BindMemberFunction(
			"format",
			MemberOverloadFunc(createTimeID("format_timestamp_string_string"), cel.TimestampType, []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.String(self.(types.Timestamp).Time.Format(string(args[0].(types.String))))
				},
			),
		),
		BindMemberFunction(
			"hour",
			MemberOverloadFunc(createTimeID("hour_timestamp_int"), cel.TimestampType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Hour())
				},
			),
		),
		BindMemberFunction(
			"withLocation",
			MemberOverloadFunc(createTimeID("with_location_timestamp_location_timestamp"), cel.TimestampType, []*cel.Type{LocationType}, cel.TimestampType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.In(args[0].(*Location).Location),
					}
				},
			),
		),
		BindMemberFunction(
			"isDST",
			MemberOverloadFunc(createTimeID("is_dst_timestamp_bool"), cel.TimestampType, []*cel.Type{}, cel.BoolType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Bool(self.(types.Timestamp).Time.IsDST())
				},
			),
		),
		BindMemberFunction(
			"isZero",
			MemberOverloadFunc(createTimeID("is_zero_timestamp_bool"), cel.TimestampType, []*cel.Type{}, cel.BoolType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Bool(self.(types.Timestamp).Time.IsZero())
				},
			),
		),
		BindMemberFunction(
			"local",
			MemberOverloadFunc(createTimeID("local_timestamp_timestamp"), cel.TimestampType, []*cel.Type{}, cel.TimestampType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.Local(),
					}
				},
			),
		),
		BindMemberFunction(
			"location",
			MemberOverloadFunc(createTimeID("location_timestamp_location"), cel.TimestampType, []*cel.Type{}, LocationType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return &Location{Location: self.(types.Timestamp).Time.Location()}
				},
			),
		),
		BindMemberFunction(
			"minute",
			MemberOverloadFunc(createTimeID("minute_timestamp_int"), cel.TimestampType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Minute())
				},
			),
		),
		BindMemberFunction(
			"month",
			MemberOverloadFunc(createTimeID("month_timestamp_int"), cel.TimestampType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Month())
				},
			),
		),
		BindMemberFunction(
			"nanosecond",
			MemberOverloadFunc(createTimeID("nanosecond_timestamp_int"), cel.TimestampType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Nanosecond())
				},
			),
		),
		BindMemberFunction(
			"round",
			MemberOverloadFunc(createTimeID("round_timestamp_int_timestamp"), cel.TimestampType, []*cel.Type{cel.IntType}, cel.TimestampType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.Round(time.Duration(args[0].(types.Int))),
					}
				},
			),
			MemberOverloadFunc(createTimeID("round_timestamp_duration_timestamp"), cel.TimestampType, []*cel.Type{cel.DurationType}, cel.TimestampType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.Round(args[0].(types.Duration).Duration),
					}
				},
			),
		),
		BindMemberFunction(
			"second",
			MemberOverloadFunc(createTimeID("second_timestamp_int"), cel.TimestampType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Second())
				},
			),
		),
		BindMemberFunction(
			"string",
			MemberOverloadFunc(createTimeID("string_timestamp_string"), cel.TimestampType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.String(self.(types.Timestamp).Time.String())
				},
			),
		),
		BindMemberFunction(
			"sub",
			MemberOverloadFunc(createTimeID("sub_timestamp_timestamp_duration"), cel.TimestampType, []*cel.Type{cel.TimestampType}, cel.DurationType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Duration{
						Duration: self.(types.Timestamp).Time.Sub(args[0].(types.Timestamp).Time),
					}
				},
			),
		),
		BindMemberFunction(
			"truncate",
			MemberOverloadFunc(createTimeID("truncate_timestamp_int_timestamp"), cel.TimestampType, []*cel.Type{cel.IntType}, cel.TimestampType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.Truncate(time.Duration(args[0].(types.Int))),
					}
				},
			),
			MemberOverloadFunc(createTimeID("truncate_timestamp_duration_timestamp"), cel.TimestampType, []*cel.Type{cel.DurationType}, cel.TimestampType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Timestamp{
						Time: self.(types.Timestamp).Time.Truncate(args[0].(types.Duration).Duration),
					}
				},
			),
		),
		BindMemberFunction(
			"utc",
			MemberOverloadFunc(createTimeID("utc_timestamp_timestamp"), cel.TimestampType, []*cel.Type{}, cel.TimestampType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Timestamp{Time: self.(types.Timestamp).Time.UTC()}
				},
			),
		),
		BindMemberFunction(
			"unix",
			MemberOverloadFunc(createTimeID("unix_timestamp_int"), cel.TimestampType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Unix())
				},
			),
		),
		BindMemberFunction(
			"unixMicro",
			MemberOverloadFunc(createTimeID("unix_micro_timestamp_int"), cel.TimestampType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.UnixMicro())
				},
			),
		),
		BindMemberFunction(
			"unixMilli",
			MemberOverloadFunc(createTimeID("unix_milli_timestamp_int"), cel.TimestampType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.UnixMilli())
				},
			),
		),
		BindMemberFunction(
			"unixNano",
			MemberOverloadFunc(createTimeID("unix_nano_timestamp_int"), cel.TimestampType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.UnixNano())
				},
			),
		),
		BindMemberFunction(
			"weekday",
			MemberOverloadFunc(createTimeID("weekday_timestamp_int"), cel.TimestampType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Weekday())
				},
			),
		),
		BindMemberFunction(
			"year",
			MemberOverloadFunc(createTimeID("year_timestamp_int"), cel.TimestampType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.Year())
				},
			),
		),
		BindMemberFunction(
			"yearDay",
			MemberOverloadFunc(createTimeID("year_day_timestamp_int"), cel.TimestampType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return types.Int(self.(types.Timestamp).Time.YearDay())
				},
			),
		),
	} {
		opts = append(opts, funcOpts...)
	}
	return opts
}

func (lib *TimeLibrary) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}
