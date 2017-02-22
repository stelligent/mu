package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapApply_SimpleMap(t *testing.T) {
	assert := assert.New(t)

	origMap := make(map[interface{}]interface{})
	origMap["StringValue"] = "hello world"
	origMap["IntValue"] = 42
	origMap["BoolValue"] = true

	newMap := make(map[interface{}]interface{})
	newMap["StringValue"] = "replaced value"
	newMap["NewValue"] = "new value"

	MapApply(origMap, newMap)

	assert.Equal(42, origMap["IntValue"])
	assert.Equal(true, origMap["BoolValue"])
	assert.Equal("replaced value", origMap["StringValue"])
	assert.Equal("new value", origMap["NewValue"])
}

func TestMapApply_DeepMap(t *testing.T) {
	assert := assert.New(t)

	origMap := make(map[interface{}]interface{})
	origMap["MapValue"] = make(map[string]string)
	origMap["MapValue"].(map[string]string)["NestedVal"] = "nested value"
	origMap["MapValue"].(map[string]string)["NestedVal2"] = "another nested value"

	newMap := make(map[interface{}]interface{})
	newMap["MapValue"] = make(map[string]string)
	newMap["MapValue"].(map[string]string)["NestedVal"] = "override nested value"
	newMap["MapValue"].(map[string]string)["NewNestedVal"] = "new nested value"
	newMap["NewValue"] = make(map[string]string)
	newMap["NewValue"].(map[string]string)["Val"] = "new map"

	MapApply(origMap, newMap)

	nestedMap := origMap["MapValue"].(map[string]string)
	nestedMap2 := origMap["NewValue"].(map[string]string)
	assert.Equal(3, len(nestedMap))
	assert.Equal("override nested value", nestedMap["NestedVal"])
	assert.Equal("another nested value", nestedMap["NestedVal2"])
	assert.Equal("new nested value", nestedMap["NewNestedVal"])
	assert.Equal(1, len(nestedMap2))
	assert.Equal("new map", nestedMap2["Val"])
}

func TestMapApply_MergeList(t *testing.T) {
	assert := assert.New(t)

	origMap := make(map[interface{}]interface{})
	origMap["ListValue"] = []string{"one", "two"}

	newMap := make(map[interface{}]interface{})
	newMap["ListValue"] = []string{"three", "four"}
	newMap["AnotherListValue"] = []string{"foo", "bar"}

	MapApply(origMap, newMap)

	aryMap := origMap["ListValue"].([]string)
	aryMap2 := origMap["AnotherListValue"].([]string)
	assert.Equal(4, len(aryMap))
	assert.Equal("one", aryMap[0])
	assert.Equal("two", aryMap[1])
	assert.Equal("three", aryMap[2])
	assert.Equal("four", aryMap[3])
	assert.Equal(2, len(aryMap2))
	assert.Equal("foo", aryMap2[0])
	assert.Equal("bar", aryMap2[1])
}
