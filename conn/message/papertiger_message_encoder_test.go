package message

import "testing"

func TestDecodeData(t *testing.T) {
	type args struct {
		data []byte
	}

	data := args{data: []byte{100, 0, 0, 0}}
	tests := []struct {
		name string
		args args
	}{
		{args: data},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DecodeData(tt.args.data)
		})
	}
}
