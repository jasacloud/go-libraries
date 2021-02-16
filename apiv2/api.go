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

package apiv2

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jasacloud/go-libraries/helper"
	"strconv"
)

// Attributes type map array
type Attributes map[string]interface{}

// Match struct
type Match struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

// RangeValue struct
type RangeValue struct {
	Gt interface{} `json:"gt" bson:"gt"`
	Lt interface{} `json:"lt" bson:"lt"`
}

// Range struct
type Range struct {
	Key   string      `json:"key" bson:"key"`
	Value *RangeValue `json:"value" bson:"value"`
}

// Filter struct
type Filter struct {
	Range     []Range `json:"range" bson:"range"`
	RangeAttr []Range `json:"range_attributes" bson:"range_attributes"`
}

// Sort struct
type Sort struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

// Limit struct
type Limit struct {
	Rows   int `json:"rows" bson:"rows"`
	Offset int `json:"offset" bson:"offset"`
}

// Query struct
type Query struct {
	Match     []Match `json:"match" bson:"match"`
	MatchAttr []Match `json:"match_attributes" bson:"match_attributes"`
}

// Params struct
type Params struct {
	Query      Query    `json:"query" bson:"query"`
	Filter     Filter   `json:"filter" bson:"filter"`
	Sort       []Sort   `json:"sort" bson:"sort"`
	Limit      Limit    `json:"limit" bson:"limit"`
	AttrFilter []Filter `json:"filter_attributes" bson:"filter_attributes"`
}

// Error struct
type Error struct {
	Code           string      `json:"code" bson:"code" binding:"required"`
	Type           string      `json:"type,omitempty" bson:"type,omitempty"`
	Status         string      `json:"status,omitempty" bson:"status,omitempty"`
	Message        string      `json:"message,omitempty" bson:"message,omitempty"`
	MessageDetails string      `json:"message_details,omitempty" bson:"message_details,omitempty"`
	Refs           interface{} `json:"refs,omitempty" bson:"refs,omitempty"`
}

// IError struct
type IError struct {
	Result  string `json:"result" bson:"result"  binding:"required"`
	Message string `json:"message,omitempty" bson:"message,omitempty"`
}

// Request struct
type Request struct {
	Kind   string      `json:"kind" bson:"kind" binding:"required"`
	Values interface{} `json:"values" bson:"values" binding:"required"`
}

// Response struct
type Response struct {
	ReturnVal bool        `json:"returnval" bson:"returnval" binding:"required"`
	RespVal   string      `json:"respval,omitempty" bson:"respval,omitempty"`
	Kind      string      `json:"kind,omitempty" bson:"kind,omitempty"`
	Values    interface{} `json:"values,omitempty" bson:"values,omitempty"`
	Error     interface{} `json:"error,omitempty" bson:"error,omitempty"`
}

// ParseRequest function
func ParseRequest(c *gin.Context) (*Request, error) {
	var request *Request
	err := c.BindJSON(&request)
	if err != nil {
		return nil, err
	}

	return request, nil
}

// Get method
func (a *Attributes) Get(key string) interface{} {
	b, err := json.Marshal(a)
	if err != nil {
		return nil
	}
	var o map[string]interface{}
	err = json.Unmarshal(b, &o)
	if err != nil {
		return nil
	}
	return o[key]
}

// GetSubAttributes method
func (a *Attributes) GetSubAttributes(key string) (*Attributes, error) {
	return NewAttributes(a.Get(key))
}

// GetArrayInterface method
func (a *Attributes) GetArrayInterface(key string) []interface{} {
	if arr := a.Get(key); arr != nil {
		if values, ok := arr.([]interface{}); ok {
			return values
		}
	}
	return nil
}

// GetArrayString method
func (a *Attributes) GetArrayString(key string) (values []string) {
	if arr := a.GetArrayInterface(key); arr != nil {
		for _, v := range arr {
			if s, ok := v.(string); ok {
				values = append(values, s)
			}
		}
	}
	return values
}

// GetArrayInt method
func (a *Attributes) GetArrayInt(key string) (values []int) {
	if arr := a.GetArrayInterface(key); arr != nil {
		for _, v := range arr {
			if s, ok := v.(int); ok {
				values = append(values, s)
			}
		}
	}
	return values
}

// GetArrayFloat32 method
func (a *Attributes) GetArrayFloat32(key string) (values []float32) {
	if arr := a.GetArrayInterface(key); arr != nil {
		for _, v := range arr {
			if s, ok := v.(float32); ok {
				values = append(values, s)
			}
		}
	}
	return values
}

// GetArrayFloat64 method
func (a *Attributes) GetArrayFloat64(key string) (values []float64) {
	if arr := a.GetArrayInterface(key); arr != nil {
		for _, v := range arr {
			if s, ok := v.(float64); ok {
				values = append(values, s)
			}
		}
	}
	return values
}

// GetString method
func (a *Attributes) GetString(key string) string {
	b, err := json.Marshal(a)
	if err != nil {
		return ""
	}
	var o map[string]interface{}
	err = json.Unmarshal(b, &o)
	if err != nil {
		return ""
	}
	switch v := o[key].(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	default:
		return ""
	}
}

// GetInt method
func (a *Attributes) GetInt(key string) int {
	b, err := json.Marshal(a)
	if err != nil {
		return 0
	}
	var o map[string]interface{}
	err = json.Unmarshal(b, &o)
	if err != nil {
		return 0
	}
	switch v := o[key].(type) {
	case string:
		i, _ := strconv.Atoi(v)
		return i
	case int:
		return v
	case float64:
		return int(v)
	case float32:
		return int(v)
	default:
		return 0
	}
}

// GetFloat64 method
func (a *Attributes) GetFloat64(key string) float64 {
	b, err := json.Marshal(a)
	if err != nil {
		return 0
	}
	var o map[string]interface{}
	err = json.Unmarshal(b, &o)
	if err != nil {
		return 0
	}
	switch v := o[key].(type) {
	case string:
		i, _ := strconv.ParseFloat(v, 64)
		return i
	case int:
		return float64(v)
	case float64:
		return v
	case float32:
		return float64(v)
	default:
		return 0
	}
}

// Set method
func (a *Attributes) Set(key string, val interface{}) error {
	ba, err := json.Marshal(a)
	if err != nil {
		return err
	}
	var o map[string]interface{}
	err = json.Unmarshal(ba, &o)
	if err != nil {
		return err
	}
	if o == nil {
		o = make(map[string]interface{})
	}
	o[key] = val
	bo, err := json.Marshal(o)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bo, &a)
	if err != nil {
		return err
	}
	return nil
}

// Delete method
func (a *Attributes) Delete(key string) error {
	ba, err := json.Marshal(a)
	if err != nil {
		return err
	}
	var o map[string]interface{}
	err = json.Unmarshal(ba, &o)
	if err != nil {
		return err
	}
	if o[key] != nil {
		delete(o, key)
	}
	bo, err := json.Marshal(o)
	if err != nil {
		return err
	}
	*a = nil
	err = json.Unmarshal(bo, &a)
	if err != nil {
		return err
	}
	return nil
}

// NewAttributes function
func NewAttributes(i interface{}) (*Attributes, error) {
	var a *Attributes
	ba, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(ba, &a)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// ParseKeyValSetAttributes method
func (a Attributes) ParseKeyValSetAttributes() Attributes {
	if a != nil {
		var attributes = make(Attributes)
		var key string
		ParseSubAttributes(a, attributes, &key)
		return attributes
	}

	return nil
}

// ParseSubAttributes function
func ParseSubAttributes(i, o Attributes, key *string) {
	for i, v := range i {
		var attributes = make(Attributes)
		err := helper.PairValues(v, &attributes)
		if err == nil {
			if *key != "" {
				k := *key + "." + i
				ParseSubAttributes(attributes, o, &k)
			} else {
				ParseSubAttributes(attributes, o, &i)
			}
		} else {
			if *key != "" {
				o[*key+"."+i] = v
			} else {
				o[i] = v
			}
		}
	}
}
