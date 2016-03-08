// Copyright 2014-2015 PagerDuty, Inc, et al. All rights reserved.
// Use of this source code is governed by the BSD 3-Clause
// license that can be found in the LICENSE file.

package godspeed

import "sync"

// AsyncGodspeed is used for asynchronous Godspeed calls.
// The AsyncGodspeed emission methods have an additional argument
// for a *sync.WaitGroup to have the method indicate when finished.
type AsyncGodspeed struct {
	// Godspeed is an instance of Godspeed
	Godspeed *Godspeed

	// W is a *sync.WaitGroup used for blocking application execution
	// when you want to wait for stats to be emitted.
	// This is here as a convenience, and you can use your own WaitGroup
	// in any AsyncGodspeed method calls.
	W *sync.WaitGroup
}

// NewAsync returns an instance of AsyncGodspeed. This is the more async-friendly version of Godspeed
// autoTruncate dictactes whether long stats emissions get auto-truncated or dropped. Unfortunately,
// Events will always be dropped. If you need monitor your events, you can access the Godspeed instance
// directly.
func NewAsync(host string, port int, autoTruncate bool) (a *AsyncGodspeed, err error) {
	gs, err := New(host, port, autoTruncate)

	if err != nil {
		return nil, err
	}

	a = &AsyncGodspeed{
		Godspeed: gs,
		W:        new(sync.WaitGroup),
	}

	return
}

// NewDefaultAsync is just like NewAsync except it uses the DefaultHost and DefaultPort
func NewDefaultAsync() (a *AsyncGodspeed, err error) {
	a, err = NewAsync(DefaultHost, DefaultPort, false)
	return
}

// AddTag is identical to that within the Godspeed client
func (a *AsyncGodspeed) AddTag(tag string) []string {
	return a.Godspeed.AddTag(tag)
}

// AddTags is identical to that within the Godspeed client
func (a *AsyncGodspeed) AddTags(tags []string) []string {
	return a.Godspeed.AddTags(tags)
}

// SetNamespace is identical to that within the Godspeed client
func (a *AsyncGodspeed) SetNamespace(ns string) {
	a.Godspeed.SetNamespace(ns)
}

// Event is almost identical to that within the Godspeed client
// The only chnage is that it has no return value, and takes a
// (sync.WaitGroup) argument
func (a *AsyncGodspeed) Event(title, body string, keys map[string]string, tags []string, y *sync.WaitGroup) {
	defer y.Done()

	a.Godspeed.Event(title, body, keys, tags)
}

// Send is almost identical to that within the Godspeed client
// with the addition of an argument and removal of the return value
func (a *AsyncGodspeed) Send(stat, kind string, delta, sampleRate float64, tags []string, y *sync.WaitGroup) {
	defer y.Done()

	a.Godspeed.Send(stat, kind, delta, sampleRate, tags)
}

// ServiceCheck is almost identical to that within the Godspeed client
// with the addition of an argument and removal of the return value
func (a *AsyncGodspeed) ServiceCheck(name string, status int, fields map[string]string, tags []string, y *sync.WaitGroup) {
	if y != nil {
		defer y.Done()
	}

	a.Godspeed.ServiceCheck(name, status, fields, tags)
}

// Count is almost identical to that within the Godspeed client
// As with the other AsyncGodpseed functions it omits a return value and
// takes a *sync.WaitGroup instance
func (a *AsyncGodspeed) Count(stat string, count float64, tags []string, y *sync.WaitGroup) {
	defer y.Done()

	a.Godspeed.Count(stat, count, tags)
}

// Incr is almost identical to that within the Godspeed client,
// except it has no return value and takes a *sync.WaitGroup argument.
func (a *AsyncGodspeed) Incr(stat string, tags []string, y *sync.WaitGroup) {
	defer y.Done()

	a.Godspeed.Incr(stat, tags)
}

// Decr is almost identical to that within the Godspeed client. It has
// no return value and takes a *sync.WaitGroup argument.
//
// Also, I've gotten tired of typing "Xxx is almost identical to that within..." so congrats
// on making it this far in to the docs.
func (a *AsyncGodspeed) Decr(stat string, tags []string, y *sync.WaitGroup) {
	defer y.Done()

	a.Godspeed.Decr(stat, tags)
}

// Gauge is almost identical to that within the Godspeed client.
// Here it has no return value, and takes a *sync.WaitGroup argument
func (a *AsyncGodspeed) Gauge(stat string, value float64, tags []string, y *sync.WaitGroup) {
	defer y.Done()

	a.Godspeed.Gauge(stat, value, tags)
}

// Histogram is almost identical to that within the Godspeed client.
// Within AsyncGodspeed it has no return value, and also takes a *sync.WaitGroup argument
func (a *AsyncGodspeed) Histogram(stat string, value float64, tags []string, y *sync.WaitGroup) {
	defer y.Done()

	a.Godspeed.Histogram(stat, value, tags)
}

// Timing is almost identical to that within the Godspeed client.
// The return value is removed, and it takes a *sync.WaitGroup argument here
func (a *AsyncGodspeed) Timing(stat string, value float64, tags []string, y *sync.WaitGroup) {
	defer y.Done()

	a.Godspeed.Timing(stat, value, tags)
}

// Set is almost identical to that within the Godspeed client
func (a *AsyncGodspeed) Set(stat string, value float64, tags []string, y *sync.WaitGroup) {
	defer y.Done()

	a.Godspeed.Set(stat, value, tags)
}
