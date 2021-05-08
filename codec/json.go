// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package codec

import (
	"encoding/json"
	"fmt"
)

func JSON() Format {
	return &jsonFormat{}
}

type jsonFormat struct{}

func (j *jsonFormat) Encode(input interface{}, output []byte) (out []byte, err error) {
	data, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("json marshalling failed: %w", err)
	}

	return data, nil
}

func (j *jsonFormat) Decode(input []byte, output interface{}) error {
	if err := json.Unmarshal(input, output); err != nil {
		return fmt.Errorf("json unmarshalling failed: %w", err)
	}

	return nil
}
