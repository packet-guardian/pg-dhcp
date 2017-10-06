package management

import (
	"github.com/packet-guardian/pg-dhcp/internal/server"
	"github.com/packet-guardian/pg-dhcp/stats"
)

type Server int

func (s *Server) GetPoolStats() []*stats.PoolStat {
	return server.GetPoolStats()
}
