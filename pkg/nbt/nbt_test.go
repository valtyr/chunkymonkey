package nbt

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type Test struct {
	serialized string
	value      ITag
}

func (test *Test) String() string {
	return fmt.Sprintf("Test{serialized=%q, value=%#v}", test.serialized, test.value)
}

func (test *Test) testRead(t *testing.T) {
	result := NewTagByType(test.value.GetType())
	err := result.Read(strings.NewReader(test.serialized))

	t.Logf("Test %v", test)

	if err != nil {
		t.Errorf("  Fail: failed to read a %T: %T%v", result, err, err)
		return
	}

	if !reflect.DeepEqual(test.value, result) {
		t.Errorf("  Fail: got result = %T%v", result, result)
	}
}


func TestSerialization(t *testing.T) {
	tests := []Test{
		{"", &End{}},
		{"\x01", &Byte{1}},
		{"\x10\x20", &Short{0x1020}},
		{"\x10\x20\x30\x40", &Int{0x10203040}},
		{"\x10\x20\x30\x40\x50\x60\x70\x80", &Long{0x1020304050607080}},
	}

	for i := range tests {
		tests[i].testRead(t)
	}
}
