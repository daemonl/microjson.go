package microjson

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var Timezone *time.Location = time.UTC
var DateTimeFormat = time.RFC3339
var DateTimeNanoFormat = time.RFC3339Nano

type DateTime time.Time

var rePostgresDateTime = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{0,6}$`)
var rePostgresShortDateTime = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}$`)

func (dt *DateTime) UnmarshalJSON(d []byte) error {
	val := ""
	if err := json.Unmarshal(d, &val); err != nil {
		return err
	}

	format := DateTimeFormat

	if rePostgresDateTime.MatchString(val) {
		//val = val + "Z00:00"
		format = "2006-01-02T15:04:05.999999"
	}

	if rePostgresShortDateTime.MatchString(val) {
		format = "2006-01-02T15:04:05"
	}

	t, err := time.Parse(format, val)
	if err != nil {
		return err
	}
	*dt = DateTime(t)
	return nil
}

func (dt DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(dt).In(Timezone).Format(DateTimeFormat))
}

func (dt DateTime) Value() (driver.Value, error) {
	return driver.Value(time.Time(dt).In(time.UTC)), nil
}

type DateTimeNano time.Time

var rePostgresDateTimeNano = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{0,6}$`)
var rePostgresShortDateTimeNano = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}$`)

func (dt *DateTimeNano) UnmarshalJSON(d []byte) error {
	val := ""
	if err := json.Unmarshal(d, &val); err != nil {
		return err
	}

	format := DateTimeNanoFormat

	if rePostgresDateTimeNano.MatchString(val) {
		//val = val + "Z00:00"
		format = "2006-01-02T15:04:05.999999"
	}

	if rePostgresShortDateTimeNano.MatchString(val) {
		format = "2006-01-02T15:04:05"
	}

	t, err := time.Parse(format, val)
	if err != nil {
		return err
	}
	*dt = DateTimeNano(t)
	return nil
}

func (dt DateTimeNano) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(dt).In(Timezone).Format(DateTimeNanoFormat))
}

func (dt DateTimeNano) Value() (driver.Value, error) {
	return driver.Value(time.Time(dt).In(time.UTC)), nil
}

type Date struct {
	Year  int
	Month int
	Day   int
}

func (date Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", date.Year, date.Month, date.Day)
}

func (date Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(date.String())
}

func (date *Date) UnmarshalJSON(data []byte) error {
	str := ""
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	d, ok := ParseDate(str)
	if !ok {
		return fmt.Errorf("invalid date format %s", str)
	}
	date.Year = d.Year
	date.Month = d.Month
	date.Day = d.Day

	return nil

}

func (date Date) Value() (driver.Value, error) {
	return driver.Value(date.String()), nil
}

var reDatePartsRFC3339 = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)

func ParseDate(str string) (Date, bool) {

	matchParts := reDatePartsRFC3339.FindStringSubmatch(str)
	if len(matchParts) != 4 {
		return Date{}, false
	}

	year, _ := strconv.ParseInt(matchParts[1], 10, 64)
	month, _ := strconv.ParseInt(matchParts[2], 10, 64)
	day, _ := strconv.ParseInt(matchParts[3], 10, 64)
	return Date{
		Year:  int(year),
		Month: int(month),
		Day:   int(day),
	}, true
}

func TimeToDate(in time.Time) Date {
	return Date{
		Year:  in.Year(),
		Month: int(in.Month()),
		Day:   in.Day(),
	}
}
