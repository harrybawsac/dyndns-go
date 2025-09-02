package lib

import (
	"context"
	"dyndns-go/schema"
	"errors"
	"fmt"
	"net"
)

// StratoRegistrar implements Registrar for Strato DNS
// Uses StratoDynDnsUrl endpoint

type StratoRegistrar struct{}

func (r *StratoRegistrar) UpdateDNS(ctx context.Context, config *Config, ipv4, ipv6 string) error {
	if config.Host == "" {
		return errors.New("host is required")
	}
	if config.User == "" || config.Password == "" {
		return errors.New("user and password required")
	}
	endpoint := schema.StratoDynDnsUrl
	if config.UpdateIpv4 && ipv4 != "" {
		res, err := UpdateDynDNS(ctx, endpoint, config.User, config.Password, config.Host, net.ParseIP(ipv4))
		fmt.Println("Strato IPv4 update:", res.Status, res.IP, res.Raw, "err:", err)
	}
	if config.UpdateIpv6 && ipv6 != "" {
		res, err := UpdateDynDNS(ctx, endpoint, config.User, config.Password, config.Host, net.ParseIP(ipv6))
		fmt.Println("Strato IPv6 update:", res.Status, res.IP, res.Raw, "err:", err)
	}
	return nil
}

// Add more registrars by implementing Registrar interface
// Example: DynDNSRegistrar, NoIPRegistrar, etc.
