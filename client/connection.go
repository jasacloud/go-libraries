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
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jasacloud/go-libraries/config"
	"github.com/jasacloud/go-libraries/system"
	"io/ioutil"
	"net/http"
	"time"
)

// ConnectionResource struct
type ConnectionResource struct {
	Name         string `json:"name" bson:"name"`
	HttpResource string `json:"HttpResource" bson:"HttpResource"`
	Cid          string `json:"cid" bson:"cid"`
	AuthResource string `json:"authResource" bson:"authResource"`
}

// ConnectionConf struct
type ConnectionConf struct {
	ConnectionResources []ConnectionResource `json:"connectionResources" bson:"connectionResources"`
}

var (
	// Conn variable
	Conn ConnectionConf
)

// GetConnectionResource function
func GetConnectionResource(resourceName string) ConnectionResource {
	c := config.GetConfig()
	config.GetConf(c.ByteConfig, &Conn)
	for _, v := range Conn.ConnectionResources {
		if v.Name == resourceName {

			return v
		}
	}

	return ConnectionResource{}
}

// Connect function
func Connect(resourceName string) ([]byte, error) {

	connectionResource := GetConnectionResource(resourceName)

	h := LoadHttpResource(connectionResource.HttpResource, &ResourceOptions{PreHeaders: true, PreParams: true})
	jsonStr, err := json.Marshal(gin.H{
		"kind": "connect#system",
		"cid":  connectionResource.Cid,
		"ver":  "1.1",
	})
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}

	body := bytes.NewBuffer(jsonStr)
	//http := Client.LoadHttp("https://localhost/api/token")
	h.SetRequest("POST", "application/json", body)
	resp, err := h.Start()
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	return b, err
}

// LoginConnection function
func LoginConnection(resourceName string, Code string) ([]byte, error) {
	connectionResource := GetConnectionResource(resourceName)
	authResource := GetIdentifierResource(connectionResource.AuthResource)
	h := LoadHttpResource(connectionResource.HttpResource, &ResourceOptions{PreHeaders: true, PreParams: true})
	jsonStr, err := json.Marshal(gin.H{
		"kind": "login#system",
		"code": Code,
		"user": authResource.Identifier,
		"pwd":  authResource.Password,
		"ver":  "1.1",
	})
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}

	body := bytes.NewBuffer(jsonStr)
	//http := Client.LoadHttp("https://localhost/api/token")
	h.SetRequest("POST", "application/json", body)
	resp, err := h.Start()
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	return b, err
}

// HeartbeatConnection function
func HeartbeatConnection(resourceName string, scode string, code string, hcode string) ([]byte, error) {
	connectionResource := GetConnectionResource(resourceName)
	authResource := GetIdentifierResource(connectionResource.AuthResource)
	h := LoadHttpResource(connectionResource.HttpResource, &ResourceOptions{PreHeaders: true, PreParams: true})

	jsonStr, err := json.Marshal(gin.H{
		"ver":    "1.1",
		"kind":   "heartbeatv3#hiq",
		"scode":  scode,
		"code":   code,
		"hcode":  hcode,
		"userid": authResource.Identifier,
	})
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}

	body := bytes.NewBuffer(jsonStr)
	h.SetRequest("POST", "application/json", body)
	h.SetCookie("_ips", system.Base64encode(system.GetAllIP()))
	resp, err := h.Start()
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	return b, err
}

// C struct
type C struct {
	State    int
	Response *http.Response
	Body     []byte
	Error    error
}

// Connection variable
var Connection C

// GoroutinesConnect function
func GoroutinesConnect() {
	for {

		if Connection.State != 1 {
			fmt.Println("Connection Sleep...")
			time.Sleep(time.Duration(6000 * time.Millisecond))
			continue
		}

		fmt.Println("Connection Reconnect...")
		Connection.Body, Connection.Error = Connect("default")
		Connection.State = 2
		time.Sleep(time.Duration(6000 * time.Millisecond))
	}
}

// L struct
type L struct {
	State    int
	Code     string
	Response *http.Response
	Body     []byte
	Error    error
	Scode    string
}

// Login variable
var Login L

// GoroutinesLogin function
func GoroutinesLogin() {
	for {
		if Login.State != 1 {
			fmt.Println("Login sleep...")
			time.Sleep(time.Duration(6000 * time.Millisecond))
			continue
		}

		if Connection.Body == nil {
			Connection.State = 1
			time.Sleep(time.Duration(6000 * time.Millisecond))
			continue
		}
		data := gin.H{}
		err := json.Unmarshal(Connection.Body, &data)
		//data, err := server.GinUnmarshal(Connection.Response.Body)
		if err != nil {
			Connection.State = 1
			time.Sleep(time.Duration(6000 * time.Millisecond))
			continue
		}

		if data["code"] != nil {
			Login.Code = data["code"].(string)
			fmt.Println("Login Reconnect...")
			Login.Body, Login.Error = LoginConnection("default", Login.Code)
			time.Sleep(time.Duration(6000 * time.Millisecond))
			Connection.State = 2
			Login.State = 2
		} else {
			Login.Code = ""
			Login.Scode = ""
			Login.State = 1
			time.Sleep(time.Duration(6000 * time.Millisecond))
		}
	}
}

// H struct
type H struct {
	State    int
	Response *http.Response
	Body     []byte
	Error    error
	Scode    string
}

// Heartbeat variable
var Heartbeat H

// GoroutinesHeartbeat function
func GoroutinesHeartbeat() {
	for {
		if Login.Body == nil {
			Login.State = 1
			time.Sleep(time.Duration(6000 * time.Millisecond))
			continue
		}
		data := gin.H{}
		err := json.Unmarshal(Login.Body, &data)
		//data, err := server.GinUnmarshal(Login.Response.Body)
		if err != nil {
			Login.State = 1
			time.Sleep(time.Duration(6000 * time.Millisecond))
			continue
		}
		if data["scode"] != nil {
			Login.Scode = data["scode"].(string)
			Login.State = 2
			Heartbeat.Body, Heartbeat.Error = HeartbeatConnection("default", Login.Scode, Login.Code, "7")
			data := gin.H{}
			err := json.Unmarshal(Heartbeat.Body, &data)
			if err != nil {
				fmt.Println("Heartbeat:", "Failed while parse data", err)
				continue
			}
			if data["returnval"] != true {
				fmt.Println("Heartbeat:", "Failed")
				Login.Scode = ""
				Login.State = 1
				Heartbeat.Body = nil
				Login.Body = nil
				time.Sleep(time.Duration(6000 * time.Millisecond))
				continue
			}
			fmt.Println("Heartbeat:", "OK")
			time.Sleep(time.Duration(6000 * time.Millisecond))
		} else {
			fmt.Println("Heartbeat:", "Failed")
			Login.Scode = ""
			Login.State = 1
			Heartbeat.Body = nil
			Login.Body = nil
			time.Sleep(time.Duration(6000 * time.Millisecond))
		}
	}
}
