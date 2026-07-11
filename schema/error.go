package schema

import "fmt"

// SyntaxError is returned when an ASN.1 schema file is syntactically invalid.
type SyntaxError struct {
	Line   int
	Column int
	Msg    string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("asn.1 schema syntax error at %d:%d: %s", e.Line, e.Column, e.Msg)
}

// ParseError wraps a syntax error with optional context.
type ParseError struct {
	Token Token
	Msg   string
}

func (e *ParseError) Error() string {
	if e.Token.Kind == TokenEOF {
		return fmt.Sprintf("asn.1 schema parse error at %d:%d: %s (got EOF)", e.Token.Line, e.Token.Column, e.Msg)
	}
	if e.Token.Value != "" {
		return fmt.Sprintf("asn.1 schema parse error at %d:%d: %s (got %q)", e.Token.Line, e.Token.Column, e.Msg, e.Token.Value)
	}
	return fmt.Sprintf("asn.1 schema parse error at %d:%d: %s (got %s)", e.Token.Line, e.Token.Column, e.Msg, e.Token.Kind)
}

func (e *ParseError) Unwrap() error {
	return &SyntaxError{Line: e.Token.Line, Column: e.Token.Column, Msg: e.Msg}
}
