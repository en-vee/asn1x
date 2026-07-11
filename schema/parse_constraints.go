package schema

func (p *parser) parseConstraints() ([]Constraint, error) {
	var constraints []Constraint
	for p.curr.Kind == TokenLParen {
		c, err := p.parseParenConstraint()
		if err != nil {
			return nil, err
		}
		constraints = append(constraints, c)
	}
	return constraints, nil
}

func (p *parser) parseParenConstraint() (Constraint, error) {
	if err := p.expectKind(TokenLParen); err != nil {
		return nil, err
	}

	var (
		c   Constraint
		err error
	)

	switch p.curr.Kind {
	case TokenSIZE:
		c, err = p.parseSizeConstraintBody()
	case TokenFROM:
		c, err = p.parsePermittedAlphabetConstraintBody()
	case TokenNumber, TokenMIN, TokenMAX:
		c, err = p.parseValueRangeConstraintBody()
	case TokenLBrace:
		c, err = p.parseObjectSetConstraintBody()
	default:
		return nil, p.error(p.curr, "expected constraint")
	}
	if err != nil {
		return nil, err
	}

	if err := p.expectKind(TokenRParen); err != nil {
		return nil, err
	}
	return c, nil
}

func (p *parser) parseSizeConstraintBody() (Constraint, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	if err := p.expectKind(TokenLParen); err != nil {
		return nil, err
	}

	lower, err := p.parseRangeEndpoint()
	if err != nil {
		return nil, err
	}

	size := SizeConstraint{Lower: lower.Value}
	if p.curr.Kind == TokenDotDot {
		if err := p.advance(); err != nil {
			return nil, err
		}
		upper, err := p.parseRangeEndpoint()
		if err != nil {
			return nil, err
		}
		size.Upper = upper.Value
		size.Fixed = false
	} else {
		size.Upper = lower.Value
		size.Fixed = true
	}

	if err := p.expectKind(TokenRParen); err != nil {
		return nil, err
	}
	return size, nil
}

func (p *parser) parseValueRangeConstraintBody() (Constraint, error) {
	lower, err := p.parseRangeEndpoint()
	if err != nil {
		return nil, err
	}
	if err := p.expectKind(TokenDotDot); err != nil {
		return nil, err
	}
	upper, err := p.parseRangeEndpoint()
	if err != nil {
		return nil, err
	}
	return ValueRangeConstraint{Lower: lower, Upper: upper}, nil
}

func (p *parser) parseRangeEndpoint() (RangeEndpoint, error) {
	switch p.curr.Kind {
	case TokenMIN:
		if err := p.advance(); err != nil {
			return RangeEndpoint{}, err
		}
		return RangeEndpoint{Kind: EndpointMin}, nil
	case TokenMAX:
		if err := p.advance(); err != nil {
			return RangeEndpoint{}, err
		}
		return RangeEndpoint{Kind: EndpointMax}, nil
	case TokenNumber:
		num, err := parseInt64(p.curr.Value)
		if err != nil {
			return RangeEndpoint{}, p.error(p.curr, "invalid integer %q", p.curr.Value)
		}
		if err := p.advance(); err != nil {
			return RangeEndpoint{}, err
		}
		return RangeEndpoint{Kind: EndpointNumber, Value: num}, nil
	default:
		return RangeEndpoint{}, p.error(p.curr, "expected range endpoint")
	}
}

func (p *parser) parsePermittedAlphabetConstraintBody() (Constraint, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	if err := p.expectKind(TokenLParen); err != nil {
		return nil, err
	}

	lower, err := p.parsePermittedAlphabetValue()
	if err != nil {
		return nil, err
	}
	if err := p.expectKind(TokenDotDot); err != nil {
		return nil, err
	}
	upper, err := p.parsePermittedAlphabetValue()
	if err != nil {
		return nil, err
	}
	if err := p.expectKind(TokenRParen); err != nil {
		return nil, err
	}
	return PermittedAlphabetConstraint{Lower: lower, Upper: upper}, nil
}

func (p *parser) parsePermittedAlphabetValue() (string, error) {
	switch p.curr.Kind {
	case TokenString:
		val := p.curr.Value
		if err := p.advance(); err != nil {
			return "", err
		}
		return val, nil
	case TokenIdent, TokenNumber:
		val := p.curr.Value
		if err := p.advance(); err != nil {
			return "", err
		}
		return val, nil
	default:
		return "", p.error(p.curr, "expected permitted alphabet value")
	}
}

func (p *parser) parseObjectSetConstraintBody() (Constraint, error) {
	constraint := ObjectSetConstraint{}
	for p.curr.Kind == TokenLBrace {
		brace, err := p.parseObjectSetBrace()
		if err != nil {
			return nil, err
		}
		constraint.Braces = append(constraint.Braces, brace)
	}
	return constraint, nil
}

func (p *parser) parseObjectSetBrace() (ObjectSetBrace, error) {
	brace := ObjectSetBrace{}
	if err := p.expectKind(TokenLBrace); err != nil {
		return ObjectSetBrace{}, err
	}

	if p.curr.Kind == TokenComma {
		brace.HasLeadingComma = true
		if err := p.advance(); err != nil {
			return ObjectSetBrace{}, err
		}
	}

	if p.curr.Kind == TokenAt {
		if err := p.advance(); err != nil {
			return ObjectSetBrace{}, err
		}
		if p.curr.Kind == TokenDot {
			if err := p.advance(); err != nil {
				return ObjectSetBrace{}, err
			}
		}
		nameTok, err := p.expectNamedLabel()
		if err != nil {
			return ObjectSetBrace{}, err
		}
		brace.PropertyRef = nameTok.Value
		if err := p.expectKind(TokenRBrace); err != nil {
			return ObjectSetBrace{}, err
		}
		return brace, nil
	}

	if p.curr.Kind == TokenEllipsis {
		brace.Extensible = true
		if err := p.advance(); err != nil {
			return ObjectSetBrace{}, err
		}
	}

	if p.curr.Kind == TokenComma {
		if err := p.advance(); err != nil {
			return ObjectSetBrace{}, err
		}
		if p.curr.Kind == TokenEllipsis {
			brace.Extensible = true
			if err := p.advance(); err != nil {
				return ObjectSetBrace{}, err
			}
		}
	}

	if err := p.expectKind(TokenRBrace); err != nil {
		return ObjectSetBrace{}, err
	}
	return brace, nil
}

func withConstraints(t Type, constraints []Constraint) Type {
	switch typ := t.(type) {
	case IntegerType:
		typ.Constraints = constraints
		return typ
	case EnumeratedType:
		typ.Constraints = constraints
		return typ
	case RealType:
		typ.Constraints = constraints
		return typ
	case BitStringType:
		typ.Constraints = constraints
		return typ
	case OctetStringType:
		typ.Constraints = constraints
		return typ
	case StringType:
		typ.Constraints = constraints
		return typ
	case ObjectIdentifierType:
		typ.Constraints = constraints
		return typ
	case RelativeOIDType:
		typ.Constraints = constraints
		return typ
	case ExternalType:
		typ.Constraints = constraints
		return typ
	case EmbeddedPDVType:
		typ.Constraints = constraints
		return typ
	case UTCTimeType:
		return UTCTimeType{Constraints: constraints}
	case GeneralizedTimeType:
		return GeneralizedTimeType{Constraints: constraints}
	case SequenceOfType:
		typ.Constraints = constraints
		return typ
	case SetOfType:
		typ.Constraints = constraints
		return typ
	case ReferenceType:
		typ.Constraints = constraints
		return typ
	default:
		return t
	}
}
