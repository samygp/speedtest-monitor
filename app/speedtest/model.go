package speedtest

import (
	"time"

	log "github.com/speedtest-monitor/pkg/Sirupsen/logrus"
)

// Server information
type Server struct {
	URL      string `xml:"url,attr"`
	Lat      string `xml:"lat,attr"`
	Lon      string `xml:"lon,attr"`
	Name     string `xml:"name,attr"`
	Country  string `xml:"country,attr"`
	Sponsor  string `xml:"sponsor,attr"`
	ID       string `xml:"id,attr"`
	URL2     string `xml:"url2,attr"`
	Host     string `xml:"host,attr"`
	Distance float64
	DLSpeed  float64
	ULSpeed  float64
	Ping     time.Duration
}

// ServerList : List of Server
type ServerList struct {
	Servers []Server `xml:"servers>server"`
}

// Servers : For sorting servers.
type Servers []Server

// ByDistance : For sorting servers.
type ByDistance struct {
	Servers
}

// LatestResult : keeps strings with the latest measurements
type LatestResult struct {
	DownSpeed    float64
	DownSpeedStr string
	Ping         string
	UpSpeed      float64
	UpSpeedStr   string
	LastQuery    string
}

func checkError(err error) {
	if err != nil {
		log.Error(err)
	}
}
