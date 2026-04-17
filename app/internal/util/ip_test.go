package util

import "testing"

func TestIsValidIPv4(t *testing.T) {
	validIPs := []string{
		"192.168.1.1",
		"8.8.8.8",
		"127.0.0.1",
		"255.255.255.255",
		"0.0.0.0",
	}

	invalidIPs := []string{
		"",
		"300.168.1.1",
		"192.168.1",
		"::1",
		"192.168.1.01", // leading zero
		"abc.def.ghi.jkl",
	}

	for _, ip := range validIPs {
		if !IsValidIPv4(ip) {
			t.Errorf("Expected valid IPv4: %s", ip)
		}
	}

	for _, ip := range invalidIPs {
		if IsValidIPv4(ip) {
			t.Errorf("Expected invalid IPv4: %s", ip)
		}
	}
}

func TestIsValidIPv6(t *testing.T) {
	validIPs := []string{
		"::1",
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		"fe80::1ff:fe23:4567:890a",
		"::",
		"2001:db8::",
	}

	invalidIPs := []string{
		"",
		"192.168.1.1",
		"12345::",
		"2001:db8:::1",
		"::g",
	}

	for _, ip := range validIPs {
		if !IsValidIPv6(ip) {
			t.Errorf("Expected valid IPv6: %s", ip)
		}
	}

	for _, ip := range invalidIPs {
		if IsValidIPv6(ip) {
			t.Errorf("Expected invalid IPv6: %s", ip)
		}
	}
}
