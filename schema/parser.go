package schema

import (
	"fmt"
	"io"
)

// Parse reads and parses an ASN.1 schema module from r.
func Parse(r io.Reader) (*Schema, error) {
	p := newParser(newLexer(r))
	return p.parseModule()
}

type parser struct {
	lex        *lexer
	curr       Token
	peek       Token
	hasPK      bool
	tagDefault TagDefault
}

func newParser(l *lexer) *parser {
	return &parser{lex: l}
}

func (p *parser) parseModule() (*Schema, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}

	nameTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, err
	}

	schema := &Schema{
		ModuleName: nameTok.Value,
		Types:      make(map[string]Type),
	}

	if p.curr.Kind == TokenLBrace {
		oid, err := p.parseModuleOID()
		if err != nil {
			return nil, err
		}
		schema.ModuleOID = oid
	}

	if err := p.expectKind(TokenDEFINITIONS); err != nil {
		return nil, err
	}
	tagDefault, err := p.parseTagDefault()
	if err != nil {
		return nil, err
	}
	p.tagDefault = tagDefault
	schema.TagDefault = tagDefault

	if err := p.expectKind(TokenColonColonEquals); err != nil {
		return nil, err
	}
	if err := p.expectKind(TokenBEGIN); err != nil {
		return nil, err
	}

	for p.curr.Kind != TokenEND && p.curr.Kind != TokenEOF {
		assign, err := p.parseAssignment()
		if err != nil {
			return nil, err
		}
		if _, exists := schema.Types[assign.Name]; exists {
			return nil, p.error(p.curr, "duplicate type assignment %q", assign.Name)
		}
		schema.Types[assign.Name] = assign.Type
	}

	if err := p.expectKind(TokenEND); err != nil {
		return nil, err
	}
	return schema, nil
}

// parseTagDefault parses an optional IMPLICIT/EXPLICIT/AUTOMATIC TAGS clause.
// ASN.1 default when omitted is EXPLICIT TAGS.
func (p *parser) parseTagDefault() (TagDefault, error) {
	switch p.curr.Kind {
	case TokenIMPLICIT:
		if err := p.advance(); err != nil {
			return TagDefaultExplicit, err
		}
		if err := p.expectKind(TokenTAGS); err != nil {
			return TagDefaultExplicit, err
		}
		return TagDefaultImplicit, nil
	case TokenEXPLICIT:
		if err := p.advance(); err != nil {
			return TagDefaultExplicit, err
		}
		if err := p.expectKind(TokenTAGS); err != nil {
			return TagDefaultExplicit, err
		}
		return TagDefaultExplicit, nil
	case TokenAUTOMATIC:
		if err := p.advance(); err != nil {
			return TagDefaultExplicit, err
		}
		if err := p.expectKind(TokenTAGS); err != nil {
			return TagDefaultExplicit, err
		}
		return TagDefaultAutomatic, nil
	default:
		return TagDefaultExplicit, nil
	}
}

func (p *parser) parseModuleOID() (string, error) {
	if err := p.expectKind(TokenLBrace); err != nil {
		return "", err
	}

	var parts []string
	for p.curr.Kind != TokenRBrace && p.curr.Kind != TokenEOF {
		switch p.curr.Kind {
		case TokenIdent, TokenNumber:
			parts = append(parts, p.curr.Value)
			if err := p.advance(); err != nil {
				return "", err
			}
		case TokenLParen:
			if err := p.advance(); err != nil {
				return "", err
			}
			num, err := p.expect(TokenNumber)
			if err != nil {
				return "", err
			}
			parts = append(parts, "("+num.Value+")")
			if err := p.expectKind(TokenRParen); err != nil {
				return "", err
			}
		default:
			return "", p.error(p.curr, "expected module OID component")
		}
	}

	if err := p.expectKind(TokenRBrace); err != nil {
		return "", err
	}
	return joinOID(parts), nil
}

func joinOID(parts []string) string {
	out := ""
	for i, part := range parts {
		if i > 0 && !stringsHasPrefixParen(part) && !stringsHasSuffixParen(out) {
			out += " "
		}
		out += part
	}
	return out
}

func stringsHasPrefixParen(s string) bool {
	return len(s) > 0 && s[0] == '('
}

func stringsHasSuffixParen(s string) bool {
	return len(s) > 0 && s[len(s)-1] == ')'
}

func (p *parser) parseAssignment() (Assignment, error) {
	// Type names may collide with constraint keywords (e.g. TAP's Min).
	nameTok, err := p.expectNamedLabel()
	if err != nil {
		return Assignment{}, err
	}
	if err := p.expectKind(TokenColonColonEquals); err != nil {
		return Assignment{}, err
	}
	typ, err := p.parseType()
	if err != nil {
		return Assignment{}, err
	}
	return Assignment{Name: nameTok.Value, Type: typ}, nil
}

func (p *parser) parseType() (Type, error) {
	switch p.curr.Kind {
	case TokenLBracket:
		return p.parseTaggedType()
	case TokenCHOICE:
		return p.parseChoiceType()
	case TokenSEQUENCE:
		if err := p.advance(); err != nil {
			return nil, err
		}
		preConstraints, err := p.parseConstraints()
		if err != nil {
			return nil, err
		}
		if p.curr.Kind == TokenOF {
			if err := p.advance(); err != nil {
				return nil, err
			}
			elem, err := p.parseType()
			if err != nil {
				return nil, err
			}
			postConstraints, err := p.parseConstraints()
			if err != nil {
				return nil, err
			}
			return SequenceOfType{Element: elem, Constraints: append(preConstraints, postConstraints...)}, nil
		}
		components, extensible, err := p.parseComponentList()
		if err != nil {
			return nil, err
		}
		return SequenceType{Components: components, Extensible: extensible, Constraints: preConstraints}, nil
	case TokenSET:
		if err := p.advance(); err != nil {
			return nil, err
		}
		preConstraints, err := p.parseConstraints()
		if err != nil {
			return nil, err
		}
		if p.curr.Kind == TokenOF {
			if err := p.advance(); err != nil {
				return nil, err
			}
			elem, err := p.parseType()
			if err != nil {
				return nil, err
			}
			postConstraints, err := p.parseConstraints()
			if err != nil {
				return nil, err
			}
			return SetOfType{Element: elem, Constraints: append(preConstraints, postConstraints...)}, nil
		}
		components, extensible, err := p.parseComponentList()
		if err != nil {
			return nil, err
		}
		return SetType{Components: components, Extensible: extensible, Constraints: preConstraints}, nil
	case TokenINTEGER:
		return p.parseIntegerType()
	case TokenENUMERATED:
		return p.parseEnumeratedType()
	case TokenREAL:
		return p.parseRealType()
	case TokenBOOLEAN:
		if err := p.advance(); err != nil {
			return nil, err
		}
		return BooleanType{}, nil
	case TokenNULL:
		if err := p.advance(); err != nil {
			return nil, err
		}
		return NullType{}, nil
	case TokenOCTET:
		return p.parseOctetStringType()
	case TokenBIT:
		return p.parseBitStringType()
	case TokenOBJECT:
		return p.parseObjectIdentifierType()
	case TokenEXTERNAL:
		return p.parseExternalType()
	case TokenEMBEDDED:
		return p.parseEmbeddedPDVType()
	case TokenUTCTime:
		return p.parseUTCTimeType()
	case TokenGeneralizedTime:
		return p.parseGeneralizedTimeType()
	case TokenIdent, TokenMIN, TokenMAX:
		// MIN/MAX are constraint keywords but also legal type names (e.g. TAP Min).
		return p.parseReferenceOrBuiltinType()
	default:
		return nil, p.error(p.curr, "expected type")
	}
}

func (p *parser) parseIntegerType() (Type, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	named, ext, err := p.parseNamedNumbers()
	if err != nil {
		return nil, err
	}
	constraints, err := p.parseConstraints()
	if err != nil {
		return nil, err
	}
	return IntegerType{NamedNumbers: named, Extensible: ext, Constraints: constraints}, nil
}

func (p *parser) parseEnumeratedType() (Type, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	values, ext, err := p.parseNamedNumbers()
	if err != nil {
		return nil, err
	}
	constraints, err := p.parseConstraints()
	if err != nil {
		return nil, err
	}
	return EnumeratedType{Values: values, Extensible: ext, Constraints: constraints}, nil
}

func (p *parser) parseRealType() (Type, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	constraints, err := p.parseConstraints()
	if err != nil {
		return nil, err
	}
	return RealType{Constraints: constraints}, nil
}

func (p *parser) parseOctetStringType() (Type, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	if err := p.expectKind(TokenSTRING); err != nil {
		return nil, err
	}
	constraints, err := p.parseConstraints()
	if err != nil {
		return nil, err
	}
	return OctetStringType{Constraints: constraints}, nil
}

func (p *parser) parseBitStringType() (Type, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	if err := p.expectKind(TokenSTRING); err != nil {
		return nil, err
	}
	named, ext, err := p.parseNamedNumbers()
	if err != nil {
		return nil, err
	}
	constraints, err := p.parseConstraints()
	if err != nil {
		return nil, err
	}
	return BitStringType{NamedBits: named, Extensible: ext, Constraints: constraints}, nil
}

func (p *parser) parseObjectIdentifierType() (Type, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	if err := p.expectKind(TokenIDENTIFIER); err != nil {
		return nil, err
	}
	constraints, err := p.parseConstraints()
	if err != nil {
		return nil, err
	}
	return ObjectIdentifierType{Constraints: constraints}, nil
}

func (p *parser) parseExternalType() (Type, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	constraints, err := p.parseConstraints()
	if err != nil {
		return nil, err
	}
	return ExternalType{Constraints: constraints}, nil
}

func (p *parser) parseEmbeddedPDVType() (Type, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	if err := p.expectKind(TokenPDV); err != nil {
		return nil, err
	}
	constraints, err := p.parseConstraints()
	if err != nil {
		return nil, err
	}
	return EmbeddedPDVType{Constraints: constraints}, nil
}

func (p *parser) parseUTCTimeType() (Type, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	constraints, err := p.parseConstraints()
	if err != nil {
		return nil, err
	}
	return UTCTimeType{Constraints: constraints}, nil
}

func (p *parser) parseGeneralizedTimeType() (Type, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	constraints, err := p.parseConstraints()
	if err != nil {
		return nil, err
	}
	return GeneralizedTimeType{Constraints: constraints}, nil
}

func (p *parser) parseReferenceOrBuiltinType() (Type, error) {
	name := p.curr.Value
	if typ, ok := builtinTypeFromName(name); ok {
		if err := p.advance(); err != nil {
			return nil, err
		}
		constraints, err := p.parseConstraints()
		if err != nil {
			return nil, err
		}
		return withConstraints(typ, constraints), nil
	}

	if err := p.advance(); err != nil {
		return nil, err
	}
	if p.curr.Kind == TokenDot {
		if err := p.advance(); err != nil {
			return nil, err
		}
		if p.curr.Kind == TokenAmpersand {
			if err := p.advance(); err != nil {
				return nil, err
			}
			fieldTok, err := p.expectNamedLabel()
			if err != nil {
				return nil, err
			}
			name = name + ".&" + fieldTok.Value
		}
	}
	constraints, err := p.parseConstraints()
	if err != nil {
		return nil, err
	}
	return ReferenceType{Name: name, Constraints: constraints}, nil
}

func (p *parser) parseNamedNumbers() (map[string]int, bool, error) {
	if p.curr.Kind != TokenLBrace {
		return nil, false, nil
	}
	if err := p.advance(); err != nil {
		return nil, false, err
	}

	values := make(map[string]int)
	extensible := false

	for p.curr.Kind != TokenRBrace && p.curr.Kind != TokenEOF {
		if p.curr.Kind == TokenEllipsis {
			extensible = true
			if err := p.advance(); err != nil {
				return nil, false, err
			}
			if p.curr.Kind == TokenComma {
				if err := p.advance(); err != nil {
					return nil, false, err
				}
			}
			continue
		}

		nameTok, err := p.expectNamedLabel()
		if err != nil {
			return nil, false, err
		}
		if err := p.expectKind(TokenLParen); err != nil {
			return nil, false, err
		}
		numTok, err := p.expect(TokenNumber)
		if err != nil {
			return nil, false, err
		}
		num, err := parseInt(numTok.Value)
		if err != nil {
			return nil, false, p.error(numTok, "invalid integer %q", numTok.Value)
		}
		if err := p.expectKind(TokenRParen); err != nil {
			return nil, false, err
		}
		values[nameTok.Value] = num

		if p.curr.Kind == TokenComma {
			if err := p.advance(); err != nil {
				return nil, false, err
			}
		}
	}

	if err := p.expectKind(TokenRBrace); err != nil {
		return nil, false, err
	}
	return values, extensible, nil
}

func (p *parser) parseChoiceType() (Type, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	components, extensible, err := p.parseComponentList()
	if err != nil {
		return nil, err
	}
	return ChoiceType{Components: components, Extensible: extensible}, nil
}

func (p *parser) parseComponentList() ([]Component, bool, error) {
	if err := p.expectKind(TokenLBrace); err != nil {
		return nil, false, err
	}

	var components []Component
	extensible := false

	for p.curr.Kind != TokenRBrace && p.curr.Kind != TokenEOF {
		if p.curr.Kind == TokenEllipsis {
			extensible = true
			if err := p.advance(); err != nil {
				return nil, false, err
			}
			if p.curr.Kind == TokenComma {
				if err := p.advance(); err != nil {
					return nil, false, err
				}
			}
			continue
		}

		comp, err := p.parseComponent()
		if err != nil {
			return nil, false, err
		}
		components = append(components, comp)

		if p.curr.Kind == TokenComma {
			if err := p.advance(); err != nil {
				return nil, false, err
			}
		}
	}

	if err := p.expectKind(TokenRBrace); err != nil {
		return nil, false, err
	}
	return components, extensible, nil
}

func (p *parser) parseTaggedType() (Type, error) {
	tag, err := p.parseTag()
	if err != nil {
		return nil, err
	}
	implicit, err := p.parseTaggingMode(true)
	if err != nil {
		return nil, err
	}
	inner, err := p.parseType()
	if err != nil {
		return nil, err
	}
	// X.680: a tagged CHOICE type uses EXPLICIT tagging.
	if _, ok := inner.(ChoiceType); ok {
		implicit = false
	}
	return TaggedType{Tag: *tag, Implicit: implicit, Type: inner}, nil
}

// parseTaggingMode reads an optional IMPLICIT/EXPLICIT keyword.
// When tagged is true and neither keyword is present, the module TagDefault applies
// (AUTOMATIC and IMPLICIT both yield implicit tagging).
func (p *parser) parseTaggingMode(tagged bool) (implicit bool, err error) {
	switch p.curr.Kind {
	case TokenIMPLICIT:
		if err := p.advance(); err != nil {
			return false, err
		}
		return true, nil
	case TokenEXPLICIT:
		if err := p.advance(); err != nil {
			return false, err
		}
		return false, nil
	default:
		if !tagged {
			return false, nil
		}
		switch p.tagDefault {
		case TagDefaultImplicit, TagDefaultAutomatic:
			return true, nil
		default:
			return false, nil
		}
	}
}

func (p *parser) parseComponent() (Component, error) {
	nameTok, err := p.expectNamedLabel()
	if err != nil {
		return Component{}, err
	}

	comp := Component{Name: nameTok.Value}

	if p.curr.Kind == TokenLBracket {
		tag, err := p.parseTag()
		if err != nil {
			return Component{}, err
		}
		comp.Tag = tag
	}

	implicit, err := p.parseTaggingMode(comp.Tag != nil)
	if err != nil {
		return Component{}, err
	}
	comp.Implicit = implicit

	typ, err := p.parseType()
	if err != nil {
		return Component{}, err
	}
	comp.Type = typ
	// X.680: a tagged CHOICE component uses EXPLICIT tagging.
	if comp.Tag != nil {
		if _, ok := typ.(ChoiceType); ok {
			comp.Implicit = false
		}
	}

	if p.curr.Kind == TokenOPTIONAL {
		comp.Optional = true
		if err := p.advance(); err != nil {
			return Component{}, err
		}
	}

	if p.curr.Kind == TokenDEFAULT {
		if err := p.advance(); err != nil {
			return Component{}, err
		}
		def, err := p.parseDefaultValue()
		if err != nil {
			return Component{}, err
		}
		comp.Default = def
	}

	return comp, nil
}

func (p *parser) parseTag() (*Tag, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}

	tag := &Tag{Class: TagClassContext}

	switch p.curr.Kind {
	case TokenIdent:
		switch p.curr.Value {
		case "UNIVERSAL":
			tag.Class = TagClassUniversal
		case "APPLICATION":
			tag.Class = TagClassApplication
		case "PRIVATE":
			tag.Class = TagClassPrivate
		default:
			return nil, p.error(p.curr, "expected tag class or number")
		}
		if err := p.advance(); err != nil {
			return nil, err
		}
		numTok, err := p.expect(TokenNumber)
		if err != nil {
			return nil, err
		}
		num, err := parseInt(numTok.Value)
		if err != nil {
			return nil, p.error(numTok, "invalid tag number %q", numTok.Value)
		}
		tag.Number = num
	case TokenNumber:
		num, err := parseInt(p.curr.Value)
		if err != nil {
			return nil, p.error(p.curr, "invalid tag number %q", p.curr.Value)
		}
		tag.Number = num
		if err := p.advance(); err != nil {
			return nil, err
		}
	default:
		return nil, p.error(p.curr, "expected tag number")
	}

	if err := p.expectKind(TokenRBracket); err != nil {
		return nil, err
	}
	return tag, nil
}

func (p *parser) parseDefaultValue() (*DefaultValue, error) {
	switch p.curr.Kind {
	case TokenNumber:
		num, err := parseInt64(p.curr.Value)
		if err != nil {
			return nil, p.error(p.curr, "invalid default integer %q", p.curr.Value)
		}
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &DefaultValue{Kind: DefaultKindInteger, Int: num}, nil
	case TokenTRUE:
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &DefaultValue{Kind: DefaultKindBoolean, Bool: true}, nil
	case TokenFALSE:
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &DefaultValue{Kind: DefaultKindBoolean, Bool: false}, nil
	case TokenNULL:
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &DefaultValue{Kind: DefaultKindNull, Null: true}, nil
	case TokenIdent:
		name := p.curr.Value
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &DefaultValue{Kind: DefaultKindReference, Ident: name}, nil
	default:
		return nil, p.error(p.curr, "expected default value")
	}
}

func (p *parser) advance() error {
	if p.hasPK {
		p.curr = p.peek
		p.hasPK = false
		return nil
	}
	tok, err := p.lex.next()
	if err != nil {
		return err
	}
	p.curr = tok
	return nil
}

func (p *parser) expect(kind TokenKind) (Token, error) {
	if p.curr.Kind != kind {
		return Token{}, p.error(p.curr, "expected %s", kind)
	}
	tok := p.curr
	if err := p.advance(); err != nil {
		return Token{}, err
	}
	return tok, nil
}

func (p *parser) expectNamedLabel() (Token, error) {
	if isStructuralToken(p.curr.Kind) || p.curr.Kind == TokenEOF {
		return Token{}, p.error(p.curr, "expected identifier")
	}
	tok := p.curr
	if err := p.advance(); err != nil {
		return Token{}, err
	}
	return tok, nil
}

func isStructuralToken(kind TokenKind) bool {
	switch kind {
	case TokenLBrace, TokenRBrace, TokenLBracket, TokenRBracket,
		TokenLParen, TokenRParen, TokenComma, TokenEllipsis, TokenColonColonEquals:
		return true
	default:
		return false
	}
}

func (p *parser) expectKind(kind TokenKind) error {
	_, err := p.expect(kind)
	return err
}

func (p *parser) error(tok Token, format string, args ...any) error {
	return &ParseError{
		Token: tok,
		Msg:   fmt.Sprintf(format, args...),
	}
}

func parseInt(s string) (int, error) {
	v, err := parseInt64(s)
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

func parseInt64(s string) (int64, error) {
	var n int64
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid integer")
		}
		n = n*10 + int64(ch-'0')
	}
	return n, nil
}
