package common

import (
	"fmt"
	"reflect"
)

type mergeOperation func(destVal reflect.Value, srcVal reflect.Value) (bool, reflect.Value)

func appendOperation(destVal reflect.Value, srcVal reflect.Value) (bool, reflect.Value) {
	switch srcVal.Kind() {
	case reflect.Slice, reflect.Array:
		if destVal.IsValid() {
			return true, reflect.AppendSlice(destVal, srcVal)
		}
	case reflect.Map:
		if destVal.IsValid() {
			MapApply(destVal.Interface(), srcVal.Interface())
			return true, destVal
		}
	}
	return true, srcVal
}

func replaceOperation(destVal reflect.Value, srcVal reflect.Value) (bool, reflect.Value) {
	if srcVal.Kind() == reflect.Map {
		newVal := valueOfInterface(srcVal.MapIndex(reflect.ValueOf("Fn::Replace")))
		return newVal.IsValid(), newVal

	}
	return false, reflect.Value{}
}

func spliceOperation(destVal reflect.Value, srcVal reflect.Value) (bool, reflect.Value) {
	if srcVal.Kind() == reflect.Map {
		newVal := valueOfInterface(srcVal.MapIndex(reflect.ValueOf("Fn::Splice")))
		if newVal.IsValid() && newVal.Len() == 3 {
			start := int(newVal.Index(0).Elem().Int())
			mergeCount := int(newVal.Index(1).Elem().Int())
			newSlice := newVal.Index(2).Elem()

			// if no existing value, just use the new value
			if !destVal.IsValid() {
				return true, newSlice
			}

			if start > destVal.Len() {
				start = destVal.Len()
			}

			// initialize with elements from before the splice
			returnVal := reflect.MakeSlice(destVal.Type(), 0, destVal.Len()+srcVal.Len())
			returnVal = reflect.AppendSlice(returnVal, destVal.Slice(0, start))

			// merge the overlapping elements
			for i := 0; i < newSlice.Len(); i++ {
				if i < mergeCount && (start+i) < destVal.Len() {
					// merge
					_, mergedElement := appendOperation(valueOfInterface(destVal.Index(start+i)), valueOfInterface(newSlice.Index(i)))
					returnVal = reflect.Append(returnVal, mergedElement)
				} else {
					// append
					returnVal = reflect.Append(returnVal, newSlice.Index(i))
				}
			}

			// append any existing element from after the splice
			if start+mergeCount < destVal.Len() {
				returnVal = reflect.AppendSlice(returnVal, destVal.Slice(start+mergeCount, destVal.Len()))
			}

			return true, returnVal
		}

	}

	return false, reflect.Value{}
}

// MapApply recursively applies map values from source to destination
func MapApply(dest, src interface{}) {
	srcMap := reflect.ValueOf(src)
	if srcMap.Kind() != reflect.Map {
		return
	}

	destMap := reflect.ValueOf(dest)
	if destMap.Kind() != reflect.Map {
		return
	}

	for _, key := range srcMap.MapKeys() {
		srcVal := valueOfInterface(srcMap.MapIndex(key))
		destVal := valueOfInterface(destMap.MapIndex(key))

		for _, op := range []mergeOperation{spliceOperation, replaceOperation, appendOperation} {
			handled, response := op(destVal, srcVal)

			if handled {
				destMap.SetMapIndex(key, response)
				break
			}
		}
	}
}

func valueOfInterface(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return v
	}
	return reflect.ValueOf(v.Interface())
}

// ConvertMapI2MapS walks the given dynamic object recursively, and
// converts maps with interface{} key type to maps with string key type.
// This function comes handy if you want to marshal a dynamic object into
// JSON where maps with interface{} key type are not allowed.
//
// Recursion is implemented into values of the following types:
//   -map[interface{}]interface{}
//   -map[string]interface{}
//   -[]interface{}
//
// When converting map[interface{}]interface{} to map[string]interface{},
// fmt.Sprint() with default formatting is used to convert the key to a string key.
func ConvertMapI2MapS(v interface{}) interface{} {
	switch x := v.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v2 := range x {
			switch k2 := k.(type) {
			case string: // Fast check if it's already a string
				m[k2] = ConvertMapI2MapS(v2)
			default:
				m[fmt.Sprint(k)] = ConvertMapI2MapS(v2)
			}
		}
		v = m

	case []interface{}:
		for i, v2 := range x {
			x[i] = ConvertMapI2MapS(v2)
		}

	case map[string]interface{}:
		for k, v2 := range x {
			x[k] = ConvertMapI2MapS(v2)
		}

	case int:
		return int64(v.(int))
	case uint:
		return int64(v.(uint))
	case uint8:
		return int64(v.(uint8))
	case uint16:
		return int64(v.(uint16))
	case uint32:
		return int64(v.(uint32))
	case uint64:
		return int64(v.(uint64))
	}

	return v
}

// MapGet returns a value denoted by the path.
//
// If path is empty or nil, v is returned.
func MapGet(v interface{}, path ...interface{}) interface{} {
	for _, el := range path {
		switch node := v.(type) {
		case map[string]interface{}:
			key, ok := el.(string)
			if !ok {
				return nil
			}
			v, ok = node[key]
			if !ok {
				return nil
			}

		case map[interface{}]interface{}:
			var ok bool
			v, ok = node[el]
			if !ok {
				return nil
			}

		case []interface{}:
			idx, ok := el.(int)
			if !ok {
				return nil
			}
			if idx < 0 || idx >= len(node) {
				return nil
			}
			v = node[idx]

		default:
			return nil
		}
	}

	return v
}

// MapGetSlice returns a value denoted by the path.
//
// If path is empty or nil, v is returned.
func MapGetSlice(v interface{}, path ...interface{}) []interface{} {
	v = MapGet(v, path...)
	if v == nil {
		return nil
	}
	s, ok := v.([]interface{})
	if !ok {
		return nil
	}
	return s
}

// MapGetString returns a string value denoted by the path.
//
// If path is empty or nil, v is returned as a string.
func MapGetString(v interface{}, path ...interface{}) string {
	v = MapGet(v, path...)
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

// MapClone returns a map copy
func MapClone(m map[string]string) map[string]string {
	rtn := make(map[string]string)
	for k, v := range m {
		rtn[k] = v
	}
	return rtn
}
