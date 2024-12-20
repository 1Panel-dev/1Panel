package geo

import (
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/oschwald/maxminddb-golang"
	"net"
	"path"
)

type Location struct {
	En string `maxminddb:"en"`
	Zh string `maxminddb:"zh"`
}

type LocationRes struct {
	Iso       string   `maxminddb:"iso"`
	Country   Location `maxminddb:"country"`
	Latitude  float64  `maxminddb:"latitude"`
	Longitude float64  `maxminddb:"longitude"`
	Province  Location `maxminddb:"province"`
}

func GetIPLocation(ip, lang string) (string, error) {
	geoPath := path.Join(global.CONF.System.BaseDir, "1panel", "geo", "GeoIP.mmdb")
	reader, err := maxminddb.Open(geoPath)
	if err != nil {
		return "", err
	}
	var geoLocation LocationRes
	ipNet := net.ParseIP(ip)
	err = reader.Lookup(ipNet, &geoLocation)
	if err != nil {
		return "", err
	}
	if lang == "zh" {
		return geoLocation.Country.Zh + " " + geoLocation.Province.Zh, nil
	}
	return geoLocation.Country.En + " " + geoLocation.Province.En, nil
}
