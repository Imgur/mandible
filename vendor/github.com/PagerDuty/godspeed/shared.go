// Copyright 2014-2015 PagerDuty, Inc, et al. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

package godspeed

import "strings"

// stats names can't include :, |, or @
func trimReserved(s string) string {
	return strings.NewReplacer(":", "_", "|", "_", "@", "_").Replace(s)
}

// function to make sure tags are unique
func uniqueTags(t []string) []string {
	// if the tag slice is empty avoid allocation
	if len(t) < 1 {
		return nil
	}

	// build a map to track which values we've seen
	s := make(map[string]bool)

	// loop over each string provided
	// if the value is not in the map then replace
	// the value at t[len(s)] so that we always have
	// only unique tags at the beginning of the slice
	for i, v := range t {
		if _, x := s[v]; !x {
			// only change the value if needed
			if i != len(s) {
				t[len(s)] = v
			}

			s[v] = true
		}
	}

	// based on the size of the map we know
	// how many unique tags there were
	// so return that slice
	return []string(t[:len(s)])
}
