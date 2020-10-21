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

package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/jasacloud/go-libraries/config"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

// Properties struct
type Properties struct {
	Name  string `json:"name" bson:"name"`
	Value string `json:"value" bson:"value"`
}

// Params struct
type Params struct {
	q []Properties
}

// ResourceOptions struct
type ResourceOptions struct {
	PreHeaders bool
	PreParams  bool
}

// HttpResource struct
type HttpResource struct {
	Name       string       `json:"name" bson:"name"`
	Url        string       `json:"url" bson:"url"`
	Uri        string       `json:"uri" bson:"uri"`
	PreHeaders []Properties `json:"preHeaders" bson:"preHeaders"`
	PreParams  []Properties `json:"preParams" bson:"preParams"`
}

// HttpServer struct
type HttpServer struct {
	HttpServer []HttpResource `json:"httpResources" bson:"httpResources"`
}

// Http struct
type Http struct {
	Client       *http.Client
	HttpResource HttpResource
	Request      *http.Request
	Err          error
}

// Config type config
type Config config.Config

// GetHttpResource function
func GetHttpResource(name string) HttpResource {
	c := config.GetConfig()
	var httpConfig HttpServer
	config.GetConf(c.ByteConfig, &httpConfig)

	return getHttpConf(httpConfig.HttpServer, name)
}

// Join method
func (h HttpResource) Join(elem ...string) string {
	u, err := url.Parse(h.Url)
	if err != nil {
		u := strings.TrimRight(h.Url, "/")
		if h.Uri != "" {
			u = u + "/" + strings.Trim(h.Uri, "/")
		}
		return strings.TrimRight(u, "/") + "/" + path.Join(elem...)
	}
	elem = append([]string{u.Path, h.Uri}, elem...)
	u.Path = path.Join(elem...)

	return u.String()
}

// GetUrl method
func (h *Http) GetUrl() string {
	return h.HttpResource.Url
}

// SetUri method
func (h *Http) SetUri(uri string) {
	h.HttpResource.Uri = uri
}

// LoadHttpResource function
func LoadHttpResource(resourceName string, options ...*ResourceOptions) Http {
	c := config.GetConfig()
	var httpConfig HttpServer
	config.GetConf(c.ByteConfig, &httpConfig)
	httpServer := httpConfig.HttpServer

	var h Http

	h.HttpResource = getHttpConf(httpServer, resourceName)

	if len(options) > 0 {
		if !options[0].PreHeaders {
			h.HttpResource.PreHeaders = nil
		}
		if !options[0].PreParams {
			h.HttpResource.PreParams = nil
		}
	} else {
		h.HttpResource.PreHeaders = nil
		h.HttpResource.PreParams = nil
	}

	return h
}

// LoadHttp function
func LoadHttp(url string) Http {

	var h Http
	h.HttpResource.Url = url

	return h
}

func getHttpConf(httpServer []HttpResource, resourceName string) HttpResource {

	for _, v := range httpServer {
		if v.Name == resourceName {
			return v
		}
	}

	return HttpResource{}
}

// SetPath method
func (h *Http) SetPath(strPath string) {
	if h.Request != nil {
		if !strings.HasPrefix(strPath, "/") {
			strPath = "/" + strPath
		}
		h.Request.URL.Path = strPath
	} else {
		u, err := url.Parse(h.HttpResource.Url)
		if err != nil {
			log.Println("url.Parse:", h.HttpResource.Url, err)
			return
		}
		u.Path = strPath
		h.HttpResource.Url = u.String()
		h.HttpResource.Uri = ""
	}
}

// AppendPath method
func (h *Http) AppendPath(strPath string) {
	if h.Request != nil {
		h.Request.URL.Path = path.Join(h.Request.URL.Path, strPath)
	} else {
		h.HttpResource.Uri = path.Join(h.HttpResource.Uri, strPath)
	}
}

// SetHeader method
func (h *Http) SetHeader(name string, value string) {

	h.Request.Header.Set(name, value)
}

// SetBasicAuthorization method
func (h *Http) SetBasicAuthorization(username, password string) {

	h.Request.SetBasicAuth(username, password)
}

// SetBearerAuthorization method
func (h *Http) SetBearerAuthorization(token string) {
	if strings.HasPrefix(token, "Bearer") {
		h.SetHeader("Authorization", token)
	} else {
		h.SetHeader("Authorization", "Bearer "+token)
	}
}

// Start method
func (h *Http) Start() (*http.Response, error) {
	return h.Client.Do(h.Request)
}

// Do method
func (h *Http) Do(i interface{}) error {
	h.SetClose(true)
	resp, err := h.Client.Do(h.Request)
	if err != nil {
		log.Println("request error:", err)
		log.Println("url:", h.Request.URL.String())

		return err
	}
	b, _ := ioutil.ReadAll(resp.Body)
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	err = json.Unmarshal(b, i)
	if err != nil {
		log.Println("parse resp json error:", err)
		log.Println("from:", h.Request.URL.String())
		log.Println("status:", resp.Status)
		log.Println("body:", string(b))

		return err
	}
	return nil
}

// SetCookie method
func (h *Http) SetCookie(name string, value string) {
	cookie := http.Cookie{Name: name, Value: value}
	h.Request.AddCookie(&cookie)
}

// SetTimeout method
func (h *Http) SetTimeout(timeoutDuration time.Duration) {
	h.Client.Timeout = timeoutDuration
}

// SetQueryParams method
func (h *Http) SetQueryParams(p *Params) {
	q := h.Request.URL.Query()
	for _, v := range p.q {
		q.Add(v.Name, v.Value)
	}
	h.Request.URL.RawQuery = q.Encode()
}

// AddQueryParam method
func (h *Http) AddQueryParam(key, value string) {
	q := h.Request.URL.Query()
	q.Add(key, value)
	h.Request.URL.RawQuery = q.Encode()
}

// SetQueryParam method
func (h *Http) SetQueryParam(key, value string) {
	q := h.Request.URL.Query()
	q.Set(key, value)
	h.Request.URL.RawQuery = q.Encode()
}

// DeleteQueryParam method
func (h *Http) DeleteQueryParam(key string) {
	q := h.Request.URL.Query()
	q.Del(key)
	h.Request.URL.RawQuery = q.Encode()
}

// GetQueryParam method
func (h *Http) GetQueryParam(key string) string {
	q := h.Request.URL.Query()

	return q.Get(key)
}

// Add method
func (p *Params) Add(key, val string) {
	if key != "" {
		p.q = append(p.q, Properties{key, val})
	}
}

// SetRequestJSON method
func (h *Http) SetRequestJSON(method string, b []byte) {
	h.SetRequest(method, "application/json", bytes.NewReader(b))
}

// SetClose method
func (h *Http) SetClose(close bool) {
	if h.Request != nil {
		h.Request.Close = close
	}
}

// SetRequest method
func (h *Http) SetRequest(method string, contentTyppe string, body io.Reader) {
	var httpPolicy RedirectPolicyFunc
	h.Client = &http.Client{
		CheckRedirect: httpPolicy,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	strUrl := ""
	if h.HttpResource.Url != "" {
		strUrl += h.HttpResource.Url
	}
	if h.HttpResource.Uri != "" {
		strUrl = strings.TrimSuffix(strUrl, "/") + "/" + strings.TrimPrefix(h.HttpResource.Uri, "/") + "?"
	}
	h.Request, h.Err = http.NewRequest(method, strUrl, body)
	if h.Err != nil {
		log.Println("client.Http.SetRequest() http.NewRequest init error:", method, strUrl)
	}
	if contentTyppe != "" {
		h.SetHeader("Content-Type", contentTyppe)
	}
	if h.HttpResource.PreParams != nil {
		q := h.Request.URL.Query()
		for _, v := range h.HttpResource.PreParams {
			q.Add(v.Name, v.Value)
		}
		h.Request.URL.RawQuery = q.Encode()
	}
	if h.HttpResource.PreHeaders != nil {
		for _, v := range h.HttpResource.PreHeaders {
			h.Request.Header.Add(v.Name, v.Value)
		}
	}
	h.Request.Header.Set("User-Agent", "JCClient/1.0")
}
