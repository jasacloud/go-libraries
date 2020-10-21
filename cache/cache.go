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

package cache

import (
	"github.com/gin-contrib/cache/persistence"
	"github.com/jasacloud/go-libraries/config"
	"time"
)

// MemcachedConf struct
type MemcachedConf struct {
	CacheOptions []MemcachedOption `json:"memcachedResources" bson:"memcachedResources"`
}

// RedisConf struct
type RedisConf struct {
	CacheOptions []RedisOption `json:"redisResources" bson:"redisResources"`
}

// MemcachedOption struct
type MemcachedOption struct {
	Name          string   `json:"name" bson:"name"`
	Host          []string `json:"host" bson:"host"`
	ExpirationSec int      `json:"expiration" bson:"expiration"`
}

// RedisOption struct
type RedisOption struct {
	Name          string `json:"name" bson:"name"`
	Host          string `json:"host" bson:"host"`
	Password      string `json:"password" bson:"password"`
	ExpirationSec int    `json:"expiration" bson:"expiration"`
}

var (
	// Memcached variable
	Memcached MemcachedConf
	// Redis variable
	Redis RedisConf
)

//Mc variable
var Mc = make(map[string]*persistence.MemcachedStore)

//Rd variable
var Rd = make(map[string]*persistence.RedisStore)

// GetMemcachedResource function
func GetMemcachedResource(resourceName string) MemcachedOption {
	c := config.GetConfig()
	config.GetConf(c.ByteConfig, &Memcached)
	for _, v := range Memcached.CacheOptions {
		if v.Name == resourceName {
			return v
		}
	}

	return MemcachedOption{}
}

// GetRedisResource function
func GetRedisResource(resourceName string) RedisOption {
	c := config.GetConfig()
	config.GetConf(c.ByteConfig, &Redis)
	for _, v := range Redis.CacheOptions {
		if v.Name == resourceName {
			return v
		}
	}

	return RedisOption{}
}

// McConnect function
func McConnect(resourceName string) *persistence.MemcachedStore {
	c := GetMemcachedResource(resourceName)

	if Mc[resourceName] == nil {
		Mc[resourceName] = persistence.NewMemcachedStore(c.Host, time.Second*time.Duration(c.ExpirationSec))
	}

	return Mc[resourceName]
}

// RdConnect function
func RdConnect(resourceName string) *persistence.RedisStore {
	c := GetRedisResource(resourceName)

	if Rd[resourceName] == nil {
		Rd[resourceName] = persistence.NewRedisCache(c.Host, c.Password, time.Second*time.Duration(c.ExpirationSec))
	}

	return Rd[resourceName]
}
