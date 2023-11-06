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

package broker

import (
	"fmt"
	"github.com/jasacloud/go-libraries/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

// AmqpConf struct
type AmqpConf struct {
	AmqpOption []AmqpOption `json:"rabitmqResources" bson:"rabitmqResources"`
}

// AmqpOption struct
type AmqpOption struct {
	Name       string `json:"name" bson:"name"`
	ConnString string `json:"connString" bson:"connString"`
}

// AmqpConsumerConf struct
type AmqpConsumerConf struct {
	AmqpConsumerOption []AmqpConsumerOption `json:"amqpConsumerResources" bson:"amqpConsumerResources"`
}

/*
var (
	amqpResource = flag.String("uri", "amqp://username:password@example.com:5672", "The rabbitmq resource or URI endpoint")
	httpResource = flag.String("form_url", "auth", "The Resource or URL that requests are sent to")
	threads      = flag.Int("threads", 1, "The max amount of go routines that you would like the process to use")
	maxprocs     = flag.Int("max_procs", 1, "The max amount of processors that your application should use")
	exchange     = flag.String("exchange", "something", "The exchange we will be binding to")
	exchangeType = flag.String("exchange_type", "topic", "Type of exchange we are binding to | topic | direct| etc..")
	queue        = flag.String("queue", "some.queue", "Name of the queue that you would like to connect to")
	routingKey   = flag.String("routing_key", "some.queue", "queue to route messages to")
	workerName   = flag.String("worker_name", "worker.name", "name to identify worker by")
	verbosity    = flag.Bool("verbos", true, "Set true if you would like to log EVERYTHING")
	reConnect    = flag.Int("autoReconnectSec", 10, "Set number on second of reconnect interval when connection closed")
)
*/

// AmqpConsumerOption struct
type AmqpConsumerOption struct {
	Name         string `json:"name" bson:"name"`
	ConnString   string `json:"connString" bson:"connString"`
	HttpResource string `json:"httpResource" bson:"httpResource"`
	Threads      int    `json:"threads" bson:"threads"`
	Exchange     string `json:"exchange" bson:"exchange"`
	ExchangeType string `json:"exchangeType" bson:"exchangeType"`
	Queue        string `json:"queue" bson:"queue"`
	RoutingKey   string `json:"routingKey" bson:"routingKey"`
	WorkerName   string `json:"workerName" bson:"workerName"`
	Verbosity    bool   `json:"verbosity" bson:"verbosity"`
	ReConnect    int    `json:"reConnectSecInterval" bson:"reConnectSecInterval"`
}

var (
	// Amqp type
	Amqp AmqpConf
	// AmqpConsumer type
	AmqpConsumer AmqpConsumerConf
)

// AmqpConn variable
var AmqpConn = make(map[string]*amqp.Connection)

//var AmqpCloseErr = make(map[string] chan *amqp.Error)
//var AmqpErr = make(map[string] chan *amqp.Error)

// GetAmqpResource function
func GetAmqpResource(resourceName string) AmqpOption {
	c := config.GetConfig()
	config.GetConf(c.ByteConfig, &Amqp)
	for _, v := range Amqp.AmqpOption {
		if v.Name == resourceName {

			return v
		}
	}

	return AmqpOption{}
}

// GetConsumerResource function
func GetConsumerResource(resourceName string) AmqpConsumerOption {
	c := config.GetConfig()
	config.GetConf(c.ByteConfig, &AmqpConsumer)
	for _, v := range AmqpConsumer.AmqpConsumerOption {
		if v.Name == resourceName {

			return v
		}
	}

	return AmqpConsumerOption{}
}

// AmqpConnect function
func AmqpConnect(resourceName string) *amqp.Connection {
	c := GetAmqpResource(resourceName)

	if AmqpConn[resourceName] != nil {
		ch, err := AmqpConn[resourceName].Channel()
		if err != nil {
			fmt.Println("Reconnect... on error :", err)
			AmqpConn[resourceName] = nil
			//AmqpErr[resourceName] = make(chan *amqp.Error)
			return AmqpConnect(resourceName)
		}
		ch.Close()

	} else {
		conn, err := amqp.Dial(c.ConnString)
		if err != nil {
			fmt.Println("Failed to connect to RabbitMQ:", err)
		}
		AmqpConn[resourceName] = conn
		//AmqpErr[resourceName] = make(chan *amqp.Error)
	}

	return AmqpConn[resourceName]
}
