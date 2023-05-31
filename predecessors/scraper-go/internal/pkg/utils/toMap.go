package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func ToMap(p interface{}) ([]map[string]string, error) {
	v := reflect.ValueOf(p)
	t := v.Type()

	if t.Kind() == reflect.Slice {
		if t.Elem().Kind() != reflect.Struct {
			return nil, fmt.Errorf("ToMap: expected slice of structs")
		}

		retData := make([]map[string]string, 0, v.Len())

		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			data, err := structToMap(elem)
			if err != nil {
				return nil, err
			}
			retData = append(retData, data)
		}

		return retData, nil
	} else if t.Kind() == reflect.Struct {
		data, err := structToMap(v)
		if err != nil {
			return nil, err
		}

		return []map[string]string{data}, nil
	}

	return nil, fmt.Errorf("ToMap: expected struct or slice of structs")
}

func structToMap(v reflect.Value) (map[string]string, error) {
	data := make(map[string]string)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		fieldValue := v.Field(i)
		fieldType := fieldValue.Type()

		switch fieldType.Kind() {
		case reflect.String:
			data[fieldName] = CleanStringCSV(fieldValue.String())
		case reflect.Slice:
			if fieldType.Elem().Kind() == reflect.String {
				var sliceValues []string
				if fieldValue.Len() == 0 {
					data[fieldName] = CleanStringCSV("")
					continue
				}
				for j := 0; j < fieldValue.Len(); j++ {
					tStr := CleanStringCSV(fieldValue.Index(j).String())
					sliceValues = append(sliceValues, tStr)
				}
				data[fieldName] = "[" + strings.Join(sliceValues, ",") + "]"
			} else {
				return nil, fmt.Errorf("ToMap: unsupported slice type")
			}
		case reflect.Float64:
			data[fieldName] = fmt.Sprintf("%f", fieldValue.Float())
		case reflect.Int:
			data[fieldName] = strconv.Itoa(int(fieldValue.Int()))
		case reflect.Struct:
			if fieldType == reflect.TypeOf(time.Time{}) {
				if t, ok := fieldValue.Interface().(time.Time); ok {
					data[fieldName] = t.Format("2006/01/02 15:04:05")
				}
			} else {
				nestedData, err := structToMap(fieldValue)
				if err != nil {
					return nil, err
				}
				for key, value := range nestedData {
					fieldNameWithPrefix := fmt.Sprintf("%s.%s", fieldName, key)
					data[fieldNameWithPrefix] = value
				}
			}
		default:
			return nil, fmt.Errorf("ToMap: unsupported field type")
		}
	}

	return data, nil
}
