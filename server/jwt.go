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

import "github.com/jasacloud/go-libraries/config"

// JwtConf struct
type JwtConf struct {
	JwtOptions []JwtOption `json:"jwtResources" bson:"jwtResources"`
}

// JwtOption struct
type JwtOption struct {
	Name          string `json:"name" bson:"name"`
	Secret        string `json:"secret" bson:"secret"`
	Algorithm     string `json:"algorithm" bson:"algorithm"`
	ExpirationSec int    `json:"expiration" bson:"expiration"`
	PrivateKey    string `json:"privateKey" bson:"privateKey"`
	PublicKey     string `json:"publicKey" bson:"publicKey"`
}

var (
	// Secret variable
	Secret JwtConf
)

// PrivateKey variable
var PrivateKey = make(map[string]interface{})

// PublicKey variable
var PublicKey = make(map[string]interface{})

// DefaultResourceName variable
var DefaultResourceName = ""

// GetJwtOption function
func GetJwtOption(resourceName string) JwtOption {
	if DefaultResourceName == "" {
		DefaultResourceName = resourceName
	}
	c := config.GetConfig()
	config.GetConf(c.ByteConfig, &Secret)
	for _, v := range Secret.JwtOptions {
		if v.Name == resourceName {
			return v
		}
	}

	return JwtOption{}
}
