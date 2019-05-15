package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	configuration "github.com/speedtest-monitor/app/configuration"
	"github.com/speedtest-monitor/app/server/router"
	"github.com/speedtest-monitor/app/slack"
	"github.com/speedtest-monitor/app/speedtest"
	log "github.com/speedtest-monitor/pkg/Sirupsen/logrus"

	"gopkg.in/alecthomas/kingpin.v2"
)

func setTimeout() {
	if *timeoutOpt != 0 {
		timeout = *timeoutOpt
	}
}

var (
	showList      = kingpin.Flag("list", "Show available speedtest.net servers").Short('l').Bool()
	serverIds     = kingpin.Flag("server", "Select server id to speedtest").Short('s').Ints()
	timeoutOpt    = kingpin.Flag("timeout", "Define timeout seconds. Default: 10 sec").Short('t').Int()
	timeout       = 10
	serviceName   = "Speed Test"
	port          = "8080"
	latestResults = &speedtest.LatestResult{}
	config        = configuration.LoadConfiguration()
)

func main() {
	kingpin.Version("1.0.3")
	kingpin.Parse()

	setTimeout()

	if config.LoggingLevel == "debug" {
		log.SetLevel(log.DebugLevel)
	}

	user := speedtest.FetchUserInfo()
	user.Show()

	list := speedtest.FetchServerList(user)
	if *showList {
		list.Show()
		return
	}

	targets := list.FindServer(*serverIds)
	if config.ServerMode == true {
		startServer(&targets, latestResults, config)
	} else {
		targets.TestNow(latestResults)
	}
}

func startServer(targets *speedtest.Servers, latestResults *speedtest.LatestResult, config *configuration.Configuration) {
	router := router.NewRouter(targets, latestResults)
	var sc *slack.SlackClient
	if config.SlackEndpoint != "" {
		sc = slack.NewSlackClient(config)
	}
	// Create a new server and set timeout values.
	server := http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    900 * time.Second,
		WriteTimeout:   900 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// We want to report the listener is closed.
	var wg sync.WaitGroup
	wg.Add(1)

	// Start the listener.
	go func() {
		log.Infof("%s running!", config.AppName)
		if config.SlackEndpoint != "" {
			sc.AddMessage(fmt.Sprintf("Running %s", config.AppName))
			sc.SendMessages()
		}
		log.Infof("Listener closed : %v", server.ListenAndServe())
		if config.SlackEndpoint != "" {
			sc.AddMessage(fmt.Sprintf("Shutting down %s", config.AppName))
			sc.SendMessages()
		}
		wg.Done()
	}()

	startPoller(targets, sc)
	// Listen for an interrupt signal from the OS.
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt)

	// Wait for a signal to shutdown.
	<-osSignals

	// Create a context to attempt a graceful 5 second shutdown.
	const timeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Attempt the graceful shutdown by closing the listener and
	// completing all inflight requests.
	if err := server.Shutdown(ctx); err != nil {

		log.WithFields(log.Fields{
			"Method":  "main",
			"Action":  "shutdown",
			"Timeout": timeout,
			"Message": err.Error(),
		}).Error("Graceful shutdown did not complete")

		// Looks like we timedout on the graceful shutdown. Kill it hard.
		if err := server.Close(); err != nil {
			log.WithFields(log.Fields{
				"Method":  "main",
				"Action":  "shutdown",
				"Message": err.Error(),
			}).Error("Error killing server")
		}
	}

	// Wait for the listener to report it is closed.
	wg.Wait()
}

func testNow(targets *speedtest.Servers, sc *slack.SlackClient) {
	targets.TestNow(latestResults)
	if config.SlackEndpoint != "" {
		if latestResults.DownSpeed < config.DownloadThreshold {
			sc.AddAlert(fmt.Sprintf("Current Download speed (%s) lower than %v", latestResults.DownSpeedStr, config.DownloadThreshold))
		}
		if latestResults.UpSpeed < config.UploadThreshold {
			sc.AddAlert(fmt.Sprintf("Current Upload speed (%s) lower than %v", latestResults.UpSpeedStr, config.UploadThreshold))
		}
		if len(sc.Messages) > 0 {
			go sc.SendMessages()
		}
	}
}

func startPoller(targets *speedtest.Servers, sc *slack.SlackClient) {
	testNow(targets, sc)
	go func() {
		interval := config.Interval
		pollPeriod := time.Second * time.Duration(interval)
		for {
			log.Debugf("Waiting %d seconds before next poll.", interval)

			select {
			case <-time.After(pollPeriod):
				go testNow(targets, sc)
			}
		}
	}()
}
