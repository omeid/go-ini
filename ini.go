package ini

import "io"

func NewDecoder(r io.Reader) decoder {
	return decoder{r: r}
}

type decoder struct {
	r io.Reader
}

func (i decoder) Decode(v interface{}) error {
	return Decoder(i.r, v)
}

func NewEncoder(w io.Writer) encoder {
	return encoder{w: w}
}

type encoder struct {
	w io.Writer
}

func (i encoder) Encode(v interface{}) error {
	return Encoder(v, i.w)
}
