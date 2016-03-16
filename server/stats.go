package server

import (
	"fmt"
	"net"
	"time"

	"github.com/PagerDuty/godspeed"
)

type RuntimeStats interface {
	LogStartup()

	Request()
	ResponseTime(elapsed time.Duration)
	Thumbnail(name string)
	Upload(source string)
	Error(code int)
}

type DiscardStats struct{}

func (d *DiscardStats) LogStartup()                        {}
func (d *DiscardStats) Request()                           {}
func (d *DiscardStats) ResponseTime(elapsed time.Duration) {}
func (d *DiscardStats) Thumbnail(name string)              {}
func (d *DiscardStats) Upload(source string)               {}
func (d *DiscardStats) Error(code int)                     {}

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

func (d *DatadogStats) Request() {
	d.dog.Incr("mandible.request", nil)
}

func (d *DatadogStats) ResponseTime(elapsed time.Duration) {
	time := elapsed.Seconds()

	d.dog.Timing("mandible.responseTime", time, nil)
}

func (d *DatadogStats) Thumbnail(name string) {
	tag := fmt.Sprintf("size:%s", name)

	d.dog.Incr("mandible.thumbnail", []string{tag})
}

func (d *DatadogStats) Upload(source string) {
	tag := fmt.Sprintf("source:%s", source)

	d.dog.Incr("mandible.upload", []string{tag})
}

func (d *DatadogStats) Error(code int) {
	tag := fmt.Sprintf("code:%d", code)
	d.dog.Incr("mandible.error", []string{tag})
}
