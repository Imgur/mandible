package server

import (
	"net"

	"github.com/PagerDuty/godspeed"
)

type RuntimeStats interface {
	LogStartup()

	Thumbnail(name string)
	Upload(source string)
}

type DiscardStats struct{}

func (d *DiscardStats) LogStartup()           {}
func (d *DiscardStats) Thumbnail(name string) {}
func (d *DiscardStats) Upload(source string)  {}

type DatadogStats struct {
	dog *godspeed.Godspeed
}

func NewDatadogStats(datadogHost string) (*DatadogStats, error) {
	var ip net.IP = nil
	var err error = nil

	// Assume datadogHost is an IP and try to parse it
	ip = net.ParseIP(datadogHost)

	// Parsing failed
	if ip == nil {
		ips, _ := net.LookupIP(datadogHost)

		if len(ips) > 0 {
			ip = ips[0]
		}
	}

	if ip != nil {
		gdsp, err := godspeed.New(ip.String(), godspeed.DefaultPort, false)
		if err == nil {
			return &DatadogStats{gdsp}, nil
		}
	}

	return nil, err
}

func (d *DatadogStats) LogStartup() {
	d.dog.Incr("mandible.startup", nil)
}

func (d *DatadogStats) Thumbnail(name string) {
	d.dog.Incr("mandible.thumbnail", []string{name})
}

func (d *DatadogStats) Upload(source string) {
	d.dog.Incr("mandible.upload", []string{source})
}
