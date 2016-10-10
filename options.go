// This source file is part of the PG-DHCP project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dhcp

import "github.com/onesimus-systems/dhcp4"

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
