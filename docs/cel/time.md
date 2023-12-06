# grpc.federation.time APIs

The API for this package was created based on Go's [`time`](https://pkg.go.dev/time) package.

# Index

## Constants

- [`LAYOUT`](#layout)
- [`ANSIC`](#ansic)
- [`UNIX_DATE`](#unix_date)
- [`RUBY_DATE`](#ruby_date)
- [`RFC822`](#rfc822)
- [`RFC822Z`](#rfc822z)
- [`RFC850`](#rfc850)
- [`RFC1123`](#rfc1123)
- [`RFC1123Z`](#rfc1123z)
- [`RFC3339`](#rfc3339)
- [`RFC3339NANO`](#rfc3339nano)
- [`KITCHEN`](#kitchen)
- [`STAMP`](#stamp)
- [`STAMP_MILLI`](#stamp_milli)
- [`STAMP_MICRO`](#stamp_micro)
- [`STAMP_NANO`](#stamp_nano)
- [`DATETIME`](#datetime)
- [`DATE_ONLY`](#date_only)
- [`TIME_ONLY`](#time_only)
- [`NANOSECOND`](#nanosecond)
- [`MICROSECOND`](#microsecond)
- [`MILLISECOND`](#millisecond)
- [`SECOND`](#second)
- [`MINUTE`](#minute)
- [`HOUR`](#hour)
- [`LOCAL`](#local)
- [`UTC`](#utc)

## Duration

- [`toDuration`](#toduration)
- [`parseDuration`](#parseduration)
- [`since`](#since)
- [`until`](#until)
- [`Duration.abs`](#durationabs)
- [`Duration.hours`](#durationhours)
- [`Duration.microseconds`](#durationmicroseconds)
- [`Duration.milliseconds`](#durationmilliseconds)
- [`Duration.minutes`](#durationminutes)
- [`Duration.nanoseconds`](#durationnanoseconds)
- [`Duration.round`](#durationround)
- [`Duration.seconds`](#durationseconds)
- [`Duration.string`](#durationstring)
- [`Duration.truncate`](#durationtruncate)

## Location

- [`fixedZone`](#fixedzone)
- [`loadLocation`](#loadlocation)
- [`loadLocationFromTZData`](#loadlocationfromtzdata)
- [`Location.string`](#locationstring)

## Time

- [`date`](#date)
- [`now`](#now)
- [`parse`](#parse)
- [`unix`](#unix)
- [`unixMicro`](#unixmicro)
- [`unixMilli`](#unixmilli)
- [`Time.add`](#timeadd)
- [`Time.addDate`](#timeadddate)
- [`Time.after`](#timeafter)
- [`Time.appendFormat`](#timeappendformat)
- [`Time.before`](#timebefore)
- [`Time.compare`](#timecompare)
- [`Time.day`](#timeday)
- [`Time.equal`](#timeequal)
- [`Time.format`](#timeformat)
- [`Time.hour`](#timehour)
- [`Time.withLocation`](#timewithlocation)
- [`Time.isDST`](#timeisdst)
- [`Time.isZero`](#timeiszero)
- [`Time.local`](#timelocal)
- [`Time.location`](#timelocation)
- [`Time.minute`](#timeminute)
- [`Time.month`](#timemonth)
- [`Time.nanosecond`](#timenanosecond)
- [`Time.round`](#timeround)
- [`Time.second`](#timesecond)
- [`Time.string`](#timestring)
- [`Time.sub`](#timesub)
- [`Time.truncate`](#timetruncate)
- [`Time.utc`](#timeutc)
- [`Time.unix`](#timeunix)
- [`Time.unixMicro`](#timeunixmicro)
- [`Time.unixMilli`](#timeunixmilli)
- [`Time.unixNano`](#timeunixnano)
- [`Time.weekday`](#timeweekday)
- [`Time.year`](#timeyear)
- [`Time.yearDay`](#timeyearday)

# Types

## Duration

A `Duration` represents the elapsed time between two instants as an int64 nanosecond count. The representation limits the largest representable duration to approximately 290 years.
`Duration` type is compatible with [`google.protobuf.Duration`](https://github.com/protocolbuffers/protobuf/blob/main/src/google/protobuf/duration.proto) type.

FYI: https://pkg.go.dev/time#Duration

## Location

A `Location` maps time instants to the zone in use at that time. Typically, the Location represents the collection of time offsets in use in a geographical area. For many Locations the time offset varies depending on whether daylight savings time is in use at the time instant.

FYI: https://pkg.go.dev/time#Location

## Time

A `Time` represents an instant in time with nanosecond precision. `Time` type is compatible with [`google.protobuf.Timestamp`](https://github.com/protocolbuffers/protobuf/blob/main/src/google/protobuf/timestamp.proto) type.

# Constants

FYI: https://pkg.go.dev/time#pkg-constants

## LAYOUT

### type

`string`

### value

`01/02 03:04:05PM '06 -0700`

## ANSIC

### type

`string`

### value

`Mon Jan _2 15:04:05 2006`

## UNIX_DATE

### type

`string`

### value

`Mon Jan _2 15:04:05 MST 2006`

## RUBY_DATE

### type

`string`

### value

`Mon Jan 02 15:04:05 -0700 2006`

## RFC822

### type

`string`

### value

`02 Jan 06 15:04 MST`

## RFC822Z

### type

`string`

### value

`02 Jan 06 15:04 -0700`

## RFC850

### type

`string`

### value

`Monday, 02-Jan-06 15:04:05 MST`

## RFC1123

### type

`string`

### value

`Mon, 02 Jan 2006 15:04:05 MST`

## RFC1123Z

### type

`string`

### value

`Mon, 02 Jan 2006 15:04:05 -0700`

## RFC3339

### type

`string`

### value

`2006-01-02T15:04:05Z07:00`

## RFC3339NANO

### type

`string`

### value

`2006-01-02T15:04:05.999999999Z07:00`

## KITCHEN

### type

`string`

### value

`3:04PM`

## STAMP

### type

`string`

### value

`Jan _2 15:04:05`

## STAMP_MILLI

### type

`string`

### value

`Jan _2 15:04:05.000`

## STAMP_MICRO

### type

`string`

### value

`Jan _2 15:04:05.000000`

## STAMP_NANO

### type

`string`

### value

`Jan _2 15:04:05.000000000`

## DATETIME

### type

`string`

### value

`2006-01-02 15:04:05`

## DATE_ONLY

### type

`string`

### value

`2006-01-02`

## TIME_ONLY

### type

`string`

### value

`15:04:05`

## NANOSECOND

### type

`int`

### value

`1`

## MICROSECOND

### type

`int`

### value

`1000 * grpc.federation.time.NANOSECOND`

## MILLISECOND

### type

`int`

### value

`1000 * grpc.federation.time.MICROSECOND`

## SECOND

### type

`int`

### value

`1000 * grpc.federation.time.MILLISECOND`

## MINUTE

### type

`int`

### value

`60 * grpc.federation.time.SECOND`

## HOUR

### type

`int`

### value

`60 * grpc.federation.time.MINUTE`

## LOCAL

### type

`Location`

### value

`LOCAL` represents the system's local time zone. On Unix systems, `LOCAL` consults the TZ environment variable to find the time zone to use. No TZ means use the system default /etc/localtime. TZ="" means use UTC. TZ="foo" means use file foo in the system timezone directory.

## UTC

### type

`Location`

### value

`UTC` represents Universal Coordinated Time (UTC).

# Functions

## toDuration

`toDuration` converts int type to `Duration` type.

### Parameters

`toDuration(var int) Duration`

- `var`: source value

### Examples

```cel
grpc.federation.time.toDuration(10 * grpc.federation.time.SECOND) // 10s
```

## parseDuration

`parseDuration` parses a duration string. A duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

FYI: https://pkg.go.dev/time#ParseDuration

### Parameters

`parseDuration(s string) Duration`

- `s`: duration string

### Examples

```cel
grpc.federation.time.parseDuration("1h10m10s")
```

## since

`since` returns the time elapsed since `t`. It is shorthand for `now().sub(t)`.

FYI: https://pkg.go.dev/time#Since

### Parameters

`since(t Time) Duration`

- `t`: time value

### Examples

```cel
grpc.federation.time.since(grpc.federation.time.now())
```

## until

`until` returns the duration until `t`. It is shorthand for `t.sub(now())`.

### Parameters

`until(t Time) Duration`

- `t`: time value

### Examples

```cel
grpc.federation.time.until(grpc.federation.time.now())
```

## Duration.abs

`abs` returns the absolute value of `Duration`. As a special case, `<min-int-value>` is converted to `<max-int-value>`.

FYI: https://pkg.go.dev/time#Duration.Abs

### Parameters

`Duration.abs() Duration`

## Duration.hours

`hours` returns the duration as a double type number of hours.

FYI: https://pkg.go.dev/time#Duration.Hours

### Parameters

`Duration.hours() double`

### Examples

```cel
grpc.federation.time.parseDuration("4h30m").hours() //=> 4.5
```

## Duration.microseconds

`microseconds` returns the duration as an integer microsecond count.

FYI: https://pkg.go.dev/time#Duration.Microseconds

### Parameters

`Duration.microseconds() int`

### Examples

```cel
grpc.federation.time.parseDuration("1s").microseconds() //=> 1000000
```

## Duration.milliseconds

`milliseconds` returns the duration as an integer millisecond count.

FYI: https://pkg.go.dev/time#Duration.Milliseconds

### Parameters

`Duration.milliseconds() int`

### Examples

```cel
grpc.federation.time.parseDuration("1s").milliseconds() //=> 1000
```

## Duration.minutes

`minutes` returns the duration as a double type number of minutes.

FYI: https://pkg.go.dev/time#Duration.Minutes

### Parameters

`Duration.minutes() double`

### Examples

```cel
grpc.federation.time.parseDuration("1h30m").minutes() //=> 90
```

## Duration.nanoseconds

`nanoseconds` returns the duration as an integer nanosecond count.

FYI: https://pkg.go.dev/time#Duration.Nanoseconds

### Parameters

`Duration.nanoseconds() int`

### Examples

```cel
grpc.federation.time.parseDuration("1µs").nanoseconds() //=> 1000
```

## Duration.round

`round` returns the result of rounding `Duration` to the nearest multiple of `m`. The rounding behavior for halfway values is to round away from zero. If the result exceeds the maximum (or minimum) value that can be stored in a Duration, `round` returns the maximum (or minimum) duration. If `m` <= 0, `round` returns `Duration` unchanged.

### Parameters

`Duration.round(m Duration) Duration`

- `m`: the value for round operation

### Examples

```cel
grpc.federation.time.parseDuration("1h15m30.918273645s").round(grpc.federation.time.MILLISECOND) //=> 1h15m30.918s
```

## Duration.seconds

`seconds` returns the duration as a double type number of seconds.

FYI: https://pkg.go.dev/time#Duration.Seconds

### Parameters

`Duration.seconds() double`

### Examples

```cel
grpc.federation.time.parseDuration("1m30s").seconds() //=> 90
```

## Duration.string

`string` returns a string representing the duration in the form "72h3m0.5s". Leading zero units are omitted. As a special case, durations less than one second format use a smaller unit (milli-, micro-, or nanoseconds) to ensure that the leading digit is non-zero. The zero duration formats as 0s.

FYI: https://pkg.go.dev/time#Duration.String

### Parameters

`Duration.string() string`

### Examples

```cel
grpc.federation.time.toDuration(grpc.federation.time.HOUR + 2*grpc.federatino.time.MINUTE).string() //=> 1h2m
```

## Duration.truncate

`truncate` returns the result of rounding d toward zero to a multiple of `m`. If `m` <= 0, Truncate returns `Duration` unchanged.

FYI: https://pkg.go.dev/time#Duration.Truncate

### Parameters

`Duration.truncate(m Duration) Duration`

- `m`: the value for truncate operation

### Examples

```cel
grpc.federation.time.parseDuration("1h15m30.918273645s").truncate(grpc.federation.time.MILLISECOND) //=> 1h15m30.918s
```

## fixedZone

`fixedZone` returns a `Location` that always uses the given zone name and offset (seconds east of `UTC`).

FYI: https://pkg.go.dev/time#FixedZone

### Parameters

`fixedZone(name string, offset int) Location`

- `name`: the zone name
- `offset`: the seconds east of `UTC`

### Examples

```cel
grpc.federation.time.fixedZone("UTC-8", -8*60*60)
```

## loadLocation

`loadLocation` returns the `Location` with the given name.

If the name is `""` or `"UTC"`, `loadLocation` returns `UTC`. If the name is `"Local"`, `loadLocation` returns `LOCAL`.

Otherwise, the name is taken to be a location name corresponding to a file in the IANA Time Zone database, such as `"America/New_York"`.

FYI: https://pkg.go.dev/time#LoadLocation

### Parameters

`loadLocation(name string) Location`

### Examples

```cel
grpc.federation.time.loadLocation("America/Los_Angeles")
```

## loadLocationFromTZData

`loadLocationFromTZData` returns a `Location` with the given name initialized from the IANA Time Zone database-formatted data. The data should be in the format of a standard IANA time zone file (for example, the content of /etc/localtime on Unix systems).

FYI: https://pkg.go.dev/time#LoadLocationFromTZData

### Parameters

`loadLocationFromTZData(name string data bytes) Location`

- `name`: the zone name
- `data`: the IANA Time Zone database-formatted data

## Location.string

`string` returns a descriptive name for the time zone information, corresponding to the name argument to `loadLocation` or `fixedZone`.

FYI: https://pkg.go.dev/time#Location.String

### Parameters

`Location.string() string`

## date

`date` returns the `Time` corresponding to

```
yyyy-mm-dd hh:mm:ss + nsec nanoseconds
```

in the appropriate zone for that time in the given location.

FYI: https://pkg.go.dev/time#Date

### Parameters

`date(year int, month int, day int, hour int, min int, sec int, nsec int, loc Location) Time`

- `year`: the year number
- `month`: the month number
- `day`: the day number
- `hour`: the hour number
- `min`: the minute number
- `sec`: the second number
- `nsec`: the nanosecond number
- `loc`: the Location value

### Examples

```cel
grpc.federation.time.date(2009, time.November, 10, 23, 0, 0, 0, grpc.federation.time.UTC) //=> 2009-11-10 23:00:00 +0000 UTC
```

## now

`now` returns the current local time.

FYI: https://pkg.go.dev/time#Now

### Parameters

`now() Time`

### Examples

```cel
grpc.federation.time.now()
```

## parse

`parse` parses a formatted string and returns the time value it represents. See the documentation for the constant called `LAYOUT` to see how to represent the format. The second argument must be parseable using the format string (layout) provided as the first argument.

FYI: https://pkg.go.dev/time#Parse

### Parameters

`parse(layout string, value string) Time`

### Examples

```cel
grpc.federation.time.parse("2006-Jan-02", "2013-Feb-03")
```

## parseInLocation

`parseInLocation` is like `parse` but differs in two important ways. First, in the absence of time zone information, `parse` interprets a time as `UTC`; `parseInLocation` interprets the time as in the given location. Second, when given a zone offset or abbreviation, `parse` tries to match it against the `LOCAL` location; `parseInLocation` uses the given location.

FYI: https://pkg.go.dev/time#ParseInLocation

### Parameters

`parseInLocation(layout string, value string, loc Location) Time`

### Examples

```cel
grpc.federation.time.parseInLocation("Jan 2, 2006 at 3:04pm (MST)", "Jul 9, 2012 at 5:02am (CEST)", grpc.federation.time.loadLocation("Europe/Berlin"))
```

## unix

`unix` returns the local Time corresponding to the given Unix time, sec seconds and nsec nanoseconds since January 1, 1970 UTC. It is valid to pass nsec outside the range [0, 999999999]. Not all sec values have a corresponding time value. One such value is 1<<63-1 (the largest int64 value).

FYI: https://pkg.go.dev/time#Unix

### Parameters

`unix(sec int, nsec int) Time`

- `sec`: the second value
- `nsec`: the nanosecond value

## unixMicro

`unixMicro` returns the local Time corresponding to the given Unix time, usec microseconds since January 1, 1970 UTC.

FYI: https://pkg.go.dev/time#UnixMicro

### Parameters

`unixMicro(usec int) Time`

- `usec`: microseconds since January 1, 1970 UTC.

## unixMilli

`unixMilli` returns the local Time corresponding to the given Unix time, msec milliseconds since January 1, 1970 UTC.

FYI: https://pkg.go.dev/time#UnixMilli

### Parameters

`unixMilli(msec int) Time`

- `msec`: milliseconds since January 1, 1970 UTC.

## Time.add

`add` returns the time `Time`+`d`.

FYI: https://pkg.go.dev/time#Time.Add

### Parameters

`Time.add(d Duration) Time`

- `d`: Duration value

## Time.addDate

`addDate` returns the time corresponding to adding the given number of years, months, and days to `Time`. For example, `addDate(-1, 2, 3)` applied to January 1, 2011 returns March 4, 2010.  

`addDate` normalizes its result in the same way that Date does, so, for example, adding one month to October 31 yields December 1, the normalized form for November 31.

FYI: https://pkg.go.dev/time#Time.AddDate

### Parameters

`Time.addDate(years int, months int, days int) Time`

- `years`: year number
- `months`: month number
- `days`: day number

## Time.after

`after` reports whether the time instant `Time` is after `u`.

FYI: https://pkg.go.dev/time#Time.After

### Parameters

`Time.after(u Time) bool`

- `u`: the time value to compare

### Examples

```cel
grpc.federation.time.date(3000, 1, 1, 0, 0, 0, 0, grpc.federation.time.UTC).after(grpc.federation.time.date(2000, 1, 1, 0, 0, 0, 0, grpc.federation.time.UTC)) //=> true
```

## Time.appendFormat

`appendFormat` is like Format but appends the textual representation to `b` and returns the extended buffer.

FYI: https://pkg.go.dev/time#Time.AppendFormat

### Parameters

- `Time.appendFormat(b bytes, layout string) string`
- `Time.appendFormat(s string, layout string) string`

### Examples

```cel
grpc.federation.time.date(2017, 11, 4, 11, 0, 0, 0, grpc.federation.time.UTC).appendFormat("Time: ", grpc.federation.time.KITCHEN) //=> Time: 11:00AM
```

## Time.before

`before` reports whether the time instant `Time` is before `u`.

FYI: https://pkg.go.dev/time#Time.Before

### Parameters

`Time.before(u Time) bool`

- `u`: the time value to compare

### Examples

```cel
grpc.federation.time.date(2000, 1, 1, 0, 0, 0, 0, grpc.federation.time.UTC).before(grpc.federation.time.date(3000, 1, 1, 0, 0, 0, 0, grpc.federation.time.UTC)) //=> true
```

## Time.compare

`compare` compares the time instant `Time` with `u`. If `Time` is before `u`, it returns `-1`; if `Time` is after `u`, it returns `+1`; if they're the same, it returns `0`.

FYI: https://pkg.go.dev/time#Time.Compare

### Parameters

`Time.compare(u Time) int`

- `u`: the time value to compare

## Time.day

`day` returns the day of the month specified by `Time`.

FYI: https://pkg.go.dev/time#Time.Day

### Parameters

`Time.day() int`

### Examples

```cel
grpc.federation.time.date(2000, 2, 1, 12, 30, 0, 0, grpc.federation.time.UTC).day() //=> 1
```

## Time.equal

`equal` reports whether `Time` and `u` represent the same time instant. Two times can be equal even if they are in different locations. For example, 6:00 +0200 and 4:00 UTC are Equal. See the documentation on the Time type for the pitfalls of using == with Time values; most code should use `equal` instead.

FYI: https://pkg.go.dev/time#Time.Equal

### Parameters

`Time.equal(u Time) bool`

### Examples

```cel
grpc.federation.time.date(2000, 2, 1, 12, 30, 0, 0, grpc.federation.time.UTC).equal(grpc.federation.time.date(2000, 2, 1, 20, 30, 0, 0, grpc.federation.time.fixedZone('Beijing Time', grpc.federation.time.toDuration(8 * grpc.federation.time.HOUR).seconds()))) //=> true
```

## Time.format

`format` returns a textual representation of the time value formatted according to the layout defined by the argument. See the documentation for the constant called Layout to see how to represent the layout format.

FYI: https://pkg.go.dev/time#Time.Format

### Parameters

`Time.format(layout string) string`

- `layout`: the layout value to format

## Time.hour

`hour` returns the hour within the day specified by `Time`, in the range [0, 23].

FYI: https://pkg.go.dev/time#Time.Hour

### Parameters

`Time.hour() int`

## Time.withLocation

`withLocation` returns a copy of `Time` representing the same time instant, but with the copy's location information set to loc for display purposes.

FYI: https://pkg.go.dev/time#Time.In

### Parameters

`Time.withLocation(loc Location) Time`

## Time.isDST

`isDST` reports whether the time in the configured location is in Daylight Savings Time.

FYI: https://pkg.go.dev/time#Time.IsDST

### Parameters

`Time.isDST() bool`

## Time.isZero

`isZero` reports whether `Time` represents the zero time instant, January 1, year 1, 00:00:00 UTC.

FYI: https://pkg.go.dev/time#Time.IsZero

### Parameters

`Time.isZero() bool`

## Time.local

`local` returns `Time` with the location set to local time.

FYI: https://pkg.go.dev/time#Time.Local

### Parameters

`Time.local() Time`

## Time.location

`location` returns the time zone information associated with `Time`.

FYI: https://pkg.go.dev/time#Time.Location

### Parameters

`Time.location() Location`

## Time.minute

`minute` returns the minute offset within the hour specified by `Time`, in the range [0, 59].

FYI: https://pkg.go.dev/time#Time.Minute

### Parameters

`Time.minute() int`

## Time.month

`month` returns the month of the year specified by `Time`.

FYI: https://pkg.go.dev/time#Time.Month

### Parameters

`Time.month() int`

## Time.nanosecond

`nanosecond` returns the nanosecond offset within the second specified by `Time`, in the range [0, 999999999].

FYI: https://pkg.go.dev/time#Time.Nanosecond

### Parameters

`Time.nanosecond() int`

## Time.round

`round` returns the result of rounding `Time` to the nearest multiple of `d` (since the zero time). The rounding behavior for halfway values is to round up. If d <= 0, `round` returns `Time` stripped of any monotonic clock reading but otherwise unchanged.

`round` operates on the time as an absolute duration since the zero time; it does not operate on the presentation form of the time. Thus, `round(grpc.federation.time.HOUR)` may return a time with a non-zero minute, depending on the time's Location.

FYI: https://pkg.go.dev/time#Time.Round

### Parameters

`Time.round(d Duration) Time`

- `d`: the duration value

### Examples

```cel
grpc.federation.time.date(0, 0, 0, 12, 15, 30, 918273645, grpc.federation.time.UTC).round(grpc.federation.time.MILLISECOND).format("15:04:05.999999999") //=> 12:15:30.918
```

## Time.second

`second` returns the second offset within the minute specified by `Time`, in the range [0, 59].

FYI: https://pkg.go.dev/time#Time.Second

### Parameters

`Time.second() int`

## Time.string

`string` returns the time formatted using the format string.

```
"2006-01-02 15:04:05.999999999 -0700 MST"
```

If the time has a monotonic clock reading, the returned string includes a final field "m=±<value>", where value is the monotonic clock reading formatted as a decimal number of seconds.

FYI: https://pkg.go.dev/time#Time.String

### Parameters

`Time.string() string`

## Time.sub

`sub` returns the duration `Time`-`u`. If the result exceeds the maximum (or minimum) value that can be stored in a Duration, the maximum (or minimum) duration will be returned. To compute `Time`-`d` for a duration `d`, use `Time.add(-d)`.

FYI: https://pkg.go.dev/time#Time.Sub

### Parameters

`Time.sub(u Time) Duration`

- `u`: the time value

### Examples

```cel
grpc.federation.time.date(2000, 1, 1, 12, 0, 0, 0, grpc.federation.time.UTC).sub(grpc.federation.time.date(2000, 1, 1, 0, 0, 0, 0, grpc.federation.time.UTC)) //=> 12h0m0s
```

## Time.truncate

`truncate` returns the result of rounding `Time` down to a multiple of `d` (since the zero time). If `d` <= 0, `truncate` returns `Time` stripped of any monotonic clock reading but otherwise unchanged.

`truncate` operates on the time as an absolute duration since the zero time; it does not operate on the presentation form of the time. Thus, `truncate(grpc.federation.time.HOUR)` may return a time with a non-zero minute, depending on the time's Location.

FYI: https://pkg.go.dev/time#Time.Truncate

### Parameters

`Time.truncate(d Duration) Time`

- `d`: since the zero time

### Examples

```cel
grpc.federation.time.parse("2006 Jan 02 15:04:05", "2012 Dec 07 12:15:30.918273645").truncate(grpc.federation.time.MILLISECOND).format("15:04:05.999999999") //=> 12:15:30.918
```

## Time.utc

`utc` returns `Time` with the location set to UTC.

FYI: https://pkg.go.dev/time#Time.UTC

### Parameters

`Time.utc() Time`

## Time.unix

`unix` returns t as a Unix time, the number of seconds elapsed since January 1, 1970 UTC. The result does not depend on the location associated with `Time`. Unix-like operating systems often record time as a 32-bit count of seconds, but since the method here returns a 64-bit value it is valid for billions of years into the past or future.

FYI: https://pkg.go.dev/time#Time.Unix

### Parameters

`Time.unix() int`

## Time.unixMicro

`unixMicro` returns `Time` as a Unix time, the number of microseconds elapsed since January 1, 1970 UTC. The result is undefined if the Unix time in microseconds cannot be represented by an int64 (a date before year -290307 or after year 294246). The result does not depend on the location associated with `Time`.

FYI: https://pkg.go.dev/time#Time.UnixMicro

### Parameters

`Time.unixMicro() int`

## Time.unixMilli

`unixMilli` returns `Time` as a Unix time, the number of milliseconds elapsed since January 1, 1970 UTC. The result is undefined if the Unix time in milliseconds cannot be represented by an int64 (a date more than 292 million years before or after 1970). The result does not depend on the location associated with `Time`.

FYI: https://pkg.go.dev/time#Time.UnixMilli

### Parameters

`Time.unixMilli() int`

## Time.unixNano

`unixNano` returns `Time` as a Unix time, the number of nanoseconds elapsed since January 1, 1970 UTC. The result is undefined if the Unix time in nanoseconds cannot be represented by an int64 (a date before the year 1678 or after 2262). Note that this means the result of calling UnixNano on the zero Time is undefined. The result does not depend on the location associated with `Time`.

FYI: https://pkg.go.dev/time#Time.UnixNano

### Parameters

`Time.unixNano() int`

## Time.weekday

`weekday` returns the day of the week specified by `Time`.

FYI: https://pkg.go.dev/time#Time.Weekday

### Parameters

`Time.weekday() int`

## Time.year

`year` returns the year in which `Time` occurs.

FYI: https://pkg.go.dev/time#Time.Year

### Parameters

`Time.year() int`

## Time.yearDay

`yearDay` returns the day of the year specified by `Time`, in the range [1,365] for non-leap years, and [1,366] in leap years.

FYI: https://pkg.go.dev/time#Time.YearDay

### Parameters

`Time.yearDay() int`
