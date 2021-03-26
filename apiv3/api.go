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

package apiv3

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jasacloud/go-libraries/db"
	"github.com/jasacloud/go-libraries/helper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"log"
	"strconv"
	"strings"
)

// Attributes type
type Attributes map[string]interface{}

// Match struct
type Match struct {
	Key   string      `json:"key" bson:"key" binding:"required"`
	Value interface{} `json:"value" bson:"value"`
}

// Like struct
type Like struct {
	Key   string `json:"key" bson:"key" binding:"required"`
	Value string `json:"value" bson:"value"`
}

// ElemMatch struct
type ElemMatch struct {
	Key   string   `json:"key" bson:"key" binding:"required"`
	Value []*Match `json:"value" bson:"value"`
}

// RangeValue struct
type RangeValue struct {
	Gt  interface{} `json:"gt" bson:"gt"`
	Lt  interface{} `json:"lt" bson:"lt"`
	Gte interface{} `json:"gte" bson:"gte"`
	Lte interface{} `json:"lte" bson:"lte"`
}

// Range struct
type Range struct {
	Key   string      `json:"key" bson:"key" binding:"required"`
	Value *RangeValue `json:"value" bson:"value"`
}

// In struct
type In struct {
	Key   string        `json:"key" bson:"key" binding:"required"`
	Value []interface{} `json:"value" bson:"value"`
}

// NotIn struct
type NotIn struct {
	Key   string        `json:"key" bson:"key" binding:"required"`
	Value []interface{} `json:"value" bson:"value"`
}

// All struct
type All struct {
	Key   string        `json:"key" bson:"key" binding:"required"`
	Value []interface{} `json:"value" bson:"value"`
}

// Filter struct
type Filter struct {
	Range []Range `json:"range" bson:"range"`
	In    []In    `json:"in" bson:"in"`
	NotIn []NotIn `json:"nin" bson:"nin"`
	All   []All   `json:"all" bson:"all"`
}

// Sort struct
type Sort struct {
	Key   string `json:"key" bson:"key" binding:"required"`
	Value string `json:"value" bson:"value"`
}

// Limit struct
type Limit struct {
	Rows   int `json:"rows" bson:"rows"`
	Offset int `json:"offset" bson:"offset"`
}

// Query struct
type Query struct {
	Match     []*Match     `json:"match" bson:"match"`
	Like      []*Like      `json:"like" bson:"like"`
	ElemMatch []*ElemMatch `json:"elem_match" bson:"elem_match"`
}

// Params struct
type Params struct {
	Query  Query       `json:"query" bson:"query"`
	Filter Filter      `json:"filter" bson:"filter"`
	Sort   []Sort      `json:"sort" bson:"sort"`
	Limit  Limit       `json:"limit" bson:"limit"`
	Data   interface{} `json:"data,omitempty" bson:"data,omitempty"`
}

// Error struct
type Error struct {
	Level          string      `json:"level,omitempty" bson:"level,omitempty"`
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

// Claims struct
type Claims struct {
	Ver          string `json:"ver" bson:"ver"`
	Cre          string `json:"cre" bson:"cre"`
	Id           string `json:"id" bson:"id"`
	Jti          string `json:"jti" bson:"jti"`
	Iss          string `json:"iss" bson:"iss"`
	Aud          string `json:"aud" bson:"aud"`
	ClientId     string `json:"client_id" bson:"client_id"`
	Sub          string `json:"sub" bson:"sub"`
	Exp          int    `json:"exp" bson:"exp"`
	Expired      int    `json:"expires" bson:"expires"`
	Iat          int    `json:"iat" bson:"iat"`
	TokenType    string `json:"token_type" bson:"token_type"`
	Scope        string `json:"scope" bson:"scope"`
	UserId       string `json:"user_id" bson:"user_id"`
	CredentialId string `json:"credential_id" bson:"credential_id"`
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

// GetMatchValue function
func GetMatchValue(i interface{}) interface{} {
	switch value := i.(type) {
	case int, float64, bool:
		return value
	case string:
		return bsonx.Regex("^"+value+"$", "i")
	case map[string]interface{}:
		var match Match
		if err := helper.PairValues(value, &match); err != nil {
			return nil
		}
		if strings.HasPrefix(match.Key, "$") && len(match.Key) > 1 {
			if match.Value == nil {
				return nil
			}
			return bson.M{match.Key: GetMatchValueOp(match.Value)}
		}
		return bson.M{match.Key: GetMatchValue(match.Value)}
	case []interface{}:
		var mi map[string]interface{}
		for _, v := range value {
			var match Match
			if err := helper.PairValues(v, &match); err == nil {
				if mi == nil {
					mi = make(map[string]interface{})
				}
				if strings.HasPrefix(match.Key, "$") && len(match.Key) > 1 {
					mi[match.Key] = GetMatchValueOp(match.Value)
				} else {
					mi[match.Key] = GetMatchValue(match.Value)
				}
			}
		}
		return mi
	default:
		return nil
	}
}

// GetMatchValueOp function
func GetMatchValueOp(i interface{}) interface{} {
	switch value := i.(type) {
	case int:
		return value
	case string:
		return value
	case float64:
		return value
	case bool:
		return value
	case map[string]interface{}:
		var match Match
		err := helper.PairValues(value, &match)
		if err != nil {
			return nil
		}
		if strings.HasPrefix(match.Key, "$") && len(match.Key) > 1 {
			if match.Value == nil {
				return nil
			}
			return bson.M{match.Key: GetMatchValueOp(match.Value)}
		}
		return bson.M{match.Key: GetMatchValue(match.Value)}
	case []interface{}:
		var m []interface{}
		for _, v := range value {
			m = append(m, GetMatchValue(v))
		}
		return m
	default:
		return nil
	}
}

// ParseMatchValue function
func ParseMatchValue(match *Match, query db.Map) error {
	switch value := match.Value.(type) {
	case map[string]interface{}:
		var m Match
		err := helper.PairValues(value, &m)
		if err != nil {
			return err
		}
		if strings.HasPrefix(m.Key, "$") && len(m.Key) > 1 {
			if m.Value == nil {
				return nil
			}
			query[match.Key] = bson.M{
				m.Key: GetMatchValueOp(m.Value),
			}
			return nil
		}
		query[match.Key+"."+m.Key] = GetMatchValue(m.Value)
		return nil
	default:
		query[match.Key] = GetMatchValue(value)
		return nil
	}
}

// ParseMatchValueReturn function
func ParseMatchValueReturn(match *Match) (db.Map, error) {
	var query = make(db.Map)
	switch value := match.Value.(type) {
	case map[string]interface{}:
		var m Match
		err := helper.PairValues(value, &m)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(m.Key, "$") && len(m.Key) > 1 {
			if m.Value == nil {
				return nil, nil
			}
			query[match.Key] = bson.M{
				m.Key: GetMatchValueOp(m.Value),
			}
			return query, nil
		}
		query[match.Key+"."+m.Key] = GetMatchValue(m.Value)
		return query, nil
	default:
		query[match.Key] = GetMatchValue(value)
		return query, nil
	}
}

// ParseLikeValue function
func ParseLikeValue(like *Like, query db.Map) error {
	if like.Value != "" {
		query[like.Key] = bsonx.Regex(".*"+like.Value+".*", "i")
	}
	return nil
}

// ParseElemMatchValue function
func ParseElemMatchValue(elemMatch *ElemMatch, query db.Map) (err error) {
	q, err := GetQueryMatchParamsReturn(elemMatch.Value)
	if err != nil {

	}
	query[elemMatch.Key] = db.Map{
		"$elemMatch": q,
	}
	return err
}

// ParseInValue function
func ParseInValue(i []interface{}) db.Map {

	var in []interface{}
	for _, v := range i {
		switch value := v.(type) {
		case string:
			in = append(in, bsonx.Regex("^"+value+"$", "i"))
		case int:
			in = append(in, value)
		case float64:
			in = append(in, value)
		case bool:
			in = append(in, value)
		default:
		}
	}

	return db.Map{"$in": in}
}

// ParseNotInValue function
func ParseNotInValue(i []interface{}) db.Map {

	var in []interface{}
	for _, v := range i {
		switch value := v.(type) {
		case string:
			in = append(in, bsonx.Regex("^"+value+"$", "i"))
		case int:
			in = append(in, value)
		case float64:
			in = append(in, value)
		case bool:
			in = append(in, value)
		default:
		}
	}

	return db.Map{"$nin": in}
}

// ParseAllValue function
func ParseAllValue(i []interface{}) db.Map {

	var all []interface{}
	for _, v := range i {
		switch value := v.(type) {
		case string:
			all = append(all, bsonx.Regex("^"+value+"$", "i"))
		case int:
			all = append(all, value)
		case float64:
			all = append(all, value)
		case bool:
			all = append(all, value)
		default:
		}
	}

	return db.Map{"$all": all}
}

// GetQueryMatchParams function
func GetQueryMatchParams(match []*Match, query db.Map) (int, error) {
	parsed := 0
	for _, v := range match {
		if strings.Trim(v.Key, " ") == "" {
			continue
		}
		err := ParseMatchValue(v, query)
		if err != nil {
			return 0, err
		}
		parsed++
	}

	return parsed, nil
}

// GetQueryMatchParamsReturn function
func GetQueryMatchParamsReturn(match []*Match) (db.Map, error) {
	query := db.Map{}
	for _, v := range match {
		if strings.Trim(v.Key, " ") == "" {
			continue
		}
		err := ParseMatchValue(v, query)
		if err != nil {
			return nil, err
		}
	}

	return query, nil
}

// GetQueryLikeParams function
func GetQueryLikeParams(like []*Like, query db.Map) (int, error) {
	parsed := 0
	for _, v := range like {
		if strings.Trim(v.Key, " ") == "" || strings.Trim(v.Value, " ") == "" {
			continue
		}
		err := ParseLikeValue(v, query)
		if err != nil {
			return 0, err
		}
		parsed++
	}

	return parsed, nil
}

// GetQueryElemMatchParams function
func GetQueryElemMatchParams(elemMatch []*ElemMatch, query db.Map) (int, error) {
	parsed := 0
	for _, v := range elemMatch {
		if strings.Trim(v.Key, " ") == "" {
			continue
		}
		err := ParseElemMatchValue(v, query)
		if err != nil {
			return 0, err
		}
		parsed++
	}

	return parsed, nil
}

// GetFilterRangeParams function
func GetFilterRangeParams(ranges []Range, query db.Map) (int, error) {
	parsed := 0
	for _, v := range ranges {
		if strings.Trim(v.Key, " ") == "" || v.Value == nil {
			continue
		}
		if v.Value != nil {
			r := db.Map{}
			if v.Value.Gt != nil {
				r["$gt"] = v.Value.Gt
			}
			if v.Value.Lt != nil {
				r["$lt"] = v.Value.Lt
			}
			if v.Value.Gte != nil {
				r["$gte"] = v.Value.Gte
			}
			if v.Value.Lte != nil {
				r["$lte"] = v.Value.Lte
			}
			query[v.Key] = r
		}
		parsed++
	}

	return parsed, nil
}

// GetFilterInParams function
func GetFilterInParams(in []In, query db.Map) (int, error) {
	parsed := 0
	for _, v := range in {
		if strings.Trim(v.Key, " ") == "" || v.Value == nil {
			continue
		}
		if v.Value != nil {
			query[v.Key] = ParseInValue(v.Value)
		}
		parsed++
	}

	return parsed, nil
}

// GetFilterNotInParams function
func GetFilterNotInParams(in []NotIn, query db.Map) (int, error) {
	parsed := 0
	for _, v := range in {
		if strings.Trim(v.Key, " ") == "" || v.Value == nil {
			continue
		}
		if v.Value != nil {
			query[v.Key] = ParseNotInValue(v.Value)
		}
		parsed++
	}

	return parsed, nil
}

// GetFilterAllParams function
func GetFilterAllParams(all []All, query db.Map) (int, error) {
	parsed := 0
	for _, v := range all {
		if strings.Trim(v.Key, " ") == "" || v.Value == nil {
			continue
		}
		if v.Value != nil {
			query[v.Key] = ParseAllValue(v.Value)
		}
		parsed++
	}

	return parsed, nil
}

// ParseSearchQuery function
func ParseSearchQuery(params Params, required ...bool) (db.Map, error) {

	q := db.Map{}
	parsed := 0

	//parse query::match clause :
	queryMatchParsed, err := GetQueryMatchParams(params.Query.Match, q)
	if err != nil {
		return nil, err
	}
	parsed = parsed + queryMatchParsed

	//parse query::like clause :
	queryLikeParsed, err := GetQueryLikeParams(params.Query.Like, q)
	if err != nil {
		return nil, err
	}
	parsed = parsed + queryLikeParsed

	//parse query::elemMatch clause :
	queryElemMatchParsed, err := GetQueryElemMatchParams(params.Query.ElemMatch, q)
	if err != nil {
		return nil, err
	}
	parsed = parsed + queryElemMatchParsed

	//parse filter::range clause :
	filterRangeParsed, err := GetFilterRangeParams(params.Filter.Range, q)
	if err != nil {
		return nil, err
	}
	parsed = parsed + filterRangeParsed

	//parse filter::in clause :
	filterInParsed, err := GetFilterInParams(params.Filter.In, q)
	if err != nil {
		return nil, err
	}
	parsed = parsed + filterInParsed

	//parse filter::nin clause :
	filterNotInParsed, err := GetFilterNotInParams(params.Filter.NotIn, q)
	if err != nil {
		return nil, err
	}
	parsed = parsed + filterNotInParsed

	//parse filter::all clause :
	filterAllParsed, err := GetFilterAllParams(params.Filter.All, q)
	if err != nil {
		return nil, err
	}
	parsed = parsed + filterAllParsed

	if len(required) > 0 {
		if required[0] && parsed == 0 {
			return nil, errors.New("required at least one params query or filter")
		}
	}

	return q, nil
}

// ParseTokenClaims function
func ParseTokenClaims(c *gin.Context) *Claims {
	var claims *Claims
	a, ok := c.Get("claims")
	if ok {
		err := helper.PairValues(a, &claims)
		if err != nil {
			return nil
		}
		if claims.Exp == 0 && claims.Expired != 0 {
			claims.Exp = claims.Expired
		}
		if claims.Exp != 0 && claims.Expired == 0 {
			claims.Expired = claims.Exp
		}
	}
	b, ok := c.Get("credential_id")
	if ok {
		claims.CredentialId = b.(string)
		if claims.Cre == "" {
			claims.Cre = claims.CredentialId
		}
	}
	d, ok := c.Get("user_id")
	if ok {
		claims.UserId = d.(string)
		if claims.Sub == "" {
			claims.Sub = claims.UserId
		}
	}

	claims.ReArrangeValues()

	return claims
}

// ReArrangeValues function
func (claims *Claims) ReArrangeValues() {
	if claims.CredentialId == "" && claims.Cre != "" {
		claims.CredentialId = claims.Cre
	}
	if claims.UserId == "" && claims.Sub != "" {
		claims.UserId = claims.Sub
	}
	if claims.ClientId == "" && claims.Aud != "" {
		claims.ClientId = claims.Aud
	}
	if claims.Aud == "" && claims.ClientId != "" {
		claims.Aud = claims.ClientId
	}
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

// Error method
func (e Error) Error() error {
	b, err := json.Marshal(e)
	if err != nil {
		return errors.New("unknown error")
	}
	return errors.New(string(b))
}

// Err method
func (e Error) Err() string {
	b, err := json.Marshal(e)
	if err != nil {
		return "unknown error"
	}
	return string(b)
}

// Log method
func (e Error) Log() {
	b, err := json.Marshal(e)
	if err != nil {
		b = []byte("unknown error")
	}
	log.Println(string(b))
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

// PairTo method
func (a *Attributes) PairTo(o interface{}) error {
	return helper.PairValues(a, &o)
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
