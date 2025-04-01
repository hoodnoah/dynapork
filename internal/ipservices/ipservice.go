package ipservices

import "net/netip"

// abstraction over external IP Address reporting services
type IpService interface {
	GetIPV4() (netip.Addr, error)
	GetIPV6() (netip.Addr, error)
	HasV4() bool
	HasV6() bool
}
