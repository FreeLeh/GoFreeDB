package common

import "github.com/mitchellh/mapstructure"

func MapStructureDecode(input interface{}, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		Result:  output,
		TagName: "db",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}
