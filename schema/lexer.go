package schema

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

type lexer struct {
	reader *bufio.Reader
	line   int
	column int
	peeked *Token
}

func newLexer(r io.Reader) *lexer {
	return &lexer{
		reader: bufio.NewReader(r),
		line:   1,
		column: 1,
	}
}

func (l *lexer) next() (Token, error) {
	if l.peeked != nil {
		t := *l.peeked
		l.peeked = nil
		return t, nil
	}
	return l.scan()
}

func (l *lexer) peek() (Token, error) {
	if l.peeked != nil {
		return *l.peeked, nil
	}
	t, err := l.scan()
	if err != nil {
		return Token{}, err
	}
	l.peeked = &t
	return t, nil
}

func (l *lexer) scan() (Token, error) {
	for {
		ch, err := l.readRune()
		if err == io.EOF {
			return Token{Kind: TokenEOF, Line: l.line, Column: l.column}, nil
		}
		if err != nil {
			return Token{}, err
		}

		switch {
		case unicode.IsSpace(ch):
			continue
		case ch == '-' && l.matchNext('-'):
			if err := l.skipLineComment(); err != nil {
				return Token{}, err
			}
			continue
		case ch == '{':
			return l.makeToken(TokenLBrace, "{"), nil
		case ch == '}':
			return l.makeToken(TokenRBrace, "}"), nil
		case ch == '[':
			return l.makeToken(TokenLBracket, "["), nil
		case ch == ']':
			return l.makeToken(TokenRBracket, "]"), nil
		case ch == '(':
			return l.makeToken(TokenLParen, "("), nil
		case ch == ')':
			return l.makeToken(TokenRParen, ")"), nil
		case ch == ',':
			return l.makeToken(TokenComma, ","), nil
		case ch == ':':
			if l.matchNext(':') {
				if !l.matchNext('=') {
					return Token{}, l.syntaxError("expected '=' after '::'")
				}
				return l.makeToken(TokenColonColonEquals, "::="), nil
			}
			return Token{}, l.syntaxError("unexpected ':'")
		case ch == '@':
			return l.makeToken(TokenAt, "@"), nil
		case ch == '&':
			return l.makeToken(TokenAmpersand, "&"), nil
		case ch == '.':
			if l.matchNext('.') {
				if l.matchNext('.') {
					return l.makeToken(TokenEllipsis, "..."), nil
				}
				return l.makeToken(TokenDotDot, ".."), nil
			}
			return l.makeToken(TokenDot, "."), nil
		case ch == '"':
			return l.scanString(ch)
		case unicode.IsDigit(ch):
			return l.scanNumber(ch)
		default:
			if isIdentStart(ch) {
				return l.scanIdent(ch)
			}
			return Token{}, l.syntaxError(fmt.Sprintf("unexpected character %q", ch))
		}
	}
}

func (l *lexer) scanIdent(first rune) (Token, error) {
	var b strings.Builder
	b.WriteRune(first)
	for {
		ch, err := l.readRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Token{}, err
		}
		if !isIdentPart(ch) {
			l.unreadRune(ch)
			break
		}
		b.WriteRune(ch)
	}

	word := b.String()
	kind := keywordKind(word)
	if kind == TokenIdent {
		return l.makeToken(TokenIdent, word), nil
	}
	return l.makeToken(kind, word), nil
}

func (l *lexer) scanNumber(first rune) (Token, error) {
	var b strings.Builder
	b.WriteRune(first)
	for {
		ch, err := l.readRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Token{}, err
		}
		if !unicode.IsDigit(ch) {
			l.unreadRune(ch)
			break
		}
		b.WriteRune(ch)
	}
	return l.makeToken(TokenNumber, b.String()), nil
}

func (l *lexer) scanString(open rune) (Token, error) {
	var b strings.Builder
	for {
		ch, err := l.readRune()
		if err == io.EOF {
			return Token{}, l.syntaxError("unterminated string")
		}
		if err != nil {
			return Token{}, err
		}
		if ch == open {
			break
		}
		b.WriteRune(ch)
	}
	return l.makeToken(TokenString, b.String()), nil
}

func (l *lexer) skipLineComment() error {
	for {
		ch, err := l.readRune()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if ch == '\n' {
			return nil
		}
	}
}

func (l *lexer) makeToken(kind TokenKind, value string) Token {
	return Token{
		Kind:   kind,
		Value:  value,
		Line:   l.line,
		Column: l.column,
	}
}

func (l *lexer) readRune() (rune, error) {
	ch, size, err := l.reader.ReadRune()
	if err != nil {
		return 0, err
	}
	if ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column += size
	}
	return ch, nil
}

func (l *lexer) unreadRune(ch rune) {
	_ = l.reader.UnreadRune()
	if ch == '\n' {
		l.line--
	} else {
		size := utf8.RuneLen(ch)
		if size > 0 {
			l.column -= size
		}
	}
}

func (l *lexer) matchNext(expected rune) bool {
	ch, err := l.readRune()
	if err != nil {
		return false
	}
	if ch != expected {
		l.unreadRune(ch)
		return false
	}
	return true
}

func (l *lexer) syntaxError(msg string) error {
	return &SyntaxError{
		Line:   l.line,
		Column: l.column,
		Msg:    msg,
	}
}

func isIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func isIdentPart(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '-' || ch == '_'
}

func keywordKind(word string) TokenKind {
	switch strings.ToUpper(word) {
	case "DEFINITIONS":
		return TokenDEFINITIONS
	case "BEGIN":
		return TokenBEGIN
	case "END":
		return TokenEND
	case "OPTIONAL":
		return TokenOPTIONAL
	case "DEFAULT":
		return TokenDEFAULT
	case "IMPLICIT":
		return TokenIMPLICIT
	case "EXPLICIT":
		return TokenEXPLICIT
	case "TRUE":
		return TokenTRUE
	case "FALSE":
		return TokenFALSE
	case "NULL":
		return TokenNULL
	case "OF":
		return TokenOF
	case "SEQUENCE":
		return TokenSEQUENCE
	case "SET":
		return TokenSET
	case "CHOICE":
		return TokenCHOICE
	case "INTEGER":
		return TokenINTEGER
	case "BOOLEAN":
		return TokenBOOLEAN
	case "ENUMERATED":
		return TokenENUMERATED
	case "REAL":
		return TokenREAL
	case "BIT":
		return TokenBIT
	case "OCTET":
		return TokenOCTET
	case "STRING":
		return TokenSTRING
	case "OBJECT":
		return TokenOBJECT
	case "IDENTIFIER":
		return TokenIDENTIFIER
	case "EXTERNAL":
		return TokenEXTERNAL
	case "EMBEDDED":
		return TokenEMBEDDED
	case "PDV":
		return TokenPDV
	case "UTCTIME":
		return TokenUTCTime
	case "GENERALIZEDTIME":
		return TokenGeneralizedTime
	case "SIZE":
		return TokenSIZE
	case "FROM":
		return TokenFROM
	case "MIN":
		return TokenMIN
	case "MAX":
		return TokenMAX
	default:
		return TokenIdent
	}
}
