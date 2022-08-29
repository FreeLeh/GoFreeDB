package freedb

import "errors"

const basicCodecPrefix string = "!"

// basicCodec encodes and decodes any bytes data using a very simple encoding rule.
// A prefix (an exclamation mark "!") will be attached to the raw bytes data.
//
// This allows the library to differentiate between empty raw bytes provided by the client from
// getting an empty data from Google Sheets API.
type basicCodec struct{}

// Encode encodes the given raw bytes by using an exclamation mark "!" as a prefix.
func (c *basicCodec) Encode(value []byte) (string, error) {
	return basicCodecPrefix + string(value), nil
}

// Decode converts the string data read from Google Sheet into raw bytes data after removing the
// exclamation mark "!" prefix.
func (c *basicCodec) Decode(value string) ([]byte, error) {
	if len(value) == 0 {
		return nil, errors.New("basic decode fail: empty string")
	}
	if value[:len(basicCodecPrefix)] != basicCodecPrefix {
		return nil, errors.New("basic decode fail: first character is not an empty space")
	}
	return []byte(value[len(basicCodecPrefix):]), nil
}
