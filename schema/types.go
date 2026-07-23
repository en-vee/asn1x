package schema

// TagClass identifies the class portion of an ASN.1 tag.
type TagClass int

const (
	TagClassUniversal TagClass = iota
	TagClassApplication
	TagClassContext
	TagClassPrivate
)

// Tag describes an ASN.1 tag applied to a type or component.
type Tag struct {
	Class  TagClass
	Number int
}

// Kind identifies the ASN.1 type form.
type Kind int

const (
	KindReference Kind = iota
	KindBoolean
	KindInteger
	KindEnumerated
	KindReal
	KindNull
	KindBitString
	KindOctetString
	KindString
	KindObjectIdentifier
	KindRelativeOID
	KindExternal
	KindEmbeddedPDV
	KindUTCTime
	KindGeneralizedTime
	KindSequence
	KindSet
	KindChoice
	KindSequenceOf
	KindSetOf
	KindTagged
)

// TagDefault is the module tagging default (X.680 TagDefault).
type TagDefault int

const (
	// TagDefaultExplicit is the ASN.1 default when no TAGS clause is present.
	TagDefaultExplicit TagDefault = iota
	TagDefaultImplicit
	TagDefaultAutomatic
)

// Type is a parsed ASN.1 type definition.
type Type interface {
	TypeKind() Kind
}

// ReferenceType refers to another named type in the module.
type ReferenceType struct {
	Name        string
	Constraints []Constraint
}

func (ReferenceType) TypeKind() Kind { return KindReference }

// BooleanType is the ASN.1 BOOLEAN type.
type BooleanType struct{}

func (BooleanType) TypeKind() Kind { return KindBoolean }

// NullType is the ASN.1 NULL type.
type NullType struct{}

func (NullType) TypeKind() Kind { return KindNull }

// IntegerType is the ASN.1 INTEGER type.
type IntegerType struct {
	NamedNumbers map[string]int
	Extensible   bool
	Constraints  []Constraint
}

func (IntegerType) TypeKind() Kind { return KindInteger }

// EnumeratedType is the ASN.1 ENUMERATED type.
type EnumeratedType struct {
	Values      map[string]int
	Extensible  bool
	Constraints []Constraint
}

func (EnumeratedType) TypeKind() Kind { return KindEnumerated }

// RealType is the ASN.1 REAL type.
type RealType struct {
	Constraints []Constraint
}

func (RealType) TypeKind() Kind { return KindReal }

// BitStringType is the ASN.1 BIT STRING type.
type BitStringType struct {
	NamedBits   map[string]int
	Extensible  bool
	Constraints []Constraint
}

func (BitStringType) TypeKind() Kind { return KindBitString }

// OctetStringType is the ASN.1 OCTET STRING type.
type OctetStringType struct {
	Constraints []Constraint
}

func (OctetStringType) TypeKind() Kind { return KindOctetString }

// StringType is a built-in ASN.1 character string type.
type StringType struct {
	Name        string
	Constraints []Constraint
}

func (StringType) TypeKind() Kind { return KindString }

// ObjectIdentifierType is the ASN.1 OBJECT IDENTIFIER type.
type ObjectIdentifierType struct {
	Constraints []Constraint
}

func (ObjectIdentifierType) TypeKind() Kind { return KindObjectIdentifier }

// RelativeOIDType is the ASN.1 RELATIVE-OID type.
type RelativeOIDType struct {
	Constraints []Constraint
}

func (RelativeOIDType) TypeKind() Kind { return KindRelativeOID }

// ExternalType is the ASN.1 EXTERNAL type.
type ExternalType struct {
	Constraints []Constraint
}

func (ExternalType) TypeKind() Kind { return KindExternal }

// EmbeddedPDVType is the ASN.1 EMBEDDED PDV type.
type EmbeddedPDVType struct {
	Constraints []Constraint
}

func (EmbeddedPDVType) TypeKind() Kind { return KindEmbeddedPDV }

// UTCTimeType is the ASN.1 UTCTime type.
type UTCTimeType struct {
	Constraints []Constraint
}

func (UTCTimeType) TypeKind() Kind { return KindUTCTime }

// GeneralizedTimeType is the ASN.1 GeneralizedTime type.
type GeneralizedTimeType struct {
	Constraints []Constraint
}

func (GeneralizedTimeType) TypeKind() Kind { return KindGeneralizedTime }

// TaggedType applies an ASN.1 tag to another type (e.g. [APPLICATION 1] SEQUENCE).
type TaggedType struct {
	Tag      Tag
	Implicit bool
	Type     Type
}

func (TaggedType) TypeKind() Kind { return KindTagged }

// Component is one member of a SEQUENCE, SET, or CHOICE type.
type Component struct {
	Name     string
	Tag      *Tag
	Implicit bool
	Type     Type
	Optional bool
	Default  *DefaultValue
}

// DefaultValue holds a parsed DEFAULT clause value.
type DefaultValue struct {
	Kind  DefaultKind
	Int   int64
	Bool  bool
	Null  bool
	Ident string
}

// DefaultKind identifies the form of a DEFAULT value.
type DefaultKind int

const (
	DefaultKindInteger DefaultKind = iota
	DefaultKindBoolean
	DefaultKindNull
	DefaultKindReference
)

// SequenceType is the ASN.1 SEQUENCE type.
type SequenceType struct {
	Components  []Component
	Extensible  bool
	Constraints []Constraint
}

func (SequenceType) TypeKind() Kind { return KindSequence }

// SetType is the ASN.1 SET type.
type SetType struct {
	Components  []Component
	Extensible  bool
	Constraints []Constraint
}

func (SetType) TypeKind() Kind { return KindSet }

// ChoiceType is the ASN.1 CHOICE type.
type ChoiceType struct {
	Components  []Component
	Extensible  bool
	Constraints []Constraint
}

func (ChoiceType) TypeKind() Kind { return KindChoice }

// SequenceOfType is the ASN.1 SEQUENCE OF type.
type SequenceOfType struct {
	Element     Type
	Constraints []Constraint
}

func (SequenceOfType) TypeKind() Kind { return KindSequenceOf }

// SetOfType is the ASN.1 SET OF type.
type SetOfType struct {
	Element     Type
	Constraints []Constraint
}

func (SetOfType) TypeKind() Kind { return KindSetOf }

// Assignment binds a name to a type within a module.
type Assignment struct {
	Name string
	Type Type
}

// Schema is a parsed ASN.1 module.
type Schema struct {
	ModuleName string
	ModuleOID  string
	TagDefault TagDefault
	Types      map[string]Type
}

// Lookup returns the type assigned to name, if present.
func (s *Schema) Lookup(name string) (Type, bool) {
	t, ok := s.Types[name]
	return t, ok
}

// ConstraintsOf returns constraints attached to t, when present.
func ConstraintsOf(t Type) []Constraint {
	switch typ := t.(type) {
	case IntegerType:
		return typ.Constraints
	case EnumeratedType:
		return typ.Constraints
	case RealType:
		return typ.Constraints
	case BitStringType:
		return typ.Constraints
	case OctetStringType:
		return typ.Constraints
	case StringType:
		return typ.Constraints
	case ObjectIdentifierType:
		return typ.Constraints
	case RelativeOIDType:
		return typ.Constraints
	case ExternalType:
		return typ.Constraints
	case EmbeddedPDVType:
		return typ.Constraints
	case UTCTimeType:
		return typ.Constraints
	case GeneralizedTimeType:
		return typ.Constraints
	case SequenceType:
		return typ.Constraints
	case SetType:
		return typ.Constraints
	case ChoiceType:
		return typ.Constraints
	case SequenceOfType:
		return typ.Constraints
	case SetOfType:
		return typ.Constraints
	case ReferenceType:
		return typ.Constraints
	default:
		return nil
	}
}
