package common

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestMapApply_SimpleMap(t *testing.T) {
	assert := assert.New(t)

	origMap := loadYamlAsMap(`
StringValue: hello world
IntValue: 42
BoolValue: true
`)

	newMap := loadYamlAsMap(`
StringValue: replaced value
NewValue: new value
`)

	MapApply(origMap, newMap)

	assert.Equal(42, origMap["IntValue"])
	assert.Equal(true, origMap["BoolValue"])
	assert.Equal("replaced value", origMap["StringValue"])
	assert.Equal("new value", origMap["NewValue"])
}

func TestMapApply_DeepMap(t *testing.T) {
	assert := assert.New(t)

	origMap := loadYamlAsMap(`
MapValue:
  NestedVal: nested value
  NestedVal2: another nested value
`)

	newMap := loadYamlAsMap(`
MapValue:
  NestedVal: override nested value
  NewNestedVal: new nested value
NewValue:
  Val: new map
`)

	MapApply(origMap, newMap)

	nestedMap := origMap["MapValue"].(map[interface{}]interface{})
	nestedMap2 := origMap["NewValue"].(map[interface{}]interface{})
	assert.Equal(3, len(nestedMap))
	assert.Equal("override nested value", nestedMap["NestedVal"])
	assert.Equal("another nested value", nestedMap["NestedVal2"])
	assert.Equal("new nested value", nestedMap["NewNestedVal"])
	assert.Equal(1, len(nestedMap2))
	assert.Equal("new map", nestedMap2["Val"])
}

func TestMapApply_MergeList(t *testing.T) {
	assert := assert.New(t)

	origMap := loadYamlAsMap(`
ListValue:
  - one
  - two
`)

	newMap := loadYamlAsMap(`
ListValue:
  - three
  - four
AnotherListValue:
  - foo
  - bar
`)

	MapApply(origMap, newMap)

	aryMap := origMap["ListValue"].([]interface{})
	aryMap2 := origMap["AnotherListValue"].([]interface{})
	assert.Equal(4, len(aryMap))
	assert.Equal("one", aryMap[0])
	assert.Equal("two", aryMap[1])
	assert.Equal("three", aryMap[2])
	assert.Equal("four", aryMap[3])
	assert.Equal(2, len(aryMap2))
	assert.Equal("foo", aryMap2[0])
	assert.Equal("bar", aryMap2[1])
}

func TestMapApply_ReplaceList(t *testing.T) {
	assert := assert.New(t)

	origMap := loadYamlAsMap(`
ListValue:
  - one
  - two
`)

	newMap := loadYamlAsMap(`
ListValue:
  Fn::Replace:
    - three
    - four
`)

	MapApply(origMap, newMap)

	aryMap := origMap["ListValue"].([]interface{})
	assert.Equal(2, len(aryMap))
	assert.Equal("three", aryMap[0])
	assert.Equal("four", aryMap[1])
}

func TestMapApply_ReplaceMap(t *testing.T) {
	assert := assert.New(t)

	origMap := loadYamlAsMap(`
MapValue:
  NestedVal: nested value
  NestedVal2: another nested value
`)

	newMap := loadYamlAsMap(`
MapValue:
  Fn::Replace:
    NewNestedVal: new nested value
`)

	MapApply(origMap, newMap)

	nestedMap := origMap["MapValue"].(map[interface{}]interface{})
	assert.Equal(1, len(nestedMap))
	assert.Equal("new nested value", nestedMap["NewNestedVal"])
}

func TestMapApply_SpliceList(t *testing.T) {
	assert := assert.New(t)

	origMap := loadYamlAsMap(`
ListValue:
  - one
  - two
  - three
  - four
List2Value:
  - a
  - b
  - c
List3Value:
  - x
  - v
  - z
List4Value:
  - foo: bar
    baz: boo
`)

	newMap := loadYamlAsMap(`
ListValue:
  Fn::Splice:
    - 1
    - 2
    - - two-new
      - three-new
      - three-and-a-half
List2Value:
  Fn::Splice:
    - 0
    - 0
    - - pre-a
List3Value:
  Fn::Splice:
    - 2
    - 2
    - - new-z
      - post-z
List3Value:
  Fn::Splice:
    - 2
    - 2
    - - new-z
      - post-z
List4Value:
  Fn::Splice:
    - 0
    - 1
    - - foo: new-bar
`)

	MapApply(origMap, newMap)

	aryMap := origMap["ListValue"].([]interface{})
	assert.Equal(5, len(aryMap))
	assert.Equal("one", aryMap[0])
	assert.Equal("two-new", aryMap[1])
	assert.Equal("three-new", aryMap[2])
	assert.Equal("three-and-a-half", aryMap[3])
	assert.Equal("four", aryMap[4])

	aryMap2 := origMap["List2Value"].([]interface{})
	assert.Equal(4, len(aryMap2))
	assert.Equal("pre-a", aryMap2[0])
	assert.Equal("a", aryMap2[1])
	assert.Equal("b", aryMap2[2])
	assert.Equal("c", aryMap2[3])

	aryMap3 := origMap["List3Value"].([]interface{})
	assert.Equal(4, len(aryMap3))
	assert.Equal("x", aryMap3[0])
	assert.Equal("v", aryMap3[1])
	assert.Equal("new-z", aryMap3[2])
	assert.Equal("post-z", aryMap3[3])

	aryMap4 := origMap["List4Value"].([]interface{})
	log.Errorf("4 == %v", aryMap4)
	assert.Equal(1, len(aryMap4))
	nestedMap := aryMap4[0].(map[interface{}]interface{})
	assert.Equal("new-bar", nestedMap["foo"])
	assert.Equal("boo", nestedMap["baz"])
}

func loadYamlAsMap(yamlString string) map[interface{}]interface{} {
	rtn := make(map[interface{}]interface{})
	err := yaml.Unmarshal([]byte(yamlString), rtn)
	if err != nil {
		log.Errorf("Unable to unmarshal YAML: %s", err)
	}
	return rtn
}
