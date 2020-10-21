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

package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogProperties struct
type LogProperties struct {
	Enable bool   `json:"enable" bson:"enable"`
	Output string `json:"output" bson:"output"`
	Type   string `json:"type" bson:"type"`
}

// LogConfig struct
type LogConfig struct {
	Properties LogProperties `json:"log" bson:"log"`
}

// logWriter struct
type logWriter struct {
	sync.Mutex
	filename    string
	newFilename string
	date        string
	ext         string
	name        string
	day         int
	Output      *os.File
}

// Write method
func (w *logWriter) Write(bytes []byte) (int, error) {
	s := time.Now().Format("2006-01-02T15:04:05.000Z") + " [DEBUG] " + string(bytes)
	w.Lock()
	defer w.Unlock()
	err := w.ReopenIfNeeded()
	if err != nil {
		fmt.Println("Failed while rotate log...")

		return 0, err
	}
	return w.Output.WriteString(s)
}

// ReopenIfNeeded method
func (w *logWriter) ReopenIfNeeded() (err error) {
	t := time.Now()
	if t.YearDay() == w.day {
		return nil
	}
	w.Lock()
	defer w.Unlock()
	err = w.Output.Close()
	if err != nil {
		return err
	}
	return w.Reopen()
}

// Reopen method
func (w *logWriter) Reopen() error {

	t := time.Now()
	w.newFilename = w.name + "_" + t.Format("20060102") + w.ext
	w.day = t.YearDay()
	w.date = t.Format("2006-01-02")

	if _, err := os.Stat(filepath.Dir(w.newFilename)); os.IsNotExist(err) {
		err := os.Mkdir(filepath.Dir(w.newFilename), 0666)
		if err != nil {
			log.Printf("Error create dir log: %v", err.Error())
		}
		log.Println("Success create log dir")
	}
	f, err := os.OpenFile(w.newFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	//defer f.Close()
	if err != nil {
		log.Printf("Error opening file: %v", err.Error())
	}

	w.Output = f
	log.New(w.Output, "", log.Ldate|log.Ltime)
	log.SetOutput(w.Output)
	log.Println("--------  NEW LOG ROTATED  --------")

	return nil
}

// NewLogFile function
func NewLogFile(output string) (*logWriter, error) {
	t := time.Now()
	ext := filepath.Ext(output)
	name := output[0 : len(output)-len(ext)]

	lw := &logWriter{
		ext:         ext,
		name:        name,
		day:         t.YearDay(),
		date:        t.Format("2006-01-02"),
		filename:    output,
		newFilename: name + "_" + t.Format("20060102") + ext,
	}
	if _, err := os.Stat(filepath.Dir(lw.newFilename)); os.IsNotExist(err) {
		err := os.Mkdir(filepath.Dir(lw.newFilename), 0666)
		if err != nil {
			log.Printf("Error create dir log: %v", err.Error())
		}
		log.Println("Success create log dir")
	}
	f, err := os.OpenFile(lw.newFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	//defer f.Close()
	if err != nil {
		log.Printf("Error opening file: %v", err.Error())
	}
	lw.Output = f

	err = lw.ReopenIfNeeded()
	if err != nil {

		return nil, err
	}

	return lw, nil
}
