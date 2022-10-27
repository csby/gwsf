package gtype

import (
	"reflect"
	"time"
)

func CopyElement(dst, src interface{}) {
	if dst == nil || src == nil {
		return
	}

	copyElement(reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem())
}

func copyElement(dst, src reflect.Value) {
	kind := dst.Kind()
	if src.Kind() != kind {
		return
	}
	if src.CanInterface() == false {
		return
	}
	if src.Interface() == nil {
		return
	}

	switch kind {
	case reflect.Ptr:
		elem := src.Elem()
		if !elem.IsValid() {
			return
		}
		dst.Set(reflect.New(elem.Type()))
		copyElement(dst.Elem(), elem)
	case reflect.Interface:
		if src.IsNil() {
			return
		}
		elem := src.Elem()
		dst.Set(reflect.New(elem.Type()).Elem())
		copyElement(dst.Elem(), elem)
	case reflect.Struct:
		t, ok := src.Interface().(time.Time)
		if ok {
			dst.Set(reflect.ValueOf(t))
			return
		}

		c := dst.NumField()
		if src.NumField() != c {
			return
		}
		for i := 0; i < c; i++ {
			copyElement(dst.Field(i), src.Field(i))
		}
	case reflect.Slice:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			copyElement(dst.Index(i), src.Index(i))
		}
	case reflect.Map:
		if src.IsNil() {
			return
		}
	default:
		dst.Set(src)
	}
}
