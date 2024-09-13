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

package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"github.com/jasacloud/go-libraries/config"
	"github.com/jasacloud/go-libraries/helper"
	"github.com/jasacloud/go-libraries/system"
	"golang.org/x/crypto/acme/autocert"
	"io"
	"io/ioutil"
	"log"
	"path"
	"reflect"
	"runtime"
	"time"
)

// Listen struct
type Listen struct {
	ListenAddr string `json:"listenAddr" bson:"listenAddr"`
	ListenPort string `json:"listenPort" bson:"listenPort"`
	Ssl        bool   `json:"ssl" bson:"ssl"`
}

// ListenConf struct
type ListenConf struct {
	Listen Listen `json:"listen" bson:"listen"`
}

// ListenSsl struct
type ListenSsl struct {
	AutoTls    bool     `json:"autotls" bson:"autotls"`
	Domain     []string `json:"domain" bson:"domain"`
	ListenAddr string   `json:"listenAddr" bson:"listenAddr"`
	ListenPort string   `json:"listenPort" bson:"listenPort"`
	CertFile   string   `json:"certFile" bson:"certFile"`
	KeyFile    string   `json:"keyFile" bson:"keyFile"`
	SslOnly    bool     `json:"sslOnly" bson:"sslOnly"`
}

// ListenSslConf struct
type ListenSslConf struct {
	ListenSsl ListenSsl `json:"ssl" bson:"ssl"`
}

// ModeConfg struct
type ModeConfg struct {
	Mode string `json:"mode" bson:"mode"`
}

var (
	// Route variable
	Route *gin.Engine
	// ListenConfig variable
	ListenConfig ListenConf
	// ListenSslConfig variable
	ListenSslConfig ListenSslConf
	// Mode variable
	Mode ModeConfg
)

// Map type
type Map map[string]interface{}

// Config type config
type Config *config.Config

// DefaultHeader function
func DefaultHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Server", "JCServer")
		c.Next()
		c.Writer.Header().Set("Server", "JCServer")
	}
}

// addr method
func (l Listen) addr() string {
	if l.ListenAddr == "" {
		return ":" + l.ListenPort
	}

	if l.ListenAddr == "*" {
		return "0.0.0.0:" + l.ListenPort
	}
	return l.ListenAddr + ":" + l.ListenPort
}

// addr method
func (l ListenSsl) addr() string {
	if l.ListenAddr == "" {
		return ":" + l.ListenPort
	}

	if l.ListenAddr == "*" {
		return "0.0.0.0:" + l.ListenPort
	}
	return l.ListenAddr + ":" + l.ListenPort
}

// LoadServer function
func LoadServer() {
	c := config.GetConfig()
	config.GetConf(c.ByteConfig, &ListenConfig)
	config.GetConf(c.ByteConfig, &ListenSslConfig)
	config.GetConf(c.ByteConfig, &Mode)
	config.GetConf(c.ByteConfig, &Cors)
	setMode(Mode)
	Route.Use(latencyHandler)
	Route.Use(CorsHandler())
}

// Start function
func Start() {
	if ListenConfig.Listen.Ssl {
		if ListenSslConfig.ListenSsl.AutoTls {
			m := autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist(ListenSslConfig.ListenSsl.Domain...),
				Cache:      autocert.DirCache(".cache"),
			}
			runWithManager(Route, &m)
		} else {
			if ListenSslConfig.ListenSsl.SslOnly == true {
				if runtime.GOOS == "windows" {
					ListenSslConfig.ListenSsl.CertFile = path.Join(config.GetConfigDir(), ListenSslConfig.ListenSsl.CertFile)
					ListenSslConfig.ListenSsl.KeyFile = path.Join(config.GetConfigDir(), ListenSslConfig.ListenSsl.KeyFile)
				}
				fmt.Println("Listen ssl/tls only")
				runTLS(ListenSslConfig.ListenSsl.addr(), Route, ListenSslConfig.ListenSsl.CertFile, ListenSslConfig.ListenSsl.KeyFile)
			} else {
				go func() {
					run(Route, ListenConfig.Listen.addr())
				}()
				if runtime.GOOS == "windows" {
					ListenSslConfig.ListenSsl.CertFile = path.Join(config.GetConfigDir(), ListenSslConfig.ListenSsl.CertFile)
					ListenSslConfig.ListenSsl.KeyFile = path.Join(config.GetConfigDir(), ListenSslConfig.ListenSsl.KeyFile)
				}
				runTLS(ListenSslConfig.ListenSsl.addr(), Route, ListenSslConfig.ListenSsl.CertFile, ListenSslConfig.ListenSsl.KeyFile)
			}
		}
	} else {
		run(Route, ListenConfig.Listen.addr())
	}
}

func Stop() {
	shutdown()
}

// startX function
func startX() {
	if ListenConfig.Listen.Ssl {
		if ListenSslConfig.ListenSsl.AutoTls {
			m := autocert.Manager{
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist(ListenSslConfig.ListenSsl.Domain...),
				Cache:      autocert.DirCache(".cache"),
			}
			err := autotls.RunWithManager(Route, &m)
			if err != nil {
				log.Fatal("Error: ", err)
			}
		} else {
			if ListenSslConfig.ListenSsl.SslOnly == true {
				if runtime.GOOS == "windows" {
					ListenSslConfig.ListenSsl.CertFile = path.Join(config.GetConfigDir(), ListenSslConfig.ListenSsl.CertFile)
					ListenSslConfig.ListenSsl.KeyFile = path.Join(config.GetConfigDir(), ListenSslConfig.ListenSsl.KeyFile)
				}
				fmt.Println("Listen ssl/tls only")
				err := Route.RunTLS(ListenSslConfig.ListenSsl.addr(), ListenSslConfig.ListenSsl.CertFile, ListenSslConfig.ListenSsl.KeyFile)
				if err != nil {
					fmt.Println("Error: ", err)
				}
			} else {
				go Route.Run(ListenConfig.Listen.addr())
				if runtime.GOOS == "windows" {
					ListenSslConfig.ListenSsl.CertFile = path.Join(config.GetConfigDir(), ListenSslConfig.ListenSsl.CertFile)
					ListenSslConfig.ListenSsl.KeyFile = path.Join(config.GetConfigDir(), ListenSslConfig.ListenSsl.KeyFile)
				}
				err := Route.RunTLS(ListenSslConfig.ListenSsl.addr(), ListenSslConfig.ListenSsl.CertFile, ListenSslConfig.ListenSsl.KeyFile)
				if err != nil {
					fmt.Println("Error: ", err)
				}
			}
		}
	} else {
		err := Route.Run(ListenConfig.Listen.addr())
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

// setMode function
func setMode(modeconfig ModeConfg) {
	switch modeconfig.Mode {
	case "production":
		gin.SetMode(gin.ReleaseMode)
		Route = gin.New()
		Route.NoRoute(NotFoundResponse)
	case "release":
		gin.SetMode(gin.ReleaseMode)
		Route = gin.New()
		Route.NoRoute(NotFoundResponse)
	case "development":
		Route = gin.Default()
		Route.NoRoute(NotFoundResponse)
	default:
		Route = gin.Default()
		Route.NoRoute(NotFoundResponse)
	}
}

// NotFoundResponse function
func NotFoundResponse(c *gin.Context) {
	c.JSON(404, gin.H{
		"returnval": false,
		"error": gin.H{
			"code":            "404",
			"message":         "Page Not Found",
			"message_details": "Page you are looking is not found",
			"type":            "request",
			"status":          "PAGE_NOT_FOUND",
		},
	})
}

// latencyHandler function
func latencyHandler(c *gin.Context) {
	c.Set("x-request-received", time.Now())
	c.Writer.Header().Set("x-request-received", system.GetTimeStampString())
	c.Writer.Header().Set("Server", "JCServer")
	c.Next()
}

// GinUnmarshal function
func GinUnmarshal(r io.Reader) (gin.H, error) {
	g := gin.H{}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return g, err
	}
	err = json.Unmarshal(b, &g)
	return g, err
}

// GinReUnmarshal function
func GinReUnmarshal(i interface{}) (gin.H, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	err := enc.Encode(i)
	if err != nil {
		return nil, err
	}
	b := buf.Bytes()
	g := gin.H{}
	err = json.Unmarshal(b, &g)
	if err != nil {
		return nil, err
	}
	return g, nil
}

// ResponseJSON function
func ResponseJSON(c *gin.Context, code int, obj interface{}) {
	//start response sent :
	if val, ok := c.Get("x-request-received"); ok && val != nil {
		t, ok := val.(time.Time)
		if ok {
			c.Writer.Header().Set("x-server-execution", time.Since(t).String())
		}
	}
	c.Writer.Header().Set("x-response-sent", system.GetTimeStampString())
	if code == 0 {
		code = 200
	}
	c.JSON(code, obj)
	c.Abort()
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
	err = helper.BindValidate(o)
	if err != nil {

		return err
	}

	return nil
}

// ParseRequest function
func ParseRequest(c *gin.Context, request interface{}) error {
	err := c.BindJSON(&request)
	if err != nil {
		return err
	}

	return nil
}

// Error function
func Error(c *gin.Context, code string, message string, t ...string) gin.H {
	request := gin.H{}
	err := c.ShouldBindJSON(&request)
	if err == nil && request["kind"] != nil {
		if len(t) > 0 {
			return gin.H{
				"kind":      request["kind"],
				"returnval": false,
				"error": gin.H{
					"code":    code,
					"message": message,
					"type":    t[0],
				},
			}
		}
		return gin.H{
			"kind":      request["kind"],
			"returnval": false,
			"error": gin.H{
				"code":    code,
				"message": message,
			},
		}
	}
	if len(t) > 0 {
		return gin.H{
			"returnval": false,
			"error": gin.H{
				"code":    code,
				"message": message,
				"type":    t[0],
			},
		}
	}
	return gin.H{
		"returnval": false,
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	}
}

// ErrorResponse function
func ErrorResponse(code string, message string) gin.H {
	return gin.H{
		"code":    code,
		"message": message,
	}
}
