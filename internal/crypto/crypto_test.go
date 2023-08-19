package crypto

import (
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
			wants: "3db5c4e0a18ba507a933bac29fb9554d627d457342554db245e0c9efe5651a9b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			td := CreateHash(tt.data.data, tt.data.key)
			assert.Equal(t, tt.wants, td)
		})
	}
}
