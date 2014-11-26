package ini

import (
	"bufio"
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
	o := reflect.ValueOf(out).Elem()
	cur := &o
	line := 0
	for scanner.Scan() {
		line++
		parts := syntaxregex.FindStringSubmatch(scanner.Text())
		var (
			section = parts[1] != ""
			key     = parts[2]
			value   = parts[3]
			comment = parts[4] != ""
			empty   = parts[5] == "" && !section && key == ""
		)

		if empty || comment {
			continue
		}

		if (!section && key == "") || scanner.Err() != nil {
			return fmt.Errorf("Invalid Syntax at line %d: \"%s\"", line, scanner.Text())
		}
		if section {
			cur = &o
			key = parts[1]
		}

		f := cur.FieldByName(key)
		if !f.IsValid() || !f.CanSet() {
			continue
		}

		if section {
			cur = &f
			continue
		}

		if setter, ok := setters[f.Kind()]; ok {
			err := setter(f, value)
			if err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}

type setter func(v reflect.Value, str string) error

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

func setString(f reflect.Value, str string) error {
	f.SetString(str)
	return nil
}

func setBool(f reflect.Value, str string) error {
	f.SetBool(str == "Yes" || str == "On" || str == "True")
	return nil
}

func setint(f reflect.Value, str string) error {
	str = strings.Replace(str, ",", "", -1)
	i64, err := strconv.ParseInt(str, 10, 0)
	if err != nil {
		return err
	}

	f.SetInt(i64)
	return nil
}

func setSlice(f reflect.Value, str string) error {

	parts := strings.Fields(str)
	l := len(parts)

	set, ok := setters_short[f.Type().Elem().Kind()]
	if !ok {
		return fmt.Errorf("I don't understand type %s.", f.Kind())
	}

	f.Set(reflect.MakeSlice(f.Type(), l, l))

	for i, str := range parts {
		_, _, _ = set, i, str
		set(f.Index(i), str)

	}
	return nil
}

func setMap(f reflect.Value, str string) error {
	return fmt.Errorf("I don't understand type %s.", f.Kind())
}
