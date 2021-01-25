// Package address contains utilities for working with encoded addresses.
//
// Example: NQ19 46LK 9YHV D9LB TDJ4 8Y2P 3J4C 37HR YDL5
//
// The Nimiq user friendly address format is a 36 character
// ASCII string separated by spaces.
// Spaces (0x20) are not counted as characters and will be ignored.
// Characters 0-2 (country code) must be "NQ".
// Characters 2-4 are a IBAN-like checksum of the following chars.
// Characters 4-36 encode the 20 byte Address type in Base32.
package address

import (
	"bytes"
	"encoding/base32"
	"errors"
	"strconv"
	"strings"
)

// Encoding is the Base32-encoding used with Nimiq addresses.
var Encoding = base32.NewEncoding("0123456789ABCDEFGHJKLMNPQRSTUVXY")

// Decode decodes the encoded address into the raw byte array.
// Spaces are ignored.
// The encoded address must have exactly 36 non-space characters.
func Decode(encoded string) (addr [20]byte, err error) {
	// Encoded in ASCII (as opposed to the Unicode-encoded string type)
	// Remove all spaces from encoded and write to compact.
	var compact [36]byte
	compact, err = truncateWhitespace(encoded)
	if err != nil {
		return
	}
	// Verify string: Country code
	if compact[0] != 'N' && compact[1] != 'Q' {
		err = ErrInvalidCountryCode
		return
	}
	// Verify string: Checksum and characters
	var check uint8
	check, err = addressCheck(&compact)
	if err != nil {
		return
	}
	if check != 1 {
		err = ErrChecksum
		return
	}
	// Decode everything past the first 4 chars (data region)
	_, err = Encoding.Decode(addr[:], compact[4:])
	return
}

// Encode encodes an address into a user friendly string (ASCII).
func Encode(addr *[20]byte) string {
	// User friendly string without spaces
	var noSpaces [36]byte

	// Use "00" as temporary checksum
	copy(noSpaces[0:4], "NQ00")
	// Encode the rest of the address
	Encoding.Encode(noSpaces[4:], addr[:])

	// Checksum is (98 - "current check result")
	check, _ := addressCheck(&noSpaces)
	check = 98 - check

	var withSpaces strings.Builder
	withSpaces.WriteString("NQ")

	// Write checksum digits to buffer
	// (9 + 0x30 == '9') turns digit to ASCII character
	withSpaces.Write([]byte{
		'0' + (check % 100 / 10), // Digit X of XY
		'0' + check%10,           // Digit Y of XY
	})

	// Copy noSpaces to withSpaces
	for i := 4; i < 36; i += 4 {
		// Insert a space every 4 characters
		withSpaces.WriteByte(' ')
		withSpaces.Write(noSpaces[i : i+4])
	}

	return withSpaces.String()
}

// Removes spaces from "encodedStr" and writes it into "compact" byte array.
func truncateWhitespace(encodedStr string) (compact [36]byte, err error) {
	// Treat encodedStr as an ASCII string
	encoded := []byte(encodedStr)
	// Slice of compact that has not been written to.
	compactRemainder := compact[:]
	// Slice of encoded address that has not been read from.
	remainder := encoded
	// Walk to end of both slices.
	// If end of encoded address and end of compact representation
	// have been reached at the same time, truncating whitespaces is successful
	for {
		if len(remainder) == 0 {
			// Input has been read completely.
			if len(compactRemainder) != 0 {
				// Not all address chars have been read (underflow).
				err = ErrInvalidLength
				return
			}
			// All chars have been read, and there is no overflow.
			return
		} else if len(compactRemainder) == 0 {
			// All address chars have been read, but there are some remaining (overflow).
			err = ErrInvalidLength
			return
		}
		// Read next char from input.
		if char := remainder[0]; char != ' ' {
			// Not a space, write to encoded.
			compactRemainder[0] = char
			compactRemainder = compactRemainder[1:]
		}
		remainder = remainder[1:]
	}
}

// Calculates a checksum over the UFA
// Returns values from 0-98
func addressCheck(userFriendly *[36]byte) (uint8, error) {
	var sumBuffer bytes.Buffer

	// Writes a bunch of chars to sumBuffer for check loop later
	// e.g. "6789ABcd" will append "678910111213" to the sumBuffer
	nextChars := func(slice []byte) error {
		for _, char := range slice {
			switch {

			// Lower case character
			case char > 0x60 && char <= 0x7A:
				char -= 0x20 // convert to upper
				fallthrough

			// Upper case character
			case char > 0x40 && char <= 0x5A:
				// Subtract 0x37
				// Rationale: Numerical values are 0-9
				// Subtracting 0x37 will make letters to number 10…
				num := char - 0x37
				// Represent num as a decimal number
				numStr := strconv.FormatUint(uint64(num), 10)
				// Append to sum buffer
				sumBuffer.WriteString(numStr)
				break

			// Numerical character
			case char >= 0x30 && char <= 0x39:
				// Append numerical character (not code) to buffer
				sumBuffer.WriteByte(char)
				break

			// Unknown character
			default:
				return ErrInvalidCharacter
			}
		}

		return nil
	}

	// Iterate over the chars in addrBuf, the first 4 chars at last
	if err := nextChars(userFriendly[4:]); err != nil {
		return 0, err
	}
	if err := nextChars(userFriendly[0:4]); err != nil {
		return 0, err
	}

	// Extract string (here as bytes) from sumBuffer
	sum := sumBuffer.Bytes()

	// Create new buffer for tmp variable
	var tmpBuffer bytes.Buffer

	// Rounding up division
	blockCount := (len(sum) + 5) / 6

	// Iterate over sum in blocks of 6
	for i := 0; true; i++ {
		offset := i * 6
		var stop int
		if len(sum) <= offset+6 {
			stop = len(sum)
		} else {
			stop = offset + 6
		}

		block := sum[offset:stop]

		// Append the block to the buffer
		tmpBuffer.Write(block)

		// Read string out of buffer
		tmp := tmpBuffer.String()

		// Convert to integer, ignore errors
		tmpNum, _ := strconv.ParseUint(tmp, 10, 64)
		tmpNum %= 97 // magic :P

		if (i + 1) < blockCount {
			// Put string back into buffer
			tmpBuffer.Reset()
			tmpBuffer.WriteString(strconv.FormatUint(tmpNum, 10))
		} else {
			// No more blocks, return final value
			return uint8(tmpNum), nil
		}
	}

	panic("unreachable")
}

// Common address errors.
var (
	ErrInvalidCountryCode = errors.New("invalid address: should start with \"NQ\"")
	ErrInvalidLength      = errors.New("invalid address: length without spaces not 36")
	ErrChecksum           = errors.New("invalid address: invalid checksum")
	ErrInvalidCharacter   = errors.New("invalid address: unexpected character")
)
