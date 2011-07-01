package nbt

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	te "testencoding"
)


type Test struct {
	Serialized te.IBytesMatcher
	Value      ITag
}

func (test *Test) String() string {
	return fmt.Sprintf("Test{Serialized=%v, Value=%#v}", test.Serialized, test.Value)
}

func (test *Test) testRead(t *testing.T, ) {
	bytesBuf := new(bytes.Buffer)
	test.Serialized.Write(bytesBuf)

	result := NewTagByType(test.Value.Type())
	err := result.Read(bytesBuf)

	t.Logf("Test read %v", test)

	if err != nil {
		t.Errorf("  Fail: failed to read a %T: %T%v", result, err, err)
		return
	}

	if !reflect.DeepEqual(test.Value, result) {
		t.Errorf("  Fail: got result = %T%v", result, result)
	}
}

func (test *Test) testWrite(t *testing.T) {
	t.Logf("Test write %v", test)

	resultBuf := new(bytes.Buffer)
	err := test.Value.Write(resultBuf)
	if err != nil {
		t.Errorf("  Fail: failed to write %#v: %T%v", test.Value, err, err)
		return
	}

	result := resultBuf.Bytes()
	n, matches := test.Serialized.Match(result)

	if !matches || n != len(result) {
		t.Errorf("  Fail: got result = %T%v", result, result)
	}
}


func TestSerialization(t *testing.T) {
	tests := []Test{
		{te.LiteralString(""), &End{}},
		{te.LiteralString("\x01"), &Byte{1}},
		{te.LiteralString("\x10\x20"), &Short{0x1020}},
		{te.LiteralString("\x10\x20\x30\x40"), &Int{0x10203040}},
		{te.LiteralString("\x10\x20\x30\x40\x50\x60\x70\x80"), &Long{0x1020304050607080}},
		{te.LiteralString("\x3f\x80\x00\x00"), &Float{1.0}},
		{te.LiteralString("\x3f\xf0\x00\x00\x00\x00\x00\x00"), &Double{1.0}},
		{te.LiteralString("\x00\x00\x00\x04\x00\x01\x02\x03"), &ByteArray{[]byte{0, 1, 2, 3}}},
		{te.LiteralString("\x00\x03foo"), &String{"foo"}},
		{te.LiteralString("\x01\x00\x00\x00\x02\x01\x02"), &List{TagByte, []ITag{&Byte{1}, &Byte{2}}}},
		{te.LiteralString("\x03\x00\x00\x00\x02\x00\x00\x00\x01\x00\x00\x00\x02"), &List{TagInt, []ITag{&Int{1}, &Int{2}}}},

		{
			// Single item Compound.
			te.InOrder(
				te.LiteralString("\x01\x00\x03foo\x01"), // NamedTag "foo" Byte{1}
				te.LiteralString("\x00"), // End
			),
			&Compound{
				map[string]*NamedTag{
					"foo": &NamedTag{"foo", &Byte{1}},
				},
			},
		},
		{
			// Multiple item Compound.
			te.InOrder(
				te.AnyOrder(
					// NamedTag "Byte" Byte{1}
					te.LiteralString("\x01\x00\x04Byte\x01"),
					// NamedTag "Short" Short{2}
					te.LiteralString("\x02\x00\x05Short\x00\x02"),
					// NamedTag "Int" Int{3}
					te.LiteralString("\x03\x00\x03Int\x00\x00\x00\x03"),
					// NamedTag "Long" Long{4}
					te.LiteralString("\x04\x00\x04Long\x00\x00\x00\x00\x00\x00\x00\x04"),
					// NamedTag "Float" Float{5}
					te.LiteralString("\x05\x00\x05Float\x40\xa0\x00\x00"),
					// NamedTag "Double" Double{6}
					te.LiteralString("\x06\x00\x06Double\x40\x18\x00\x00\x00\x00\x00\x00"),
					// NamedTag "String" String{"foo"}
					te.LiteralString("\x08\x00\x06String\x00\x03foo"),
					// NamedTag "List" List{Byte{1}, Byte{2}}
					te.LiteralString("\x09\x00\x04List\x01\x00\x00\x00\x02\x01\x02"),
				),
				te.LiteralString("\x00"), // End
			),
			&Compound{
				map[string]*NamedTag{
					"Byte": &NamedTag{"Byte", &Byte{1}},
					"Short": &NamedTag{"Short", &Short{2}},
					"Int": &NamedTag{"Int", &Int{3}},
					"Long": &NamedTag{"Long", &Long{4}},
					"Float": &NamedTag{"Float", &Float{5}},
					"Double": &NamedTag{"Double", &Double{6}},
					"String": &NamedTag{"String", &String{"foo"}},
					"List": &NamedTag{"List", &List{TagByte, []ITag{&Byte{1}, &Byte{2}}}},
				},
			},
		},
	}

	for i := range tests {
		tests[i].testRead(t)
		tests[i].testWrite(t)
	}
}
