package gtype

import (
	"fmt"
	"strings"
	"time"
)

const (
	timeFormat = "2006-01-02 15:04:05"
	dateFormat = "2006-01-02"
	zoneFormat = "2006-01-02T15:04:05.000Z"
)

func GetTimeLayout(v string) string {
	zoneIndex := strings.Index(v, "Z")
	if zoneIndex < 0 {
		zoneIndex = strings.Index(v, "+")
	}
	if zoneIndex < 0 {
		return ""
	}
	nanoIndex := strings.Index(v, ".")
	if zoneIndex < 0 {
		return ""
	}

	sb := &strings.Builder{}
	sb.WriteString(`"`)
	sb.WriteString("2006-01-02T15:04:05.")
	c := zoneIndex - nanoIndex - 1
	for i := 0; i < c; i++ {
		sb.WriteString("0")
	}
	sb.WriteString(v[zoneIndex:])

	return sb.String()
}

type DateTime time.Time

func (t *DateTime) UnmarshalJSON(data []byte) (err error) {
	var now time.Time
	dataLen := len(data)

	if dataLen == len(dateFormat)+2 {
		now, err = time.ParseInLocation(`"`+dateFormat+`"`, string(data), time.Local)
	} else if dataLen == len(timeFormat)+2 {
		now, err = time.ParseInLocation(`"`+timeFormat+`"`, string(data), time.Local)
	} else if dataLen == len(zoneFormat)+2 {
		now, err = time.ParseInLocation(`"`+zoneFormat+`"`, string(data), time.UTC)
	} else {
		now, err = time.Parse(time.RFC3339, string(data))
		if err != nil {
			now, err = time.Parse(time.RFC3339Nano, string(data))
			if err != nil {
				layout := GetTimeLayout(string(data))
				if len(layout) > 0 {
					now, err = time.ParseInLocation(layout, string(data), time.UTC)
				}
			}
		}
	}

	*t = DateTime(now)
	return
}

func (t DateTime) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(timeFormat)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, timeFormat)
	b = append(b, '"')
	return b, nil
}

func (t DateTime) String() string {
	return time.Time(t).Format(timeFormat)
}

func (t DateTime) Year() int {
	return time.Time(t).Year()
}

func (t DateTime) Month() time.Month {
	return time.Time(t).Month()
}

func (t DateTime) Day() int {
	return time.Time(t).Day()
}

func (t DateTime) Hour() int {
	return time.Time(t).Hour()
}

func (t DateTime) Minute() int {
	return time.Time(t).Minute()
}

func (t DateTime) Second() int {
	return time.Time(t).Second()
}

func (t DateTime) Duration() string {
	sb := &strings.Builder{}
	duration := time.Now().Sub(time.Time(t))

	days := time.Duration(0)
	if duration >= time.Hour*24 {
		days = duration / (time.Hour * 24)
		duration = duration - days*time.Hour*24
	}
	hours := time.Duration(0)
	if duration >= time.Hour {
		hours = duration / time.Hour
		duration = duration - hours*time.Hour
	}
	minutes := time.Duration(0)
	if duration >= time.Minute {
		minutes = duration / time.Minute
		duration = duration - minutes*time.Minute
	}
	seconds := time.Duration(0)
	if duration >= time.Second {
		seconds = duration / time.Second
		duration = duration - seconds*time.Second
	}

	if days > 0 {
		sb.WriteString(fmt.Sprintf("%d天", days))
		if hours == 0 {
			sb.WriteString("0时")
		}
	}
	if hours > 0 {
		sb.WriteString(fmt.Sprintf("%d时", hours))
		if minutes == 0 {
			sb.WriteString("0分")
		}
	}
	if minutes > 0 {
		sb.WriteString(fmt.Sprintf("%d分", minutes))
	}
	sb.WriteString(fmt.Sprintf("%d秒", seconds))

	return sb.String()
}

func (t DateTime) Elapse(start DateTime) string {
	end := time.Time(t)
	now := time.Time(start)

	nanosecond := end.Sub(now)
	millisecond := nanosecond / time.Millisecond

	hour := nanosecond / time.Hour
	minute := (nanosecond - hour*time.Hour) / time.Minute
	second := (nanosecond - hour*time.Hour - minute*time.Minute) / time.Second
	millisecond = (nanosecond - hour*time.Hour - minute*time.Minute - second*time.Second) / time.Millisecond
	elapse := fmt.Sprintf("%02d:%02d:%02d.%03d",
		hour,
		minute,
		second,
		millisecond)

	return elapse
}

func (t *DateTime) ToDate(plusDays int) *time.Time {
	date := time.Time(*t)
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	if plusDays != 0 {
		date = date.AddDate(0, 0, plusDays)
	}

	return &date
}

func (t *DateTime) ToTime(location *time.Location) *time.Time {
	v := time.Time(*t)
	l := location
	if l == nil {
		l = v.Location()
	}
	v = time.Date(v.Year(), v.Month(), v.Day(), v.Hour(), v.Minute(), v.Second(), v.Nanosecond(), l)

	return &v
}

func (t *DateTime) GetDays(now time.Time) int64 {
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	end := time.Time(*t)
	end = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	duration := end.Sub(start)
	days := duration / (time.Hour * 24)

	return int64(days)
}

func (t DateTime) After(u DateTime) bool {
	return time.Time(t).After(time.Time(u))
}

func (t DateTime) Before(u DateTime) bool {
	return time.Time(t).Before(time.Time(u))
}

func (t DateTime) Equal(u DateTime) bool {
	return time.Time(t).Equal(time.Time(u))
}

type Date time.Time

func (t *Date) UnmarshalJSON(data []byte) (err error) {
	var now time.Time
	dataLen := len(data)

	if dataLen == len(dateFormat)+2 {
		now, err = time.ParseInLocation(`"`+dateFormat+`"`, string(data), time.Local)
	} else if dataLen == len(timeFormat)+2 {
		now, err = time.ParseInLocation(`"`+timeFormat+`"`, string(data), time.Local)
	} else if dataLen == len(zoneFormat)+2 {
		now, err = time.ParseInLocation(`"`+zoneFormat+`"`, string(data), time.UTC)
	} else {
		now, err = time.Parse(time.RFC3339, string(data))
	}

	*t = Date(now)
	return
}

func (t Date) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(dateFormat)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, dateFormat)
	b = append(b, '"')
	return b, nil
}

func (t Date) String() string {
	return time.Time(t).Format(dateFormat)
}

func (t *Date) ToDate(plusDays int) *time.Time {
	date := time.Time(*t)
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	if plusDays != 0 {
		date = date.AddDate(0, 0, plusDays)
	}

	return &date
}

func (t *Date) GetDays(now time.Time) int64 {
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	end := time.Time(*t)
	end = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	duration := end.Sub(start)
	days := duration / (time.Hour * 24)

	return int64(days)
}
