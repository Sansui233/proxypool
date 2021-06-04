package healthcheck

import (
	"bytes"
	"encoding/xml"
	"errors"
	C "github.com/Dreamacro/clash/constant"
)

// speedtest.net config
type User struct {
	IP  string `xml:"ip,attr"`
	Lat string `xml:"lat,attr"`
	Lon string `xml:"lon,attr"`
	Isp string `xml:"isp,attr"`
}

// Users : for decode speedtest.net xml
type Users struct {
	Users []User `xml:"client"`
}

// fetchUserInfo with proxy connection
func fetchUserInfo(clashProxy C.Proxy) (user *User, err error) {
	url := "https://www.speedtest.net/speedtest-config.php"
	body, err := HTTPGetBodyViaProxy(clashProxy, url)
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
		//log.Println("Warning: Cannot fetch user information. http://www.speedtest.net/speedtest-config.php is temporarily unavailable.")
		return nil, errors.New("No user to speedtest.net. ")
	}
	return &users.Users[0], nil
}
