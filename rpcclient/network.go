package rpcclient

type NetworkRPCRequest struct {
	client *RPCClient
}

func (n *NetworkRPCRequest) GetNameList() ([]string, error) {
	var reply []string
	if err := n.client.call("Network.GetNameList", 0, &reply); err != nil {
		return nil, err
	}
	return reply, nil
}
