package schema

// TokenKind classifies lexical tokens in an ASN.1 schema file.
type TokenKind int

const (
	TokenEOF TokenKind = iota
	TokenIdent
	TokenNumber
	TokenString

	TokenDEFINITIONS
	TokenBEGIN
	TokenEND
	TokenOPTIONAL
	TokenDEFAULT
	TokenIMPLICIT
	TokenEXPLICIT
	TokenTRUE
	TokenFALSE
	TokenNULL
	TokenOF

	TokenSEQUENCE
	TokenSET
	TokenCHOICE
	TokenINTEGER
	TokenBOOLEAN
	TokenENUMERATED
	TokenREAL
	TokenBIT
	TokenOCTET
	TokenSTRING
	TokenOBJECT
	TokenIDENTIFIER
	TokenEXTERNAL
	TokenEMBEDDED
	TokenPDV
	TokenUTCTime
	TokenGeneralizedTime
	TokenSIZE
	TokenFROM
	TokenMIN
	TokenMAX

	TokenLBrace
	TokenRBrace
	TokenLBracket
	TokenRBracket
	TokenLParen
	TokenRParen
	TokenComma
	TokenDot
	TokenDotDot
	TokenEllipsis
	TokenAmpersand
	TokenAt
	TokenColonColonEquals
)

// Token is a single lexeme with source position.
type Token struct {
	Kind   TokenKind
	Value  string
	Line   int
	Column int
}

func (k TokenKind) String() string {
	switch k {
	case TokenEOF:
		return "EOF"
	case TokenIdent:
		return "identifier"
	case TokenNumber:
		return "number"
	case TokenString:
		return "string"
	case TokenDEFINITIONS:
		return "DEFINITIONS"
	case TokenBEGIN:
		return "BEGIN"
	case TokenEND:
		return "END"
	case TokenOPTIONAL:
		return "OPTIONAL"
	case TokenDEFAULT:
		return "DEFAULT"
	case TokenIMPLICIT:
		return "IMPLICIT"
	case TokenEXPLICIT:
		return "EXPLICIT"
	case TokenTRUE:
		return "TRUE"
	case TokenFALSE:
		return "FALSE"
	case TokenNULL:
		return "NULL"
	case TokenOF:
		return "OF"
	case TokenSEQUENCE:
		return "SEQUENCE"
	case TokenSET:
		return "SET"
	case TokenCHOICE:
		return "CHOICE"
	case TokenINTEGER:
		return "INTEGER"
	case TokenBOOLEAN:
		return "BOOLEAN"
	case TokenENUMERATED:
		return "ENUMERATED"
	case TokenREAL:
		return "REAL"
	case TokenBIT:
		return "BIT"
	case TokenOCTET:
		return "OCTET"
	case TokenSTRING:
		return "STRING"
	case TokenOBJECT:
		return "OBJECT"
	case TokenIDENTIFIER:
		return "IDENTIFIER"
	case TokenEXTERNAL:
		return "EXTERNAL"
	case TokenEMBEDDED:
		return "EMBEDDED"
	case TokenPDV:
		return "PDV"
	case TokenUTCTime:
		return "UTCTime"
	case TokenGeneralizedTime:
		return "GeneralizedTime"
	case TokenSIZE:
		return "SIZE"
	case TokenFROM:
		return "FROM"
	case TokenMIN:
		return "MIN"
	case TokenMAX:
		return "MAX"
	case TokenLBrace:
		return "{"
	case TokenRBrace:
		return "}"
	case TokenLBracket:
		return "["
	case TokenRBracket:
		return "]"
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenComma:
		return ","
	case TokenDot:
		return "."
	case TokenDotDot:
		return ".."
	case TokenEllipsis:
		return "..."
	case TokenAmpersand:
		return "&"
	case TokenAt:
		return "@"
	case TokenColonColonEquals:
		return "::="
	default:
		return "unknown"
	}
}
