// The testencoding package is used in testing serialization and
// deserialization. It is particularly useful when some sub-sequences of bytes
// can acceptably be written in any order (e.g if they are generated from
// representations that do not provide order guaruntees like maps).
package testencoding

import (
	"bytes"
)


// IBytesMatcher is the interface that tests if a sequence of bytes matches
// expectations.
type IBytesMatcher interface {
	// Match tests if the given sequence of bytes matches. It returns
	// matches=true if it matches correctly, and returns the number of bytes
	// matched as n.
	Match(b []byte) (n int, matches bool)

	// Write writes a sequence of bytes to the buffer that is an acceptable match
	// for the matcher..
	Write(writer *bytes.Buffer)
}


// BytesLiteral matches a literal sequence of bytes.
type BytesLiteral struct {
	Bytes []byte
}

func LiteralString(s string) *BytesLiteral {
	return &BytesLiteral{[]byte(s)}
}

func (bm *BytesLiteral) Match(b []byte) (n int, matches bool) {
	if len(b) < len(bm.Bytes) {
		return 0, false
	}

	for i, v := range bm.Bytes {
		if v != b[i] {
			return 0, false
		}
	}

	return len(bm.Bytes), true
}

func (bm *BytesLiteral) Write(writer *bytes.Buffer) {
	writer.Write(bm.Bytes)
}


// BytesMatcherInOrder matches a sequence of BytesMatchers that must match in
// the order given.
type BytesMatcherInOrder struct {
	Matchers []IBytesMatcher
}

func InOrder(matchers ...IBytesMatcher) *BytesMatcherInOrder {
	return &BytesMatcherInOrder{matchers}
}

func (bm *BytesMatcherInOrder) Match(b []byte) (n int, matches bool) {
	var consumed int

	for _, matcher := range bm.Matchers {
		remainder := b[n:]

		if consumed, matches = matcher.Match(remainder); !matches {
			return 0, false
		}

		n += consumed
	}

	return n, true
}

func (bm *BytesMatcherInOrder) Write(writer *bytes.Buffer) {
	for _, matcher := range bm.Matchers {
		matcher.Write(writer)
	}
}


// BytesMatcherAnyOrder matches a set of BytesMatchers, but they do not have to
// match in any particular order.
type BytesMatcherAnyOrder struct {
	Matchers []IBytesMatcher
}

func AnyOrder(matchers ...IBytesMatcher) *BytesMatcherAnyOrder {
	return &BytesMatcherAnyOrder{matchers}
}

func (bm *BytesMatcherAnyOrder) Match(b []byte) (n int, matches bool) {
	toMatch := make([]IBytesMatcher, len(bm.Matchers))
	copy(toMatch, bm.Matchers)
	var consumed int

	for {
		if len(toMatch) == 0 {
			break
		}

		remainder := b[n:]

		foundMatch := false
		for i, matcher := range toMatch {
			if consumed, foundMatch = matcher.Match(remainder); foundMatch {
				n += consumed

				// Remove matcher from toMatch.
				if len(toMatch) > 1 {
					toMatch[i] = toMatch[len(toMatch)-1]
				}
				toMatch = toMatch[:len(toMatch)-1]

				break
			}
		}

		if !foundMatch {
			return 0, false
		}
	}

	return n, true
}

func (bm *BytesMatcherAnyOrder) Write(writer *bytes.Buffer) {
	for _, matcher := range bm.Matchers {
		matcher.Write(writer)
	}
}
