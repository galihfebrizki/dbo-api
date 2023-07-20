package gorm

import (
	"database/sql/driver"
	"errors"
	"strings"
	"time"
)

const (
	DateTimeLayout     = "2006-01-02 15:04:05"
	DateTimezoneLayout = "2006-01-02 15:04:05 MST"
	TimeLayout         = "15:04:05"
	TimeLayoutHHMM     = "15:04"
)

type Time time.Time

func (o *Time) UnmarshalJSON(b []byte) error {
	value := strings.Trim(string(b), `"`) //get rid of "
	if value == "" || value == "null" {
		return nil
	}

	t, err := time.Parse(TimeLayout, value) //parse time
	if err != nil {
		return err
	}

	*o = Time(t) //set result using the pointer
	return nil
}

func (o Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(o).Format(TimeLayout) + `"`), nil
}

func (o *Time) Scan(src interface{}) (err error) {
	var t time.Time

	switch src.(type) {
	case string:
		t, err = time.Parse(TimeLayout, src.(string))
	case []byte:
		t, err = time.Parse(TimeLayout, string(src.([]byte)))
	case time.Time:
		t = src.(time.Time)
	default:
		return errors.New("Incompatible type for epoch")
	}
	if err != nil {
		return
	}

	*o = Time(t) //set result using the pointer
	return nil
}

func (o Time) Value() (driver.Value, error) {
	return []byte(`"` + time.Time(o).Format(TimeLayout) + `"`), nil
}
