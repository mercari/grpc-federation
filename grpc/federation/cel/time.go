package cel

import (
	"context"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const TimePackageName = "time"

var (
	TimeType     = cel.ObjectType("grpc.federation.time.Time")
	LocationType = cel.ObjectType("grpc.federation.time.Location")
)

func (x *Time) GoTime() (time.Time, error) {
	loc, err := x.GetLoc().GoLocation()
	if err != nil {
		return time.Time{}, err
	}
	return x.GetTimestamp().AsTime().In(loc), nil
}

func (x *Location) GoLocation() (*time.Location, error) {
	if x.GetOffset() != 0 {
		return time.FixedZone(x.GetName(), int(x.GetOffset())), nil
	}
	loc, err := time.LoadLocation(x.GetName())
	if err != nil {
		return nil, err
	}
	return loc, nil
}

var _ cel.SingletonLibrary = new(TimeLibrary)

type TimeLibrary struct {
	typeAdapter types.Adapter
}

func NewTimeLibrary(typeAdapter types.Adapter) *TimeLibrary {
	return &TimeLibrary{
		typeAdapter: typeAdapter,
	}
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

func (lib *TimeLibrary) refToGoTimeValue(v ref.Val) (time.Time, error) {
	return v.Value().(*Time).GoTime()
}

func (lib *TimeLibrary) toTimeValue(v time.Time) ref.Val {
	name, offset := v.Zone()
	return lib.typeAdapter.NativeToValue(&Time{
		Timestamp: timestamppb.New(v),
		Loc: &Location{
			Name:   name,
			Offset: int64(offset),
		},
	})
}

func (lib *TimeLibrary) toLocationValue(name string, offset int) ref.Val {
	return lib.typeAdapter.NativeToValue(&Location{Name: name, Offset: int64(offset)})
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
					return lib.toLocationValue(time.Local.String(), 0)
				},
			),
		),
		BindFunction(
			createTimeName("UTC"),
			OverloadFunc(createTimeID("utc_location"), []*cel.Type{}, LocationType,
				func(_ context.Context, _ ...ref.Val) ref.Val {
					return lib.toLocationValue(time.UTC.String(), 0)
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
			OverloadFunc(createTimeID("since_timestamp_duration"), []*cel.Type{TimeType}, cel.DurationType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(args[0])
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Duration{
						Duration: time.Since(v),
					}
				},
			),
		),
		BindFunction(
			createTimeName("until"),
			OverloadFunc(createTimeID("until_timestamp_duration"), []*cel.Type{TimeType}, cel.DurationType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(args[0])
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Duration{
						Duration: time.Until(v),
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
					return lib.toLocationValue(string(args[0].(types.String)), int(args[1].(types.Int)))
				},
			),
			OverloadFunc(createTimeID("fixed_zone_string_double_location"), []*cel.Type{cel.StringType, cel.DoubleType}, LocationType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return lib.toLocationValue(string(args[0].(types.String)), int(args[1].(types.Double)))
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
					return lib.toLocationValue(loc.String(), 0)
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
					return lib.toLocationValue(loc.String(), 0)
				},
			),
		),
		BindMemberFunction(
			"string",
			MemberOverloadFunc(createTimeID("string_location_string"), LocationType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					loc, err := self.Value().(*Location).GoLocation()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(loc.String())
				},
			),
		),

		// Timestamp functions
		BindMemberFunction(
			"asTime",
			MemberOverloadFunc(createTimeID("asTime_timestamp_int_time"), cel.TimestampType, []*cel.Type{}, TimeType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					return lib.toTimeValue(self.(types.Timestamp).Time)
				},
			),
		),

		// Time functions
		BindFunction(
			createTimeName("date"),
			OverloadFunc(
				createTimeID("date_int_int_int_int_int_int_int_location_time"),
				[]*cel.Type{cel.IntType, cel.IntType, cel.IntType, cel.IntType, cel.IntType, cel.IntType, cel.IntType, LocationType},
				TimeType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					loc, err := args[7].Value().(*Location).GoLocation()
					if err != nil {
						return types.NewErr(err.Error())
					}
					return lib.toTimeValue(
						time.Date(
							int(args[0].(types.Int)),
							time.Month(args[1].(types.Int)),
							int(args[2].(types.Int)),
							int(args[3].(types.Int)),
							int(args[4].(types.Int)),
							int(args[5].(types.Int)),
							int(args[6].(types.Int)),
							loc,
						),
					)
				},
			),
		),
		BindFunction(
			createTimeName("now"),
			OverloadFunc(createTimeID("now_time"), nil, TimeType,
				func(_ context.Context, _ ...ref.Val) ref.Val {
					return lib.toTimeValue(time.Now())
				},
			),
		),
		BindFunction(
			createTimeName("parse"),
			OverloadFunc(createTimeID("parse_string_string_time"), []*cel.Type{cel.StringType, cel.StringType}, TimeType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					t, err := time.Parse(string(args[0].(types.String)), string(args[1].(types.String)))
					if err != nil {
						return types.NewErr(err.Error())
					}
					return lib.toTimeValue(t)
				},
			),
		),
		BindFunction(
			createTimeName("parseInLocation"),
			OverloadFunc(createTimeID("parse_in_location_string_string_location_time"), []*cel.Type{cel.StringType, cel.StringType, LocationType}, TimeType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					loc, err := args[2].Value().(*Location).GoLocation()
					if err != nil {
						return types.NewErr(err.Error())
					}
					t, err := time.ParseInLocation(string(args[0].(types.String)), string(args[1].(types.String)), loc)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return lib.toTimeValue(t)
				},
			),
		),
		BindFunction(
			createTimeName("unix"),
			OverloadFunc(createTimeID("unix_int_int_time"), []*cel.Type{cel.IntType, cel.IntType}, TimeType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return lib.toTimeValue(time.Unix(int64(args[0].(types.Int)), int64(args[1].(types.Int))))
				},
			),
		),
		BindFunction(
			createTimeName("unixMicro"),
			OverloadFunc(createTimeID("unix_micro_int_time"), []*cel.Type{cel.IntType}, TimeType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return lib.toTimeValue(time.UnixMicro(int64(args[0].(types.Int))))
				},
			),
		),
		BindFunction(
			createTimeName("unixMilli"),
			OverloadFunc(createTimeID("unix_milli_int_time"), []*cel.Type{cel.IntType}, TimeType,
				func(_ context.Context, args ...ref.Val) ref.Val {
					return lib.toTimeValue(time.UnixMilli(int64(args[0].(types.Int))))
				},
			),
		),
		BindMemberFunction(
			"add",
			MemberOverloadFunc(createTimeID("add_time_int_time"), TimeType, []*cel.Type{cel.IntType}, TimeType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return lib.toTimeValue(
						v.Add(time.Duration(args[0].(types.Int))),
					)
				},
			),
			MemberOverloadFunc(createTimeID("add_time_duration_time"), TimeType, []*cel.Type{cel.DurationType}, TimeType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return lib.toTimeValue(
						v.Add(args[0].(types.Duration).Duration),
					)
				},
			),
		),
		BindMemberFunction(
			"addDate",
			MemberOverloadFunc(createTimeID("add_date_time_int_int_int_time"), TimeType, []*cel.Type{cel.IntType, cel.IntType, cel.IntType}, TimeType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return lib.toTimeValue(
						v.AddDate(int(args[0].(types.Int)), int(args[1].(types.Int)), int(args[2].(types.Int))),
					)
				},
			),
		),
		BindMemberFunction(
			"after",
			MemberOverloadFunc(createTimeID("after_time_time_bool"), TimeType, []*cel.Type{TimeType}, cel.BoolType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					a0, err := lib.refToGoTimeValue(args[0])
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Bool(v.After(a0))
				},
			),
		),
		BindMemberFunction(
			"appendFormat",
			MemberOverloadFunc(createTimeID("append_format_time_string_string_bytes"), TimeType, []*cel.Type{cel.StringType, cel.StringType}, cel.BytesType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Bytes(
						v.AppendFormat(
							[]byte(args[0].(types.String)),
							string(args[1].(types.String)),
						),
					)
				},
			),
			MemberOverloadFunc(createTimeID("append_format_time_bytes_string_bytes"), TimeType, []*cel.Type{cel.BytesType, cel.StringType}, cel.BytesType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Bytes(
						v.AppendFormat(
							[]byte(args[0].(types.Bytes)),
							string(args[1].(types.String)),
						),
					)
				},
			),
		),
		BindMemberFunction(
			"before",
			MemberOverloadFunc(createTimeID("before_time_time_bool"), TimeType, []*cel.Type{TimeType}, cel.BoolType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					a0, err := lib.refToGoTimeValue(args[0])
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Bool(v.Before(a0))
				},
			),
		),
		BindMemberFunction(
			"compare",
			MemberOverloadFunc(createTimeID("compare_time_time_int"), TimeType, []*cel.Type{TimeType}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					a0, err := lib.refToGoTimeValue(args[0])
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Int(v.Compare(a0))
				},
			),
		),
		BindMemberFunction(
			"day",
			MemberOverloadFunc(createTimeID("day_time_int"), TimeType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Int(v.Day())
				},
			),
		),
		BindMemberFunction(
			"equal",
			MemberOverloadFunc(createTimeID("equal_time_time_bool"), TimeType, []*cel.Type{TimeType}, cel.BoolType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					a0, err := lib.refToGoTimeValue(args[0])
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Bool(v.Equal(a0))
				},
			),
		),
		BindMemberFunction(
			"format",
			MemberOverloadFunc(createTimeID("format_time_string_string"), TimeType, []*cel.Type{cel.StringType}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.Format(string(args[0].(types.String))))
				},
			),
		),
		BindMemberFunction(
			"hour",
			MemberOverloadFunc(createTimeID("hour_time_int"), TimeType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Int(v.Hour())
				},
			),
		),
		BindMemberFunction(
			"withLocation",
			MemberOverloadFunc(createTimeID("with_location_time_location_time"), TimeType, []*cel.Type{LocationType}, TimeType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					loc, err := args[0].Value().(*Location).GoLocation()
					if err != nil {
						return types.NewErr(err.Error())
					}
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return lib.toTimeValue(v.In(loc))
				},
			),
		),
		BindMemberFunction(
			"isDST",
			MemberOverloadFunc(createTimeID("is_dst_time_bool"), TimeType, []*cel.Type{}, cel.BoolType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Bool(v.IsDST())
				},
			),
		),
		BindMemberFunction(
			"isZero",
			MemberOverloadFunc(createTimeID("is_zero_time_bool"), TimeType, []*cel.Type{}, cel.BoolType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Bool(v.IsZero())
				},
			),
		),
		BindMemberFunction(
			"local",
			MemberOverloadFunc(createTimeID("local_time_time"), TimeType, []*cel.Type{}, TimeType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return lib.toTimeValue(v.Local())
				},
			),
		),
		BindMemberFunction(
			"location",
			MemberOverloadFunc(createTimeID("location_time_location"), TimeType, []*cel.Type{}, LocationType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					name, offset := v.Zone()
					return lib.toLocationValue(name, offset)
				},
			),
		),
		BindMemberFunction(
			"minute",
			MemberOverloadFunc(createTimeID("minute_time_int"), TimeType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Int(v.Minute())
				},
			),
		),
		BindMemberFunction(
			"month",
			MemberOverloadFunc(createTimeID("month_time_int"), TimeType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Int(v.Month())
				},
			),
		),
		BindMemberFunction(
			"nanosecond",
			MemberOverloadFunc(createTimeID("nanosecond_time_int"), TimeType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Int(v.Nanosecond())
				},
			),
		),
		BindMemberFunction(
			"round",
			MemberOverloadFunc(createTimeID("round_time_int_time"), TimeType, []*cel.Type{cel.IntType}, TimeType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return lib.toTimeValue(v.Round(time.Duration(args[0].(types.Int))))
				},
			),
			MemberOverloadFunc(createTimeID("round_time_duration_time"), TimeType, []*cel.Type{cel.DurationType}, TimeType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return lib.toTimeValue(v.Round(args[0].(types.Duration).Duration))
				},
			),
		),
		BindMemberFunction(
			"second",
			MemberOverloadFunc(createTimeID("second_time_int"), TimeType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Int(v.Second())
				},
			),
		),
		BindMemberFunction(
			"string",
			MemberOverloadFunc(createTimeID("string_time_string"), TimeType, []*cel.Type{}, cel.StringType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.String(v.String())
				},
			),
		),
		BindMemberFunction(
			"sub",
			MemberOverloadFunc(createTimeID("sub_time_time_duration"), TimeType, []*cel.Type{TimeType}, cel.DurationType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					a0, err := lib.refToGoTimeValue(args[0])
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Duration{
						Duration: v.Sub(a0),
					}
				},
			),
		),
		BindMemberFunction(
			"truncate",
			MemberOverloadFunc(createTimeID("truncate_time_int_time"), TimeType, []*cel.Type{cel.IntType}, TimeType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return lib.toTimeValue(v.Truncate(time.Duration(args[0].(types.Int))))
				},
			),
			MemberOverloadFunc(createTimeID("truncate_time_duration_time"), TimeType, []*cel.Type{cel.DurationType}, TimeType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return lib.toTimeValue(v.Truncate(args[0].(types.Duration).Duration))
				},
			),
		),
		BindMemberFunction(
			"utc",
			MemberOverloadFunc(createTimeID("utc_time_time"), TimeType, []*cel.Type{}, TimeType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return lib.toTimeValue(v.UTC())
				},
			),
		),
		BindMemberFunction(
			"unix",
			MemberOverloadFunc(createTimeID("unix_time_int"), TimeType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Int(v.Unix())
				},
			),
		),
		BindMemberFunction(
			"unixMicro",
			MemberOverloadFunc(createTimeID("unix_micro_time_int"), TimeType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Int(v.UnixMicro())
				},
			),
		),
		BindMemberFunction(
			"unixMilli",
			MemberOverloadFunc(createTimeID("unix_milli_time_int"), TimeType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Int(v.UnixMilli())
				},
			),
		),
		BindMemberFunction(
			"unixNano",
			MemberOverloadFunc(createTimeID("unix_nano_time_int"), TimeType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Int(v.UnixNano())
				},
			),
		),
		BindMemberFunction(
			"weekday",
			MemberOverloadFunc(createTimeID("weekday_time_int"), TimeType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Int(v.Weekday())
				},
			),
		),
		BindMemberFunction(
			"year",
			MemberOverloadFunc(createTimeID("year_time_int"), TimeType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Int(v.Year())
				},
			),
		),
		BindMemberFunction(
			"yearDay",
			MemberOverloadFunc(createTimeID("year_day_time_int"), TimeType, []*cel.Type{}, cel.IntType,
				func(_ context.Context, self ref.Val, args ...ref.Val) ref.Val {
					v, err := lib.refToGoTimeValue(self)
					if err != nil {
						return types.NewErr(err.Error())
					}
					return types.Int(v.YearDay())
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
