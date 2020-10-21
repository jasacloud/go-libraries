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

package system

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

// Random function
func Random(min, max int) int {
	rand.Seed(time.Now().Unix())

	return rand.Intn(max-min) + min
}

// GetUID function
func GetUID() string {
	a1 := "ABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890"
	a2 := "ABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890"
	a3 := "ABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890"

	a11 := "ABCDEFGHIJKLMNOPQRSTUVWXYZ09876543210"
	a12 := "ABCDEFGHIJKLMNOPQRSTUVWXYZ09876543210"
	a13 := "ABCDEFGHIJKLMNOPQRSTUVWXYZ09876543210"

	r1 := string(a1[rand.Intn(37)])
	r2 := string(a2[rand.Intn(37)])
	r3 := string(a3[rand.Intn(37)])

	r11 := string(a11[rand.Intn(37)])
	r12 := string(a12[rand.Intn(37)])
	r13 := string(a13[rand.Intn(37)])

	s1 := r3 + r2 + r1
	s2 := r11 + r12 + r13

	t := GetTimeStamp()
	dt := time.Now().Format("060102150405")
	hex := fmt.Sprintf("%x", t)

	return dt + strings.ToUpper(hex)[11-6:] + s1 + s2
}

// GetAllIP function
func GetAllIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Read IP Error:", err)
	}
	ip := ""
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				//fmt.Println(ipnet.IP.String())
				ip = ",\"" + ipnet.IP.String() + "\"" + ip
			}
		}
	}

	return "[" + strings.Trim(ip, ",") + "]"
}
