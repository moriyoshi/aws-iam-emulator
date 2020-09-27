package main

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

func UnmarshalParams(recv interface{}, v url.Values, isEC2 bool) error {
	return (&queryBuilder{isEC2: isEC2}).buildValue(reflect.ValueOf(recv), "", "", false, v)
}

func elemOf(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Ptr {
		if value.IsNil() {
			if !value.CanSet() {
				panic("???")
			}
			rv := reflect.New(value.Type().Elem())
			value.Set(rv)
		}
		value = value.Elem()
	}
	return value
}

type queryBuilder struct {
	isEC2 bool
}

func (q *queryBuilder) buildValue(value reflect.Value, prefix string, tag reflect.StructTag, parentCollection bool, v url.Values) error {
	value = elemOf(value)

	t := tag.Get("type")
	if t == "" {
		switch value.Kind() {
		case reflect.Struct:
			t = "structure"
		case reflect.Slice:
			t = "list"
		case reflect.Map:
			t = "map"
		}
	}

	var err error
	switch t {
	case "structure":
		err = q.buildStruct(value, prefix, v)
	case "list":
		err = q.buildList(value, prefix, tag, v)
	case "map":
		err = q.buildMap(value, prefix, tag, v)
	default:
		err = q.buildScalar(value, prefix, tag, parentCollection, v)
	}

	return err
}

func (q *queryBuilder) buildStruct(value reflect.Value, prefix string, v url.Values) error {
	if !value.IsValid() {
		return nil
	}

	t := value.Type()
	for i := 0; i < value.NumField(); i++ {
		elemValue := value.Field(i).Addr()
		field := t.Field(i)

		if field.PkgPath != "" {
			continue // ignore unexported fields
		}
		if field.Tag.Get("ignore") != "" {
			continue
		}

		var name string
		if q.isEC2 {
			name = field.Tag.Get("queryName")
		}
		if name == "" {
			if field.Tag.Get("flattened") != "" && field.Tag.Get("locationNameList") != "" {
				name = field.Tag.Get("locationNameList")
			} else if locName := field.Tag.Get("locationName"); locName != "" {
				name = locName
			}
			if name != "" && q.isEC2 {
				name = strings.ToUpper(name[0:1]) + name[1:]
			}
		}
		if name == "" {
			name = field.Name
		}

		if prefix != "" {
			name = prefix + "." + name
		}

		required, _ := strconv.ParseBool(field.Tag.Get("required"))

		_, ok := v[name]
		if !ok && !required {
			continue
		}

		if err := q.buildValue(elemValue, name, field.Tag, false, v); err != nil {
			return err
		}
	}
	return nil
}

func (q *queryBuilder) buildList(value reflect.Value, prefix string, tag reflect.StructTag, v url.Values) error {
	// If it's empty, generate an empty value
	if v.Get(prefix) == "" {
		if value.Type().Kind() == reflect.Slice {
			value.Set(reflect.MakeSlice(value.Type(), 0, 0))
		}
		return nil
	}

	t := value.Type()
	if t.Kind() == reflect.Array && t.Elem().Kind() == reflect.Uint8 {
		return q.buildScalar(value, prefix, tag, true, v)
	}

	// check for unflattened list member
	if !q.isEC2 && tag.Get("flattened") == "" {
		if listName := tag.Get("locationNameList"); listName == "" {
			prefix += ".member"
		} else {
			prefix += "." + listName
		}
	}

	for i := 0; i < value.Len(); i++ {
		slicePrefix := prefix
		if slicePrefix == "" {
			slicePrefix = strconv.Itoa(i + 1)
		} else {
			slicePrefix = slicePrefix + "." + strconv.Itoa(i+1)
		}
		if err := q.buildValue(value.Index(i), slicePrefix, "", true, v); err != nil {
			return err
		}
	}
	return nil
}

func (q *queryBuilder) buildMap(value reflect.Value, prefix string, tag reflect.StructTag, v url.Values) error {
	// If it's empty, generate an empty value
	if v.Get(prefix) == "" {
		value.Set(reflect.MakeMap(value.Type()))
		return nil
	}

	// check for unflattened list member
	if !q.isEC2 && tag.Get("flattened") == "" {
		prefix += ".entry"
	}

	// sort keys for improved serialization consistency.
	// this is not strictly necessary for protocol support.
	mapKeyValues := value.MapKeys()
	mapKeys := map[string]reflect.Value{}
	mapKeyNames := make([]string, len(mapKeyValues))
	for i, mapKey := range mapKeyValues {
		name := mapKey.String()
		mapKeys[name] = mapKey
		mapKeyNames[i] = name
	}
	sort.Strings(mapKeyNames)

	for i, mapKeyName := range mapKeyNames {
		mapKey := mapKeys[mapKeyName]
		mapValue := reflect.New(value.Type().Elem())

		kname := tag.Get("locationNameKey")
		if kname == "" {
			kname = "key"
		}
		vname := tag.Get("locationNameValue")
		if vname == "" {
			vname = "value"
		}

		// serialize key
		var keyName string
		if prefix == "" {
			keyName = strconv.Itoa(i+1) + "." + kname
		} else {
			keyName = prefix + "." + strconv.Itoa(i+1) + "." + kname
		}

		if err := q.buildValue(mapKey, keyName, "", true, v); err != nil {
			return err
		}

		// deserialize value
		var valueName string
		if prefix == "" {
			valueName = strconv.Itoa(i+1) + "." + vname
		} else {
			valueName = prefix + "." + strconv.Itoa(i+1) + "." + vname
		}

		if err := q.buildValue(mapValue, valueName, "", true, v); err != nil {
			return err
		}

		value.SetMapIndex(mapKey, mapValue)
	}

	return nil
}

var timeType = reflect.TypeOf(time.Time{})

func (q *queryBuilder) buildScalar(r reflect.Value, name string, tag reflect.StructTag, parentCollection bool, v url.Values) error {
	value := v.Get(name)
	t := r.Type()
	switch t.Kind() {
	case reflect.String:
		r.SetString(value)
		return nil
	case reflect.Bool:
		vv, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		r.SetBool(vv)
		return nil
	case reflect.Int64:
		vv, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		r.SetInt(vv)
		return nil
	case reflect.Int:
		vv, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return err
		}
		r.SetInt(vv)
		return nil
	case reflect.Float64:
		vv, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		r.SetFloat(vv)
		return nil
	case reflect.Float32:
		vv, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return err
		}
		r.SetFloat(vv)
		return nil
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			vv, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				return err
			}
			r.Set(reflect.ValueOf(vv))
			return nil
		}
	case reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			vv, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				return err
			}
			if r.Len() != len(vv) {
				return fmt.Errorf("length of array does not match (%d != %d)", r.Len(), len(vv))
			}
			rv := r.Slice(0, r.Len()).Interface().([]byte)
			copy(rv, vv)
			return nil
		}
	case reflect.Struct:
		if t.Elem() == timeType {
			vv, err := time.Parse("2006-01-02T15:04:05Z", value)
			if err != nil {
				return err
			}
			r.Set(reflect.ValueOf(vv))
			return nil
		}
	default:
	}
	return fmt.Errorf("unsupported value for param %s: %v (%s)", name, r.Interface(), r.Type().Name())
}
