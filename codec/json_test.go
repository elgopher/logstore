// (c) 2021 Jacek Olszak
// This code is licensed under MIT license (see LICENSE for details)

package codec_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJsonFormat_Encode(t *testing.T) {
	t.Run("should marshal json", func(t *testing.T) {
		msg := &JSONMessage{Text: "message"}
		// when
		bytes, err := json.Marshal(msg)
		// then
		require.NoError(t, err)
		assert.JSONEq(t, `{"Text":"message"}`, string(bytes))
	})
}

func TestJsonFormat_Decode(t *testing.T) {
	t.Run("should unmarshal json", func(t *testing.T) {
		bytes := []byte(`{"Text":"message"}`)
		msg := JSONMessage{}
		// when
		err := json.Unmarshal(bytes, &msg)
		// then
		require.NoError(t, err)
		assert.Equal(t, JSONMessage{Text: "message"}, msg)
	})
}

type JSONMessage struct {
	Text string
}
