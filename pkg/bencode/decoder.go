// from https://github.com/anacrolix/torrent/blob/master/bencode/decode.go, change panic to return error explicit

package bencode

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strconv"
	"sync"
	"unsafe"
)

const DefaultDecodeMaxStrLen = 1<<27 - 1 // ~128MiB

type MaxStrLen = int64

type Decoder struct {
	MaxStrLen MaxStrLen
	r         interface {
		io.ByteScanner
		io.Reader
	}
	Offset int64
	buf    bytes.Buffer
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: bufio.NewReader(r)}
}

func (d *Decoder) Decode(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &UnmarshalInvalidArgError{reflect.TypeOf(v)}
	}

	ok, err := d.parseValue(rv.Elem())
	if err != nil {
		return err
	}

	if !ok {
		return d.makeSyntaxError(d.Offset-1, errors.New("unexpected 'e'"))
	}

	return nil
}

func (d *Decoder) parseValue(v reflect.Value) (bool, error) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		v = v.Elem()
	}

	ok, err := d.parseUnmarshaler(v)

	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}

	if v.Kind() == reflect.Interface && v.NumMethod() == 0 {
		iface, ok, err := d.parseValueInterface()
		if err != nil {
			return false, err
		}

		v.Set(reflect.ValueOf(iface))
		return ok, nil
	}

	b, err := d.r.ReadByte()
	if err != nil {
		return false, err
	}

	d.Offset++

	switch b {
	case 'e':
		return false, nil
	case 'd':
		return true, d.parseDict(v)
	case 'l':
		return true, d.parseList(v)
	case 'i':
		return true, d.parseInt(v)
	default:
		if b >= '0' && b <= '9' {
			d.buf.Reset()
			d.buf.WriteByte(b)
			return true, d.parseString(v)
		}

		return false, d.unknowValueType(b, d.Offset-1)
	}
}

func (d *Decoder) parseUnmarshaler(v reflect.Value) (bool, error) {
	if !v.Type().Implements(unmarshalerType) {
		if v.Addr().Type().Implements(unmarshalerType) {
			v = v.Addr()
		} else {
			return false, nil
		}
	}
	d.buf.Reset()
	ok, err := d.readOneValue()

	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	m := v.Interface().(Unmarshaler)
	err = m.UnmarshalBencode(d.buf.Bytes())

	if err != nil {
		return false, err
	}

	return true, nil
}

func (d *Decoder) readOneValue() (bool, error) {
	b, err := d.r.ReadByte()

	if err != nil {
		return false, err
	}

	if b == 'e' {
		err := d.r.UnreadByte()

		if err != nil {
			return false, err
		}

		return false, nil
	} else {
		d.Offset++
		d.buf.WriteByte(b)
	}

	switch b {
	case 'd', 'l':
		// read until there is nothing to read
		for {
			ok, err := d.readOneValue()
			if err != nil {
				return false, err
			}
			if !ok {
				break
			}
		}
		// consume 'e' as well
		b, err = d.readByte()
		if err != nil {
			return false, err
		}
		d.buf.WriteByte(b)
	case 'i':
		if err := d.readUntil('e'); err != nil {
			return false, err
		}
		d.buf.WriteString("e")
	default:
		if b >= '0' && b <= '9' {
			start := d.buf.Len() - 1

			if err := d.readUntil(':'); err != nil {
				return false, err
			}

			length, err := strconv.ParseInt(bytesAsString(d.buf.Bytes()[start:]), 10, 64)

			if err := checkForIntParseError(err, d.Offset-1); err != nil {
				return false, err
			}

			d.buf.WriteString(":")
			n, err := io.CopyN(&d.buf, d.r, length)
			d.Offset += n

			if err != nil {
				return false, checkForUnexpectedEOF(err, d.Offset)
			}
			break
		}

		return false, d.unknowValueType(b, d.Offset-1)
	}

	return true, nil
}

func (d *Decoder) parseInt(v reflect.Value) error {
	start := d.Offset - 1

	if err := d.readInt(); err != nil {
		return err
	}

	s := bytesAsString(d.buf.Bytes())

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err := checkForIntParseError(err, start); err != nil {
			return err
		}

		if v.OverflowInt(n) {
			return &UnmarshalTypeError{
				BencodeTypeName:     "int",
				UnmarshalTargetType: v.Type(),
			}
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 10, 64)
		if err := checkForIntParseError(err, start); err != nil {
			return err
		}

		if v.OverflowUint(n) {
			return &UnmarshalTypeError{
				BencodeTypeName:     "int",
				UnmarshalTargetType: v.Type(),
			}
		}
		v.SetUint(n)
	case reflect.Bool:
		v.SetBool(s != "0")
	default:
		return &UnmarshalTypeError{
			BencodeTypeName:     "int",
			UnmarshalTargetType: v.Type(),
		}
	}

	d.buf.Reset()
	return nil
}

func (d *Decoder) readInt() error {
	if err := d.readUntil('e'); err != nil {
		return err
	}

	if err := d.checkBufferedInt(); err != nil {
		return err
	}

	return nil
}

func (d *Decoder) readUntil(c byte) error {
	for {
		b, err := d.readByte()

		if err != nil {
			return err
		}

		if b == c {
			return nil
		}

		d.buf.WriteByte(b)
	}
}

func (d *Decoder) readByte() (byte, error) {
	b, err := d.r.ReadByte()
	if err != nil {
		err = checkForUnexpectedEOF(err, d.Offset)
		return 0, err
	}

	d.Offset++
	return b, nil
}

func (d *Decoder) readBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	n, err := io.ReadFull(d.r, b)

	if err != nil {
		return nil, err
	}

	if n != length {
		return nil, fmt.Errorf("read %v bytes expected %v", n, length)
	}

	return b, nil
}

func (d *Decoder) checkBufferedInt() error {
	b := d.buf.Bytes()
	if len(b) <= 1 {
		return nil
	}

	if b[0] == '-' {
		b = b[1:]
	}

	if b[0] < '1' || b[0] > '9' {
		return errors.New("Invalid leading digit")
	}

	return nil
}

func (d *Decoder) parseString(v reflect.Value) error {
	length, err := d.parseStringLength()

	if err != nil {
		return err
	}

	defer d.buf.Reset()

	read := func(b []byte) error {
		n, err := io.ReadFull(d.r, b)
		d.Offset += int64(n)
		if err != nil {
			return checkForUnexpectedEOF(err, d.Offset)
		}

		return nil
	}

	switch v.Kind() {
	case reflect.String:
		b := make([]byte, length)
		if err := read(b); err != nil {
			return err
		}
		v.SetString(bytesAsString(b))
		return nil
	case reflect.Slice:
		if v.Type().Elem().Kind() != reflect.Uint8 {
			break
		}
		b := make([]byte, length)
		if err := read(b); err != nil {
			return err
		}
		v.SetBytes(b)
		return nil
	case reflect.Array:
		if v.Type().Elem().Kind() != reflect.Uint8 {
			break
		}
		d.buf.Grow(length)
		b := d.buf.Bytes()[:length]
		if err := read(b); err != nil {
			return err
		}
		reflect.Copy(v, reflect.ValueOf(b))
		return nil
	case reflect.Bool:
		d.buf.Grow(length)
		b := d.buf.Bytes()[:length]
		if err := read(b); err != nil {
			return err
		}
		x, err := strconv.ParseBool(bytesAsString(b))
		if err != nil {
			x = length != 0
		}
		v.SetBool(x)
		return nil
	}

	d.buf.Grow(length)
	b := d.buf.Bytes()[:length]
	if err := read(b); err != nil {
		return err
	}

	return &UnmarshalTypeError{
		BencodeTypeName:     "string",
		UnmarshalTargetType: v.Type(),
	}
}

func (d *Decoder) parseStringLength() (int, error) {
	start := d.Offset - 1

	if err := d.readUntil(':'); err != nil {
		return 0, err
	}

	if err := d.checkBufferedInt(); err != nil {
		return 0, err
	}

	length, err := strconv.ParseInt(bytesAsString(d.buf.Bytes()), 10, 0)
	if err := checkForIntParseError(err, start); err != nil {
		return 0, err
	}

	if int64(length) > d.getMaxStrLen() {
		err = fmt.Errorf("parsed string length %v exceeds limit [%v]", length, DefaultDecodeMaxStrLen)
	}
	d.buf.Reset()
	return int(length), err
}

func (d *Decoder) getMaxStrLen() int64 {
	if d.MaxStrLen == 0 {
		return DefaultDecodeMaxStrLen
	}
	return d.MaxStrLen
}

func (d *Decoder) parseList(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
	default:
		l := reflect.New(reflect.SliceOf(v.Type()))
		if err := d.parseList(l.Elem()); err != nil {
			return err
		}
		if l.Elem().Len() != 1 {
			return &UnmarshalTypeError{
				BencodeTypeName:     "list",
				UnmarshalTargetType: v.Type(),
			}
		}
		v.Set(l.Elem().Index(0))
		return nil
	}

	i := 0

	for ; ; i++ {
		if v.Kind() == reflect.Slice && i >= v.Len() {
			v.Set(reflect.Append(v, reflect.Zero(v.Type().Elem())))
		}

		if i < v.Len() {
			ok, err := d.parseValue(v.Index(i))
			if err != nil {
				return err
			}

			// leave invalid values
			if !ok {
				break
			}
		} else {
			_, ok, err := d.parseValueInterface()
			if err != nil {
				return err
			}

			if !ok {
				break
			}
		}
	}

	// fulfill remaining element with default value
	if i < v.Len() {
		if v.Kind() == reflect.Array {
			z := reflect.Zero(v.Type().Elem())
			for n := v.Len(); i < n; i++ {
				v.Index(i).Set(z)
			}
		} else {
			v.SetLen(i)
		}
	}

	// make empty slice
	if i == 0 && v.Kind() == reflect.Slice {
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	}

	return nil
}

func (d *Decoder) parseDict(v reflect.Value) error {
	keyType := keyType(v)

	if keyType == nil {
		return fmt.Errorf("cannot parse dicts into %v", v.Type())
	}

	for {
		keyValue := reflect.New(keyType).Elem()
		ok, err := d.parseValue(keyValue)

		if err != nil {
			return fmt.Errorf("error parsing dict key: %w", err)
		}
		if !ok {
			return nil
		}

		df, err := getDictField(v.Type(), keyValue)
		if err != nil {
			return fmt.Errorf("parsing bencode dict into %v: %w", v.Type(), err)
		}

		if df.Type == nil {
			_, ok, err = d.parseValueInterface()
			if err != nil {
				return err
			}

			if !ok {
				return fmt.Errorf("missing value for key %q", keyValue)
			}

			continue
		}

		setValue := reflect.New(df.Type).Elem()
		ok, err = d.parseValue(setValue)
		if err != nil {
			var target *UnmarshalTypeError
			if !(errors.As(err, &target) && df.Tags.IgnoreUnmarshalTypeError()) {
				return fmt.Errorf("parsing value for key %q: %w", keyValue, err)
			}
		}

		if !ok {
			return fmt.Errorf("missing value for key %q", keyValue)
		}
		df.Get(v)(setValue)
	}
}

var structKeyType = reflect.TypeFor[string]()

func keyType(v reflect.Value) reflect.Type {
	switch v.Kind() {
	case reflect.Map:
		return v.Type().Key()
	case reflect.Struct:
		return structKeyType
	default:
		return nil
	}
}

type dictField struct {
	Type reflect.Type
	Get  func(value reflect.Value) func(reflect.Value)
	Tags tag
}

func getDictField(dict reflect.Type, key reflect.Value) (_ dictField, err error) {
	switch k := dict.Kind(); k {
	case reflect.Map:
		return dictField{
			Type: dict.Elem(),
			Get: func(mapValue reflect.Value) func(reflect.Value) {
				return func(value reflect.Value) {
					if mapValue.IsNil() {
						mapValue.Set(reflect.MakeMap(dict))
					}
					mapValue.SetMapIndex(key, value)
				}
			},
		}, nil
	case reflect.Struct:
		if key.Kind() != reflect.String {
			return dictField{}, errors.New("struct keys must be strings")
		}
		return getStructFieldForKey(dict, key.String()), nil
	default:
		err = fmt.Errorf("cannot assign bencode dict items into a %v", k)
		return
	}
}

var (
	structFieldsMu sync.Mutex
	structFields   = map[reflect.Type]map[string]dictField{}
)

func parseStructFields(struct_ reflect.Type, each func(key string, df dictField)) {
	for _i, n := 0, struct_.NumField(); _i < n; _i++ {
		i := _i
		f := struct_.Field(i)

		if f.Anonymous {
			t := f.Type
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			parseStructFields(t, func(key string, df dictField) {
				innerGet := df.Get
				df.Get = func(value reflect.Value) func(reflect.Value) {
					anonPtr := value.Field(1)
					if anonPtr.Kind() == reflect.Ptr && anonPtr.IsNil() {
						anonPtr.Set(reflect.New(f.Type.Elem()))
						anonPtr = anonPtr.Elem()
					}

					return innerGet(anonPtr)
				}
				each(key, df)
			})
		}

		tagStr := f.Tag.Get("bencode")
		if tagStr == "-" {
			continue
		}

		tag := parseTag(tagStr)
		key := tag.Key()
		if key == "" {
			key = f.Name
		}

		each(key, dictField{f.Type, func(value reflect.Value) func(reflect.Value) {
			return value.Field(i).Set
		}, tag})
	}
}

func saveStructFields(struct_ reflect.Type) {
	m := make(map[string]dictField)
	parseStructFields(struct_, func(key string, df dictField) {
		m[key] = df
	})
	structFields[struct_] = m
}

func getStructFieldForKey(struct_ reflect.Type, key string) (f dictField) {
	structFieldsMu.Lock()
	if _, ok := structFields[struct_]; !ok {
		saveStructFields(struct_)
	}
	f, ok := structFields[struct_][key]
	structFieldsMu.Unlock()

	if !ok {
		var discard interface{}
		return dictField{
			Type: reflect.TypeOf(discard),
			Get:  func(reflect.Value) func(reflect.Value) { return func(reflect.Value) {} },
			Tags: nil,
		}
	}

	return
}

func (d *Decoder) makeSyntaxError(offset int64, err error) *SyntaxError {
	return &SyntaxError{
		Offset: offset,
		Err:    err,
	}
}

func (d *Decoder) unknowValueType(b byte, offset int64) *SyntaxError {
	return &SyntaxError{
		Offset: offset,
		Err:    fmt.Errorf("bencode: unknow value type %+q", b),
	}
}

func (d *Decoder) parseValueInterface() (interface{}, bool, error) {
	b, err := d.r.ReadByte()
	if err != nil {
		return nil, false, err
	}

	d.Offset++

	switch b {
	case 'e':
		return nil, false, nil
	case 'd':
		val, err := d.parseDictInterface()
		return val, true, err
	case 'l':
		val, err := d.parseListInterface()
		return val, true, err
	case 'i':
		val, err := d.parseIntInterface()
		return val, true, err
	default:
		if b >= '0' && b <= '9' {
			d.buf.Reset()
			d.buf.WriteByte(b)
			val, err := d.parseStringInterface()
			return val, true, err
		}

		return nil, false, d.unknowValueType(b, d.Offset-1)
	}
}

func (d *Decoder) parseIntInterface() (ret interface{}, err error) {
	start := d.Offset - 1

	if err := d.readInt(); err != nil {
		return nil, err
	}
	n, err := strconv.ParseInt(d.buf.String(), 10, 64)
	if ne, ok := err.(*strconv.NumError); ok && ne.Err == strconv.ErrRange {
		i := new(big.Int)
		_, ok := i.SetString(d.buf.String(), 10)
		if !ok {
			return nil, &SyntaxError{
				Offset: start,
				Err:    errors.New("failed to parse integer"),
			}
		}
		ret = i
	} else {
		if err := checkForIntParseError(err, start); err != nil {
			return nil, err
		}
		ret = n
	}

	d.buf.Reset()
	return ret, nil
}

func (d *Decoder) parseStringInterface() (string, error) {
	length, err := d.parseStringLength()

	if err != nil {
		return "", err
	}

	b, err := d.readBytes(int(length))

	if err != nil {
		return "", &SyntaxError{Offset: d.Offset, Err: err}
	}

	d.Offset += int64(len(b))
	return bytesAsString(b), nil
}

func (d *Decoder) parseListInterface() (list []interface{}, err error) {
	list = []interface{}{}

	for {
		valuei, ok, err := d.parseValueInterface()
		if err != nil {
			return nil, err
		}

		if !ok {
			break
		}

		list = append(list, valuei)
	}
	return list, nil
}

func (d *Decoder) parseDictInterface() (interface{}, error) {
	dict := make(map[string]interface{})
	var lastKey string
	lastKeyOk := false

	for {
		start := d.Offset
		keyi, ok, err := d.parseValueInterface()

		if err != nil {
			return nil, err
		}

		if !ok {
			break
		}

		key, ok := keyi.(string)

		if !ok {
			return nil, &SyntaxError{
				Offset: d.Offset,
				Err:    errors.New("non-string key in a dict"),
			}
		}

		if lastKeyOk && key <= lastKey {
			return nil, d.makeSyntaxError(start, fmt.Errorf("dict keys unsorted: %q <= %q", key, lastKey))
		}

		start = d.Offset
		valuei, ok, err := d.parseValueInterface()

		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, d.makeSyntaxError(start, fmt.Errorf("dict elem missing value [key=%v]", key))
		}

		lastKey = key
		lastKeyOk = true
		dict[key] = valuei
	}

	return dict, nil
}

func (d *Decoder) ReadEOF() error {
	_, err := d.r.ReadByte()
	if err == nil {
		err := d.r.UnreadByte()
		if err != nil {
			return err
		}

		return errors.New("expected EOF")
	}

	if err == io.EOF {
		return nil
	}

	return fmt.Errorf("expected EOF, got %w", err)
}

func checkForIntParseError(err error, offset int64) error {
	if err != nil {
		return &SyntaxError{
			Offset: offset,
			Err:    err,
		}
	}

	return nil
}

func checkForUnexpectedEOF(err error, offset int64) error {
	if err == io.EOF {
		return &SyntaxError{
			Offset: offset,
			Err:    io.ErrUnexpectedEOF,
		}
	}
	return err
}

func bytesAsString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
