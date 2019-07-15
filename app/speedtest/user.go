package speedtest

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// User information
type User struct {
	IP  string `xml:"ip,attr"`
	Lat string `xml:"lat,attr"`
	Lon string `xml:"lon,attr"`
	Isp string `xml:"isp,attr"`
}

// Users : for decode xml
type Users struct {
	Users []User `xml:"client"`
}

//FetchUserInfo retrieves information related to the user based on their
//geographical location
func FetchUserInfo() User {
	// Fetch xml user data
	resp, err := http.Get("http://speedtest.net/speedtest-config.php")
	checkError(err)
	body, err := ioutil.ReadAll(resp.Body)
	checkError(err)
	defer resp.Body.Close()

	// Decode xml
	decoder := xml.NewDecoder(bytes.NewReader(body))
	users := Users{}
	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			decoder.DecodeElement(&users, &se)
		}
	}
	if users.Users == nil {
		log.Warn("Warning: Cannot fetch user information. http://www.speedtest.net/speedtest-config.php is temporarily unavailable.")
		return User{}
	}
	return users.Users[0]
}

// Show user location
func (u *User) Show() {
	if u.IP != "" {
		log.Infof("Testing From IP: %s (%s) [%s, %s]", u.IP, u.Isp, u.Lat, u.Lon)
	}
}
