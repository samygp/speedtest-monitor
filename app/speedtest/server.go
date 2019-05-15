package speedtest

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	log "github.com/speedtest-monitor/pkg/Sirupsen/logrus"
)

// Len : length of servers. For sorting servers.
func (svrs Servers) Len() int {
	return len(svrs)
}

// Swap : swap i-th and j-th. For sorting servers.
func (svrs Servers) Swap(i, j int) {
	svrs[i], svrs[j] = svrs[j], svrs[i]
}

// Less : compare the distance. For sorting servers.
func (b ByDistance) Less(i, j int) bool {
	return b.Servers[i].Distance < b.Servers[j].Distance
}

//FetchServerList retrieves information of the available
//servers based on the user's geolocation
func FetchServerList(user User) ServerList {
	// Fetch xml server data
	resp, err := http.Get("http://www.speedtest.net/speedtest-servers-static.php")
	checkError(err)
	body, err := ioutil.ReadAll(resp.Body)
	checkError(err)
	defer resp.Body.Close()

	if len(body) == 0 {
		resp, err = http.Get("http://c.speedtest.net/speedtest-servers-static.php")
		checkError(err)
		body, err = ioutil.ReadAll(resp.Body)
		checkError(err)
		defer resp.Body.Close()
	}

	// Decode xml
	decoder := xml.NewDecoder(bytes.NewReader(body))
	list := ServerList{}
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			decoder.DecodeElement(&list, &se)
		}
	}

	// Calculate distance
	for i := range list.Servers {
		server := &list.Servers[i]
		sLat, _ := strconv.ParseFloat(server.Lat, 64)
		sLon, _ := strconv.ParseFloat(server.Lon, 64)
		uLat, _ := strconv.ParseFloat(user.Lat, 64)
		uLon, _ := strconv.ParseFloat(user.Lon, 64)
		server.Distance = distance(sLat, sLon, uLat, uLon)
	}

	// Sort by distance
	sort.Sort(ByDistance{list.Servers})

	return list
}

func distance(lat1 float64, lon1 float64, lat2 float64, lon2 float64) float64 {
	radius := 6378.137

	a1 := lat1 * math.Pi / 180.0
	b1 := lon1 * math.Pi / 180.0
	a2 := lat2 * math.Pi / 180.0
	b2 := lon2 * math.Pi / 180.0

	x := math.Sin(a1)*math.Sin(a2) + math.Cos(a1)*math.Cos(a2)*math.Cos(b2-b1)
	return radius * math.Acos(x)
}

// FindServer : find server by serverID
func (l *ServerList) FindServer(serverID []int) Servers {
	servers := Servers{}

	for _, sid := range serverID {
		for _, s := range l.Servers {
			id, _ := strconv.Atoi(s.ID)
			if sid == id {
				servers = append(servers, s)
			}
		}
	}

	if len(servers) == 0 {
		servers = append(servers, l.Servers[0])
	}

	return servers
}

// Show : show server list
func (l ServerList) Show() {
	for _, s := range l.Servers {
		log.Infof("[%4s] %8.2fkm ", s.ID, s.Distance)
		log.Info(s.Name + " (" + s.Country + ") by " + s.Sponsor + "\n")
	}
}

// Show : show server information
func (s Server) Show() {
	log.Infof("\nTarget Server: [%4s] %8.2fkm ", s.ID, s.Distance)
	log.Info(s.Name + " (" + s.Country + ") by " + s.Sponsor + "\n")
}

// TestNow starts the test and shows the results
func (svrs *Servers) TestNow(latestResults *LatestResult) {
	svrs.StartTest()
	svrs.ShowResult(latestResults)
}

// StartTest : start testing to the servers.
func (svrs Servers) StartTest() {
	for i, s := range svrs {
		s.Show()
		latency := pingTest(s.URL)
		dlSpeed := downloadTest(s.URL, latency)
		ulSpeed := uploadTest(s.URL, latency)
		svrs[i].DLSpeed = dlSpeed
		svrs[i].ULSpeed = ulSpeed
		svrs[i].Ping = latency
	}
}

// ShowResult : show testing result
func (svrs Servers) ShowResult(latestResult *LatestResult) {
	fmt.Printf(" \n")
	if len(svrs) == 1 {
		latestResult.DownSpeed = svrs[0].DLSpeed
		latestResult.DownSpeedStr = fmt.Sprintf("Download: %5.2f Mbit/s", svrs[0].DLSpeed)
		latestResult.UpSpeed = svrs[0].ULSpeed
		latestResult.UpSpeedStr = fmt.Sprintf("Upload: %5.2f Mbit/s", svrs[0].ULSpeed)
		latestResult.Ping = fmt.Sprint("Latency: ", svrs[0].Ping)
	} else {
		avgDL := 0.0
		avgUL := 0.0
		avgPing := 0.0
		for _, s := range svrs {
			log.Infof("[%4s] Download: %5.2f Mbit/s, Upload: %5.2f Mbit/s\n", s.ID, s.DLSpeed, s.ULSpeed)
			avgDL = avgDL + s.DLSpeed
			avgUL = avgUL + s.ULSpeed
			avgPing = avgPing + float64(s.Ping)
		}
		latestResult.DownSpeed = avgDL / float64(len(svrs))
		latestResult.DownSpeedStr = fmt.Sprintf("Download Avg: %5.2f Mbit/s", avgDL/float64(len(svrs)))
		latestResult.UpSpeed = avgUL / float64(len(svrs))
		latestResult.UpSpeedStr = fmt.Sprintf("Upload Avg: %5.2f Mbit/s", avgUL/float64(len(svrs)))
		latestResult.Ping = fmt.Sprintf("Upload Avg: %5.2f ms", avgPing/float64(len(svrs)))
	}
	latestResult.LastQuery = time.Now().String()
	log.Info(latestResult.DownSpeedStr)
	log.Info(latestResult.UpSpeedStr)
	err := svrs.checkResult()
	if err {
		log.Warn("Warning: Result seems to be wrong. Please speedtest again.")
	}
}

func (svrs Servers) checkResult() bool {
	errFlg := false
	if len(svrs) == 1 {
		s := svrs[0]
		errFlg = (s.DLSpeed*100 < s.ULSpeed) || (s.DLSpeed > s.ULSpeed*100)
	} else {
		for _, s := range svrs {
			errFlg = errFlg || (s.DLSpeed*100 < s.ULSpeed) || (s.DLSpeed > s.ULSpeed*100)
		}
	}
	return errFlg
}
