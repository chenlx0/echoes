package balance

import (
	"github.com/chenlx0/echoes/internal/config"
)

const (
	// RR round robin
	RR = "RR"
	// HASH ip hash
	HASH = "HASH"
)

// LoadBalanceFunc load balance function
type LoadBalanceFunc func([]config.Upstream) *config.Upstream

// GetLoadBalanceFunc return load balance func by the method specified
func GetLoadBalanceFunc(method string) LoadBalanceFunc {
	switch method {
	case RR:
		return RoundRobin
	case HASH:
		return Hash
	}
	return nil
}
