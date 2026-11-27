package geo

import (
	"net"
	"path"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/oschwald/maxminddb-golang"
)

var carrierNATNet = &net.IPNet{
	IP:   net.IPv4(100, 64, 0, 0),
	Mask: net.CIDRMask(10, 32),
}

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

func NewGeo() (*maxminddb.Reader, error) {
	geoPath := path.Join(global.Dir.DataDir, "geo", "GeoIP.mmdb")
	return maxminddb.Open(geoPath)
}

// GetIPLocation resolves known address ranges without a database. A nil reader
// leaves other locations empty; the caller owns the database reader.
func GetIPLocation(reader *maxminddb.Reader, ip, lang string) (string, error) {
	ipNet := net.ParseIP(ip)
	if carrierNATNet.Contains(ipNet) {
		if lang == "zh" {
			return "运营商级NAT", nil
		}
		return "Carrier-grade NAT", nil
	}

	if reader == nil {
		return "", nil
	}
	var geoLocation LocationRes
	if err := reader.Lookup(ipNet, &geoLocation); err != nil {
		return "", err
	}
	if lang == "zh" {
		return geoLocation.Country.Zh + " " + geoLocation.Province.Zh, nil
	}
	return geoLocation.Country.En + " " + geoLocation.Province.En, nil
}
