package bencode

import "bytes"

func Marshal(v interface{}) ([]byte, error) {
  return nil, nil
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
