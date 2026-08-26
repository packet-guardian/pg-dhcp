package stats

type StatusResp struct {
	GoRoutines *GoRoutineStatusResp `json:"go_routines"`
	Memory     *MemoryStatusResp    `json:"memory"`
}

type GoRoutineStatusResp struct {
	RoutineNum int `json:"routine_num"`
}

type MemoryStatusResp struct {
	Alloc        uint64 `json:"alloc"`
	TotalAlloc   uint64 `json:"total_alloc"`
	Sys          uint64 `json:"sys"`
	Mallocs      uint64 `json:"mallocs"`
	Frees        uint64 `json:"frees"`
	PauseTotalNs uint64 `json:"pause_total_ns"`
	NumGC        uint32 `json:"num_gc"`
	HeapObjects  uint64 `json:"head_objects"`
	LastGC       string `json:"last_gc"`
}
