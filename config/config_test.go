package config

import (
	"context"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
)

type Foo struct {
	Field1 string `json,yaml:"field1"`
	Field2 bool   `json:"field2" yaml:"field2"`
}

type TestStruct struct {
	Field1 string `json,yaml:"field1"`
	Field2 bool   `json,yaml:"field2"`
	Field3 []Foo  `json,yaml:"field3"`
}

func creator() interface{} {
	return &TestStruct{}
}

func TestJSONConfig(t *testing.T) {
	RegisterConfigCreator("test", creator)
	data := []byte(`
	{
		"field1": "test1",
		"field2": true,
		"field3": [
			{
				"field1": "aaaa",
				"field2": true
			}
		]
	}
	`)
	ctx, err := WithJSONConfig(context.Background(), data)
	common.Must(err)
	c := FromContext(ctx, "test").(*TestStruct)
	if c.Field1 != "test1" || c.Field2 != true {
		t.Fail()
	}
}

func TestYAMLConfig(t *testing.T) {
	RegisterConfigCreator("test", creator)
	data := []byte(`
field1: 012345678
field2: true
field3:
  - field1: test
    field2: true
`)
	ctx, err := WithYAMLConfig(context.Background(), data)
	common.Must(err)
	c := FromContext(ctx, "test").(*TestStruct)
	if c.Field1 != "012345678" || c.Field2 != true || c.Field3[0].Field1 != "test" {
		t.Fail()
	}
}
