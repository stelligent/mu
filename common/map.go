package common

import (
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
