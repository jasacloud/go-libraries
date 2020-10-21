# go-libraries
[![Go Report Card](https://goreportcard.com/badge/github.com/jasacloud/go-libraries)](https://goreportcard.com/report/github.com/jasacloud/go-libraries)  

How to usage


Example for **`/path/to/config.json`** :

```json
{
	"mode":"debug",
	"log":{
		"enable":false,
		"type":"file",
		"output":"log/log.txt"
	},
	"listen":{
		"listenAddr":"*",
		"listenPort": "8800",
		"ssl":false
	},
	"ssl": {
		"autotls":false,
		"domain":["example.com","www.example.com"],
		"listenAddr":"*",
		"listenPort":"443",
		"certFile":"ssl/fullchain.cer",
		"keyFile":"ssl/example.com.key",
		"sslOnly":false
	},
	"cors":{
		"useSetting":true,
		"allowOrigins": ["*"],
		"allowMethods": ["GET","POST","OPTIONS","PUT","DELETE"],
		"allowHeaders": ["DNT","X-Mx-ReqToken","Keep-Alive","User-Agent","X-Requested-With","If-Modified-Since","Cache-Control","Content-Type","origin","content-type","accept","authorization"],
		"allowCredentials": true,
		"allowWildcard": true,
		"exposeHeaders": ["Content-Length"],
		"maxAgeSec": 3600
	},
	"jwtResources":[
		{
			"name" : "default",
			"secret":"secret",
			"algorithm": "HS256",
			"expiration":3600
		},
		{
			"name" : "RS512",
			"algorithm": "RS512",
			"expiration":3600,
			"privateKey":"keys/privatekey.pem",
			"publicKey":"keys/publickey.pem"
		}
	],
	"httpResources": [
		{
			"name":"default",
			"url": "https://www.example.com",
			"uri": "api/user",
			"preHeaders" : [
				{"name":"X-Agent","value":"JCClient"}
			],
			"preParams": [
				{"name":"sender","value":"SENDER1"},
				{"name":"authusr","value":"user123"},
				{"name":"authpwd", "value":"password123"}
			]
		}
	],
	"mongoResources": [
		{
			"name":"default",
			"host": "db.example.com",
			"port": "27017",
			"username": "username",
			"password": "password",
			"db": "dbname",
			"ssl": true
		},
		{
			"name": "secondary",
			"uri": "mongodb+srv://username:password@example-x1nlu.mongodb.net/test?retryWrites=true&w=majority&readPreference=secondaryPreferred"
		}
	]
}
```
  
  
Example for **`main.go`** :
```go

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jasacloud/go-libraries/config"
	"github.com/jasacloud/go-libraries/db"
	"github.com/jasacloud/go-libraries/db/mongoc"
	"github.com/jasacloud/go-libraries/server"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	config.LoadConfig("/path/to/config.json")
}

func main() {

	server.LoadServer()

	resources := server.Route.Group("/api")
	resources.GET("/ping/:name", func(c *gin.Context) {
		name := c.Param("name")
		server.ResponseJSON(c, 200, gin.H{
			"message": "pong",
			"name":    name,
		})
	})

	resources.GET("/db/:resource", func(c *gin.Context) {
		resource := c.Param("resource")
		var data []interface{}
		d := db.DbConnect(resource).CopyAll()
		defer d.Sess.Close()
		err := d.Conn.C("colectionX").Find(nil).All(&data)
		if err != nil {
			server.ResponseJSON(c, 200, gin.H{
				"message": "error",
			})
			return
		}

		server.ResponseJSON(c, 200, gin.H{
			"message": "success",
			"data":    data,
		})
	})

	resources.GET("/dbUriConnect", func(c *gin.Context) {
		d := db.UriConnect("mongodb://username:password@db.example.com:27017/dbname?ssl=true&maxPoolSize=3")
		var data []map[string]interface{}
		err := d.Conn.C(d.Conn.Name).Find(nil).All(&data)
		if err != nil {
			server.ResponseJSON(c, 200, gin.H{
				"error": err,
			})
			return
		}

		server.ResponseJSON(c, 200, gin.H{
			"status": "OK",
			"data":   data,
		})
	})

	resources.GET("/dbUriConnectV2", func(c *gin.Context) {
		conn, err := mongoc.NewConnectionURI("mongodb+srv://username:password@example-x1nlu.mongodb.net/test?retryWrites=true&w=majority&readPreference=secondaryPreferred")
		if err != nil {
			server.ResponseJSON(c, 200, gin.H{
				"error": err,
			})
			return
		}

		conn.C("mycollections")
		collection := conn.Collection
		opt := options.Find()
		opt.SetLimit(int64(10))
		opt.SetSkip(int64(0))
		q := db.Map{"name": "Dwi BudUt"}

		cur, err := collection.Find(c.Request.Context(), q, opt)
		if err != nil {
			server.ResponseJSON(c, 200, gin.H{
				"error": err,
			})
			return
		}

		defer cur.Close(c.Request.Context())

		var data []map[string]interface{}
		err = cur.All(c.Request.Context(), &data)
		if err != nil {
			server.ResponseJSON(c, 200, gin.H{
				"error": err,
			})
			return
		}

		server.ResponseJSON(c, 200, gin.H{
			"status": "OK",
			"data":   data,
		})
	})

	server.Start()
}

```