package webanalyzer

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

type TokenType string

const (
	NOT      TokenType = "not"
	AND      TokenType = "and"
	OR       TokenType = "or"
	LP       TokenType = "("
	RP       TokenType = ")"
	VARIABLE TokenType = "variable"
	EOF      TokenType = "EOF"
)

const (
	AllowChars = "abcdefghijklmnopqrstuvwxyz0123456789_"
)

type Token struct {
	Type  TokenType
	Name  string
	Value bool
}

type result struct {
	Name  string
	Value bool
}

type Parser struct {
	cond       string
	symbolTab  map[string]bool
	index      int
	backTokens []*Token
}

func (p *Parser) getToken() (*Token, error) {
	for p.index < len(p.cond) {
		switch {
		case p.cond[p.index] == ' ' || p.cond[p.index] == '\t':
			p.index += 1
			continue
		case p.index + 2 < len(p.cond) &&
			p.cond[p.index:p.index+2] == "or" &&
			!strings.ContainsRune(AllowChars, rune(p.cond[p.index+2])):
			p.index += 2
			return &Token{Type: OR}, nil
		case p.index + 3 < len(p.cond) &&
			p.cond[p.index:p.index+3] == "and" &&
			!strings.ContainsRune(AllowChars, rune(p.cond[p.index+3])):
			p.index += 3
			return &Token{Type: AND}, nil
		case p.index + 3 < len(p.cond) &&
			p.cond[p.index:p.index+3] == "not" &&
			!strings.ContainsRune(AllowChars, rune(p.cond[p.index+3])):
			p.index += 3
			return &Token{Type: NOT}, nil
		case p.cond[p.index] == '(':
			p.index += 1
			return &Token{Type: LP}, nil
		case p.cond[p.index] == ')':
			p.index += 1
			return &Token{Type: RP}, nil
		default:
			nameStart := p.index
			for p.index < len(p.cond) && strings.ContainsRune(AllowChars, rune(p.cond[p.index])) {
				p.index += 1
			}

			name := p.cond[nameStart:p.index]
			if v, ok := p.symbolTab[name]; ok {
				return &Token{Type: VARIABLE, Name: name, Value: v}, nil
			}

			return nil, errors.New(fmt.Sprintf("%s does not exists", name))
		}
	}
	return &Token{Type: EOF}, nil
}

func (p *Parser) popToken() (*Token, error) {
	if len(p.backTokens) > 0 {
		return p.backTokens[0], nil
	}

	return p.getToken()
}

func (p *Parser) pushToken(token *Token) {
	p.backTokens = append(p.backTokens, token)
}

func (p *Parser) parseVarExpression() (*result, error) {
	token, err := p.popToken()
	if err != nil {
		return nil, err
	}

	switch token.Type {
	case EOF:
		return nil, nil
	case VARIABLE:
		return &result{Name: token.Name, Value: token.Value}, nil
	}

	return nil, errors.New(fmt.Sprintf("invalid condition, expect VARIABLE, got %s(%s)", token.Name, token.Type))
}

func (p *Parser) parsePrimaryExpression() (*result, error) {
	token, err := p.popToken()
	if err != nil {
		return nil, err
	}

	switch {
	case token.Type == EOF:
		return nil, nil
	case token.Type != LP:
		p.pushToken(token)
		return p.parseVarExpression()
	}

	r, err := p.parseExpression()
	switch {
	case err != nil:
		return nil, err
	case r == nil:
		return nil, nil
	}

	token, err = p.popToken()
	if err != nil {
		return nil, err
	}
	if token.Type != RP {
		return nil, errors.New(fmt.Sprintf("invalid condition, expect RP, got %s(%s)", token.Name, token.Type))
	}

	return &result{Name: fmt.Sprintf("(%s)", token.Name), Value: token.Value}, nil
}

func (p *Parser) parseNotExpression() (*result, error) {
	token, err := p.popToken()
	if err != nil {
		return nil, err
	}
	switch {
	case token.Type == EOF:
		return nil, nil
	case token.Type != NOT:
		p.pushToken(token)
		return p.parsePrimaryExpression()
	}

	r1, err := p.parseNotExpression()
	switch {
	case err != nil:
		return nil, err
	case r1 == nil:
		return nil, nil
	}

	return &result{Name: fmt.Sprintf("not %s", r1.Name), Value: !r1.Value}, nil
}

func (p *Parser) parseAndExpression() (*result, error) {
	r1, err := p.parseNotExpression()
	switch {
	case err != nil:
		return nil, err
	case r1 == nil:
		return nil, nil
	}

	for {
		token, err := p.popToken()
		if err != nil {
			return nil, err
		}
		switch {
		case token.Type == EOF:
			return r1, nil
		case token.Type != AND:
			p.pushToken(token)
			return r1, nil
		}

		r2, err := p.parseNotExpression()
		switch {
		case err != nil:
			return nil, err
		case r2 == nil:
			return nil, nil
		}

		r1 = &result{Name: fmt.Sprintf("%s and %s", r1.Name, r2.Name), Value: r1.Value && r2.Value}
	}
}

func (p *Parser) parseOrExpression() (*result, error) {
	r1, err := p.parseAndExpression()
	switch {
	case err != nil:
		return nil, err
	case r1 == nil:
		return nil, nil
	}

	for {
		token, err := p.popToken()
		if err != nil {
			return nil, err
		}
		switch {
		case token.Type == EOF:
			return r1, nil
		case token.Type != OR:
			p.pushToken(token)
			return r1, nil
		}

		r2, err := p.parseAndExpression()
		switch {
		case err != nil:
			return nil, err
		case r2 == nil:
			return nil, nil
		}

		r1 = &result{Name: fmt.Sprintf("%s or %s", r1.Name, r2.Name), Value: r1.Value || r2.Value}
	}
}

func (p *Parser) parseExpression() (*result, error) {
	return p.parseOrExpression()
}

func (p *Parser) Parse(cond string, symbolTab map[string]bool) (bool, error) {
	p.cond = cond
	p.symbolTab = symbolTab
	p.index = 0

	result, err := p.parseExpression()
	if err != nil {
		return false, err
	}

	if result == nil {
		return false, errors.New("invalid condition")
	}

	return result.Value, nil
}
