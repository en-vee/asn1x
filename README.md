# libasn1 - Go library for parsing ASN.1 BER encoded files

## Features
- Parse ASN.1 Schema files
- Convert from ASN.1 to JSON and vice-versa
- All mappings between ASN.1 <-> JSON are dynamic

## Usage
```go
import libasn1 "github.com/en-vee/libasn1"

func main() {
    
    // Converting ASN.1 to JSON
    asn1Schema, err := libasn1.Parse(asn1SchemaFileReader) // Argument is an ASN.1 Syntax Schema io.Reader
    // error handling
    // ...
    asn1Reader, err := libasn1.NewReader(asn1Schema, asn1File) // asn1File is an io.Reader on the ASN.1 BER encoded file
    // Print the JSON of the ASN.1 to a strings.Builder
    err := io.Copy(sb, asn1Reader)
   
    // JSON to ASN.1
    asn1Writer, err := libasn1.NewWriter(asn1Schema)
    fmt.Fprintf(sb,asn1JsonString)

}
```