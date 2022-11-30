// Copyright (c) 2019 JasaCloud.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helper

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
	"sync"
)

var once sync.Once

var validate *validator.Validate

// ValidateStruct function
func ValidateStruct(o interface{}) error {

	validate = validator.New()
	err := validate.Struct(o)

	return err
}

// ValidateJSON function
func ValidateJSON(o interface{}) error {
	validate = validator.New()
	validate.SetTagName("binding")
	err := validate.Struct(o)

	return err
}

// BindValidate function
func BindValidate(o interface{}) error {
	once.Do(JsonTagNameFunc)

	err := binding.Validator.ValidateStruct(o)
	if err != nil {
		return err
	}

	validate = validator.New()
	err = validate.Struct(o)

	return err
}

// PairValues function
func PairValues(i, o interface{}) error {
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

	//validate struct :
	err = BindValidate(o)
	if err != nil {

		return err
	}

	return nil
}

func JsonTagNameFunc() {
	if f := reflect.ValueOf(binding.Validator.Engine()).MethodByName("RegisterTagNameFunc"); f.IsValid() {
		var args []reflect.Value
		args = append(args, reflect.ValueOf(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		}))
		f.Call(args)
	}
}
