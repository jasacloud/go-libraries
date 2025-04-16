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
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memcached"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/jasacloud/go-libraries/config"
)

// SessionOption struct
type SessionOption struct {
	Name          string `json:"name" bson:"name"`
	Type          string `json:"type" bson:"type"`
	Host          string `json:"host" bson:"host"`
	Secret        string `json:"secret" bson:"secret"`
	Username      string `json:"username" bson:"username"`
	Password      string `json:"password" bson:"password"`
	ExpirationSec int    `json:"expiration" bson:"expiration"`
	SessionName   string `json:"sessionName" bson:"sessionName"`
}

// SessionConf struct
type SessionConf struct {
	SessionOption []SessionOption `json:"sessionResources" bson:"sessionResources"`
}

var (
	// Sess variable
	Sess SessionConf
)

// sessionStore variable
var sessionStore = make(map[string]gin.HandlerFunc)

// GetSessionResource function
func GetSessionResource(resourceName string) SessionOption {
	c := config.GetConfig()
	config.GetConf(c.ByteConfig, &Sess)
	for _, v := range Sess.SessionOption {
		if v.Name == resourceName {

			return v
		}
	}

	return SessionOption{}
}

// GetSessionStore function
func GetSessionStore(option SessionOption) sessions.Store {
	switch option.Type {
	case "memcached":
		store := memcached.NewStore(memcache.New(option.Host), "", []byte(option.Secret))

		return store
	case "redis":
		store, _ := redis.NewStore(10, "tcp", option.Host, option.Username, option.Password, []byte(option.Secret))

		return store
	default:

		return nil
	}
}

// LoadSession function
func LoadSession(resourceName string) gin.HandlerFunc {
	option := GetSessionResource(resourceName)
	if sessionStore[resourceName] == nil {
		store := GetSessionStore(option)
		sessionStore[resourceName] = sessions.Sessions(option.SessionName, store)
	}

	return sessionStore[resourceName]
}

// Session function
func Session(c *gin.Context) sessions.Session {

	return c.MustGet(sessions.DefaultKey).(sessions.Session)
}
