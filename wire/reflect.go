package wire

import (
	"errors"
	"fmt"
	"reflect"
)

// Unmarshal parses a wire-format message in b and places the decoded results
// in args.
func (b *Buffer) Unmarshal(args ...interface{}) error {
	var err error
	for i := 0; i < len(args) && err == nil; i++ {
		v := reflect.ValueOf(args[i])
		if v.Kind() != reflect.Ptr {
			return fmt.Errorf("arg of type %q must be a pointer", v.Type())
		}
		if v.Kind() == reflect.Ptr && v.IsNil() {
			return fmt.Errorf("cannot decode <nil> pointer %q", v.Type())
		}
		if v.Kind() == reflect.Invalid {
			return errors.New("cannot decode <nil> value")
		}
		v = v.Elem()
		err = b.unmarshalType(v)
	}
	b.setErr(err)
	return b.Err()
}

func (b *Buffer) unmarshalType(v reflect.Value) (err error) {
	switch v.Kind() {
	default:
		err = fmt.Errorf("cannot decode type %q", v.Type())

	case reflect.Slice:
		switch v.Type().Elem().Kind() {
		case reflect.Uint8:
			v.SetBytes(b.Bytes())
		case reflect.String, reflect.Struct:
			size := int(b.Uint16())
			elemType := v.Type().Elem()
			for i := 0; i < size; i++ {
				obj := reflect.New(elemType)
				if err = b.unmarshalType(obj.Elem()); err != nil {
					break
				}
				v.Set(reflect.Append(v, obj.Elem()))
			}
		case reflect.Ptr:
			panic("decode: pointer to slices not supported")
		}

	case reflect.Struct:
		fields := v.NumField()
		for i := 0; i < fields; i++ {
			if err = b.unmarshalType(v.Field(i)); err != nil {
				break
			}
		}

	case reflect.String:
		v.SetString(b.String())
	case reflect.Uint64:
		v.SetUint(b.Uint64())
	case reflect.Uint32:
		v.SetUint(uint64(b.Uint32()))
	case reflect.Uint16:
		v.SetUint(uint64(b.Uint16()))
	case reflect.Uint8:
		v.SetUint(uint64(b.Uint8()))
	}
	return err
}

// Marshal returns the wire-format encoding of args.
func (b *Buffer) Marshal(args ...interface{}) error {
	var err error
	for i := 0; i < len(args) && err == nil; i++ {
		v := reflect.ValueOf(args[i])
		if v.Kind() == reflect.Ptr && v.IsNil() {
			return fmt.Errorf("cannot encode <nil> pointer %q", v.Type())
		}
		if v.Kind() == reflect.Invalid {
			return errors.New("cannot encode <nil> value")
		}
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		err = b.marshalType(v)
	}
	b.setErr(err)
	return b.Err()
}

func (b *Buffer) marshalType(v reflect.Value) (err error) {
	switch v.Kind() {
	default:
		err = fmt.Errorf("cannot encode type %q", v.Type())

	case reflect.Slice:
		switch v.Type().Elem().Kind() {
		case reflect.Uint8:
			b.PutBytes(v.Bytes())
		case reflect.String, reflect.Struct:
			size := v.Len()
			b.PutUint16(uint16(size))
			for i := 0; i < size; i++ {
				if err = b.marshalType(v.Index(i)); err != nil {
					break
				}
			}
		case reflect.Ptr:
			panic("encode: pointer to slices not supported")
		}

	case reflect.Struct:
		fields := v.NumField()
		for i := 0; i < fields; i++ {
			if err = b.marshalType(v.Field(i)); err != nil {
				break
			}
		}

	case reflect.String:
		b.PutString(v.String())
	case reflect.Uint64:
		b.PutUint64(v.Uint())
	case reflect.Uint32:
		b.PutUint32(uint32(v.Uint()))
	case reflect.Uint16:
		b.PutUint16(uint16(v.Uint()))
	case reflect.Uint8:
		b.PutUint8(uint8(v.Uint()))
	}
	return err
}

// SizeOf returns the size of args encoded as 9P types and data information.
func SizeOf(args ...interface{}) (n int) {
	for _, arg := range args {
		v := reflect.ValueOf(arg)
		if v.Kind() == reflect.Invalid {
			continue
		}
		if v.Kind() == reflect.Ptr && v.IsNil() {
			continue
		}
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		n += sizeOfType(v)
	}
	return
}

func sizeOfType(v reflect.Value) (n int) {
	switch v.Kind() {
	case reflect.Slice:
		switch v.Type().Elem().Kind() {
		case reflect.Uint8: // bytes slice
			n += 4 + v.Len()
		case reflect.String, reflect.Struct:
			size := v.Len()
			n += 2
			for i := 0; i < size; i++ {
				n += sizeOfType(reflect.Indirect(v.Index(i)))
			}
		}
	case reflect.Struct:
		fields := v.NumField()
		for i := 0; i < fields; i++ {
			n += sizeOfType(v.Field(i))
		}
	case reflect.String:
		n += 2 + len(v.String())
	case reflect.Uint64:
		n += 8
	case reflect.Uint32:
		n += 4
	case reflect.Uint16:
		n += 2
	case reflect.Uint8:
		n++
	}
	return n
}
