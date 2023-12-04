package cel_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	cellib "github.com/mercari/grpc-federation/grpc/federation/cel"
)

func TestTime(t *testing.T) {
	tests := []struct {
		name string
		expr string
		args map[string]any
		cmp  func(any) error
	}{
		// constants
		{
			name: "constant_layout",
			expr: "grpc.federation.time.LAYOUT",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Layout
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_ansic",
			expr: "grpc.federation.time.ANSIC",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.ANSIC
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_unix_date",
			expr: "grpc.federation.time.UNIX_DATE",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.UnixDate
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_ruby_date",
			expr: "grpc.federation.time.RUBY_DATE",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.RubyDate
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_rfc822",
			expr: "grpc.federation.time.RFC822",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.RFC822
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_rfc822z",
			expr: "grpc.federation.time.RFC822Z",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.RFC822Z
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_rfc850",
			expr: "grpc.federation.time.RFC850",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.RFC850
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_rfc1123",
			expr: "grpc.federation.time.RFC1123",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.RFC1123
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_rfc1123z",
			expr: "grpc.federation.time.RFC1123Z",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.RFC1123Z
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_rfc3339",
			expr: "grpc.federation.time.RFC3339",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.RFC3339
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_rfc3339nano",
			expr: "grpc.federation.time.RFC3339NANO",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.RFC3339Nano
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_kitchen",
			expr: "grpc.federation.time.KITCHEN",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Kitchen
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_stamp",
			expr: "grpc.federation.time.STAMP",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Stamp
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_stamp_milli",
			expr: "grpc.federation.time.STAMP_MILLI",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.StampMilli
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_stamp_micro",
			expr: "grpc.federation.time.STAMP_MICRO",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.StampMicro
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_stamp_nano",
			expr: "grpc.federation.time.STAMP_NANO",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.StampNano
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_datetime",
			expr: "grpc.federation.time.DATETIME",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.DateTime
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_date_only",
			expr: "grpc.federation.time.DATE_ONLY",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.DateOnly
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_time_only",
			expr: "grpc.federation.time.TIME_ONLY",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.TimeOnly
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_nanosecond",
			expr: "grpc.federation.time.NANOSECOND",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Nanosecond
				if diff := cmp.Diff(time.Duration(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_microsecond",
			expr: "grpc.federation.time.MICROSECOND",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Microsecond
				if diff := cmp.Diff(time.Duration(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_millisecond",
			expr: "grpc.federation.time.MILLISECOND",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Millisecond
				if diff := cmp.Diff(time.Duration(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_second",
			expr: "grpc.federation.time.SECOND",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Second
				if diff := cmp.Diff(time.Duration(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_minute",
			expr: "grpc.federation.time.MINUTE",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Minute
				if diff := cmp.Diff(time.Duration(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_hour",
			expr: "grpc.federation.time.HOUR",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Hour
				if diff := cmp.Diff(time.Duration(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_local",
			expr: "grpc.federation.time.LOCAL",
			cmp: func(got any) error {
				gotV, ok := got.(*cellib.Location)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Local.String()
				if diff := cmp.Diff(gotV.Location.String(), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "constant_utc",
			expr: "grpc.federation.time.UTC",
			cmp: func(got any) error {
				gotV, ok := got.(*cellib.Location)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.UTC.String()
				if diff := cmp.Diff(gotV.Location.String(), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},

		// duration functions
		{
			name: "parseDuration",
			expr: "grpc.federation.time.parseDuration('10h')",
			cmp: func(got any) error {
				gotV, ok := got.(types.Duration)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected, err := time.ParseDuration("10h")
				if err != nil {
					return err
				}
				if diff := cmp.Diff(gotV.Duration.String(), expected.String()); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "since",
			expr: "grpc.federation.time.since(grpc.federation.time.now())",
			cmp: func(got any) error {
				gotV, ok := got.(types.Duration)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if !strings.Contains(gotV.Duration.String(), "s") {
					return fmt.Errorf("failed to evaluate time.since: %v", gotV)
				}
				return nil
			},
		},
		{
			name: "until",
			expr: "grpc.federation.time.until(grpc.federation.time.now())",
			cmp: func(got any) error {
				gotV, ok := got.(types.Duration)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if !strings.Contains(gotV.Duration.String(), "s") {
					return fmt.Errorf("failed to evaluate time.until: %v", gotV)
				}
				return nil
			},
		},
		{
			name: "abs",
			expr: "grpc.federation.time.parseDuration('4h30m').abs()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Duration)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				d, err := time.ParseDuration("4h30m")
				if err != nil {
					return err
				}
				expected := d.Abs()
				if diff := cmp.Diff(gotV.Duration.String(), expected.String()); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "hours",
			expr: "grpc.federation.time.parseDuration('4h30m').hours()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Double)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				d, err := time.ParseDuration("4h30m")
				if err != nil {
					return err
				}
				expected := d.Hours()
				if diff := cmp.Diff(float64(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "microseconds",
			expr: "grpc.federation.time.parseDuration('1s').microseconds()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				d, err := time.ParseDuration("1s")
				if err != nil {
					return err
				}
				expected := d.Microseconds()
				if diff := cmp.Diff(int64(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "milliseconds",
			expr: "grpc.federation.time.parseDuration('1s').milliseconds()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				d, err := time.ParseDuration("1s")
				if err != nil {
					return err
				}
				expected := d.Milliseconds()
				if diff := cmp.Diff(int64(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "minutes",
			expr: "grpc.federation.time.parseDuration('1h30m').minutes()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Double)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				d, err := time.ParseDuration("1h30m")
				if err != nil {
					return err
				}
				expected := d.Minutes()
				if diff := cmp.Diff(float64(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "nanoseconds",
			expr: "grpc.federation.time.parseDuration('1µs').nanoseconds()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				d, err := time.ParseDuration("1µs")
				if err != nil {
					return err
				}
				expected := d.Nanoseconds()
				if diff := cmp.Diff(int64(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "round",
			expr: "grpc.federation.time.parseDuration('1h15m30.918273645s').round(grpc.federation.time.MICROSECOND)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Duration)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				d, err := time.ParseDuration("1h15m30.918273645s")
				if err != nil {
					return err
				}
				expected := d.Round(time.Microsecond)
				if diff := cmp.Diff(gotV.Duration, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "seconds",
			expr: "grpc.federation.time.parseDuration('1m30s').seconds()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Double)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				d, err := time.ParseDuration("1m30s")
				if err != nil {
					return err
				}
				expected := d.Seconds()
				if diff := cmp.Diff(float64(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "duration_string",
			expr: "grpc.federation.time.parseDuration('1m30s').string()",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				d, err := time.ParseDuration("1m30s")
				if err != nil {
					return err
				}
				expected := d.String()
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "truncate",
			expr: "grpc.federation.time.parseDuration('1h15m30.918273645s').truncate(grpc.federation.time.MICROSECOND)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Duration)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				d, err := time.ParseDuration("1h15m30.918273645s")
				if err != nil {
					return err
				}
				expected := d.Truncate(time.Microsecond)
				if diff := cmp.Diff(gotV.Duration, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},

		// location functions
		{
			name: "fixedZone",
			expr: "grpc.federation.time.fixedZone('UTC-8', -8*60*60)",
			cmp: func(got any) error {
				gotV, ok := got.(*cellib.Location)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.FixedZone("UTC-8", -8*60*60).String()
				if diff := cmp.Diff(gotV.Location.String(), expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "loadLocation",
			expr: "grpc.federation.time.loadLocation('America/Los_Angeles')",
			cmp: func(got any) error {
				gotV, ok := got.(*cellib.Location)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected, err := time.LoadLocation("America/Los_Angeles")
				if err != nil {
					return err
				}
				if diff := cmp.Diff(gotV.Location.String(), expected.String(), cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "location_string",
			expr: "grpc.federation.time.loadLocation('America/Los_Angeles').string()",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected, err := time.LoadLocation("America/Los_Angeles")
				if err != nil {
					return err
				}
				if diff := cmp.Diff(string(gotV), expected.String(), cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},

		// time functions
		{
			name: "date",
			expr: "grpc.federation.time.date(2009, 11, 10, 23, 0, 0, 0, grpc.federation.time.UTC)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "now",
			expr: "grpc.federation.time.now()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now()
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse",
			expr: "grpc.federation.time.parse(grpc.federation.time.RFC3339, '2006-01-02T15:04:05Z')",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
				if err != nil {
					return err
				}
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "parse_in_location",
			expr: "grpc.federation.time.parseInLocation('Jan 2, 2006 at 3:04pm (MST)', 'Jul 9, 2012 at 5:02am (CEST)', grpc.federation.time.loadLocation('Europe/Berlin'))",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				loc, err := time.LoadLocation("Europe/Berlin")
				if err != nil {
					return err
				}
				expected, err := time.ParseInLocation("Jan 2, 2006 at 3:04pm (MST)", "Jul 9, 2012 at 5:02am (CEST)", loc)
				if err != nil {
					return err
				}
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "unix",
			expr: "grpc.federation.time.unix(1257894000, 0)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Unix(1257894000, 0)
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "unix_micro",
			expr: "grpc.federation.time.unixMicro(1257894000000000)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.UnixMicro(1257894000000000)
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "unix_milli",
			expr: "grpc.federation.time.unixMilli(1257894000000)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.UnixMilli(1257894000000)
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "add",
			expr: "grpc.federation.time.now().add(10 * grpc.federation.time.HOUR)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().Add(10 * time.Hour)
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "addDate",
			expr: "grpc.federation.time.now().addDate(1, 1, 1)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().AddDate(1, 1, 1)
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "after",
			expr: "grpc.federation.time.date(3000, 1, 1, 0, 0, 0, 0, grpc.federation.time.UTC).after(grpc.federation.time.date(2000, 1, 1, 0, 0, 0, 0, grpc.federation.time.UTC))",
			cmp: func(got any) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC).After(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
				if diff := cmp.Diff(bool(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "appendFormat",
			expr: "grpc.federation.time.date(2017, 11, 4, 11, 0, 0, 0, grpc.federation.time.UTC).appendFormat('Time: ', grpc.federation.time.KITCHEN)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Bytes)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Date(2017, time.November, 4, 11, 0, 0, 0, time.UTC).AppendFormat([]byte("Time: "), time.Kitchen)
				if diff := cmp.Diff([]byte(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "before",
			expr: "grpc.federation.time.date(2000, 1, 1, 0, 0, 0, 0, grpc.federation.time.UTC).before(grpc.federation.time.date(3000, 1, 1, 0, 0, 0, 0, grpc.federation.time.UTC))",
			cmp: func(got any) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Before(time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC))
				if diff := cmp.Diff(bool(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "compare",
			expr: "grpc.federation.time.date(2000, 1, 1, 0, 0, 0, 0, grpc.federation.time.UTC).compare(grpc.federation.time.date(3000, 1, 1, 0, 0, 0, 0, grpc.federation.time.UTC))",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Compare(time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC))
				if diff := cmp.Diff(int(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "day",
			expr: "grpc.federation.time.now().day()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().Day()
				if diff := cmp.Diff(int(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "equal",
			expr: "grpc.federation.time.date(2000, 2, 1, 12, 30, 0, 0, grpc.federation.time.UTC).equal(grpc.federation.time.date(2000, 2, 1, 20, 30, 0, 0, grpc.federation.time.fixedZone('Beijing Time', grpc.federation.time.toDuration(8 * grpc.federation.time.HOUR).seconds())))",
			cmp: func(got any) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Date(2000, 2, 1, 12, 30, 0, 0, time.UTC).Equal(
					time.Date(2000, 2, 1, 20, 30, 0, 0, time.FixedZone("Beijing Time", int((8*time.Hour).Seconds()))),
				)
				if diff := cmp.Diff(bool(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "hour",
			expr: "grpc.federation.time.now().hour()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().Hour()
				if diff := cmp.Diff(int(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "withLocation",
			expr: "grpc.federation.time.now().withLocation(grpc.federation.time.UTC)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().In(time.UTC)
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "isDST",
			expr: "grpc.federation.time.now().isDST()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().IsDST()
				if diff := cmp.Diff(bool(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "isZero",
			expr: "grpc.federation.time.now().isDST()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Bool)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().IsZero()
				if diff := cmp.Diff(bool(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "local",
			expr: "grpc.federation.time.now().local()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().Local()
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "location",
			expr: "grpc.federation.time.now().location()",
			cmp: func(got any) error {
				gotV, ok := got.(*cellib.Location)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().Location()
				if diff := cmp.Diff(gotV.Location.String(), expected.String()); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "minute",
			expr: "grpc.federation.time.now().minute()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().Minute()
				if diff := cmp.Diff(int(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "month",
			expr: "grpc.federation.time.now().month()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().Month()
				if diff := cmp.Diff(time.Month(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "nanosecond",
			expr: "grpc.federation.time.now().nanosecond()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				if gotV <= 0 {
					return fmt.Errorf("failed to get nanosecond: %v", got)
				}
				return nil
			},
		},
		{
			name: "round",
			expr: "grpc.federation.time.date(0, 0, 0, 12, 15, 30, 918273645, grpc.federation.time.UTC).round(grpc.federation.time.MILLISECOND)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Date(0, 0, 0, 12, 15, 30, 918273645, time.UTC).Round(time.Millisecond)
				if diff := cmp.Diff(gotV.Time, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "second",
			expr: "grpc.federation.time.date(0, 0, 0, 12, 15, 30, 10, grpc.federation.time.UTC).second()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Date(0, 0, 0, 12, 15, 30, 10, time.UTC).Second()
				if diff := cmp.Diff(int(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "string",
			expr: "grpc.federation.time.date(2000, 11, 12, 12, 15, 30, 10, grpc.federation.time.UTC).string()",
			cmp: func(got any) error {
				gotV, ok := got.(types.String)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Date(2000, 11, 12, 12, 15, 30, 10, time.UTC).String()
				if diff := cmp.Diff(string(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "sub",
			expr: "grpc.federation.time.date(2000, 1, 1, 0, 0, 0, 0, grpc.federation.time.UTC).sub(grpc.federation.time.date(2000, 1, 1, 12, 0, 0, 0, grpc.federation.time.UTC))",
			cmp: func(got any) error {
				gotV, ok := got.(types.Duration)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Sub(time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC))
				if diff := cmp.Diff(gotV.Duration, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "truncate",
			expr: "grpc.federation.time.date(2012, 12, 7, 12, 15, 30, 918273645, grpc.federation.time.UTC).truncate(grpc.federation.time.MILLISECOND)",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Date(2012, 12, 7, 12, 15, 30, 918273645, time.UTC).Truncate(time.Millisecond)
				if diff := cmp.Diff(gotV.Time, expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "utc",
			expr: "grpc.federation.time.now().utc()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Timestamp)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().UTC()
				if diff := cmp.Diff(gotV.Time, expected, cmpopts.EquateApproxTime(1*time.Second)); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "time_unix",
			expr: "grpc.federation.time.date(2012, 12, 7, 12, 15, 30, 0, grpc.federation.time.UTC).unix()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Date(2012, 12, 7, 12, 15, 30, 0, time.UTC).Unix()
				if diff := cmp.Diff(int64(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "time_unixMilli",
			expr: "grpc.federation.time.date(2012, 12, 7, 12, 15, 30, 0, grpc.federation.time.UTC).unixMilli()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Date(2012, 12, 7, 12, 15, 30, 0, time.UTC).UnixMilli()
				if diff := cmp.Diff(int64(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "time_unixMicro",
			expr: "grpc.federation.time.date(2012, 12, 7, 12, 15, 30, 0, grpc.federation.time.UTC).unixMicro()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Date(2012, 12, 7, 12, 15, 30, 0, time.UTC).UnixMicro()
				if diff := cmp.Diff(int64(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "time_unixNano",
			expr: "grpc.federation.time.date(2012, 12, 7, 12, 15, 30, 0, grpc.federation.time.UTC).unixNano()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Date(2012, 12, 7, 12, 15, 30, 0, time.UTC).UnixNano()
				if diff := cmp.Diff(int64(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "weekday",
			expr: "grpc.federation.time.now().weekday()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().Weekday()
				if diff := cmp.Diff(time.Weekday(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "year",
			expr: "grpc.federation.time.now().year()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().Year()
				if diff := cmp.Diff(int(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
		{
			name: "yearDay",
			expr: "grpc.federation.time.now().yearDay()",
			cmp: func(got any) error {
				gotV, ok := got.(types.Int)
				if !ok {
					return fmt.Errorf("invalid result type: %T", got)
				}
				expected := time.Now().YearDay()
				if diff := cmp.Diff(int(gotV), expected); diff != "" {
					return fmt.Errorf("(-got, +want)\n%s", diff)
				}
				return nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env, err := cel.NewEnv(cel.Lib(new(cellib.TimeLibrary)))
			if err != nil {
				t.Fatal(err)
			}
			ast, iss := env.Compile(test.expr)
			if iss.Err() != nil {
				t.Fatal(iss.Err())
			}
			program, err := env.Program(ast)
			if err != nil {
				t.Fatal(err)
			}
			out, _, err := program.Eval(test.args)
			if err != nil {
				t.Fatal(err)
			}
			if err := test.cmp(out); err != nil {
				t.Fatal(err)
			}
		})
	}
}
