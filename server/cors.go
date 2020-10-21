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
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

// CorsOptions is option of cors that defined from config
type CorsOptions struct {
	UseSetting       bool     `json:"useSetting" bson:"useSetting"`
	AllowAllOrigins  bool     `json:"allowAllOrigins" bson:"allowAllOrigins"`
	AllowOrigins     []string `json:"allowOrigins" bson:"allowOrigins"`
	AllowMethods     []string `json:"allowMethods" bson:"allowMethods"`
	AllowHeaders     []string `json:"allowHeaders" bson:"allowHeaders"`
	AllowCredentials bool     `json:"allowCredentials" bson:"allowCredentials"`
	ExposeHeaders    []string `json:"exposeHeaders" bson:"exposeHeaders"`
	AllowWildcard    bool     `json:"allowWildcard" bson:"allowWildcard"`
	MaxAgeSec        int      `json:"maxAgeSec" bson:"maxAgeSec"`
}

// CorsConf struct
type CorsConf struct {
	CorsOptions CorsOptions `json:"cors" bson:"cors"`
}

// Cors variable
var Cors CorsConf

// CorsHandler function
func CorsHandler() gin.HandlerFunc {
	if !Cors.CorsOptions.UseSetting {
		return func(c *gin.Context) {
			return
		}
	}
	for _, v := range Cors.CorsOptions.AllowOrigins {
		if v == "*" {
			return cors.New(cors.Config{
				AllowMethods:     Cors.CorsOptions.AllowMethods,
				AllowHeaders:     Cors.CorsOptions.AllowHeaders,
				ExposeHeaders:    Cors.CorsOptions.ExposeHeaders,
				AllowCredentials: Cors.CorsOptions.AllowCredentials,
				MaxAge:           time.Duration(Cors.CorsOptions.MaxAgeSec) * time.Second,
				AllowWildcard:    Cors.CorsOptions.AllowWildcard,
				AllowOriginFunc: func(origin string) bool {
					return true
				},
			})
		}
	}
	return cors.New(cors.Config{
		AllowOrigins:     Cors.CorsOptions.AllowOrigins,
		AllowMethods:     Cors.CorsOptions.AllowMethods,
		AllowHeaders:     Cors.CorsOptions.AllowHeaders,
		ExposeHeaders:    Cors.CorsOptions.ExposeHeaders,
		AllowCredentials: Cors.CorsOptions.AllowCredentials,
		MaxAge:           time.Duration(Cors.CorsOptions.MaxAgeSec) * time.Second,
		AllowWildcard:    Cors.CorsOptions.AllowWildcard,
	})
}
