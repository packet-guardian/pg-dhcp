package management

import "testing"

func TestNetworkListRPC(t *testing.T) {
	_, s := setUpTest(t)
	defer tearDownStore(s)

	n := new(Network)
	list := n.GetNameList()

	expected := []string{"network1", "network2", "network3", "network4"}

	if !stringSliceEqual(expected, list) {
		t.Fatalf("Network list wrong. Expected %#v, got %#v", expected, list)
	}
}
