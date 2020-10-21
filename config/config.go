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

package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jasacloud/go-libraries/logger"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	// DefaultConfig constant
	DefaultConfig string = "config"
)

// Properties variable
var Properties *Config

// AppVersion variable
var AppVersion string

// Ver variable
var Ver string

// Config struct
type Config struct {
	ConfigPath string
	ByteConfig []byte
}

// LoadConfig function
func LoadConfig(configPath string) {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	fmt.Println("PROCESS DIR :", dir)

	configArg := flag.String("config", "", "config file")
	logArg := flag.String("log", "", "log output file")
	flag.Parse()

	// handling config file :
	if *configArg != "" {
		configPath = *configArg
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			log.Panic("failed read file config in "+configPath+", error: ", err)
		}
	} else if configPath != "" {
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			processPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
			if err != nil {
				log.Panic("failed read file config in "+configPath+", error: ", err)
			}
			configPath = path.Join(processPath, configPath)
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				log.Println("failed read file config in "+configPath+", error: ", err)
				log.Println("Try to load default config...")
				configPath = DefaultConfig
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					processPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
					if err != nil {
						log.Panic("failed read default file config in "+configPath+", error: ", err)
					}
					configPath = path.Join(processPath, configPath)
					if _, err := os.Stat(configPath); os.IsNotExist(err) {
						log.Panic("failed read default file config in "+configPath+", error: ", err)
					}
				}
			}
		}
	} else {
		log.Panic("failed read file config, error: config path not defined")
	}
	fmt.Println("CONFIG : ", configPath)

	Properties = &Config{
		configPath,
		GetByteConf(configPath),
	}
	var Log logger.LogConfig
	GetConf(Properties.ByteConfig, &Log)

	if runtime.GOOS == "windows" {
		Log.Properties.Output = path.Join(GetConfigDir(), Log.Properties.Output)
	}

	//check if log from cmd params :
	if *logArg != "" {
		Log.Properties.Enable = true
		Log.Properties.Output = *logArg
		Log.Properties.Type = "file"
	}

	// Open a system file to start logging to
	if Log.Properties.Enable == true {
		// Start logger function
		StartLogger(Log)
	}
}

// StartLogger function
func StartLogger(Log logger.LogConfig) {
	fmt.Println("LOG : ", Log.Properties.Output)

	lw, err := logger.NewLogFile(Log.Properties.Output)
	if err != nil {
		log.Printf("Error opening file: %v", err.Error())
	}

	log.New(lw.Output, "", log.Ldate|log.Ltime)
	log.SetOutput(lw.Output)
	log.Println("--------  APPLICATION STARTED  --------")
	pidFile := "PID"
	if runtime.GOOS == "windows" {
		pidFile = path.Join(GetConfigDir(), pidFile)
	}
	err = WritePidFile(pidFile)
	if err != nil {
		log.Println("error while write PID file: ", err)
	}
	stopped := make(chan bool, 1)
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				if lw.Output != nil {
					lw.ReopenIfNeeded()
				}
			case <-stopped:
				return
			}
		}
	}()
}

// getMajorVersion function
func getMajorVersion(major interface{}) string {
	version := ""
	switch v := major.(type) {
	case int:
		version = version + strconv.Itoa(v)
	case string:
		if v == "" {
			v = "0"
		}
		version = version + v
	default:
		version = version + "0"
	}

	return version
}

// getMinorVersion function
func getMinorVersion(minor interface{}) string {
	version := ""
	switch v := minor.(type) {
	case int:
		version = version + "." + strconv.Itoa(v)
	case string:
		if v == "" {
			v = ".0"
		}
		version = version + "." + v
	default:
		version = version + ".0"
	}

	return version
}

func getBuildVersion(build interface{}) string {
	version := ""
	switch v := build.(type) {
	case int:
		version = version + "." + strconv.Itoa(v)
	case string:
		if v == "" {
			v = ".0"
		}
		version = version + "." + v
	default:
		version = version + ".0"
	}

	return version
}

// getRevisionVersion function
func getRevisionVersion(revision interface{}) string {
	version := ""
	switch v := revision.(type) {
	case int:
		version = version + "." + strconv.Itoa(v)
	case string:
		if v == "" {
			v = ".0"
		}
		version = version + "." + v
	default:
		version = version + ".0"
	}

	return version
}

// Version function
func Version(major, minor, build, revision interface{}) {
	version := ""
	version = version + getMajorVersion(major)
	version = version + getMinorVersion(minor)
	version = version + getBuildVersion(build)
	version = version + getRevisionVersion(revision)

	log.Println("--------  AppVersion: " + version + "  --------")
	AppVersion = version
	v := strings.Split(version, ".")
	Ver = v[0] + "." + v[1]
}

// GetConf function
func GetConf(byteConfig []byte, nodeConfig interface{}) {
	if err := json.Unmarshal(byteConfig, nodeConfig); err != nil {
		log.Fatal(err)
	}
}

// GetByteConf function
func GetByteConf(configPath string) []byte {
	jsonFile, err := os.Open(configPath)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		fmt.Println("Error load config : ", err)
	}

	return byteValue
}

// GetConfig function
func GetConfig() *Config {

	return Properties
}

// WritePidFile Write a pid file, but first make sure it doesn't exist with a running pid.
func WritePidFile(pidFile string) error {
	// Read in the pid file as a slice of bytes.
	if piddata, err := ioutil.ReadFile(pidFile); err == nil {
		// Convert the file contents to an integer.
		if pid, err := strconv.Atoi(string(piddata)); err == nil {
			// Look for the pid in the process list.
			if process, err := os.FindProcess(pid); err == nil {
				// Send the process a signal zero kill.
				if err := process.Signal(syscall.Signal(0)); err == nil {
					// We only get an error if the pid isn't running, or it's not ours.
					return fmt.Errorf("pid already running: %d", pid)
				}
			}
		}
	}
	// If we get here, then the pidfile didn't exist,
	// or the pid in it doesn't belong to the user running this app.
	return ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0664)
}

// GetProcessPath function
func GetProcessPath() string {
	processPath, err := filepath.Abs(os.Args[0])
	if err != nil {
		log.Println("failed read process path, error: ", err)
	}

	return processPath
}

// GetProcessDir function
func GetProcessDir() string {
	processPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Println("failed read process path, error: ", err)
	}

	return processPath
}

// GetConfigDir function
func GetConfigDir() string {
	configDir, err := filepath.Abs(filepath.Dir(Properties.ConfigPath))
	if err != nil {
		log.Println("failed read process path, error: ", err)
	}

	return configDir
}
