package management

import (
	"github.com/packet-guardian/pg-dhcp/internal/server"
	"github.com/packet-guardian/pg-dhcp/stats"
)

type Server int

func (s *Server) GetPoolStats(_ int, reply *[]*stats.PoolStat) error {
	*reply = server.GetPoolStats()
	return nil
}
