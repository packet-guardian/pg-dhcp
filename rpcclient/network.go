package rpcclient

type NetworkRequest struct {
	client *Client
}

func (n *NetworkRequest) GetNameList() ([]string, error) {
	reply := make([]string, 0)
	if err := n.client.c.Call("Network.GetNameList", nil, reply); err != nil {
		return nil, err
	}
	return reply, nil
}
