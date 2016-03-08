// Copyright 2014-2015 PagerDuty, Inc, et al. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

package godspeed

import (
	"bytes"
	"fmt"
	"strings"
)

var scKeys = []string{"service_check_message", "timestamp", "hostname"}
var scMark = []string{"m", "d", "h"}

// ServiceCheck is a function to emit DogStatsD service checks
// to the local DD agent. It takes the name of the service,
// which must NOT contain a pipe (|) character, and the numeric
// status for the service. The status values are the same as Nagios:
//
// OK = 0, WARNING = 1, CRITICAL = 2, UNKNOWN = 3
//
// This functionality is an extension to the statsd
// protocol by Datadog (DogStatsD):
//
// http://docs.datadoghq.com/guides/dogstatsd/#service-checks
func (g *Godspeed) ServiceCheck(name string, status int, fields map[string]string, tags []string) error {
	if len(name) == 0 {
		return fmt.Errorf("service name must have at least one character")
	}

	if status < 0 || status > 3 {
		return fmt.Errorf("unknown service status (%d); known values: 0,1,2,3", status)
	}

	if strings.ContainsAny("|", name) {
		return fmt.Errorf("service name '%s' may not include pipe character ('|')", name)
	}

	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("_sc|%s|%d", name, status))

	if len(fields) > 0 {
		for i, v := range scKeys {
			if mv, ok := fields[v]; ok {
				buf.WriteString(fmt.Sprintf("|%s:%s", scMark[i], removePipes(mv)))
			}
		}
	}

	tags = uniqueTags(append(g.Tags, tags...))

	if len(tags) > 0 {
		for i, v := range tags {
			tags[i] = strings.Replace(v, "|", "", -1)
		}
		buf.WriteString(fmt.Sprintf("|#%s", strings.Join(tags, ",")))
	}

	if bufLen := buf.Len(); bufLen > MaxBytes {
		return fmt.Errorf("error sending %s service check, packet larger than %d (%d)", name, MaxBytes, bufLen)
	}

	_, err := g.Conn.Write(buf.Bytes())
	return err
}
