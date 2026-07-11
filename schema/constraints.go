package schema

// ConstraintKind identifies a parsed ASN.1 constraint.
type ConstraintKind int

const (
	ConstraintKindSize ConstraintKind = iota
	ConstraintKindValueRange
	ConstraintKindPermittedAlphabet
	ConstraintKindObjectSet
)

// Constraint is a parsed constraint applied to an ASN.1 type.
type Constraint interface {
	Kind() ConstraintKind
}

// SizeConstraint is a SIZE(l..u) or SIZE(n) constraint.
type SizeConstraint struct {
	Lower int64
	Upper int64
	Fixed bool
}

func (SizeConstraint) Kind() ConstraintKind { return ConstraintKindSize }

// RangeEndpointKind identifies a value-range endpoint.
type RangeEndpointKind int

const (
	EndpointNumber RangeEndpointKind = iota
	EndpointMin
	EndpointMax
)

// RangeEndpoint is one bound of a value range constraint.
type RangeEndpoint struct {
	Kind  RangeEndpointKind
	Value int64
}

// ValueRangeConstraint is a (lower..upper) constraint on numeric types.
type ValueRangeConstraint struct {
	Lower RangeEndpoint
	Upper RangeEndpoint
}

func (ValueRangeConstraint) Kind() ConstraintKind { return ConstraintKindValueRange }

// PermittedAlphabetConstraint is a FROM(...) constraint on string types.
type PermittedAlphabetConstraint struct {
	Lower string
	Upper string
}

func (PermittedAlphabetConstraint) Kind() ConstraintKind { return ConstraintKindPermittedAlphabet }

// ObjectSetBrace is one {...} block inside an object-set constraint.
type ObjectSetBrace struct {
	Extensible      bool
	HasLeadingComma bool
	PropertyRef     string
}

// ObjectSetConstraint is a parenthesized {...} object-set constraint.
type ObjectSetConstraint struct {
	Braces []ObjectSetBrace
}

func (ObjectSetConstraint) Kind() ConstraintKind { return ConstraintKindObjectSet }

// SizeConstraintFrom returns the first SIZE constraint, if any.
func SizeConstraintFrom(constraints []Constraint) (*SizeConstraint, bool) {
	for _, c := range constraints {
		if s, ok := c.(SizeConstraint); ok {
			return &s, true
		}
	}
	return nil, false
}

// ValueRangeConstraintFrom returns the first value-range constraint, if any.
func ValueRangeConstraintFrom(constraints []Constraint) (*ValueRangeConstraint, bool) {
	for _, c := range constraints {
		if r, ok := c.(ValueRangeConstraint); ok {
			return &r, true
		}
	}
	return nil, false
}
