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
	"math"
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

// GetTimeStamp function return milliseconds int64
func GetTimeStamp(ts ...time.Time) int64 {
	t := time.Now()
	if len(ts) > 0 {
		t = ts[0]
	}
	return t.UnixNano() / int64(time.Millisecond)
}

// GetTimeStampString function return milliseconds string
func GetTimeStampString(ts ...time.Time) string {
	t := time.Now()
	if len(ts) > 0 {
		t = ts[0]
	}
	return strconv.FormatInt(t.UnixNano()/int64(time.Millisecond), 10)
}

// GetTimeFromUnixString function parameter is milliseconds string
func GetTimeFromUnixString(unixTime string) (time.Time, error) {
	i, err := strconv.ParseInt(unixTime, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	a := float64(i) / 1000
	b := float64(i / 1000)
	nsec := ToFixed(a-b, 3) * 1000000000
	return time.Unix(i/1000, int64(nsec)), nil
}

// GetTimeFromUnix function parameter is milliseconds int64
func GetTimeFromUnix(unixTime int64) time.Time {
	a := float64(unixTime) / 1000
	b := float64(unixTime / 1000)
	nsec := ToFixed(a-b, 3) * 1000000000
	return time.Unix(unixTime/1000, int64(nsec))
}

// Round function
func Round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

// ToFixed function
func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(Round(num*output)) / output
}
