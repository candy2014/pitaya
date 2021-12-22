package papertiger

import (
	"reflect"
	"testing"
)

func TestParterreBinary_Marshal(t *testing.T) {

	type Data struct {
		Year int8
		Old  bool
		Cat  int16
		Name string
	}

	dataVa := &Data{
		Year: -8,
		Old:  true,
		Cat:  -32,
		Name: "hello world",
	}
	type args struct {
		v interface{}
	}

	a := args{v: dataVa}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			args: a,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ParterreBinary{}
			got, err := s.Marshal(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Marshal() got = %v, want %v", got, tt.want)
			}
		})
	}
}
