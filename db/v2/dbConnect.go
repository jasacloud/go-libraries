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

package db

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/jasacloud/go-libraries/config"
	"github.com/jasacloud/go-libraries/utils/masker"
	"github.com/juju/mgo"
	"log"
	"net"
	"strings"
	"sync"
)

// Resource struct
type Resource struct {
	Name     string `json:"name" bson:"name"`
	Engine   string `json:"engine" bson:"engine"`
	Host     string `json:"host" bson:"host"`
	Port     string `json:"port" bson:"port"`
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
	Db       string `json:"db" bson:"db"`
	Ssl      bool   `json:"ssl" bson:"ssl"`
}

// DBResources struct
type DBResources struct {
	Resources []*Resource `json:"dbResources" bson:"dbResources"`
}

// Connections struct
type Connections struct {
	sync.RWMutex
	Session    *mgo.Session
	Database   *mgo.Database
	Collection *mgo.Collection
	Indexed    bool
	Option     *Resource
	URI        string
}

// ConnectionBuffers struct
type ConnectionBuffers struct {
	sync.RWMutex
	Connections map[string]*Connections
}

// Map type map
type Map map[string]interface{}

// AllConnection variable
var AllConnection = ConnectionBuffers{
	Connections: make(map[string]*Connections),
}

var (
	// ErrNotFound variable
	ErrNotFound = mgo.ErrNotFound
	// ErrCursor variable
	ErrCursor = mgo.ErrCursor
)

// GetResource function
func GetResource(resources []*Resource, resourceName string) *Resource {

	for _, v := range resources {
		if v.Name == resourceName {

			return v
		}
	}
	log.Println("DB resourceName not found in config. resourceName: ", resourceName)

	return nil
}

// GenerateURI function
func GenerateURI(option *Resource, db bool) string {
	uri := ""
	if option == nil {

		return uri
	}
	sslString := ""
	if option.Ssl == true {
		sslString = "ssl=true&maxPoolSize=3&authSource=admin"
	} else {
		sslString = ""
	}
	dbString := ""
	if db == true {
		dbString = option.Db
	} else {
		dbString = ""
	}
	if strings.Index(option.Host, "mongodb://") != -1 || strings.Index(option.Host, "mongodb+srv://") != -1 {
		uri += option.Host
		if strings.Index(option.Host, "@") == -1 {
			if option.Username != "" {
				if option.Password != "" {
					uri = strings.Replace(uri, "://", "://"+option.Username+":"+option.Password+"@", 1)
				} else {
					uri = strings.Replace(uri, "://", "://"+option.Username+"@", 1)
				}
			}
		}
	} else {
		uri += "mongodb://" + option.Username + ":" + option.Password + "@" + option.Host
	}

	if option.Port != "" {
		uri += ":" + option.Port
	}

	uri += "/" + dbString + "?" + sslString

	return uri
}

// MgoParseURI function
func MgoParseURI(url string, rootCAs *x509.CertPool) (*mgo.DialInfo, error) {

	isSSL := strings.Index(url, "ssl=true") > -1
	url = strings.Replace(url, "ssl=true", "", 1)

	dialInfo, err := mgo.ParseURL(url)

	if err != nil {

		return nil, err
	}

	if isSSL {
		tlsConfig := &tls.Config{}
		if rootCAs != nil {
			tlsConfig.RootCAs = rootCAs
		} else {
			tlsConfig.InsecureSkipVerify = true
		}

		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)

			return conn, err
		}
	}

	dialInfo.Mechanism = "SCRAM-SHA-1"

	return dialInfo, err
}

func connectURI(uri string) (*Connections, error) {
	if uri == "" {

		return nil, errors.New("invalid connection URI")
	}
	dialInfo, err := MgoParseURI(uri, nil)
	if err != nil {
		log.Println("error while ParseURI: ", err)

		return nil, err
	}
	//mgo.SetDebug(true)
	session, err := mgo.DialWithInfo(dialInfo)

	if err != nil {

		return nil, err
	}

	log.Println("Connected to MongoDB on: " + masker.UriPassword(uri))
	session.SetMode(mgo.Primary, true)
	database := session.DB(dialInfo.Database)
	collection := database.C(dialInfo.Database)
	resource := &Resource{
		Name:     uri,
		Engine:   "mongodb",
		Host:     strings.Join(dialInfo.Addrs, ","),
		Port:     "",
		Username: dialInfo.Username,
		Password: dialInfo.Password,
		Db:       dialInfo.Database,
		Ssl:      strings.Index(uri, "ssl=true") > -1,
	}

	return &Connections{
		Session:    session,
		Database:   database,
		Collection: collection,
		URI:        uri,
		Option:     resource,
	}, nil
}

func connect(resourceName string) (*Connections, error) {
	c := config.GetConfig()
	var dbResources DBResources
	config.GetConf(c.ByteConfig, &dbResources)
	var resource = GetResource(dbResources.Resources, resourceName)
	uri := GenerateURI(resource, true)

	connection, err := connectURI(uri)
	if err != nil {

		return nil, err
	}
	connection.Option = resource

	return connection, err
}

// NewConnectionURI function
func NewConnectionURI(uri string) (*Connections, error) {
	AllConnection.Lock()
	defer AllConnection.Unlock()
	if AllConnection.Connections[uri] != nil {
		err := AllConnection.Connections[uri].Session.Ping()
		if err != nil {
			AllConnection.Connections[uri].Session.Refresh()

			return AllConnection.Connections[uri], nil
		}
	} else {
		log.Println("initiate new connection to:", uri)
		connection, err := connectURI(uri)
		if err != nil {

			return nil, err
		}
		AllConnection.Connections[uri] = connection
	}

	return AllConnection.Connections[uri], nil
}

// NewConnection function
func NewConnection(resourceName string) (*Connections, error) {
	AllConnection.Lock()
	defer AllConnection.Unlock()
	if AllConnection.Connections[resourceName] != nil {
		err := AllConnection.Connections[resourceName].Session.Ping()
		if err != nil {
			AllConnection.Connections[resourceName].Session.Refresh()

			return AllConnection.Connections[resourceName], nil
		}
	} else {
		connection, err := connect(resourceName)
		if err != nil {

			return nil, err
		}
		AllConnection.Connections[resourceName] = connection
	}

	return AllConnection.Connections[resourceName], nil
}

// DB method
func (c *Connections) DB(name string) {
	if c != nil {
		c.Lock()
		c.Database = c.Session.DB(name)
		c.Database.Name = name
		c.Collection = c.Database.C(c.Collection.Name)
		c.Option.Db = name
		c.Unlock()
	}
}

// C method
func (c *Connections) C(name string) {
	if c != nil {
		c.Lock()
		defer c.Unlock()
		c.Collection = c.Database.C(name)
		c.Collection.Name = name
	}
}

// Copy method
func (c *Connections) Copy() *mgo.Session {
	if c != nil {
		if c.Session != nil {

			return c.Session.Copy()
		}
	}

	return nil
}

// CopyAll method
func (c *Connections) CopyAll() (*Connections, error) {
	if c != nil {
		c.Lock()
		defer c.Unlock()
		if c.Session != nil {
			session := c.Session.Copy()
			database := session.DB(c.Option.Db)
			option := *c.Option

			return &Connections{
				Session:    session,
				Database:   database,
				Collection: database.C(c.Collection.Name),
				Option:     &option,
				Indexed:    c.Indexed,
				URI:        c.URI,
			}, nil
		}
	}

	return nil, errors.New("nil connections")
}

// Refresh method
func (c *Connections) Refresh() {

	c.Session.Refresh()
}

// CloseAll method
func (c *Connections) CloseAll() {
	c.Lock()
	defer c.Unlock()
	AllConnection.Connections[c.Option.Name] = nil
	c.Session.Close()
}

// Close method
func (c *Connections) Close() {
	c.Lock()
	defer c.Unlock()
	c.Session.Close()
}

// Close function
func Close(resourceName string) {
	AllConnection.Lock()
	defer AllConnection.Unlock()
	if AllConnection.Connections[resourceName] != nil {
		AllConnection.Connections[resourceName].Session.Close()
	}
	if AllConnection.Connections[resourceName].Database != nil {
		AllConnection.Connections[resourceName].Database = nil
	}
	if AllConnection.Connections[resourceName].Collection != nil {
		AllConnection.Connections[resourceName].Collection = nil
	}
	AllConnection.Connections[resourceName] = nil
}
