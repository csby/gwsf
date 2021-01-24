package gdoc

import (
	"encoding/json"
	"testing"
)

func TestArgument_FromExample(t *testing.T) {
	a := &argument{}

	example := TestArgumentBase{
		Data: json.MarshalerError{},
	}

	args := a.FromExample(example)
	t.Logf("%v", toJson(args))

	model := args.ToModel()
	t.Logf("%v", toJson(model))
}

func TestArgument_FromExample_Anonymous(t *testing.T) {
	a := &argument{}

	example := TestArgument{
		TestArgumentBase: TestArgumentBase{
			Data: uint64(64),
		},
		ChildId: int64(64),
	}

	args1 := a.FromExample(example)
	t.Logf("%v", toJson(args1))

	args2 := a.FromExample(&example)
	t.Logf("%v", toJson(args2))

	if toJson(args1) != toJson(args2) {
		t.Error("different between addr and value")
	}
}

func toJson(v interface{}) string {
	bytes, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return ""
	}

	return string(bytes[:])
}

type TestArgumentBase struct {
	Name   string      `json:"name" note:"名称(Base)"`
	Value  int         `json:"value" note:"值(base)-int"`
	Point  *float64    `json:"point" note:"值(base)-*float64"`
	Data   interface{} `json:"data" note:"数据"`
	Arrays []string    `json:"arrays" note:"数组"`
}

type TestArgument struct {
	TestArgumentBase

	ChildId interface{} `json:"childId" required:"true" note:"子ID"`
}
