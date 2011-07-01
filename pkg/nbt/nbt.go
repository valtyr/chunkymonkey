package nbt

import (
	"os"
	"io"
	"fmt"
	"strings"
	"encoding/binary"
)

const (
	// Tag types
	TagEnd       = 0
	TagByte      = 1
	TagShort     = 2
	TagInt       = 3
	TagLong      = 4
	TagFloat     = 5
	TagDouble    = 6
	TagByteArray = 7
	TagString    = 8
	TagList      = 9
	TagCompound  = 10
	TagNamed     = 0x80
)

type ITag interface {
	Type() byte
	Read(io.Reader) os.Error
	Write(io.Writer) os.Error
	Lookup(path string) ITag
}

func NewTagByType(tagType byte) (tag ITag) {
	switch tagType {
	case TagEnd:
		tag = new(End)
	case TagByte:
		tag = new(Byte)
	case TagShort:
		tag = new(Short)
	case TagInt:
		tag = new(Int)
	case TagLong:
		tag = new(Long)
	case TagFloat:
		tag = new(Float)
	case TagDouble:
		tag = new(Double)
	case TagByteArray:
		tag = new(ByteArray)
	case TagString:
		tag = new(String)
	case TagList:
		tag = new(List)
	case TagCompound:
		tag = new(Compound)
	default:
		// TODO Don't panic, produce an error.
		panic(fmt.Sprintf("Invalid NBT tag type %#x", tagType))
	}
	return
}

type End struct{}

func (end *End) Type() byte {
	return TagEnd
}

func (end *End) Read(io.Reader) os.Error {
	return nil
}

func (end *End) Write(io.Writer) os.Error {
	return nil
}

func (end *End) Lookup(path string) ITag {
	return nil
}

type NamedTag struct {
	Name string
	Tag  ITag
}

func (n *NamedTag) Type() byte {
	return TagNamed | n.Tag.Type()
}

func (n *NamedTag) Read(reader io.Reader) (err os.Error) {
	var tagType byte
	err = binary.Read(reader, binary.BigEndian, &tagType)
	if err != nil {
		return
	}

	var name String
	if tagType != TagEnd {
		err = name.Read(reader)
		if err != nil {
			return
		}
	}

	var value = NewTagByType(tagType)
	err = value.Read(reader)
	if err != nil {
		return
	}

	n.Name = name.Value
	n.Tag = value
	return
}

func (n *NamedTag) Write(writer io.Writer) (err os.Error) {
	if err = binary.Write(writer, binary.BigEndian, n.Tag.Type()); err != nil {
		return
	}

	name := String{n.Name}
	if err = name.Write(writer); err != nil {
		return
	}

	return n.Tag.Write(writer)
}

func (n *NamedTag) Lookup(path string) ITag {
	components := strings.Split(path, "/", 2)
	if components[0] != n.Name {
		return nil
	}

	if len(components) == 1 {
		return n.Tag
	}

	return n.Tag.Lookup(components[1])
}

type Byte struct {
	Value int8
}

func (*Byte) Type() byte {
	return TagByte
}

func (*Byte) Lookup(path string) ITag {
	return nil
}

func (b *Byte) Read(reader io.Reader) (err os.Error) {
	return binary.Read(reader, binary.BigEndian, &b.Value)
}

func (b *Byte) Write(writer io.Writer) (err os.Error) {
	return binary.Write(writer, binary.BigEndian, &b.Value)
}

type Short struct {
	Value int16
}

func (*Short) Type() byte {
	return TagShort
}

func (s *Short) Read(reader io.Reader) (err os.Error) {
	return binary.Read(reader, binary.BigEndian, &s.Value)
}

func (s *Short) Write(writer io.Writer) (err os.Error) {
	return binary.Write(writer, binary.BigEndian, &s.Value)
}

func (*Short) Lookup(path string) ITag {
	return nil
}

type Int struct {
	Value int32
}

func (*Int) Type() byte {
	return TagInt
}

func (i *Int) Read(reader io.Reader) (err os.Error) {
	return binary.Read(reader, binary.BigEndian, &i.Value)
}

func (i *Int) Write(writer io.Writer) (err os.Error) {
	return binary.Write(writer, binary.BigEndian, &i.Value)
}

func (*Int) Lookup(path string) ITag {
	return nil
}

type Long struct {
	Value int64
}

func (*Long) Type() byte {
	return TagLong
}

func (l *Long) Read(reader io.Reader) (err os.Error) {
	return binary.Read(reader, binary.BigEndian, &l.Value)
}

func (l *Long) Write(writer io.Writer) (err os.Error) {
	return binary.Write(writer, binary.BigEndian, &l.Value)
}

func (*Long) Lookup(path string) ITag {
	return nil
}

type Float struct {
	Value float32
}

func (*Float) Type() byte {
	return TagFloat
}

func (f *Float) Read(reader io.Reader) (err os.Error) {
	return binary.Read(reader, binary.BigEndian, &f.Value)
}

func (f *Float) Write(writer io.Writer) (err os.Error) {
	return binary.Write(writer, binary.BigEndian, &f.Value)
}

func (*Float) Lookup(path string) ITag {
	return nil
}

type Double struct {
	Value float64
}

func (*Double) Type() byte {
	return TagDouble
}

func (d *Double) Read(reader io.Reader) (err os.Error) {
	return binary.Read(reader, binary.BigEndian, &d.Value)
}

func (d *Double) Write(writer io.Writer) (err os.Error) {
	return binary.Write(writer, binary.BigEndian, &d.Value)
}

func (*Double) Lookup(path string) ITag {
	return nil
}

type ByteArray struct {
	Value []byte
}

func (*ByteArray) Type() byte {
	return TagByteArray
}

func (b *ByteArray) Read(reader io.Reader) (err os.Error) {
	var length Int

	err = length.Read(reader)
	if err != nil {
		return
	}

	bs := make([]byte, length.Value)
	_, err = io.ReadFull(reader, bs)
	if err != nil {
		return
	}

	b.Value = bs
	return
}

func (b *ByteArray) Write(writer io.Writer) (err os.Error) {
	length := Int{int32(len(b.Value))}

	if err = length.Write(writer); err != nil {
		return
	}

	_, err = writer.Write(b.Value)
	return
}

func (*ByteArray) Lookup(path string) ITag {
	return nil
}

type String struct {
	Value string
}

func (*String) Type() byte {
	return TagString
}

func (s *String) Read(reader io.Reader) (err os.Error) {
	var length Short

	err = length.Read(reader)
	if err != nil {
		return
	}

	bs := make([]byte, length.Value)
	_, err = io.ReadFull(reader, bs)
	if err != nil {
		return
	}

	s.Value = string(bs)
	return
}

func (s *String) Write(writer io.Writer) (err os.Error) {
	length := Short{int16(len(s.Value))}

	if err = length.Write(writer); err != nil {
		return
	}

	_, err = writer.Write([]byte(s.Value))
	return
}

func (*String) Lookup(path string) ITag {
	return nil
}

type List struct {
	TagType byte
	Value   []ITag
}

func (*List) Type() byte {
	return TagList
}

func (l *List) Read(reader io.Reader) (err os.Error) {
	var tagType Byte
	err = tagType.Read(reader)
	if err != nil {
		return
	}

	l.TagType = byte(tagType.Value)

	var length Int
	err = length.Read(reader)
	if err != nil {
		return
	}

	list := make([]ITag, length.Value)
	for i, _ := range list {
		tag := NewTagByType(byte(tagType.Value))
		err = tag.Read(reader)
		if err != nil {
			return
		}

		list[i] = tag
	}

	l.Value = list
	return
}

func (l *List) Write(writer io.Writer) (err os.Error) {
	tagType := Byte{int8(l.TagType)}
	if err = tagType.Write(writer); err != nil {
		return
	}

	length := Int{int32(len(l.Value))}
	if err = length.Write(writer); err != nil {
		return
	}

	for _, tag := range l.Value {
		if err = tag.Write(writer); err != nil {
			return
		}
	}

	return
}

func (*List) Lookup(path string) ITag {
	return nil
}

type Compound struct {
	Tags map[string]*NamedTag
}

func (*Compound) Type() byte {
	return TagCompound
}

func (c *Compound) Read(reader io.Reader) (err os.Error) {
	tags := make(map[string]*NamedTag)
	for {
		tag := &NamedTag{}
		err = tag.Read(reader)
		if err != nil {
			return
		}

		if tag.Type() == TagNamed|TagEnd {
			break
		}

		tags[tag.Name] = tag
	}

	c.Tags = tags
	return
}

func (c *Compound) Write(writer io.Writer) (err os.Error) {
	for _, tag := range c.Tags {
		if err = tag.Write(writer); err != nil {
			return
		}
	}

	return binary.Write(writer, binary.BigEndian, byte(TagEnd))
}

func (c *Compound) Lookup(path string) (tag ITag) {
	components := strings.Split(path, "/", 2)
	tag, ok := c.Tags[components[0]]
	if !ok {
		return nil
	}

	return tag.Lookup(path)
}

func Read(reader io.Reader) (compound *NamedTag, err os.Error) {
	compound = &NamedTag{}
	err = compound.Read(reader)
	if err != nil {
		return
	}

	if compound.Type() != TagNamed|TagCompound {
		return nil, os.NewError("Expected named compound tag")
	}
	return
}
