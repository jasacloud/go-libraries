package db

import (
	"database/sql"
	"errors"
	"github.com/jasacloud/go-libraries/config"
	"log"
	"sync"

	// Register mysql driver
	_ "github.com/go-sql-driver/mysql"
)

// SqlConnections struct
type SqlConnections struct {
	sync.RWMutex
	Database *sql.DB
	Indexed  bool
	Option   *Resource
	URI      string
}

// SqlConnectionBuffers struct
type SqlConnectionBuffers struct {
	sync.RWMutex
	Connections map[string]*SqlConnections
}

// AllSqlConnection variable
var AllSqlConnection = SqlConnectionBuffers{
	Connections: make(map[string]*SqlConnections),
}

// GenerateDSN function
func GenerateDSN(resource *Resource, includeDb bool) string {

	dsn := resource.Username + ":" + resource.Password + "@tcp(" + resource.Host + ":" + resource.Port + ")/"
	if includeDb {
		dsn = dsn + resource.Db
	}

	return dsn
}

// sqlConnectURI function
func sqlConnectURI(uri string) (*SqlConnections, error) {

	if uri == "" {
		return nil, errors.New("invalid connection URI")
	}
	db, err := sql.Open("mysql", uri)
	if err != nil {
		log.Println("cannot reconnect sql resources")
		return nil, err
	}

	return &SqlConnections{
		Database: db,
	}, nil
}

// sqlConnect function
func sqlConnect(resourceName string) (*SqlConnections, error) {
	c := config.GetConfig()
	var dbResources DBResources
	config.GetConf(c.ByteConfig, &dbResources)
	var resource = GetResource(dbResources.Resources, resourceName)
	uri := GenerateDSN(resource, true)

	connection, err := sqlConnectURI(uri)
	if err != nil {

		return nil, err
	}
	connection.Option = resource
	connection.URI = uri
	return connection, err
}

// NewSqlConnectionURI function
func NewSqlConnectionURI(uri string) (*SqlConnections, error) {
	AllSqlConnection.Lock()
	defer AllSqlConnection.Unlock()
	if AllSqlConnection.Connections[uri] != nil {
		err := AllSqlConnection.Connections[uri].CheckConnection()
		if err != nil {
			connection, err := sqlConnectURI(uri)
			if err != nil {

				return nil, err
			}
			AllSqlConnection.Connections[uri] = connection

			return AllSqlConnection.Connections[uri], nil
		}
	} else {
		log.Println("initiate new connection to:", uri)
		connection, err := sqlConnectURI(uri)
		if err != nil {

			return nil, err
		}
		AllSqlConnection.Connections[uri] = connection
	}

	return AllSqlConnection.Connections[uri], nil
}

// NewSqlConnection function
func NewSqlConnection(resourceName string) (*SqlConnections, error) {
	AllSqlConnection.Lock()
	defer AllSqlConnection.Unlock()
	if AllSqlConnection.Connections[resourceName] != nil {
		err := AllSqlConnection.Connections[resourceName].CheckConnection()
		if err != nil {
			connection, err := sqlConnect(resourceName)
			if err != nil {

				return nil, err
			}
			AllSqlConnection.Connections[resourceName] = connection

			return AllSqlConnection.Connections[resourceName], nil
		}
	} else {
		connection, err := sqlConnect(resourceName)
		if err != nil {

			return nil, err
		}
		AllSqlConnection.Connections[resourceName] = connection
	}

	return AllSqlConnection.Connections[resourceName], nil
}

// CheckConnection method
func (c *SqlConnections) CheckConnection() error {
	if c != nil {
		return c.Database.Ping()
	}

	return errors.New("no sql connection")
}
