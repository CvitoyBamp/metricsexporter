package crypto

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type data struct {
	data []byte
	key  string
}

func TestCreateHash(t *testing.T) {

	tests := []struct {
		testName string
		data     data
		wants    string
	}{
		{
			testName: "Correct encrypting",
			data: data{
				data: []byte("Это тест функции"),
				key:  "тест",
			},
			wants: "a605da4faa96f24b8323cd8aea5842164d13ecc71900e0cac7976c0d110a5438",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			td := CreateHash(tt.data.data, tt.data.key)
			assert.Equal(t, tt.wants, fmt.Sprintf("%x", td))
		})
	}
}
