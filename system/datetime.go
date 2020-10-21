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
	"strconv"
	"time"
)

const (
	// UTCDateLayout constant
	UTCDateLayout = "2006-01-02T15:04:05Z07:00"
)

// AddDateUnitName type
type AddDateUnitName string

const (
	// AddYears constant
	AddYears AddDateUnitName = "years"
	// AddMonths constant
	AddMonths AddDateUnitName = "months"
	// AddDays constant
	AddDays AddDateUnitName = "days"
	// AddHours constant
	AddHours AddDateUnitName = "hours"
	// AddMinutes constant
	AddMinutes AddDateUnitName = "minutes"
	// AddSeconds constant
	AddSeconds AddDateUnitName = "seconds"
)

// GetCurDateTimeZ function
func GetCurDateTimeZ() string {

	return time.Now().UTC().Format(UTCDateLayout)
}

// ParseDateTimeZ function
func ParseDateTimeZ(value string) (time.Time, error) {

	return time.Parse(UTCDateLayout, value)
}

// GetCurDateTimeZAdd function
func GetCurDateTimeZAdd(unit AddDateUnitName, value int) string {
	t := time.Now()
	switch unit {
	case AddYears:
		return t.AddDate(value, 0, 0).UTC().Format(UTCDateLayout)
	case AddMonths:
		return t.AddDate(0, value, 0).UTC().Format(UTCDateLayout)
	case AddDays:
		return t.AddDate(0, 0, value).UTC().Format(UTCDateLayout)
	case AddHours:
		return t.Add(time.Duration(value) * time.Hour).UTC().Format(UTCDateLayout)
	case AddMinutes:
		return t.Add(time.Duration(value) * time.Minute).UTC().Format(UTCDateLayout)
	case AddSeconds:
		return t.Add(time.Duration(value) * time.Second).UTC().Format(UTCDateLayout)
	default:
	}
	return time.Now().UTC().Format(UTCDateLayout)
}

// GetTimeStamp function
func GetTimeStamp() int64 {

	return time.Now().UnixNano() / int64(time.Millisecond)
}

// GetTimeStampString function
func GetTimeStampString() string {

	return strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
}
