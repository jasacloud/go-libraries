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

package client

// HttpClientPoll struct
type HttpClientPoll struct {
	State chan int
	Func  func()
}

// PollResources variable
var PollResources = make(map[string]*HttpClientPoll)

// StartPolling function
func StartPolling(resource string, f func()) *HttpClientPoll {
	if PollResources[resource] != nil {
		x := PollResources[resource]
		return x
	}
	return PollResources[resource]
}
