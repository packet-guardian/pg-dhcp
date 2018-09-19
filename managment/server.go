package management

import (
	"runtime"
	"time"

	"github.com/packet-guardian/pg-dhcp/internal/server"
	"github.com/packet-guardian/pg-dhcp/stats"
)

type Server int

func (s *Server) GetPoolStats(_ int, reply *[]*stats.PoolStat) error {
	*reply = server.GetPoolStats()
	return nil
}

func (s *Server) MemStatus(_ int, reply *stats.StatusResp) error {
	*reply = stats.StatusResp{
		GoRoutines: goRoutineStatus(),
		Memory:     memoryStatus(),
	}
	return nil
}

func goRoutineStatus() *stats.GoRoutineStatusResp {
	return &stats.GoRoutineStatusResp{
		RoutineNum: runtime.NumGoroutine(),
	}
}

func memoryStatus() *stats.MemoryStatusResp {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	return &stats.MemoryStatusResp{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		Mallocs:      m.Mallocs,
		Frees:        m.Frees,
		PauseTotalNs: m.PauseTotalNs,
		NumGC:        m.NumGC,
		HeapObjects:  m.HeapObjects,
		LastGC:       time.Unix(0, int64(m.LastGC)).Format(time.RFC3339),
	}
}
