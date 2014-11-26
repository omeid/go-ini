package ini

import (
	"errors"
	"io"
)

func Encoder(i interface{}, out io.Writer) error {
	return errors.New("Sorry. filebase/codec/ini is read only for now. :/ Pull requests welcome.")
}
