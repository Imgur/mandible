// Copyright 2014-2015 PagerDuty, Inc, et al. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

package godspeed

import (
	"bytes"
	"fmt"
	"strings"
)

var eventKeys = []string{"date_happened", "hostname", "aggregation_key", "priority", "source_type_name", "alert_type"}
var eventMarkers = []rune{'d', 'h', 'k', 'p', 's', 't'}

func escapeEvent(s string) string {
	return strings.NewReplacer("\n", "\\n").Replace(s)
}

func removePipes(s string) string {
	return strings.Replace(s, "|", "", -1)
}

// Event is the function for submitting a Datadog event.
// This is a Datadog-specific emission and most likely will not work on other statsd implementations.
// title and body are both strings, and are the title and body of the event respectively.
// field can be used to send the optional keys.
func (g *Godspeed) Event(title, text string, fields map[string]string, tags []string) error {
	if len(title) < 1 {
		return fmt.Errorf("title must have at least one character")
	}

	if len(text) < 1 {
		return fmt.Errorf("body must have at least one character")
	}

	var buf bytes.Buffer

	title = escapeEvent(title)
	text = escapeEvent(text)

	buf.WriteString(fmt.Sprintf("_e{%d,%d}:%v|%v", len(title), len(text), title, text))

	// if some fields were passed in convert them to their proper format
	// and write that to the buffer
	if len(fields) > 0 {
		for i, v := range eventKeys {
			if mv, ok := fields[v]; ok {
				buf.WriteString(fmt.Sprintf("|%v:%v", string(eventMarkers[i]), removePipes(mv)))
			}
		}
	}

	tags = uniqueTags(append(g.Tags, tags...))

	if len(tags) > 0 {
		for i, v := range tags {
			tags[i] = strings.Replace(v, "|", "", -1)
		}

		buf.WriteString(fmt.Sprintf("|#%v", strings.Join(tags, ",")))
	}

	// this handles the logic for truncation
	// if the buffer length is larger than the max, return an error
	// else just write it
	if bufLen := buf.Len(); bufLen > MaxBytes {
		return fmt.Errorf("error sending %v, packet larger than %d (%d)", string(title), MaxBytes, buf.Len())
	}

	_, err := g.Conn.Write(buf.Bytes())
	return err
}
