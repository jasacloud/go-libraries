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

package mongoc

import (
	"context"
	"errors"
	"github.com/jasacloud/go-libraries/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"log"
	"strings"
	"sync"
)

// Resource struct
type Resource struct {
	Name     string `json:"name" bson:"name"`
	Uri      string `json:"uri" bson:"uri"`
	Host     string `json:"host" bson:"host"`
	Port     string `json:"port" bson:"port"`
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
	Db       string `json:"db" bson:"db"`
	Ssl      bool   `json:"ssl" bson:"ssl"`
}

// DBResources struct
type DBResources struct {
	Resources []*Resource `json:"mongoResources" bson:"mongoResources"`
}

// Connections struct
type Connections struct {
	sync.RWMutex
	Client     *mongo.Client
	Database   *mongo.Database
	Collection *mongo.Collection
	Option     *Resource
	URI        string
}

// Index struct
type Index struct {
	IndexModel           mongo.IndexModel
	CreateIndexesOptions *options.CreateIndexesOptions
}

// Indexes struct
type Indexes struct {
	IndexModels []*Index
}

// ConnectionBuffers struct
type ConnectionBuffers struct {
	sync.RWMutex
	Connections map[string]*Connections
}

// AllConnection variable
var AllConnection = ConnectionBuffers{
	Connections: make(map[string]*Connections),
}

// GetResource function
func GetResource(resources []*Resource, resourceName string) *Resource {

	for _, v := range resources {
		if v.Name == resourceName {
			return v
		}
	}
	log.Println("DB resourceName not found in config. resourceName:", resourceName)

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
		sslString = "ssl=true&maxPoolSize=3&SSLInsecure=true&authSource=admin"
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

// connectURI function
func connectURI(uri string) (*Connections, error) {

	if uri == "" {
		return nil, errors.New("invalid connection URI")
	}
	// Set client options
	clientOptions := options.Client()
	clientOptions.ApplyURI(uri)
	//clientOptions.SetConnectTimeout(60 * time.Second)
	//clientOptions.SetServerSelectionTimeout(60 * time.Second)
	err := clientOptions.Validate()
	if err != nil {
		return nil, err
	}
	databaseName := clientOptions.Auth.AuthSource
	cs, err := connstring.Parse(uri)
	if err != nil {
		return nil, err
	}
	if cs.Database != "" {
		databaseName = cs.Database
	}

	//clientOptions.Auth.AuthSource = ""

	// connect to MongoDB
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {

		return nil, err
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)

	if err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB on: " + strings.ReplaceAll(uri, clientOptions.Auth.Password, "xxxxxx"))
	database := client.Database(databaseName)
	collection := database.Collection(databaseName)

	return &Connections{
		Client:     client,
		Database:   database,
		Collection: collection,
	}, nil
}

// connect function
func connect(resourceName string) (*Connections, error) {
	c := config.GetConfig()
	var dbResources DBResources
	config.GetConf(c.ByteConfig, &dbResources)
	var resource = GetResource(dbResources.Resources, resourceName)
	uri := ""
	if resource.Uri != "" {
		uri = resource.Uri
	} else {
		uri = GenerateURI(resource, true)
	}
	connection, err := connectURI(uri)
	if err != nil {

		return nil, err
	}
	connection.Option = resource

	return connection, err
}

// NewConnectionURI function
func NewConnectionURI(uri string) (*Connections, error) {
	connection, err := GetConnectionURI(uri)
	if err != nil {
		return SetNewConnectionURI(uri)
	}

	return connection, nil
}

// GetConnectionURI function
func GetConnectionURI(uri string) (*Connections, error) {
	AllConnection.RLock()
	defer AllConnection.RUnlock()
	if AllConnection.Connections[uri] != nil {
		err := AllConnection.Connections[uri].CheckConnection()
		if err != nil {
			return nil, err
		}
		return AllConnection.Connections[uri], nil
	}

	return nil, errors.New("connection not found")
}

// SetNewConnectionURI function
func SetNewConnectionURI(uri string) (*Connections, error) {
	AllConnection.Lock()
	defer AllConnection.Unlock()
	connection, err := connectURI(uri)
	if err != nil {
		return nil, err
	}
	AllConnection.Connections[uri] = connection

	return AllConnection.Connections[uri], nil
}

// NewConnection function
func NewConnection(resourceName string) (*Connections, error) {
	connection, err := GetConnection(resourceName)
	if err != nil {
		return SetNewConnection(resourceName)
	}

	return connection, nil
}

// GetConnection function
func GetConnection(resourceName string) (*Connections, error) {
	AllConnection.RLock()
	defer AllConnection.RUnlock()
	if AllConnection.Connections[resourceName] != nil {
		err := AllConnection.Connections[resourceName].CheckConnection()
		if err != nil {
			return nil, err
		}
		return AllConnection.Connections[resourceName], nil
	}

	return nil, errors.New("connection not found")
}

// SetNewConnection function
func SetNewConnection(resourceName string) (*Connections, error) {
	AllConnection.Lock()
	defer AllConnection.Unlock()
	connection, err := connect(resourceName)
	if err != nil {
		return nil, err
	}
	AllConnection.Connections[resourceName] = connection

	return AllConnection.Connections[resourceName], nil
}

// DB method
func (c *Connections) DB(name string) {
	c.Lock()
	defer c.Unlock()
	c.Database = c.Client.Database(name)
}

// C method
func (c *Connections) C(name string) {
	c.Lock()
	defer c.Unlock()
	c.Collection = c.Database.Collection(name)
}

// CheckConnection method
func (c *Connections) CheckConnection() error {
	command := bson.D{{Key: "ping", Value: 1}}
	err := c.Client.Database("admin").RunCommand(context.Background(), command).Err()
	if err != nil {
		log.Println("check connection error:", err)

		return err
	}

	return nil
}

// EnsureIndex method
func (c *Connections) EnsureIndex(index mongo.IndexModel) (string, error) {

	return c.Collection.Indexes().CreateOne(context.Background(), index)
}

// EnsureIndexes method
func (c *Connections) EnsureIndexes(indexes ...mongo.IndexModel) ([]string, error) {

	return c.Collection.Indexes().CreateMany(context.Background(), indexes)
}

// CreateIndex method
func (c *Connections) CreateIndex(index *Index) (string, error) {

	return c.Collection.Indexes().CreateOne(context.Background(), index.IndexModel, index.CreateIndexesOptions)
}

// CreateIndexes method
func (c *Connections) CreateIndexes(index ...*Index) ([]string, error) {
	var sliceIndex []mongo.IndexModel
	var sliceOpts []*options.CreateIndexesOptions
	for _, v := range index {
		sliceIndex = append(sliceIndex, v.IndexModel)
		sliceOpts = append(sliceOpts, v.CreateIndexesOptions)
	}

	return c.Collection.Indexes().CreateMany(context.Background(), sliceIndex, sliceOpts...)
}

// NewIndex function
func NewIndex() *Index {

	return &Index{
		IndexModel:           mongo.IndexModel{},
		CreateIndexesOptions: &options.CreateIndexesOptions{},
	}
}

// AddKeys method
func (i *Index) AddKeys(keys bson.D) *Index {
	i.IndexModel.Keys = keys

	return i
}

// SetUnique method
func (i *Index) SetUnique(unique bool) *Index {
	if i.IndexModel.Options != nil {
		i.IndexModel.Options.SetUnique(unique)

		return i
	}
	i.IndexModel.Options = options.Index()
	i.IndexModel.Options.SetUnique(unique)

	return i
}

// SetBackground method
func (i *Index) SetBackground(background bool) *Index {
	if i.IndexModel.Options != nil {
		i.IndexModel.Options.SetBackground(background)

		return i
	}
	i.IndexModel.Options = options.Index()
	i.IndexModel.Options.SetBackground(background)

	return i
}

// SetSparse method
func (i *Index) SetSparse(sparse bool) *Index {
	if i.IndexModel.Options != nil {
		i.IndexModel.Options.SetSparse(sparse)

		return i
	}
	i.IndexModel.Options = options.Index()
	i.IndexModel.Options.SetSparse(sparse)

	return i
}

// SetName method
func (i *Index) SetName(name string) *Index {
	if i.IndexModel.Options != nil {
		i.IndexModel.Options.SetName(name)

		return i
	}
	i.IndexModel.Options = options.Index()
	i.IndexModel.Options.SetName(name)

	return i
}

// example IndexModel :
func yieldIndexModel() mongo.IndexModel {
	indexKeys := bson.D{
		{Key: "name", Value: 1},
		{Key: "type", Value: 1},
		{Key: "parent", Value: 1},
	}
	indexOptions := options.Index()
	indexOptions.SetUnique(true)
	indexOptions.SetName("_name_1_type_1_parent_1_")
	indexOptions.SetBackground(true)
	indexOptions.SetSparse(true)

	index := mongo.IndexModel{}
	index.Keys = indexKeys
	index.Options = indexOptions

	return index
}
