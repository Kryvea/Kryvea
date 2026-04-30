package model

type Target struct {
	Model
	IPv4     string   `json:"ipv4,omitempty"`
	IPv6     string   `json:"ipv6,omitempty"`
	Port     int      `json:"port,omitempty"`
	Protocol string   `json:"protocol,omitempty"`
	FQDN     string   `json:"fqdn"`
	Tag      string   `json:"tag,omitempty"`
	Customer Customer `json:"customer,omitempty"`
}
