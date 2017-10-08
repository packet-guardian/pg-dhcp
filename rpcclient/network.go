package rpcclient

type NetworkRequest struct {
	client *Client
}

func (n *NetworkRequest) GetNameList() ([]string, error) {
	var reply []string
	if err := n.client.c.Call("Network.GetNameList", 0, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}
