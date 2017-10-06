package rpcclient

import "github.com/packet-guardian/pg-dhcp/stats"

type ServerRequest struct {
	client *Client
}

func (s *ServerRequest) GetPoolStats() ([]*stats.PoolStat, error) {
	reply := make([]*stats.PoolStat, 0)
	if err := s.client.c.Call("Server.GetPoolStats", nil, reply); err != nil {
		return nil, err
	}
	return reply, nil
}
