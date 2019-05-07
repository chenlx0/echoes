package balance

import (
	"github.com/chenlx0/echoes/internal/config"
)

// Hash implement LoadBlanceFunc,
func Hash(ups []config.Upstream) *config.Upstream {
	return &ups[0]
}
