// Copyright 2014-2015 PagerDuty, Inc, et al. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

package godspeed

import (
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

// Send is the function for emitting the metrics to statsd
// It takes the name of the stat as a string, as well as the kind.
// The kind is "g" for gauge, "c" for count, "ms" for timing, etc.
// This returns any error hit during the flushing of the stat
func (g *Godspeed) Send(stat, kind string, delta, sampleRate float64, tags []string) (err error) {
	// if the connection hasn't been set up yet
	if g.Conn == nil {
		return fmt.Errorf("socket not created")
	}

	// return if the sample rate is less than 1 and the random number is less than the sample rate
	if sampleRate < 1 && rand.Float64() >= sampleRate {
		return nil
	}

	var buffer bytes.Buffer

	// if we have a namespace write it to the byte buffer
	if len(g.Namespace) > 0 {
		buffer.WriteString(fmt.Sprintf("%v.", g.Namespace))
	}

	floatStr := strconv.FormatFloat(delta, 'f', -1, 64)

	// write the name of the metric to the byte buffer as well as the metric itself
	buffer.WriteString(fmt.Sprintf("%v:%v|%v", string(trimReserved(stat)), floatStr, kind))

	// if the sample rate is less than 1 add it too
	if sampleRate < 1 {
		floatStr = strconv.FormatFloat(sampleRate, 'f', -1, 64)
		buffer.WriteString(fmt.Sprintf("|@%v", floatStr))
	}

	// add any provided tags to the metric
	tags = uniqueTags(append(g.Tags, tags...))
	if len(tags) > 0 {
		buffer.WriteString(fmt.Sprintf("|#%v", strings.Join(tags, ",")))
	}

	// this handles the logic for truncation
	// if the buffer length is smaller than the max, just write it
	// else if AutoTruncate is enabled truncate/write the bytes
	// else generate an error to return
	if buffer.Len() <= MaxBytes {
		_, err = g.Conn.Write(buffer.Bytes())
	} else if g.AutoTruncate {
		_, err = g.Conn.Write(buffer.Bytes()[0:MaxBytes])
	} else {
		err = fmt.Errorf("error sending %v, packet larger than %d (%d)", stat, MaxBytes, buffer.Len())
	}

	return
}

// Count wraps Send() and simplifies the interface for Count stats
func (g *Godspeed) Count(stat string, count float64, tags []string) error {
	return g.Send(stat, "c", count, 1, append(g.Tags, tags...))
}

// Incr wraps Send() and simplifies the interface for incrementing a counter
// It only takes the name of the stat, and tags
func (g *Godspeed) Incr(stat string, tags []string) error {
	return g.Count(stat, 1, append(g.Tags, tags...))
}

// Decr wraps Send() and simplifies the interface for decrementing a counter
// It only takes the name of the stat, and tags
func (g *Godspeed) Decr(stat string, tags []string) error {
	return g.Count(stat, -1, append(g.Tags, tags...))
}

// Gauge wraps Send() and simplifies the interface for Gauge stats
func (g *Godspeed) Gauge(stat string, value float64, tags []string) error {
	return g.Send(stat, "g", value, 1, append(g.Tags, tags...))
}

// Histogram wraps Send() and simplifies the interface for Histogram stats
func (g *Godspeed) Histogram(stat string, value float64, tags []string) error {
	return g.Send(stat, "h", value, 1, append(g.Tags, tags...))
}

// Timing wraps Send() and simplifies the interface for Timing stats
func (g *Godspeed) Timing(stat string, value float64, tags []string) error {
	return g.Send(stat, "ms", value, 1, append(g.Tags, tags...))
}

// Set wraps Send() and simplifies the interface for Timing stats
func (g *Godspeed) Set(stat string, value float64, tags []string) error {
	return g.Send(stat, "s", value, 1, append(g.Tags, tags...))
}
