package bencode

import "bytes"

func Marshal(v interface{}) ([]byte, error) {
  var buf bytes.Buffer
  e := Encoder{w: &buf}
  err := e.Encode(v)
  if err != nil {
    return nil, err
  }

  return buf.Bytes(), nil
}

func Unmarshal(data []byte, v interface{}) (err error) {
  buf := bytes.NewReader(data)
  decoder := Decoder{r: buf}
  err = decoder.Decode(v)
  
  if err != nil {
    return
  }

  if buf.Len() != 0 {
    return ErrUnusedTrailingBytes{buf.Len()}
  }

  return decoder.ReadEOF()
}
