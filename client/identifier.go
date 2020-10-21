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
	"github.com/jasacloud/go-libraries/config"
)

// IdentifierResource struct
type IdentifierResource struct {
	Name       string `json:"name" bson:"name"`
	Identifier string `json:"identifier" bson:"identifier"`
	Password   string `json:"password" bson:"password"`
}

// IdentifierConf struct
type IdentifierConf struct {
	IdentifierResources []IdentifierResource `json:"authResources" bson:"authResources"`
}

var (
	// Identifier variable
	Identifier IdentifierConf
)

// GetIdentifierResource function
func GetIdentifierResource(resourceName string) IdentifierResource {
	c := config.GetConfig()
	config.GetConf(c.ByteConfig, &Identifier)
	for _, v := range Identifier.IdentifierResources {
		if v.Name == resourceName {

			return v
		}
	}

	return IdentifierResource{}
}
