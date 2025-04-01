package internal

import (
	"net/netip"
	"testing"
)

type MockIPService struct {
	ReturnV4  netip.Addr
	ReturnV6  netip.Addr
	V4Enabled bool
	V6Enabled bool
}

// implement IPService interface on mock
func (m *MockIPService) GetIPV4() (netip.Addr, error) {
	return m.ReturnV4, nil
}

func (m *MockIPService) GetIPV6() (netip.Addr, error) {
	return m.ReturnV6, nil
}

func (m *MockIPService) HasV4() bool {
	return m.V4Enabled
}

func (m *MockIPService) HasV6() bool {
	return m.V6Enabled
}

// mock optional func
func WithMockIPServices(services []IpService) Option {
	return func(m *IPMonitor) {
		m.ipServices = services
	}
}

func TestIpMonitorV4(t *testing.T) {

	initialAddr := netip.AddrFrom4([4]byte{0xff, 0xff, 0xff, 0xff})

	mockSvc := MockIPService{}
	mockSvc.ReturnV4 = initialAddr
	mockSvc.V4Enabled = true
	mockSvc.V6Enabled = false

	monitor, err := NewIPMonitor(WithIpServices([]IpService{&mockSvc}))

	if err != nil {
		t.Errorf("failed to instantiate a new IPMonitor: %v", err)
	}

	// should show a change for the initial address
	v4Chan := monitor.SubscribeV4()

	// blocks
	addr := <-v4Chan

	if addr != initialAddr {
		t.Fatalf("expected %v, received %v", initialAddr, addr)
	}

	// Change the return address
	newAddr := netip.AddrFrom4([4]byte{0xff, 0xff, 0xff, 0xfe})
	mockSvc.ReturnV4 = newAddr

	// blocks
	addr = <-v4Chan
	if addr != newAddr {
		t.Fatalf("expected %v, received %v", newAddr, addr)
	}
}

func TestIpMonitorV6(t *testing.T) {

	initialAddr := netip.AddrFrom16([16]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})

	mockSvc := MockIPService{}
	mockSvc.ReturnV6 = initialAddr
	mockSvc.V4Enabled = false
	mockSvc.V6Enabled = true

	monitor, err := NewIPMonitor(WithIpServices([]IpService{&mockSvc}))

	if err != nil {
		t.Errorf("failed to instantiate a new IPMonitor: %v", err)
	}

	// should show a change for the initial address
	v6Chan := monitor.SubscribeV6()

	// blocks
	addr := <-v6Chan

	if addr != initialAddr {
		t.Fatalf("expected %v, received %v", initialAddr, addr)
	}

	// Change the return address
	newAddr := netip.AddrFrom16([16]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe})
	mockSvc.ReturnV6 = newAddr

	// blocks
	addr = <-v6Chan
	if addr != newAddr {
		t.Fatalf("expected %v, received %v", newAddr, addr)
	}
}
