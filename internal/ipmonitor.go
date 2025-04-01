package internal

import (
	// std
	"fmt"
	"net/http"
	"net/netip"
	"sync"
	"time"

	// external
	ipsvc "github.com/hoodnoah/dynapork/internal/ipservices"
	tb "github.com/hoodnoah/token_bucket"
)

type IPMonitor struct {
	client     http.Client
	ipServices []ipsvc.IpService
	v4Changes  chan netip.Addr
	v6Changes  chan netip.Addr
	limiter    tb.TokenBucket
	lastV4     netip.Addr
	lastV6     netip.Addr
	mutex      sync.Mutex
}

// option type; allow configuration of the IPMonitor flexibly
type Option func(*IPMonitor)

// option which applies a provided slice of ipServices.
// dependency injection.
func WithIpServices(services []ipsvc.IpService) Option {
	return func(m *IPMonitor) {
		m.ipServices = services
	}
}

// default option
func WithDefaultIpServices() Option {
	return func(m *IPMonitor) {
		services := make([]ipsvc.IpService, 0)

		canHaz, err := ipsvc.NewICanHazIp(&m.client)
		if err != nil {
		} // ignore error

		services = append(services, canHaz)

		m.ipServices = services
	}
}

// constructor for a new IP address monitor
func NewIPMonitor(options ...Option) (*IPMonitor, error) {
	client := http.Client{
		Timeout: time.Second * 2, // restrictive timeout since there are redundant sources
	}

	limiter := tb.NewTokenBucket(2, time.Second)

	monitor := IPMonitor{
		client:     client,
		ipServices: make([]ipsvc.IpService, 0),
		limiter:    *limiter,
		v4Changes:  make(chan netip.Addr),
		v6Changes:  make(chan netip.Addr),
	}

	for _, opt := range options {
		opt(&monitor)
	}

	// start polling ip endpoints
	go monitor.run()

	return &monitor, nil
}

// for all provided IPServices, poll them for v4 and v6 indefinitely
func (i *IPMonitor) run() {
	idx := 0

	for {
		i.mutex.Lock()
		svc := i.ipServices[idx]

		// wait for limiter, request v4 if available
		if svc.HasV4() {
			i.limiter.Wait()
			v4, err := svc.GetIPV4()
			if err != nil {
				fmt.Printf("unexpected error retrieving v4: %v", err)
			} // ignore error

			// if new, publish update
			if v4 != i.lastV4 {
				i.lastV4 = v4
				i.v4Changes <- v4
			}
		}

		// wait for limiter, request v6 if available
		if svc.HasV6() {
			i.limiter.Wait()
			v6, err := svc.GetIPV6()
			if err != nil {
				fmt.Printf("unexpected error retrieving v6: %v", err)
			} // ignore error

			// if new, publish update
			if v6 != i.lastV6 {
				i.lastV6 = v6
				i.v6Changes <- v6
			}

		}
		// if we've reached the last ipService, restart at 0
		if idx == len(i.ipServices)-1 {
			idx = 0
		} else { // otherwise move to the next
			idx++
		}

		i.mutex.Unlock()
	}
}

// receive a channel, to which observed v4 changes are pushed
func (i *IPMonitor) SubscribeV4() chan netip.Addr {
	return i.v4Changes
}

// receive a channel, to which observed v6 changes are pushed
func (i *IPMonitor) SubscribeV6() chan netip.Addr {
	return i.v6Changes
}
