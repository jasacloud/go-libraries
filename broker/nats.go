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
	"github.com/nats-io/nats.go"
	"log"
	"time"
)

// NatsConf is config name of nats resources
type NatsConf struct {
	NatsOption []NatsOption `json:"natsResources" bson:"natsResources"`
}

// NatsOption struct
type NatsOption struct {
	Name              string `json:"name" bson:"name"`
	ConnString        string `json:"connString" bson:"connString"`
	MaxReconnectSec   int    `json:"maxReconnectSec" bson:"maxReconnectSec"`
	MaxReconnect      *int   `json:"maxReconnect,omitempty" bson:"maxReconnect,omitempty"`
	ReconnectDelaySec int    `json:"reconnectDelaySec" bson:"reconnectDelaySec"`
}

var (
	// Nats variable
	Nats NatsConf
)

// NatsConn variable
var NatsConn = make(map[string]*nats.Conn)

// GetNatsResource function
func GetNatsResource(resourceName string) NatsOption {
	c := config.GetConfig()
	config.GetConf(c.ByteConfig, &Nats)
	for _, v := range Nats.NatsOption {
		if v.Name == resourceName {

			return v
		}
	}
	return NatsOption{}
}

// NatsConnect function
func NatsConnect(resourceName string) *nats.Conn {

	if NatsConn[resourceName] == nil {
		c := GetNatsResource(resourceName)
		nc, err := nats.Connect(c.ConnString, setupConnOptions(c)...)
		if err != nil {
			fmt.Println("Failed to connect to NATS:", err)
		}
		NatsConn[resourceName] = nc
	}

	return NatsConn[resourceName]
}

// setupConnOptions function
func setupConnOptions(opt NatsOption) []nats.Option {
	opts := []nats.Option{
		nats.Name("NATS-Client"),
		nats.Timeout(5 * time.Second), // Good practice: prevent hanging on initial startup
	}

	// Default delays if not provided or set to 0
	reconnectDelay := time.Duration(opt.ReconnectDelaySec) * time.Second
	if reconnectDelay <= 0 {
		reconnectDelay = 1 * time.Second // Fallback to 1s delay
	}
	opts = append(opts, nats.ReconnectWait(reconnectDelay))

	// Handle Max Reconnects safely
	if opt.MaxReconnect != nil {
		opts = append(opts, nats.MaxReconnects(*opt.MaxReconnect))
	} else if opt.MaxReconnectSec > 0 {
		totalWait := time.Duration(opt.MaxReconnectSec) * time.Second
		maxAttempts := int(totalWait / reconnectDelay)
		if maxAttempts < 1 {
			maxAttempts = 1
		}
		opts = append(opts, nats.MaxReconnects(maxAttempts))
	}

	// Logging Handlers
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		if err != nil {
			log.Printf("Disconnected: with error (%v)", err)
			return
		}
		log.Println("Disconnected from NATS cluster")
	}))

	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("Reconnected to [%s]", nc.ConnectedUrl())
	}))

	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Fatal("Connection closed permanently, exiting...")
	}))

	return opts
}
