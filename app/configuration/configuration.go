package configuration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	log "github.com/speedtest-monitor/pkg/Sirupsen/logrus"
)

// Configuration represents the config variables to be loaded from a
// configuration.json file
type Configuration struct {
	AppName           string
	DownloadThreshold float64
	Interval          int
	LoggingLevel      string
	UploadThreshold   float64
	ServerMode        bool
	SlackEndpoint     string
}

// LoadConfiguration loads the json config file into a Configuration struct
func LoadConfiguration() *Configuration {
	_, filename, _, ok := runtime.Caller(2)
	if !ok {
		log.Error("No caller information")
		return nil
	}
	log.Debug(filename)

	filePath := "./app/configuration/configuration.json"

	if _, err := os.Stat(filePath); err != nil {
		log.Errorf("Error while reading JSON config file: %s. Trying with absolute path", err)
		filePath = "/run/config/configuration.json"
		if _, err := os.Stat(filePath); err != nil {
			log.Errorf("Error while reading JSON config file: %s", err)
			panic(fmt.Sprintf("Error while reading JSON config file: %s", err))
		}
	}

	jFile, _ := ioutil.ReadFile(filePath)
	var data map[string]interface{}
	err := json.Unmarshal(jFile, &data)

	if err != nil {
		log.Errorf("Error while reading JSON config file: %s", err)
		panic(fmt.Sprintf("Error while reading JSON config file: %s", err))
	}

	conf := &Configuration{}
	conf.AppName = fmt.Sprintf("%v", data["appName"])
	conf.LoggingLevel = fmt.Sprintf("%v", data["loggingLevel"])
	conf.DownloadThreshold = data["downloadThreshold"].(float64)
	conf.Interval = int(data["interval"].(float64))
	conf.UploadThreshold = data["uploadThreshold"].(float64)
	conf.ServerMode = data["serverMode"].(bool)
	conf.SlackEndpoint = fmt.Sprintf("%v", data["slackEndpoint"])
	return conf
}
