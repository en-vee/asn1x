package decode

import "fmt"

// decodePLMNID decodes a 3GPP PLMN identity (OCTET STRING SIZE(3)) from BCD.
// See 3GPP TS 24.008 / TS 23.003 and semi-octet layout described at
// https://nickvsnetworking.com/plmn-identifier-calculation-mcc-mnc-to-plmn/
func decodePLMNID(content []byte) (map[string]string, error) {
	if len(content) == 0 {
		return nil, nil
	}
	if len(content) != 3 {
		return nil, fmt.Errorf("PLMNID length %d, want 3", len(content))
	}

	mccDigit1 := nibble(content[0], false)
	mccDigit2 := nibble(content[0], true)
	mccDigit3 := nibble(content[1], false)
	mncDigit3 := nibble(content[1], true)
	mncDigit1 := nibble(content[2], false)
	mncDigit2 := nibble(content[2], true)

	if err := validatePLMNDigit(mccDigit1); err != nil {
		return nil, fmt.Errorf("PLMNID mcc digit 1: %w", err)
	}
	if err := validatePLMNDigit(mccDigit2); err != nil {
		return nil, fmt.Errorf("PLMNID mcc digit 2: %w", err)
	}
	if err := validatePLMNDigit(mccDigit3); err != nil {
		return nil, fmt.Errorf("PLMNID mcc digit 3: %w", err)
	}

	mcc := fmt.Sprintf("%d%d%d", mccDigit1, mccDigit2, mccDigit3)

	var mnc string
	if mncDigit3 == 0x0F {
		if err := validatePLMNDigit(mncDigit1); err != nil {
			return nil, fmt.Errorf("PLMNID mnc digit 1: %w", err)
		}
		if err := validatePLMNDigit(mncDigit2); err != nil {
			return nil, fmt.Errorf("PLMNID mnc digit 2: %w", err)
		}
		mnc = fmt.Sprintf("%d%d", mncDigit1, mncDigit2)
	} else {
		if err := validatePLMNDigit(mncDigit3); err != nil {
			return nil, fmt.Errorf("PLMNID mnc digit 3: %w", err)
		}
		if err := validatePLMNDigit(mncDigit1); err != nil {
			return nil, fmt.Errorf("PLMNID mnc digit 1: %w", err)
		}
		if err := validatePLMNDigit(mncDigit2); err != nil {
			return nil, fmt.Errorf("PLMNID mnc digit 2: %w", err)
		}
		mnc = fmt.Sprintf("%d%d%d", mncDigit3, mncDigit2, mncDigit1)
	}

	return map[string]string{
		"mcc": mcc,
		"mnc": mnc,
	}, nil
}

func nibble(b byte, high bool) int {
	if high {
		return int(b >> 4)
	}
	return int(b & 0x0F)
}

func validatePLMNDigit(d int) error {
	if d < 0 || d > 9 {
		return fmt.Errorf("invalid BCD digit %d", d)
	}
	return nil
}
