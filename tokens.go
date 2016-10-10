// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import (
	"fmt"
	"strconv"

	"github.com/onesimus-systems/dhcp4"
)

type token int

type lexToken struct {
	token     token
	value     interface{}
	line, pos int
}

const (
	ILLEGAL token = iota
	EOF
	COMMENT

	literal_beg
	NUMBER
	STRING
	IP_ADDRESS
	literal_end

	keyword_beg
	END
	GLOBAL
	NETWORK
	SUBNET
	POOL
	REGISTERED
	UNREGISTERED
	SERVER_IDENTIFIER
	RANGE

	setting_beg
	OPTION
	FREE_LEASE_AFTER
	DEFAULT_LEASE_TIME
	MAX_LEASE_TIME
	setting_end
	keyword_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	COMMENT: "COMMENT",

	NUMBER:     "NUMBER",
	STRING:     "STRING",
	IP_ADDRESS: "IP_ADDRESS",

	END:               "end",
	GLOBAL:            "global",
	NETWORK:           "network",
	SUBNET:            "subnet",
	POOL:              "pool",
	REGISTERED:        "registered",
	UNREGISTERED:      "unregistered",
	SERVER_IDENTIFIER: "server-identifier",
	RANGE:             "range",

	OPTION:             "option",
	FREE_LEASE_AFTER:   "free-lease-after",
	DEFAULT_LEASE_TIME: "default-lease-time",
	MAX_LEASE_TIME:     "max-lease-time",
}

var keywords map[string]token

func (tok token) string() string {
	s := ""
	if 0 <= tok && tok < token(len(tokens)) {
		s = tokens[tok]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(tok)) + ")"
	}
	return s
}

func (tok *lexToken) string() string {
	return fmt.Sprintf("%s: %v", tok.token.string(), tok.value)
}

func init() {
	keywords = make(map[string]token)
	for i := keyword_beg + 1; i < keyword_end-1; i++ {
		keywords[tokens[i]] = i
	}
}

func lookup(ident string) token {
	if tok, valid := keywords[ident]; valid {
		return tok
	}
	return STRING
}

func (tok token) isSetting() bool { return setting_beg < tok && tok < setting_end }

var options = map[string]*dhcpOptionBlock{
	"subnet-mask": &dhcpOptionBlock{
		code:   dhcp4.OptionSubnetMask,
		schema: &optionSchema{token: IP_ADDRESS, multi: 1},
	},
	"router": &dhcpOptionBlock{
		code:   dhcp4.OptionRouter,
		schema: &optionSchema{token: IP_ADDRESS, multi: oneOrMore},
	},
	"domain-name-server": &dhcpOptionBlock{
		code:   dhcp4.OptionDomainNameServer,
		schema: &optionSchema{token: IP_ADDRESS, multi: oneOrMore},
	},
	"domain-name": &dhcpOptionBlock{
		code:   dhcp4.OptionDomainName,
		schema: &optionSchema{token: STRING, multi: 1},
	},
	"broadcast-address": &dhcpOptionBlock{
		code:   dhcp4.OptionBroadcastAddress,
		schema: &optionSchema{token: IP_ADDRESS, multi: 1},
	},
	"network-time-protocol-servers": &dhcpOptionBlock{
		code:   dhcp4.OptionNetworkTimeProtocolServers,
		schema: &optionSchema{token: IP_ADDRESS, multi: oneOrMore},
	},
}

type multiple int

const (
	oneOrMore multiple = -1
)

type optionSchema struct {
	token token
	multi multiple
}

type dhcpOptionBlock struct {
	code   dhcp4.OptionCode
	schema *optionSchema
}
