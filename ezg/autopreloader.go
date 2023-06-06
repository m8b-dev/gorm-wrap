package ezg

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	maxRecursion = 200
	nopreload1   = "no-preload"
	nopreload2   = "nopreload"
)

type fieldInfo struct {
	name string
	typ  reflect.Type
}

func autoPreloads(model interface{}) []string {
	return autoPreloadsInternal(model, "", 0)
}

func autoPreloadsInternal(model interface{}, prefix string, i uint) []string {
	if i > maxRecursion {
		panic(fmt.Sprintf("max recursion treshold of %d exceeded. Infinitely recursive preloads are not supported.", maxRecursion))
	}
	flds := exportedFields(model)
	out := make([]string, 0)
	for _, fld := range flds {
		if fld.typ.Kind() == reflect.Slice {
			out = append(out, prefix+fld.name)
			out = append(out, autoPreloadsInternal(fld.typ, prefix+fld.name+".", i+1)...)
		}
	}
	return out
}

func exportedFields(i interface{}) []fieldInfo {
	typ, ok := i.(reflect.Type)
	if !ok {
		typ = reflect.TypeOf(i)
	}
	for typ.Kind() != reflect.Struct {
		switch typ.Kind() {
		case reflect.Ptr:
			typ = typ.Elem()
		case reflect.Slice, reflect.Array:
			typ = typ.Elem()
		default:
			return make([]fieldInfo, 0)
		}
	}

	fields := make([]fieldInfo, 0)

	// Loop over struct fields.
	for i := 0; i < typ.NumField(); i++ {
		tag, ok := typ.Field(i).Tag.Lookup("ezg")
		if ok && (strings.EqualFold(tag, nopreload1) || strings.EqualFold(tag, nopreload2)) {
			continue
		}
		// Only take exported fields (name starts with an uppercase letter).
		if typ.Field(i).PkgPath == "" {
			field := fieldInfo{
				name: typ.Field(i).Name,
				typ:  typ.Field(i).Type,
			}
			fields = append(fields, field)
		}
	}

	return fields
}
