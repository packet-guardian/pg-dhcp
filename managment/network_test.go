package management

import "testing"

func TestNetworkListRPC(t *testing.T) {
	_, s := setUpTest(t)
	defer tearDownStore(s)

	var list []string
	n := new(Network)
	if err := n.GetNameList(0, &list); err != nil {
		t.Fatal(err)
	}

	expected := []string{"network1", "network2", "network3", "network4"}

	if !stringSliceEqual(expected, list) {
		t.Fatalf("Network list wrong. Expected %#v, got %#v", expected, list)
	}
}
