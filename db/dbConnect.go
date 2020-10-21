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
	"database/sql"
	"fmt"
	"github.com/jasacloud/go-libraries/config"
	"gopkg.in/mgo.v2"
	"log"
	"net"
	"strings"
	"sync"
)

// DbConfig struct
type DbConfig struct {
	Name     string `json:"name" bson:"name"`
	Engine   string `json:"engine" bson:"engine"`
	Host     string `json:"host" bson:"host"`
	Port     string `json:"port" bson:"port"`
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
	Db       string `json:"db" bson:"db"`
	Ssl      bool   `json:"ssl" bson:"ssl"`
}

// DBServer struct
type DBServer struct {
	DbServer []DbConfig `json:"dbResources" bson:"dbResources"`
}

// Properties variable
var Properties *Mongo

// SqlProperties variable
var SqlProperties *Sql

// Resources variable
var Resources = make(map[string]*Mongo)

// SqlResources variable
var SqlResources = make(map[string]*Sql)

// Mongo struct
type Mongo struct {
	m          sync.RWMutex
	DbResource DbConfig
	Sess       *mgo.Session
	Conn       *mgo.Database
	Uri        string
	indexed    bool
}

// Sql struct
type Sql struct {
	DbResource DbConfig
	Sess       *sql.DB
	Conn       *sql.DB
}

// Map type
type Map map[string]interface{}

var (
	// ErrNotFound variable
	ErrNotFound = mgo.ErrNotFound
	// ErrCursor variable
	ErrCursor = mgo.ErrCursor
)

// getDBConf function
func getDbConf(dbServer []DbConfig, resourceName string) DbConfig {

	for _, v := range dbServer {
		if v.Name == resourceName {
			return v
		}
	}
	fmt.Println("DB resourceName not found in config. resourceName: ", resourceName)

	return DbConfig{}
}

// getDBUri function
func getDBUri(dbConfig DbConfig, db bool) string {

	uri := ""
	sslString := ""
	if dbConfig.Ssl == true {

		sslString = "ssl=true&maxPoolSize=3"
	} else {
		sslString = ""
	}
	dbString := ""
	if db == true {
		dbString = dbConfig.Db
	} else {
		dbString = ""
	}
	if strings.Index(dbConfig.Host, "mongodb://") != -1 || strings.Index(dbConfig.Host, "mongodb+srv://") != -1 {
		uri += dbConfig.Host
		if strings.Index(dbConfig.Host, "@") == -1 {
			if dbConfig.Username != "" {
				if dbConfig.Password != "" {
					uri = strings.Replace(uri, "://", "://"+dbConfig.Username+":"+dbConfig.Password+"@", 1)
				} else {
					uri = strings.Replace(uri, "://", "://"+dbConfig.Username+"@", 1)
				}
			}
		}
	} else {
		uri += "mongodb://" + dbConfig.Username + ":" + dbConfig.Password + "@" + dbConfig.Host
	}

	if dbConfig.Port != "" {
		uri += ":" + dbConfig.Port
	}
	if dbString != "" {
		uri += "/" + dbString
	}
	if sslString != "" {
		if strings.Index(uri, "ssl=") == -1 {
			if strings.Index(uri, "?") != -1 {
				uri += "&" + sslString
			} else {
				uri += "?" + sslString
			}
		}
	}

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

	return dialInfo, err
}

// UriConnect function
func UriConnect(uri string) *Mongo {

	if Resources[uri] != nil {
		err := Resources[uri].Sess.Ping()
		if err != nil {
			Resources[uri].Sess.Refresh()

			return Resources[uri]
		}
	} else {
		dialInfo, err := MgoParseURI(uri, nil)
		if err != nil {
			fmt.Println("MgoParseURI Error: ", err)
			return nil
		}
		DbResource := DbConfig{
			Name:     uri,
			Host:     dialInfo.Addrs[0],
			Username: dialInfo.Username,
			Password: dialInfo.Password,
			Db:       dialInfo.Database,
		}
		dialInfo.Database = ""
		//session, err := mgo.Dial("jasacloud.com")
		Sess, err := mgo.DialWithInfo(dialInfo)
		if err != nil {
			fmt.Println("DBConnect Error: ", err)
			return nil
		}
		fmt.Printf("DB Conneted to uri %s \n", uri)
		//defer dbConn.sess.Close()
		if len(dialInfo.Addrs) > 1 {
			Sess.SetMode(mgo.SecondaryPreferred, true)
		} else {
			Sess.SetMode(mgo.Primary, true)
		}
		db := Sess.DB(DbResource.Db)
		Conn := db
		Properties = &Mongo{
			DbResource: DbResource,
			Sess:       Sess,
			Conn:       Conn,
			Uri:        uri,
		}
		Resources[uri] = Properties
	}

	return Resources[uri]
}

// DbConnect function
func DbConnect(resourceName string) *Mongo {

	if Resources[resourceName] != nil {
		err := Resources[resourceName].Sess.Ping()
		if err != nil {
			Resources[resourceName].Sess.Refresh()

			return Resources[resourceName]
		}
	} else {
		c := config.GetConfig()
		dbServer := DBServer{}
		config.GetConf(c.ByteConfig, &dbServer)
		var DbResource = getDbConf(dbServer.DbServer, resourceName)
		uri := getDBUri(DbResource, false)
		dialInfo, err := MgoParseURI(uri, nil)
		if err != nil {
			fmt.Println("MgoParseURI Error: ", err)
			return nil
		}
		//session, err := mgo.Dial("jasacloud.com")
		//mgo.SetDebug(true)
		Sess, err := mgo.DialWithInfo(dialInfo)

		if err != nil {
			fmt.Println("DBConnect Error: ", err)
			return nil
		}
		fmt.Printf("DB Conneted to %s on Port %s \n", DbResource.Host, DbResource.Port)
		//defer dbConn.sess.Close()
		if len(dialInfo.Addrs) > 1 {
			Sess.SetMode(mgo.SecondaryPreferred, true)
		} else {
			Sess.SetMode(mgo.Primary, true)
		}

		db := Sess.DB(DbResource.Db)
		Conn := db
		Properties = &Mongo{
			DbResource: DbResource,
			Sess:       Sess,
			Conn:       Conn,
		}
		Resources[resourceName] = Properties
	}

	return Resources[resourceName]
}

// DB method
func (d *Mongo) DB(name string) *Mongo {
	if d != nil {
		d.m.Lock()
		d.Conn = d.Sess.DB(name)
		d.DbResource.Name = name
		d.m.Unlock()
		return d
	}

	return nil
}

// C method
func (d *Mongo) C(name string, index ...mgo.Index) *mgo.Collection {
	if d != nil {
		if !d.indexed && len(index) > 0 {
			for _, v := range index {
				err := d.Conn.C(name).EnsureIndex(v)
				if err != nil {
					log.Println("db::EnsureIndex error: ", err)
				}
			}
			d.m.Lock()
			d.indexed = true
			d.m.Unlock()
		}

		return d.Conn.C(name)
	}

	return nil
}

// Copy method
func (d *Mongo) Copy() *mgo.Session {
	if d == nil {
		log.Println("d is nil")
		return nil
	}
	if d.Sess != nil {
		return d.Sess.Copy()
	}
	return nil
}

// CopyAll method
func (d *Mongo) CopyAll() *Mongo {
	if d == nil {
		log.Println("d is nil")
		return nil
	}
	d.m.Lock()
	defer d.m.Unlock()
	if d.Sess != nil {
		s := d.Sess.Copy()
		return &Mongo{
			DbResource: d.DbResource,
			Sess:       s,
			Conn:       s.DB(d.DbResource.Db),
			Uri:        d.Uri,
		}
	}
	return nil
}

// Refresh method
func (d *Mongo) Refresh() {
	d.Sess.Refresh()
}

// CloseAll method
func (d *Mongo) CloseAll() {
	d.m.Lock()
	defer d.m.Unlock()
	Resources[d.DbResource.Name] = nil
	d.Sess.Close()
}

// Close method
func (d *Mongo) Close() {
	d.m.Lock()
	defer d.m.Unlock()
	d.Sess.Close()
}

// Close function
func Close(resourceName string) {
	if Resources[resourceName] != nil {
		Resources[resourceName].Sess.Close()
	}
	if Resources[resourceName].Conn != nil {
		Resources[resourceName].Conn = nil
	}
	Resources[resourceName] = nil
}

// GetDB function
func GetDB() *Mongo {

	return Properties
}

// SqlConnect function
func SqlConnect(resourceName string) *Sql {

	if SqlResources[resourceName] != nil {
		err := SqlResources[resourceName].Sess.Ping()
		if err != nil {
			log.Println("Error while ping sql:", err)
			log.Println("Try to reconnect sql...")
			dsn := getDSN(SqlResources[resourceName].DbResource, true)
			db, err := sql.Open("mysql", dsn)
			if err != nil {
				log.Println("cannot reconnect sql resources")
				return nil
			}
			SqlProperties = &Sql{
				DbResource: SqlResources[resourceName].DbResource,
				Sess:       db,
				Conn:       db,
			}
			SqlResources[resourceName] = SqlProperties
			return SqlResources[resourceName]
		}
	} else {
		c := config.GetConfig()
		dbServer := DBServer{}
		config.GetConf(c.ByteConfig, &dbServer)
		var DbResource = getDbConf(dbServer.DbServer, resourceName)
		dsn := getDSN(DbResource, true)
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			fmt.Println("DBConnect Error: ", err)
		}
		fmt.Printf("DB Conneted to %s on Port %s \n", DbResource.Host, DbResource.Port)
		SqlProperties = &Sql{
			DbResource,
			db,
			db,
		}
		SqlResources[resourceName] = SqlProperties
	}

	return SqlResources[resourceName]
}

// getDSN function
func getDSN(dbConfig DbConfig, db bool) string {

	dsn := dbConfig.Username + ":" + dbConfig.Password + "@tcp(" + dbConfig.Host + ":" + dbConfig.Port + ")/"
	if db {
		dsn = dsn + dbConfig.Db
	}

	return dsn
}
