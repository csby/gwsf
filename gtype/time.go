package gtype

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"strings"
	"time"
)

const (
	timeFormat = "2006-01-02 15:04:05"
	dateFormat = "2006-01-02"
	zoneFormat = "2006-01-02T15:04:05.000Z"

	xmlDateFormat = "20060102"
	xmlTimeFormat = "20060102150405"
)

type DateTime time.Time

func (t *DateTime) UnmarshalText(data []byte) (err error) {
	var now time.Time
	dataLen := len(data)
	if dataLen == len(dateFormat)+2 {
		now, err = time.ParseInLocation(`"`+dateFormat+`"`, string(data), time.Local)
	} else if dataLen == len(timeFormat)+2 {
		now, err = time.ParseInLocation(`"`+timeFormat+`"`, string(data), time.Local)
	} else if dataLen == len(zoneFormat)+2 {
		now, err = time.ParseInLocation(`"`+zoneFormat+`"`, string(data), time.UTC)
	} else if dataLen == len(xmlTimeFormat) {
		now, err = time.ParseInLocation(``+xmlTimeFormat+``, string(data), time.Local)
	} else if dataLen == len(xmlDateFormat) {
		now, err = time.ParseInLocation(``+xmlDateFormat+``, string(data), time.Local)
	} else {
		now, err = time.Parse(time.RFC3339, string(data))
	}

	*t = DateTime(now)

	return
}

func (t DateTime) MarshalText() ([]byte, error) {
	text := time.Time(t).Format(xmlTimeFormat)
	return []byte(text), nil
}

func (t *DateTime) UnmarshalJSON(data []byte) (err error) {
	var now time.Time
	dataLen := len(data)

	if dataLen == len(dateFormat)+2 {
		now, err = time.ParseInLocation(`"`+dateFormat+`"`, string(data), time.Local)
	} else if dataLen == len(timeFormat)+2 {
		now, err = time.ParseInLocation(`"`+timeFormat+`"`, string(data), time.Local)
	} else if dataLen == len(zoneFormat)+2 {
		now, err = time.ParseInLocation(`"`+zoneFormat+`"`, string(data), time.UTC)
	} else if dataLen == len(xmlTimeFormat) {
		now, err = time.ParseInLocation(``+xmlTimeFormat+``, string(data), time.Local)
	} else if dataLen == len(xmlDateFormat) {
		now, err = time.ParseInLocation(``+xmlDateFormat+``, string(data), time.Local)
	} else {
		now, err = time.Parse(time.RFC3339, string(data))
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

func (t *DateTime) UnmarshalBSON(data []byte) (err error) {
	var now time.Time
	err = bson.Unmarshal(data, &now)

	*t = DateTime(now)
	return
}

func (t DateTime) MarshalBSON() ([]byte, error) {
	v := time.Time(t)
	return bson.Marshal(&v)
}

func (t *DateTime) UnmarshalBSONValue(dataType bsontype.Type, data []byte) error {
	rv := bson.RawValue{
		Type:  dataType,
		Value: data,
	}
	now := time.Time{}
	err := rv.Unmarshal(&now)
	if err == nil {
		*t = DateTime(now)
	}
	return err
}

func (t *DateTime) MarshalBSONValue() (bsontype.Type, []byte, error) {
	v := time.Time(*t)
	return bson.MarshalValue(&v)
}

func (t DateTime) String() string {
	return time.Time(t).Format(timeFormat)
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

func (t *DateTime) ToDate(plusDays int) *time.Time {
	date := time.Time(*t)
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	if plusDays != 0 {
		date = date.AddDate(0, 0, plusDays)
	}

	return &date
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

func (t *Date) UnmarshalText(data []byte) (err error) {
	var now time.Time
	dataLen := len(data)
	if dataLen == len(dateFormat)+2 {
		now, err = time.ParseInLocation(`"`+dateFormat+`"`, string(data), time.Local)
	} else if dataLen == len(timeFormat)+2 {
		now, err = time.ParseInLocation(`"`+timeFormat+`"`, string(data), time.Local)
	} else if dataLen == len(zoneFormat)+2 {
		now, err = time.ParseInLocation(`"`+zoneFormat+`"`, string(data), time.UTC)
	} else if dataLen == len(xmlTimeFormat) {
		now, err = time.ParseInLocation(``+xmlTimeFormat+``, string(data), time.Local)
	} else if dataLen == len(xmlDateFormat) {
		now, err = time.ParseInLocation(``+xmlDateFormat+``, string(data), time.Local)
	} else {
		now, err = time.Parse(time.RFC3339, string(data))
	}

	*t = Date(now)

	return
}

func (t Date) MarshalText() ([]byte, error) {
	text := time.Time(t).Format(xmlDateFormat)
	return []byte(text), nil
}

func (t *Date) UnmarshalJSON(data []byte) (err error) {
	var now time.Time
	dataLen := len(data)

	if dataLen == len(dateFormat)+2 {
		now, err = time.ParseInLocation(`"`+dateFormat+`"`, string(data), time.Local)
	} else if dataLen == len(timeFormat)+2 {
		now, err = time.ParseInLocation(`"`+timeFormat+`"`, string(data), time.Local)
	} else if dataLen == len(zoneFormat)+2 {
		now, err = time.ParseInLocation(`"`+zoneFormat+`"`, string(data), time.UTC)
	} else if dataLen == len(xmlTimeFormat) {
		now, err = time.ParseInLocation(``+xmlTimeFormat+``, string(data), time.Local)
	} else if dataLen == len(xmlDateFormat) {
		now, err = time.ParseInLocation(``+xmlDateFormat+``, string(data), time.Local)
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

func (t *Date) UnmarshalBSON(data []byte) (err error) {
	var now time.Time
	err = bson.Unmarshal(data, &now)

	*t = Date(now)
	return
}

func (t Date) MarshalBSON() ([]byte, error) {
	v := time.Time(t)
	return bson.Marshal(&v)
}

func (t *Date) UnmarshalBSONValue(dataType bsontype.Type, data []byte) error {
	rv := bson.RawValue{
		Type:  dataType,
		Value: data,
	}
	now := time.Time{}
	err := rv.Unmarshal(&now)
	if err == nil {
		*t = Date(now)
	}
	return err
}

func (t *Date) MarshalBSONValue() (bsontype.Type, []byte, error) {
	v := time.Time(*t)
	return bson.MarshalValue(&v)
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
