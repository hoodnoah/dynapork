package ipservices

import (
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
)

type iCanHazIp struct {
	client *http.Client
	v4url  url.URL
	v6url  url.URL
}

func NewICanHazIp(client *http.Client) (IpService, error) {
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
