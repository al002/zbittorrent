package bencode

import (
	"io"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"sync"
)

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(v interface{}) error {
	if v == nil {
		return nil
	}

	return e.encodeValue(reflect.ValueOf(v))
}

var bigIntType = reflect.TypeOf((*big.Int)(nil)).Elem()

func (e *Encoder) encodeValue(v reflect.Value) error {
	if v.Type() == bigIntType {
		if err := e.writeString("i"); err != nil {
			return err
		}

		bi := v.Interface().(big.Int)
		if err := e.writeString(bi.String()); err != nil {
			return err
		}

		return e.writeString("e")
	}

	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return e.writeString("i1e")
		} else {
			return e.writeString("i0e")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s := "i" + strconv.FormatInt(v.Int(), 10) + "e"
		return e.writeString(s)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s := "i" + strconv.FormatUint(v.Uint(), 10) + "e"
		return e.writeString(s)
	case reflect.String:
		return e.encodeString(v.String())
	case reflect.Struct:
		if err := e.writeString("d"); err != nil {
			return err
		}

		for _, ef := range getEncodeFields(v.Type()) {
			fieldValue := ef.i(v)
			if !fieldValue.IsValid() {
				continue
			}

			if ef.omitEmpty && isEmptyValue(fieldValue) {
				continue
			}

			if err := e.encodeString(ef.tag); err != nil {
				return err
			}
			if err := e.encodeValue(fieldValue); err != nil {
				return err
			}
		}

		return e.writeString("e")
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return &MarshalTypeError{v.Type()}
		}

		if v.IsNil() {
			return e.writeString("de")
		}

		if err := e.writeString("d"); err != nil {
			return err
		}

		sv := stringValues(v.MapKeys())
		sort.Sort(sv)
		for _, key := range sv {
			if err := e.encodeString(key.String()); err != nil {
				return err
			}
			if err := e.encodeValue(v.MapIndex(key)); err != nil {
				return err
			}
		}
		return e.writeString("e")
	case reflect.Slice, reflect.Array:
		return e.encodeSequence(v)
	case reflect.Interface:
		return e.encodeValue(v.Elem())
	case reflect.Ptr:
		if v.IsNil() {
			v = reflect.Zero(v.Type().Elem())
		} else {
			v = v.Elem()
		}

		return e.encodeValue(v)
	default:
		return &MarshalTypeError{
			v.Type(),
		}
	}
}

func (e *Encoder) encodeString(s string) error {
	if err := e.writeStringPrefix(int64(len(s))); err != nil {
		return err
	}

	return e.writeString(s)
}

func (e *Encoder) encodeSequence(v reflect.Value) error {
	if v.Type().Elem().Kind() == reflect.Uint8 {
		if v.Kind() != reflect.Slice {
			if !v.CanAddr() {
				if err := e.writeStringPrefix(int64(v.Len())); err != nil {
					return err
				}

				for i := 0; i < v.Len(); i++ {
					var b [1]byte
					b[0] = byte(v.Index(i).Uint())

					if err := e.write(b[:]); err != nil {
						return err
					}
				}

				return nil
			}

			v = v.Slice(0, v.Len())
		}

		return e.encodeByteSlice(v.Bytes())
	}

	if v.IsNil() {
		return e.writeString("le")
	}

	if err := e.writeString("l"); err != nil {
		return err
	}

	for i, n := 0, v.Len(); i < n; i++ {
		if err := e.encodeValue(v.Index(i)); err != nil {
			return err
		}
	}

	return e.writeString("e")
}

func (e *Encoder) encodeByteSlice(s []byte) error {
	if err := e.writeStringPrefix(int64(len(s))); err != nil {
		return err
	}

	return e.write(s)
}

func (e *Encoder) write(s []byte) error {
	_, err := e.w.Write(s)
	return err
}

func (e *Encoder) writeString(s string) error {
	_, err := io.WriteString(e.w, s)
	return err
}

func (e *Encoder) writeStringPrefix(l int64) error {
	s := strconv.FormatInt(l, 10)
	if err := e.writeString(s); err != nil {
		return err
	}
	return e.writeString(":")
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true

		for i := 0; i < v.Len(); i++ {
			z = z && isEmptyValue(v.Index(i))
		}

		return z
	case reflect.Struct:
		z := true
		vType := v.Type()

		for i := 0; i < v.NumField(); i++ {
			// ignore unexported fields to avoid reflection panics
			if !vType.Field(i).IsExported() {
				continue
			}
			z = z && isEmptyValue(v.Field(i))
		}

		return z
	}
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	return v.Interface() == z.Interface()
}

type stringValues []reflect.Value

func (sv stringValues) Len() int {
	return len(sv)
}

func (sv stringValues) Swap(i, j int) {
	sv[i], sv[j] = sv[j], sv[i]
}

func (sv stringValues) Less(i, j int) bool {
	return sv.get(i) < sv.get(j)
}

func (sv stringValues) get(i int) string {
	return sv[i].String()
}

type encodeField struct {
	i         func(v reflect.Value) reflect.Value
	tag       string
	omitEmpty bool
}

type encodeFieldsSortType []encodeField

func (ef encodeFieldsSortType) Len() int {
	return len(ef)
}

func (ef encodeFieldsSortType) Swap(i, j int) {
	ef[i], ef[j] = ef[j], ef[i]
}

func (ef encodeFieldsSortType) Less(i, j int) bool {
	return ef[i].tag < ef[i].tag
}

var (
	typeCacheLock     sync.RWMutex
	encodeFieldsCache = make(map[reflect.Type][]encodeField)
)

func getEncodeFields(t reflect.Type) []encodeField {
	typeCacheLock.RLock()
	fs, ok := encodeFieldsCache[t]
	typeCacheLock.RUnlock()

	if ok {
		return fs
	}

	fs = makeEncodeFields(t)
	typeCacheLock.Lock()
	defer typeCacheLock.Unlock()
	encodeFieldsCache[t] = fs

	return fs
}

func makeEncodeFields(t reflect.Type) (fs []encodeField) {
	for _i, n := 0, t.NumField(); _i < n; _i++ {
		i := _i
		f := t.Field(i)
		// only exported field
		if f.PkgPath != "" {
			continue
		}

		if f.Anonymous {
			t := f.Type
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			anonEFs := makeEncodeFields(t)
			for aefi := range anonEFs {
				anonEF := anonEFs[aefi]
				bottomField := anonEF
				bottomField.i = func(v reflect.Value) reflect.Value {
					v = v.Field(i)
					if v.Kind() == reflect.Ptr {
						if v.IsNil() {
							// This will skip serializing this value.
							return reflect.Value{}
						}
						v = v.Elem()
					}
					return anonEF.i(v)
				}
				fs = append(fs, bottomField)
			}
			continue
		}

		var ef encodeField

		ef.i = func(v reflect.Value) reflect.Value {
			return v.Field(i)
		}
		ef.tag = f.Name

		tv := getTag(f.Tag)

		if tv.Ignore() {
			continue
		}

		if tv.Key() != "" {
			ef.tag = tv.Key()
		}

		ef.omitEmpty = tv.OmitEmpty()
		fs = append(fs, ef)
	}

	fss := encodeFieldsSortType(fs)
	sort.Sort(fss)
	return fs
}
