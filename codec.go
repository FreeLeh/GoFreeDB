package freeleh

import "errors"

const basicCodecPrefix string = "!"

type BasicCodec struct{}

func (c *BasicCodec) Encode(value []byte) (string, error) {
	return basicCodecPrefix + string(value), nil
}

func (c *BasicCodec) Decode(value string) ([]byte, error) {
	if len(value) == 0 {
		return nil, errors.New("basic decode fail: empty string")
	}
	if value[:len(basicCodecPrefix)] != basicCodecPrefix {
		return nil, errors.New("basic decode fail: first character is not an empty space")
	}
	return []byte(value[len(basicCodecPrefix):]), nil
}
