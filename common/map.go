package common

import (
	"reflect"
)

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

		switch srcVal.Kind() {
		default:
			destMap.SetMapIndex(key, srcVal)
		case reflect.Slice, reflect.Array:
			if destVal.IsValid() {
				newAry := reflect.AppendSlice(destVal, srcVal)
				destMap.SetMapIndex(key, newAry)
			} else {
				destMap.SetMapIndex(key, srcVal)
			}
		case reflect.Map:
			if destVal.IsValid() {
				MapApply(destVal.Interface(), srcVal.Interface())
			} else {
				destMap.SetMapIndex(key, srcVal)
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
