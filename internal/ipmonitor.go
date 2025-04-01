package internal

import (
	// std
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"sync"
	"time"

	// external
	tb "github.com/hoodnoah/token_bucket"
)

// abstraction over external IP Address reporting services
type IpService interface {
	GetIPV4() (netip.Addr, error)
	GetIPV6() (netip.Addr, error)
	HasV4() bool
	HasV6() bool
}

type IPMonitor struct {
	client     http.Client
	ipServices []IpService
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
func WithIpServices(services []IpService) Option {
	return func(m *IPMonitor) {
		m.ipServices = services
	}
}

// default option
func WithDefaultIpServices() Option {
	return func(m *IPMonitor) {
		services := make([]IpService, 0)

		canHaz, err := newICanHazIp(&m.client)
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
		ipServices: make([]IpService, 0),
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

type iCanHazIp struct {
	client *http.Client
	v4url  url.URL
	v6url  url.URL
}

func newICanHazIp(client *http.Client) (IpService, error) {
	v4Url, err := url.Parse("https://icanhazip.com")
	if err != nil {
		return nil, err
	}

	v6Url, err := url.Parse("https://ipv6.icanhazip.com")
	if err != nil {
		return nil, err
	}

	return iCanHazIp{
		client: client,
		v4url:  *v4Url,
		v6url:  *v6Url,
	}, nil
}

func (i iCanHazIp) HasV4() bool {
	return true
}

func (i iCanHazIp) HasV6() bool {
	return true
}

func (i iCanHazIp) GetIPV4() (netip.Addr, error) {
	response, err := i.client.Get(i.v4url.Host)
	if err != nil {
		return netip.IPv4Unspecified(), err
	}

	if response.StatusCode != http.StatusOK {
		return netip.IPv4Unspecified(), fmt.Errorf("received non-200 status code: %v", response.Status)
	}

	// read bytes of body, should be an IPV4 Address
	bodyBytes := make([]byte, 0, 4)
	_, err = response.Body.Read(bodyBytes)
	if err != nil {
		return netip.IPv4Unspecified(), err
	}

	// parse v4 address from 4 bytes
	ip := netip.AddrFrom4([4]byte(bodyBytes))

	return ip, nil
}

func (i iCanHazIp) GetIPV6() (netip.Addr, error) {
	response, err := i.client.Get(i.v6url.Host)
	if err != nil {
		return netip.IPv6Unspecified(), err
	}

	if response.StatusCode != http.StatusOK {
		return netip.IPv6Unspecified(), fmt.Errorf("received non-200 status code: %v", response.Status)
	}

	// read bytes of body, should be an IPV4 Address
	bodyBytes := make([]byte, 0, 16)
	_, err = response.Body.Read(bodyBytes)
	if err != nil {
		return netip.IPv6Unspecified(), err
	}

	// parse v6 address from 16 bytes
	ip := netip.AddrFrom16(([16]byte(bodyBytes)))

	return ip, nil
}
