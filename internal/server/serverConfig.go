// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"github.com/lfkeitel/verbose"
	"github.com/packet-guardian/pg-dhcp/store"
)

type Environment string

const (
	EnvTesting Environment = "testing"
	EnvDev     Environment = "dev"
	EnvProd    Environment = "prod"
)

type ServerConfig struct {
	Env            Environment
	Log            *verbose.Logger
	Store          store.Store
	BlockBlacklist bool
	Workers        int
}

func (s *ServerConfig) IsTesting() bool {
	return (s.Env == EnvTesting)
}

func (s *ServerConfig) IsProd() bool {
	return (s.Env == EnvProd)
}

func (s *ServerConfig) IsDev() bool {
	return (s.Env == EnvDev)
}
