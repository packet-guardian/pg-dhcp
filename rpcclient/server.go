package rpcclient

import "github.com/packet-guardian/pg-dhcp/stats"

type ServerRPCRequest struct {
	client *RPCClient
}

func (s *ServerRPCRequest) GetPoolStats() ([]*stats.PoolStat, error) {
	var reply []*stats.PoolStat
	if err := s.client.c.Call("Server.GetPoolStats", 0, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}
