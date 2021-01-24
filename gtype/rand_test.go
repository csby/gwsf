package gtype

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestRandNumber_New(t *testing.T) {
	r := &randNumber{
		index: 9,
	}

	n := r.New()
	t.Logf("%d: %d", n, len(fmt.Sprint(n)))

	n = r.New()
	t.Logf("%d: %d", n, len(fmt.Sprint(n)))
}

func Test_NewRand(t *testing.T) {
	r := NewRand(1)

	n := r.New()
	t.Logf("%d: %d", n, len(fmt.Sprint(n)))

	n = r.New()
	t.Logf("%d: %d", n, len(fmt.Sprint(n)))

	rs := &Result{
		Code: 10,
	}
	rs.Error.Summary = "s"
	t.Logf("%v", rs.FormatString())

	rsd, _ := rs.Marshal()
	rs2 := &Result{}
	rs2.Unmarshal(rsd)
	t.Logf("%v", rs2.FormatString())
}

func toJson(v interface{}) string {
	bytes, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return ""
	}

	return string(bytes[:])
}
