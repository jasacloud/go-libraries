package utils

import (
	"encoding/json"
	"errors"
	"github.com/jasacloud/go-libraries/helper"
	"reflect"
	"strings"
)

// Map type
type Map map[string]interface{}

// InArray function
func InArray(arr interface{}, value interface{}) bool {
	switch v := value.(type) {
	case int:
		return InArrayInt(arr, v)
	case string:
		return InArrayString(arr, v)
	case float32:
		return InArrayFloat32(arr, v)
	case float64:
		return InArrayFloat64(arr, v)
	default:
		return false
	}
}

// InArrayInt function
func InArrayInt(arr interface{}, v int) bool {
	values, ok := arr.([]int)
	if !ok {
		if i, ok := arr.([]interface{}); ok {
			for _, iValues := range i {
				if a, ok := iValues.(int); ok {
					if a == v {
						return true
					}
				}
			}
		}
		return false
	}
	for _, a := range values {
		if !ok {
			continue
		}
		if a == v {
			return true
		}
	}
	return false
}

// InArrayString function
func InArrayString(arr interface{}, v string) bool {
	values, ok := arr.([]string)
	if !ok {
		if i, ok := arr.([]interface{}); ok {
			for _, iValues := range i {
				if a, ok := iValues.(string); ok {
					if a == v {
						return true
					}
				}
			}
		}
		return false
	}
	for _, a := range values {
		if a == v {
			return true
		}
	}
	return false
}

// InArrayFloat32 function
func InArrayFloat32(arr interface{}, v float32) bool {
	values, ok := arr.([]float32)
	if !ok {
		if i, ok := arr.([]interface{}); ok {
			for _, iValues := range i {
				if a, ok := iValues.(float32); ok {
					if a == v {
						return true
					}
				}
			}
		}
		return false
	}
	for _, a := range values {
		if a == v {
			return true
		}
	}
	return false
}

// InArrayFloat64 function
func InArrayFloat64(arr interface{}, v float64) bool {
	values, ok := arr.([]float64)
	if !ok {
		if i, ok := arr.([]interface{}); ok {
			for _, iValues := range i {
				if a, ok := iValues.(float64); ok {
					if a == v {
						return true
					}
				}
			}
		}
		return false
	}
	for _, a := range values {
		if a == v {
			return true
		}
	}
	return false
}

// FindString function
func FindString(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// UnionArrayString function
func UnionArrayString(a, b []string) []string {
	m := make(map[string]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; !ok {
			a = append(a, item)
		}
	}
	return a
}

// Intersect function
func Intersect(l1, l2 interface{}) (interface{}, error) {
	l1v := reflect.ValueOf(l1)
	l2v := reflect.ValueOf(l2)

	switch l1v.Kind() {
	case reflect.Array, reflect.Slice:
		switch l2v.Kind() {
		case reflect.Array, reflect.Slice:
			return processIntersect(l1v, l2v)
		default:
			return nil, errors.New("can't iterate over " + reflect.ValueOf(l2).Type().String())
		}
	default:
		return nil, errors.New("can't iterate over " + reflect.ValueOf(l1).Type().String())
	}
}

// processIntersect function
func processIntersect(l1v, l2v reflect.Value) (interface{}, error) {
	r := reflect.MakeSlice(l1v.Type(), 0, 0)
	for i := 0; i < l1v.Len(); i++ {
		l1vv := l1v.Index(i)
		for j := 0; j < l2v.Len(); j++ {
			l2vv := l2v.Index(j)
			switch l1vv.Kind() {
			case reflect.String:
				if l1vv.Type() == l2vv.Type() && l1vv.String() == l2vv.String() && !ValueIn(r, l2vv) {
					r = reflect.Append(r, l2vv)
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				switch l2vv.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if l1vv.Int() == l2vv.Int() && !ValueIn(r, l2vv) {
						r = reflect.Append(r, l2vv)
					}
				}
			case reflect.Float32, reflect.Float64:
				switch l2vv.Kind() {
				case reflect.Float32, reflect.Float64:
					if l1vv.Float() == l2vv.Float() && !ValueIn(r, l2vv) {
						r = reflect.Append(r, l2vv)
					}
				}
			}
		}
	}
	return r.Interface(), nil
}

// ValueIn function
func ValueIn(l interface{}, v interface{}) bool {
	lv := reflect.ValueOf(l)
	vv := reflect.ValueOf(v)

	switch lv.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < lv.Len(); i++ {
			lvv := lv.Index(i)
			switch lvv.Kind() {
			case reflect.String:
				if vv.Type() == lvv.Type() && vv.String() == lvv.String() {
					return true
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				switch vv.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if vv.Int() == lvv.Int() {
						return true
					}
				}
			case reflect.Float32, reflect.Float64:
				switch vv.Kind() {
				case reflect.Float32, reflect.Float64:
					if vv.Float() == lvv.Float() {
						return true
					}
				}
			}
		}
	case reflect.String:
		if vv.Type() == lv.Type() && strings.Contains(lv.String(), vv.String()) {
			return true
		}
	}
	return false
}

// UniqueArrayString function
func UniqueArrayString(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// PairValues function
func PairValues(i, o interface{}, validate ...bool) error {
	if i == nil {
		return errors.New("error while pair values, values is nil")
	}
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &o)
	if err != nil {
		return err
	}

	//check type of o
	r := reflect.ValueOf(o)
	if r.Kind() == reflect.Ptr && !r.IsNil() {
		r = r.Elem()
	}
	if r.Kind() != reflect.Struct && r.Kind() != reflect.Interface {

		return nil
	}

	//validate struct required values :
	if len(validate) > 0 {
		if validate[0] {
			err = helper.BindValidate(o)
			if err != nil {

				return err
			}
		}
	}

	return nil
}

// InterfaceSlice function
func InterfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		return nil
	}

	// Keep the distinction between nil and empty slice input
	if s.IsNil() {
		return nil
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}
