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

package serial

import (
	"github.com/jasacloud/go-libraries/config"
	"github.com/tarm/serial"
	"log"
)

// SerialConf struct
type SerialConf struct {
	SerialOption []SerialOption `json:"serialResources" bson:"serialResources"`
}

// SerialOption struct
type SerialOption struct {
	Name    string `json:"name" bson:"name"`
	ComName string `json:"comName" bson:"comName"`
	Baud    int    `json:"baud" bson:"baud"`
}

var (
	// Serial variable
	Serial SerialConf
)

// SerialConn variable
var SerialConn = make(map[string]*serial.Port)

// GetSerialResource function
func GetSerialResource(resourceName string) SerialOption {

	c := config.GetConfig()
	config.GetConf(c.ByteConfig, &Serial)
	for _, v := range Serial.SerialOption {
		if v.Name == resourceName {

			return v
		}
	}

	return SerialOption{}
}

// SerialConnect function
func SerialConnect(resourceName string) (*serial.Port, error) {

	opt := GetSerialResource(resourceName)

	if SerialConn[resourceName] == nil {
		cfg := &serial.Config{Name: opt.ComName, Baud: opt.Baud}
		s, err := serial.OpenPort(cfg)
		if err != nil {
			return s, err
		}
		log.Println("serial connected to", opt.ComName)
		SerialConn[resourceName] = s
	}

	return SerialConn[resourceName], nil
}

// Close function
func Close(resourceName string) {

	err := SerialConn[resourceName].Close()
	if err != nil {
		log.Println("error while close the serial COM:", resourceName)
	}
	SerialConn[resourceName] = nil
}
