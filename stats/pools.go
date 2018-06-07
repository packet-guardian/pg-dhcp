package stats

type PoolStat struct {
	NetworkName, Subnet                     string
	Start                                   string
	End                                     string
	Registered                              bool
	Total, Active, Claimed, Abandoned, Free int
}
