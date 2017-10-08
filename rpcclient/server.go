package rpcclient

import "github.com/packet-guardian/pg-dhcp/stats"

type ServerRequest struct {
	client *Client
}

func (s *ServerRequest) GetPoolStats() ([]*stats.PoolStat, error) {
	var reply []*stats.PoolStat
	if err := s.client.c.Call("Server.GetPoolStats", 0, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}
