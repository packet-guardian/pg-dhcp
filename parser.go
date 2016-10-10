package dhcp

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/onesimus-systems/dhcp4"
)

// ParseFile takes the file name to a configuration file.
// It will attempt to parse the file using the PG-DHCP configuration
// format. If an error occures config will be nil.
func ParseFile(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return newParser(bufio.NewReader(file)).parse()
}

// type parseError struct {
// 	line    int
// 	message string
// }
//
// func newError(line int, message string, v ...interface{}) parseError {
// 	return parseError{
// 		line:    line,
// 		message: fmt.Sprintf(message, v...),
// 	}
// }
//
// func (p parseError) Error() string {
// 	return fmt.Sprintf("Error %s on line %d", p.message, p.line)
// }

type parser struct {
	r *bufio.Reader
	c *Config
}

func newParser(r *bufio.Reader) *parser {
	return &parser{r: r}
}

func (p *parser) parse() (*Config, error) {
	p.c = newConfig()
	toks := (&lexer{}).lex(p.r)

	for i := 0; i < len(toks); i++ {
		tok := toks[i]
		var err error
		var next int
		switch tok.token {
		case COMMENT:
			continue
		case GLOBAL:
			next, err = p.parseGlobal(toks, i+1)
		case NETWORK:
			next, err = p.parseNetwork(toks, i+1)
		default:
			return nil, fmt.Errorf("Invalid token on line %d: %s", tok.line, tok.string())
		}
		if err != nil {
			return nil, err
		}
		i = next - 1
	}

	for _, n := range p.c.networks {
		n.global = p.c.global
	}
	return p.c, nil
}

func (p *parser) parseGlobal(toks []*lexToken, start int) (int, error) {
	var i int
mainLoop:
	for i = start; i < len(toks); i++ {
		tok := toks[i]
		switch tok.token {
		case COMMENT:
			continue
		case SERVER_IDENTIFIER:
			if i+1 > len(toks) || toks[i+1].token != IP_ADDRESS {
				return 0, fmt.Errorf("Expected IP address on line %d", toks[i+1].line)
			}
			p.c.global.serverIdentifier = toks[i+1].value.(net.IP)
			i++
		case REGISTERED:
			s, next, err := p.parseSettingsBlock(toks, i+1)
			if err != nil {
				return 0, err
			}
			p.c.global.registeredSettings = s
			i = next
		case UNREGISTERED:
			s, next, err := p.parseSettingsBlock(toks, i+1)
			if err != nil {
				return 0, err
			}
			p.c.global.unregisteredSettings = s
			i = next
		case END:
			break mainLoop
		default:
			if tok.token.isSetting() {
				next, err := p.parseSetting(toks, i, p.c.global.settings)
				if err != nil {
					return 0, err
				}
				i = next - 1
				continue
			}
			return 0, fmt.Errorf("Unexpected token %s on line %d in global", tok.string(), tok.line)
		}
	}
	return i + 1, nil
}

func (p *parser) parseNetwork(toks []*lexToken, start int) (int, error) {
	nameToken := toks[start]
	if nameToken.token != STRING {
		return 0, fmt.Errorf("Expected STRING on line %d", nameToken.line)
	}
	name := nameToken.value.(string)

	if _, exists := p.c.networks[name]; exists {
		return 0, fmt.Errorf("Network %s already declared, line %d", name, nameToken.line)
	}
	netBlock := newNetwork(name)
	start++   // Skip name token
	mode := 0 // 0 = root, 1 = registered, 2 = unregistered
	var i int
mainLoop:
	for i = start; i < len(toks); i++ {
		tok := toks[i]
		switch tok.token {
		case COMMENT:
			continue
		case SUBNET:
			shortSyntax := false
			if mode == 0 {
				mode = 2
				shortSyntax = true
			}
			subnet, next, err := p.parseSubnet(toks, i+1)
			if err != nil {
				return 0, err
			}
			if mode == 2 { // Unregistered block
				subnet.allowUnknown = true
			}
			subnet.network = netBlock
			netBlock.subnets = append(netBlock.subnets, subnet)
			i = next - 1
			if shortSyntax {
				mode = 0
			}
		case REGISTERED:
			if mode == 0 {
				mode = 1
				continue
			}
			return 0, fmt.Errorf("Registered block not allowed on line %d", tok.line)
		case UNREGISTERED:
			if mode == 0 {
				mode = 2
				continue
			}
			return 0, fmt.Errorf("Unregistered block not allowed on line %d", tok.line)
		case END:
			if mode == 0 { // Exit from root network block
				break mainLoop
			} else { // Exit from un/registered block
				mode = 0
			}
		default:
			if tok.token.isSetting() {
				block := netBlock.settings
				if mode == 1 {
					block = netBlock.registeredSettings
				} else if mode == 2 {
					block = netBlock.unregisteredSettings
				}
				next, err := p.parseSetting(toks, i, block)
				if err != nil {
					return 0, err
				}
				i = next - 1
				continue
			}
			return 0, fmt.Errorf("Unexpected token %s on line %d in network", tok.string(), tok.line)
		}
	}
	p.c.networks[name] = netBlock
	return i + 1, nil
}

func (p *parser) parseSubnet(toks []*lexToken, start int) (*subnet, int, error) {
	if start+2 > len(toks) {
		return nil, 0, errors.New("Unexpected end of file")
	}
	ipAddr := toks[start]
	netmask := toks[start+1]
	start += 2
	if ipAddr.token != IP_ADDRESS {
		return nil, 0, fmt.Errorf("Expected IP address on line %d", ipAddr.line)
	}
	if netmask.token != IP_ADDRESS {
		return nil, 0, fmt.Errorf("Expected IP address on line %d", netmask.line)
	}
	sub := newSubnet()
	sub.net = &net.IPNet{
		IP:   ipAddr.value.(net.IP),
		Mask: net.IPMask(netmask.value.(net.IP)),
	}

	var i int
mainLoop:
	for i = start; i < len(toks); i++ {
		tok := toks[i]
		switch tok.token {
		case COMMENT:
			continue
		case POOL:
			subPool, next, err := p.parsePool(toks, i+1)
			if err != nil {
				return nil, 0, err
			}
			subPool.subnet = sub
			sub.pools = append(sub.pools, subPool)
			i = next - 1
		case RANGE:
			subPool, next, err := p.parsePool(toks, i) // Start with range statement
			if err != nil {
				return nil, 0, err
			}
			subPool.subnet = sub
			sub.pools = append(sub.pools, subPool)
			i = next - 2 // Get End token again
		case END:
			break mainLoop
		default:
			if tok.token.isSetting() {
				next, err := p.parseSetting(toks, i, sub.settings)
				if err != nil {
					return nil, 0, err
				}
				i = next - 1
				continue
			}
			return nil, 0, fmt.Errorf("Unexpected token %s on line %d in subnet", tok.string(), tok.line)
		}
	}
	if _, ok := sub.settings.options[dhcp4.OptionSubnetMask]; !ok {
		sub.settings.options[dhcp4.OptionSubnetMask] = []byte(sub.net.Mask)
	}
	return sub, i + 1, nil
}

func (p *parser) parsePool(toks []*lexToken, start int) (*pool, int, error) {
	nPool := newPool()

	var i int
mainLoop:
	for i = start; i < len(toks); i++ {
		tok := toks[i]
		switch tok.token {
		case COMMENT:
			continue
		case RANGE:
			if nPool.rangeStart != nil {
				return nil, 0, fmt.Errorf("Range redeclared on line %d", tok.line)
			}
			startIP := toks[i+1]
			endIP := toks[i+2]
			i += 2
			if startIP.token != IP_ADDRESS {
				return nil, 0, fmt.Errorf("Expected IP address on line %d, got %s", startIP.line, startIP.string())
			}
			if endIP.token != IP_ADDRESS {
				return nil, 0, fmt.Errorf("Expected IP address on line %d, got %s", endIP.line, endIP.string())
			}
			nPool.rangeStart = startIP.value.(net.IP)
			nPool.rangeEnd = endIP.value.(net.IP)
		case END:
			break mainLoop
		default:
			if tok.token.isSetting() {
				next, err := p.parseSetting(toks, i, nPool.settings)
				if err != nil {
					return nil, 0, err
				}
				i = next - 1
				continue
			}
			return nil, 0, fmt.Errorf("Unexpected token %s on line %d in pool", tok.string(), tok.line)
		}
	}
	return nPool, i + 1, nil
}

func (p *parser) parseSettingsBlock(toks []*lexToken, start int) (*settings, int, error) {
	s := newSettingsBlock()

	var i int
	for i = start; i < len(toks); i++ {
		if !toks[i].token.isSetting() {
			break
		}
		next, err := p.parseSetting(toks, i, s)
		if err != nil {
			return nil, 0, err
		}
		i = next - 1
	}
	return s, i, nil
}

func (p *parser) parseSetting(toks []*lexToken, start int, setBlock *settings) (int, error) {
	tok := toks[start]

	switch tok.token {
	case COMMENT:
		return start + 1, nil
	case OPTION:
		code, data, next, err := p.parseOption(toks, start+1)
		if err != nil {
			return 0, err
		}
		setBlock.options[code] = data
		return next, nil
	case DEFAULT_LEASE_TIME:
		if start+1 > len(toks) || toks[start+1].token != NUMBER {
			return 0, fmt.Errorf("Expected number on line %d", toks[start+1].line)
		}
		setBlock.defaultLeaseTime = time.Duration(toks[start+1].value.(int)) * time.Second
		return start + 2, nil
	case MAX_LEASE_TIME:
		if start+1 > len(toks) || toks[start+1].token != NUMBER {
			return 0, fmt.Errorf("Expected number on line %d", toks[start+1].line)
		}
		setBlock.maxLeaseTime = time.Duration(toks[start+1].value.(int)) * time.Second
		return start + 2, nil
	case FREE_LEASE_AFTER:
		if start+1 > len(toks) || toks[start+1].token != NUMBER {
			return 0, fmt.Errorf("Expected number on line %d", toks[start+1].line)
		}
		setBlock.freeLeaseAfter = time.Duration(toks[start+1].value.(int)) * time.Second
		return start + 2, nil
	default:
		return 0, fmt.Errorf("Unexpected token %s on line %d in settings", tok.string(), tok.line)
	}

	return start + 1, nil
}

func (p *parser) parseOption(toks []*lexToken, start int) (dhcp4.OptionCode, []byte, int, error) {
	option := toks[start].value.(string)
	block, exists := options[option]
	if !exists {
		return dhcp4.OptionCode(0), nil, 0, fmt.Errorf("Option %s is not supported", option)
	}

	optionData := make([]byte, 0)

	if block.schema.multi == oneOrMore {
		for i := start + 1; i < len(toks); i++ {
			start++
			if toks[i].token != block.schema.token {
				break
			}
			switch t := toks[i].value.(type) {
			case string:
				optionData = append(optionData, []byte(t)...)
			case []byte:
				optionData = append(optionData, t...)
			case net.IP:
				optionData = append(optionData, []byte(t.To4())...)
			}
		}
	} else {
		for i := 0; i < int(block.schema.multi); i++ {
			start++
			tok := toks[start]
			if tok.token != block.schema.token {
				return 0, nil, 0, fmt.Errorf("Expected %s, got %s on line %d", block.schema.token, tok.token, tok.line)
			}
			switch t := tok.value.(type) {
			case string:
				optionData = append(optionData, []byte(t)...)
			case []byte:
				optionData = append(optionData, t...)
			case net.IP:
				optionData = append(optionData, []byte(t.To4())...)
			}
		}
		start++
	}

	return block.code, optionData, start, nil
}
