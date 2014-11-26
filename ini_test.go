package ini

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

type testType struct {
	Name        string
	Description string
	After       []string
	Boot        bool
	Service     testTypeSection
}

type testTypeSection struct {
	Type             string
	User             string
	Group            string
	WorkingDirectory string
	ExecStart        string
	Wait             string
	Retry            int
	WaitTry          []int
}

var (
	test_input = `[Unit]
Description=Gogs (Go Git Service) server daemon
After=syslog.target network.target #mysqld.service postgresql.service memcached.service redis.service
Boot=Yes
[Service]
Type=simple
User=git
Group=git
WorkingDirectory=/home/git/gogs
ExecStart=/home/git/gogs/start.sh

Wait=23s
Retry=234
WaitTry=1 3 5 12 30 120 1,300
`

	test_bad_input = `[Unit]
Okay=pass
// bad line
`

	reader     = bytes.NewBuffer([]byte(test_input))
	bad_reader = bytes.NewBuffer([]byte(test_bad_input))

	test_expect = testType{
		"",
		"Gogs (Go Git Service) server daemon",
		[]string{"syslog.target", "network.target", "#mysqld.service", "postgresql.service", "memcached.service", "redis.service"},
		true,
		testTypeSection{
			"simple",
			"git",
			"git",
			"/home/git/gogs",
			"/home/git/gogs/start.sh",
			"23s",
			234,
			[]int{1, 3, 5, 12, 30, 120, 1300},
		}}
)

func TestDecoder(t *testing.T) {

	result := testType{}
	err := Decoder(reader, &result)

	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, test_expect) {
		t.Fatalf("BAD.\nGot:      %+v\nExpected: %+v\n", result, test_expect)
	}

	result = testType{}
	err = Decoder(bad_reader, &result)
	if err == nil {
		t.Fatal("Expected Error. Passed BAD:\n%s\n:ENDBAD", test_bad_input)
	}
	fmt.Printf("GOOD.\nGot:      %+v\n\nExpected: %+v\n", result, test_expect)
}
