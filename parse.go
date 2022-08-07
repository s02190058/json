package json

import (
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

type Parser struct {
	scanner scanner.Scanner
	token   rune
}

func NewParser(input string) *Parser {
	p := new(Parser)
	p.scanner.Init(strings.NewReader(input))
	p.scanner.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats | scanner.ScanStrings

	return p
}

func (p *Parser) next() {
	p.token = p.scanner.Scan()
}

func (p *Parser) text() string {
	return p.scanner.TokenText()
}

func (p *Parser) describe() string {
	switch p.token {
	case scanner.EOF:
		return "end of file"
	case scanner.Ident:
		return fmt.Sprintf("identidfier %s", p.text())
	case scanner.Int, scanner.Float:
		return fmt.Sprintf("number %s", p.text())
	case scanner.String:
		return fmt.Sprintf("string %s", p.text())
	}

	return fmt.Sprintf("%q", p.token)
}

func addParserPrefixToError(err error) error {
	if err != nil {
		err = fmt.Errorf("parser: %s", err.Error())
	}

	return err
}

func (p *Parser) Parse() (Decoder, error) {
	value, err := p.parse()
	err = addParserPrefixToError(err)
	return Decoder{value}, err
}

func (p *Parser) parse() (element, error) {
	p.next()
	value, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	if p.token != scanner.EOF {
		return nil, fmt.Errorf("unexpected token %s", p.describe())
	}

	return value, nil
}

func (p *Parser) parsePrimary() (element, error) {
	switch p.token {
	case '{':
		p.next()
		return p.parseObject()

	case '[':
		p.next()
		return p.parseArray()
	}

	return nil, fmt.Errorf("unexpected %s", p.describe())
}

func (p *Parser) parseObject() (element, error) {
	pairs := make(map[string]element)
	if p.token != '}' {
		for {
			if p.token != scanner.String {
				return nil, fmt.Errorf("expected string, got %s", p.describe())
			}
			key := strings.TrimSuffix(strings.TrimPrefix(p.text(), "\""), "\"")

			p.next()
			if p.token != ':' {
				return nil, fmt.Errorf("expected ':', got %s", p.describe())
			}

			p.next()
			value, err := p.parseValue()
			if err != nil {
				return nil, err
			}
			pairs[key] = value

			if p.token != ',' {
				break
			}

			p.next()
		}
	}

	if p.token != '}' {
		return nil, fmt.Errorf("expected '}', got %s", p.describe())
	}

	p.next()
	return tObject(pairs), nil
}

func (p *Parser) parseArray() (element, error) {
	values := make([]element, 0)
	if p.token != ']' {
		for {
			value, err := p.parseValue()
			if err != nil {
				return nil, err
			}
			values = append(values, value)

			if p.token != ',' {
				break
			}

			p.next()
		}
	}

	if p.token != ']' {
		return nil, fmt.Errorf("expected ']', got %s", p.describe())
	}

	p.next()
	return tArray(values), nil
}

func (p *Parser) parseValue() (element, error) {
	switch p.token {
	case scanner.Ident:
		switch p.text() {
		case "true":
			p.next()
			return tBoolean(true), nil

		case "false":
			p.next()
			return tBoolean(false), nil

		case "null":
			p.next()
			return tNull("null"), nil
		}

		return nil, fmt.Errorf("unexpected %s", p.describe())

	case scanner.Int, scanner.Float:
		n, _ := strconv.ParseFloat(p.text(), 64)
		p.next()
		return tNumber(n), nil

	case scanner.String:
		s := strings.TrimSuffix(strings.TrimPrefix(p.text(), "\""), "\"")
		p.next()
		return tString(s), nil

	case '{':
		p.next()
		return p.parseObject()

	case '[':
		p.next()
		return p.parseArray()
	}

	return nil, fmt.Errorf("unexpected %s", p.describe())
}
