package decode

import "testing"

func TestDecodeOctetString(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want any
	}{
		{
			name: "generalized time text",
			in:   []byte("2022-04-29T01:04:42Z"),
			want: "2022-04-29T01:04:42Z",
		},
		{
			name: "binary plmn",
			in:   []byte{0x05, 0xf5, 0x30},
			want: "05f530",
		},
		{
			name: "ascii identifier",
			in:   []byte("505038309039675"),
			want: "505038309039675",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decodeOctetString(tt.in)
			if got != tt.want {
				t.Fatalf("decodeOctetString() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
