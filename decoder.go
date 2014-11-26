package ini

import (
	"bufio"
	"encoding"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Decoder
var syntaxregex = regexp.MustCompile(`^(?:\[(?P<section>[a-zA-Z-]+)\]\s*$)|(?:(?P<key>[a-zA-Z]+)=(?P<value>.*))|(?:#(?P<comment>.*))|(?P<bad>[^\s]*)$`)

func Decoder(i io.Reader, out interface{}) error {

	scanner := bufio.NewScanner(i)

	scanner = bufio.NewScanner(i)
	o := reflect.ValueOf(out)
	cur := &o
	line := 0
	for scanner.Scan() {
		line++
		parts := syntaxregex.FindSubmatch(scanner.Bytes())
		var (
			section = parts[1] != nil
			key     = parts[2]
			value   = parts[3]
			comment = parts[4] != nil
			empty   = len(parts[5]) == 0 && !section && key == nil
		)

		if empty || comment {
			continue
		}

		if (!section && key == nil) || scanner.Err() != nil {
			return fmt.Errorf("Invalid Syntax at line %d: \"%s\"", line, scanner.Bytes())
		}
		if section {
			cur = &o
			key = parts[1]
		}

		f := reflect.Indirect(*cur).FieldByName(string(key))
		if !f.IsValid() || !f.CanSet() {
			continue
		}

		if section {
			cur = &f
			continue
		}

		if cur.CanAddr() {
			if u, ok := cur.Addr().Interface().(encoding.BinaryUnmarshaler); ok {
				err := u.UnmarshalBinary(nil)
				if err != nil {
					return err
				}
				continue
			}
		}

		if setter, ok := setters[f.Kind()]; ok {
			err := setter(f, value)
			if err != nil {
				return err
			}
			continue
		}

	}
	return scanner.Err()
}

type setter func(v reflect.Value, val []byte) error

var (
	setters_short = map[reflect.Kind]setter{
		reflect.String:  setString,
		reflect.Bool:    setBool,
		reflect.Int:     setint,
		reflect.Int8:    setint,
		reflect.Int16:   setint,
		reflect.Int32:   setint,
		reflect.Int64:   setint,
		reflect.Uint:    setint,
		reflect.Uint8:   setint,
		reflect.Uint16:  setint,
		reflect.Uint32:  setint,
		reflect.Uint64:  setint,
		reflect.Uintptr: setint}

	setters = setters_short
)

func init() {
	setters[reflect.Slice] = setSlice
	setters[reflect.Map] = setMap
}

func setString(f reflect.Value, val []byte) error {
	f.SetString(string(val))
	return nil
}

func setBool(f reflect.Value, val []byte) error {
	str := string(val)
	f.SetBool(str == "Yes" || str == "On" || str == "True")
	return nil
}

func setint(f reflect.Value, val []byte) error {
	str := strings.Replace(string(val), ",", "", -1)
	i64, err := strconv.ParseInt(str, 10, 0)
	if err != nil {
		return err
	}

	f.SetInt(i64)
	return nil
}

func setSlice(f reflect.Value, val []byte) error {

	parts := strings.Fields(string(val))
	l := len(parts)

	set, ok := setters_short[f.Type().Elem().Kind()]
	if !ok {
		return fmt.Errorf("I don't understand type %s.", f.Kind())
	}

	f.Set(reflect.MakeSlice(f.Type(), l, l))

	for i, part := range parts {
		set(f.Index(i), []byte(part))

	}
	return nil
}

func setMap(f reflect.Value, val []byte) error {
	return fmt.Errorf("I don't understand type %s.", f.Kind())
}
