package dhcp

import (
	"bufio"
	"bytes"
	"net"
	"strconv"
	"unicode"
)

type lexer struct {
	line int
}

func (l *lexer) lex(r *bufio.Reader) []*lexToken {
	tokens := make([]*lexToken, 0)
	l.line = 1

	for {
		c, err := r.ReadByte()
		if err != nil {
			break
		}
		var tok []*lexToken

		if c == '"' {
			tok = l.consumeString(r) // Start after double quote
		} else if isNumber(c) {
			r.UnreadByte()
			tok = l.consumeNumeric(r)
		} else if c == '\n' {
			l.line++
			continue
		} else if c == '#' {
			line := l.consumeLine(r)
			tok = []*lexToken{
				&lexToken{
					token: COMMENT,
					value: string(line),
				},
			}
		} else if isLetter(c) {
			r.UnreadByte()
			tok = l.consumeIdent(r)
		} else if isWhitespace(c) {
			continue
		} else {
			continue
		}

		for _, t := range tok {
			t.line = l.line
			//fmt.Println(t.String())
		}
		tokens = append(tokens, tok...)
	}
	return tokens
}

func (l *lexer) consumeString(r *bufio.Reader) []*lexToken {
	buf := bytes.Buffer{}
	for {
		b, err := r.ReadByte()
		if err != nil {
			return nil
		}
		if b == '"' {
			break
		}
		buf.WriteByte(b)
	}
	return []*lexToken{&lexToken{token: STRING, value: buf.String()}}
}

func (l *lexer) consumeLine(r *bufio.Reader) []byte {
	buf := bytes.Buffer{}
	for {
		b, err := r.ReadByte()
		if err != nil {
			return nil
		}
		if b == '\n' {
			r.UnreadByte()
			break
		}
		buf.WriteByte(b)
	}
	return buf.Bytes()
}

func (l *lexer) consumeNumeric(r *bufio.Reader) []*lexToken {
	buf := bytes.Buffer{}
	dotCount := 0
	hasSlash := false

	for {
		b, err := r.ReadByte()
		if err != nil {
			return nil
		}
		if isNumber(b) {
			buf.WriteByte(b)
			continue
		} else if b == '.' {
			buf.WriteByte(b)
			dotCount++
			continue
		} else if b == '/' {
			buf.WriteByte(b)
			hasSlash = true
			continue
		}
		r.UnreadByte()
		break
	}

	toks := make([]*lexToken, 1)
	toks[0] = &lexToken{}
	if hasSlash && dotCount == 3 { // CIDR notation
		ip, network, err := net.ParseCIDR(buf.String())
		if err != nil {
			toks[0].token = ILLEGAL
		} else {
			toks[0].token = IP_ADDRESS
			toks[0].value = ip
			t := &lexToken{
				token: IP_ADDRESS,
				value: net.IP(network.Mask),
			}
			toks = append(toks, t)
		}
	} else if dotCount == 3 { // IP Address
		ip := net.ParseIP(buf.String())
		if ip == nil {
			toks[0].token = ILLEGAL
		} else {
			toks[0].token = IP_ADDRESS
			toks[0].value = ip
		}
	} else if dotCount == 0 { // Number
		num, err := strconv.Atoi(buf.String())
		if err != nil {
			toks[0].token = ILLEGAL
		} else {
			toks[0].token = NUMBER
			toks[0].value = num
		}
	}
	return toks
}

func (l *lexer) consumeIdent(r *bufio.Reader) []*lexToken {
	buf := bytes.Buffer{}
	for {
		b, err := r.ReadByte()
		if err != nil {
			return nil
		}
		if isWhitespace(b) {
			r.UnreadByte()
			break
		}
		buf.WriteByte(b)
	}
	tok := &lexToken{token: lookup(buf.String()), value: buf.String()}
	return []*lexToken{tok}
}

func isNumber(b byte) bool     { return unicode.IsDigit(rune(b)) }
func isLetter(b byte) bool     { return unicode.IsLetter(rune(b)) }
func isWhitespace(b byte) bool { return unicode.IsSpace(rune(b)) }
