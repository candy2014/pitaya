package papertiger

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
)

type Category uint

const (
	Unknown Category = iota
	Bool
	Int8
	Int16
	Int32
	Int64
	Float32
	Float64
	String
)

// ParterreBinary implements the serialize.Serializer interface
type ParterreBinary struct{}

// NewPaperTigerBinary returns a new Serializer.
func NewPaperTigerBinary() *ParterreBinary {
	return &ParterreBinary{}
}

// Marshal returns the JSON encoding of v.
func (s *ParterreBinary) Marshal(v interface{}) ([]byte, error) {
	mutable := reflect.ValueOf(v)
	fieldNum := mutable.Elem().NumField()
	buff := make([]byte, 2)
	binary.LittleEndian.PutUint16(buff, uint16(fieldNum))

	for i := 0; i < fieldNum; i++ {
		tp := mutable.Elem().Field(i)
		valType := tp.Type()

		if tp.IsValid() {
			if valType.Kind() == reflect.Bool {
				bf := make([]byte, 2)
				bf[0] = 1

				if tp.Interface().(bool) {
					bf[1] = 1
				} else {
					bf[1] = 0
				}
				buff = append(buff, bf...)
			} else if valType.Kind() == reflect.Int8 {
				bf := make([]byte, 2)
				bf[0] = 2
				bf[1] = byte(tp.Interface().(int8))
				buff = append(buff, bf...)
			} else if valType.Kind() == reflect.Int16 {
				bf := make([]byte, 3)
				bf[0] = 3
				binary.LittleEndian.PutUint16(bf[1:], uint16(tp.Interface().(int16)))
				buff = append(buff, bf...)
			} else if valType.Kind() == reflect.Int32 {
				bf := make([]byte, 5)
				bf[0] = 4
				binary.LittleEndian.PutUint32(bf[1:], uint32(tp.Interface().(int32)))
				buff = append(buff, bf...)
			} else if valType.Kind() == reflect.String {
				strLen := len(tp.Interface().(string))
				bf := make([]byte, 4+strLen)
				bf[0] = 8
				binary.LittleEndian.PutUint16(bf[1:], uint16(strLen+1))
				data := []byte(tp.Interface().(string))
				copy(bf[3:], data)
				buff = append(buff, bf...)
			}
		}
	}
	return buff, nil
}

// Unmarshal parses the JSON-encoded data and stores the result
// in the value pointed to by v.
func (s *ParterreBinary) Unmarshal(data []byte, v interface{}) error {
	length := reflect.ValueOf(v).Elem().NumField()
	offset := 0

	for i := 0; i < length; i++ {
		dataType := data[offset]
		offset++

		if Category(dataType) == Bool {
			value := data[offset]
			offset++
			s.writeBool(v, i, value == 1)
		} else if Category(dataType) == Int8 {
			value := data[offset]
			offset++
			s.writeInt(v, i, int64(value))
		} else if Category(dataType) == Int16 {
			value := int16(binary.LittleEndian.Uint16(data[offset:]))
			offset = offset + 2
			s.writeInt(v, i, int64(value))
		} else if Category(dataType) == Int32 {
			value := int32(binary.LittleEndian.Uint32(data[offset:]))
			offset = offset + 4
			s.writeInt(v, i, int64(value))
		} else if Category(dataType) == Int64 {
			value := int64(binary.LittleEndian.Uint64(data[offset:]))
			offset = offset + 8
			s.writeInt(v, i, value)
		} else if Category(dataType) == Float32 {
			value := math.Float32frombits(binary.LittleEndian.Uint32(data[offset:]))
			offset = offset + 4
			s.writeFloat(v, i, float64(value))
		} else if Category(dataType) == Float64 {
			value := math.Float64frombits(binary.LittleEndian.Uint64(data[offset:]))
			offset = offset + 8
			s.writeFloat(v, i, value)
		} else if Category(dataType) == String {
			value := int16(binary.LittleEndian.Uint16(data[offset:]))
			offset = offset + 2
			var end = value + int16(offset)
			content := string(data[offset : end-1])
			offset = offset + int(value)
			s.writeString(v, i, content)
		}
	}
	return nil
}

// GetName returns the name of the serializer.
func (s *ParterreBinary) GetName() string {
	return "parpertigerbinary"
}

func (s *ParterreBinary) writeInt(param interface{}, postion int, value int64) error {
	mutable := reflect.ValueOf(param).Elem()
	vv := mutable.Field(postion)

	if vv.IsValid() && vv.CanSet() {
		kind := vv.Kind()
		if kind != reflect.Int8 && kind != reflect.Uint8 &&
			kind != reflect.Int16 && kind != reflect.Uint16 &&
			kind != reflect.Int32 && kind != reflect.Uint32 &&
			kind != reflect.Int64 && kind != reflect.Uint64 {
			yp := reflect.TypeOf(param)
			return fmt.Errorf(
				"%q 数据的第 %d 类型不匹配 需要 int 但是当前是 %q",
				yp.Name(),
				postion,
				yp.Field(postion).Type.Name())
		}
		vv.SetInt(value)
	}
	return nil
}

func (s *ParterreBinary) writeFloat(param interface{}, postion int, value float64) error {
	mutable := reflect.ValueOf(param).Elem()
	vv := mutable.Field(postion)

	if vv.IsValid() && vv.CanSet() {
		kind := vv.Type().Kind()
		if kind != reflect.Float32 && kind != reflect.Float64 {
			yp := reflect.TypeOf(param)
			return fmt.Errorf(
				"%q 数据的第 %d 类型不匹配 需要 int 但是当前是 %q",
				yp.Name(),
				postion,
				yp.Field(postion).Type.Name())
		}
		vv.SetFloat(value)
	}
	return nil
}

func (s *ParterreBinary) writeBool(param interface{}, potion int, value bool) error {
	mutable := reflect.ValueOf(param).Elem()
	vv := mutable.Field(potion)

	if vv.IsValid() && vv.CanSet() {
		kind := vv.Type().Kind()
		if kind != reflect.Bool {
			yp := reflect.TypeOf(param)
			return fmt.Errorf(
				"%q 数据的第 %d 类型不匹配 需要 int 但是当前是 %q",
				yp.Name(),
				potion,
				yp.Field(potion).Type.Name())
		}
		vv.SetBool(value)
	}
	return nil
}

func (s *ParterreBinary) writeString(param interface{}, potion int, value string) error {
	mutable := reflect.ValueOf(param).Elem()
	vv := mutable.Field(potion)

	if vv.IsValid() && vv.CanSet() {
		kind := vv.Type().Kind()
		if kind != reflect.String {
			yp := reflect.TypeOf(param)
			return fmt.Errorf(
				"%q 数据的第 %d 类型不匹配 需要 int 但是当前是 %q",
				yp.Name(),
				potion,
				yp.Field(potion).Type.Name())
		}
		vv.SetString(value)
	}
	return nil
}
