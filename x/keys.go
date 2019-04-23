/*
 * Copyright 2016-2018 Dgraph Labs, Inc. and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package x

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

const (
	// TODO(pawan) - Make this 2 bytes long. Right now ParsedKey has byteType and
	// bytePrefix. Change it so that it just has one field which has all the information.
	ByteData     = byte(0x00)
	ByteIndex    = byte(0x02)
	ByteReverse  = byte(0x04)
	ByteCount    = byte(0x08)
	ByteCountRev = ByteCount | ByteReverse
	// same prefix for data, index and reverse keys so that relative order of data doesn't change
	// keys of same attributes are located together
	DefaultPrefix = byte(0x00)
	byteSchema    = byte(0x01)
	byteType      = byte(0x02)

	// Constant to specify a given key corresponds to a posting list split into multiple parts.
	ByteSplit = byte(0x01)
)

func writeAttr(buf []byte, attr string) []byte {
	AssertTrue(len(attr) < math.MaxUint16)
	binary.BigEndian.PutUint16(buf[:2], uint16(len(attr)))

	rest := buf[2:]
	AssertTrue(len(attr) == copy(rest, attr))

	return rest[len(attr):]
}

// SchemaKey returns schema key for given attribute. Schema keys are stored
// separately with unique prefix, since we need to iterate over all schema keys.
// The structure of a schema key is as follows:
//
// byte 0: key type prefix (set to byteSchema)
// byte 1-2: length of attr
// next len(attr) bytes: value of attr
func SchemaKey(attr string) []byte {
	buf := make([]byte, 1+2+len(attr))
	buf[0] = byteSchema
	rest := buf[1:]

	writeAttr(rest, attr)
	return buf
}

// TypeKey returns type key for given type name. Type keys are stored separately
// with a unique prefix, since we need to iterate over all type keys.
// The structure of a type key is as follows:
//
// byte 0: key type prefix (set to byteType)
// byte 1-2: length of typeName
// next len(attr) bytes: value of typeName
func TypeKey(typeName string) []byte {
	buf := make([]byte, 1+2+len(typeName))
	buf[0] = byteType
	rest := buf[1:]

	writeAttr(rest, typeName)
	return buf
}

// DataKey generates a data key with the given attribute and UID.
// The structure of a data key is as follows:
//
// byte 0: key type prefix (set to DefaultPrefix)
// byte 1-2: length of attr
// next len(attr) bytes: value of attr
// next byte: data type prefix (set to ByteData)
// next byte: byte to determine if this key corresponds to a list that has been split
//   into multiple parts
// next eight bytes: value of uid
// next eight bytes (optional): if the key corresponds to a split list, the startUid of
//   the split stored in this key.
func DataKey(attr string, uid uint64) []byte {
	buf := make([]byte, 1+2+len(attr)+1+1+8)
	buf[0] = DefaultPrefix
	rest := buf[1:]

	rest = writeAttr(rest, attr)
	rest[0] = ByteData

	// By default, this key does not correspond to a part of a split key.
	rest = rest[1:]
	rest[0] = 0

	rest = rest[1:]
	binary.BigEndian.PutUint64(rest, uid)
	return buf
}

// ReverseKey generates a reverse key with the given attribute and UID.
// The structure of a reverse key is as follows:
//
// byte 0: key type prefix (set to DefaultPrefix)
// byte 1-2: length of attr
// next len(attr) bytes: value of attr
// next byte: data type prefix (set to ByteReverse)
// next byte: byte to determine if this key corresponds to a list that has been split
//   into multiple parts
// next eight bytes: value of uid
// next eight bytes (optional): if the key corresponds to a split list, the startUid of
//   the split stored in this key.
func ReverseKey(attr string, uid uint64) []byte {
	buf := make([]byte, 1+2+len(attr)+1+1+8)
	buf[0] = DefaultPrefix
	rest := buf[1:]

	rest = writeAttr(rest, attr)
	rest[0] = ByteReverse

	// By default, this key does not correspond to a part of a split key.
	rest = rest[1:]
	rest[0] = 0

	rest = rest[1:]
	binary.BigEndian.PutUint64(rest, uid)
	return buf
}

// IndexKey generates a index key with the given attribute and term.
// The structure of an index key is as follows:
//
// byte 0: key type prefix (set to DefaultPrefix)
// byte 1-2: length of attr
// next len(attr) bytes: value of attr
// next byte: data type prefix (set to ByteIndex)
// next byte: byte to determine if this key corresponds to a list that has been split
//   into multiple parts
// next len(term) bytes: value of term
// next eight bytes (optional): if the key corresponds to a split list, the startUid of
//   the split stored in this key.
func IndexKey(attr, term string) []byte {
	buf := make([]byte, 1+2+len(attr)+1+1+len(term))
	buf[0] = DefaultPrefix
	rest := buf[1:]

	rest = writeAttr(rest, attr)
	rest[0] = ByteIndex

	// By default, this key does not correspond to a part of a split key.
	rest = rest[1:]
	rest[0] = 0

	rest = rest[1:]
	AssertTrue(len(term) == copy(rest, term))
	return buf
}

// CountKey generates a count key with the given attribute and uid.
// The structure of a count key is as follows:
//
// byte 0: key type prefix (set to DefaultPrefix)
// byte 1-2: length of attr
// next len(attr) bytes: value of attr
// next byte: data type prefix (set to ByteCount or ByteCountRev)
// next byte: byte to determine if this key corresponds to a list that has been split
//   into multiple parts
// next four bytes: value of count.
// next eight bytes (optional): if the key corresponds to a split list, the startUid of
//   the split stored in this key.
func CountKey(attr string, count uint32, reverse bool) []byte {
	buf := make([]byte, 1+2+len(attr)+1+1+4)
	buf[0] = DefaultPrefix
	rest := buf[1:]

	rest = writeAttr(rest, attr)
	if reverse {
		rest[0] = ByteCountRev
	} else {
		rest[0] = ByteCount
	}

	// By default, this key does not correspond to a part of a split key.
	rest = rest[1:]
	rest[0] = 0

	rest = rest[1:]
	binary.BigEndian.PutUint32(rest, count)
	return buf
}

// ParsedKey represents a key that has been parsed into its multiple attributes.
type ParsedKey struct {
	byteType    byte
	Attr        string
	Uid         uint64
	HasStartUid bool
	StartUid    uint64
	Term        string
	Count       uint32
	bytePrefix  byte
}

func (p ParsedKey) IsData() bool {
	return p.bytePrefix == DefaultPrefix && p.byteType == ByteData
}

func (p ParsedKey) IsReverse() bool {
	return p.bytePrefix == DefaultPrefix && p.byteType == ByteReverse
}

func (p ParsedKey) IsCount() bool {
	return p.bytePrefix == DefaultPrefix && (p.byteType == ByteCount ||
		p.byteType == ByteCountRev)
}

func (p ParsedKey) IsIndex() bool {
	return p.bytePrefix == DefaultPrefix && p.byteType == ByteIndex
}

func (p ParsedKey) IsSchema() bool {
	return p.bytePrefix == byteSchema
}

func (p ParsedKey) IsType() bool {
	return p.bytePrefix == byteType
}

func (p ParsedKey) IsOfType(typ byte) bool {
	switch typ {
	case ByteCount, ByteCountRev:
		return p.IsCount()
	case ByteReverse:
		return p.IsReverse()
	case ByteIndex:
		return p.IsIndex()
	case ByteData:
		return p.IsData()
	default:
	}
	return false
}

func (p ParsedKey) SkipPredicate() []byte {
	buf := make([]byte, 1+2+len(p.Attr)+1)
	buf[0] = p.bytePrefix
	rest := buf[1:]
	k := writeAttr(rest, p.Attr)
	AssertTrue(len(k) == 1)
	k[0] = 0xFF
	return buf
}

func (p ParsedKey) SkipRangeOfSameType() []byte {
	buf := make([]byte, 1+2+len(p.Attr)+1)
	buf[0] = p.bytePrefix
	rest := buf[1:]
	k := writeAttr(rest, p.Attr)
	AssertTrue(len(k) == 1)
	k[0] = p.byteType + 1
	return buf
}

func (p ParsedKey) SkipSchema() []byte {
	var buf [1]byte
	buf[0] = byteSchema + 1
	return buf[:]
}

func (p ParsedKey) SkipType() []byte {
	var buf [1]byte
	buf[0] = byteType + 1
	return buf[:]
}

// DataPrefix returns the prefix for data keys.
func (p ParsedKey) DataPrefix() []byte {
	buf := make([]byte, 1+2+len(p.Attr)+1+1)
	buf[0] = p.bytePrefix
	rest := buf[1:]
	k := writeAttr(rest, p.Attr)
	AssertTrue(len(k) == 2)
	k[0] = ByteData
	k[1] = 0
	return buf
}

// IndexPrefix returns the prefix for index keys.
func (p ParsedKey) IndexPrefix() []byte {
	buf := make([]byte, 1+2+len(p.Attr)+1+1)
	buf[0] = p.bytePrefix
	rest := buf[1:]
	k := writeAttr(rest, p.Attr)
	AssertTrue(len(k) == 2)
	k[0] = ByteIndex
	k[1] = 0
	return buf
}

// ReversePrefix returns the prefix for index keys.
func (p ParsedKey) ReversePrefix() []byte {
	buf := make([]byte, 1+2+len(p.Attr)+1+1)
	buf[0] = p.bytePrefix
	rest := buf[1:]
	k := writeAttr(rest, p.Attr)
	AssertTrue(len(k) == 2)
	k[0] = ByteReverse
	k[1] = 0
	return buf
}

// CountPrefix returns the prefix for count keys.
func (p ParsedKey) CountPrefix(reverse bool) []byte {
	buf := make([]byte, 1+2+len(p.Attr)+1+1)
	buf[0] = p.bytePrefix
	rest := buf[1:]
	k := writeAttr(rest, p.Attr)
	AssertTrue(len(k) == 2)
	if reverse {
		k[0] = ByteCountRev
	} else {
		k[0] = ByteCount
	}
	k[1] = 0
	return buf
}

// SchemaPrefix returns the prefix for Schema keys.
func SchemaPrefix() []byte {
	var buf [1]byte
	buf[0] = byteSchema
	return buf[:]
}

// TypePrefix returns the prefix for Schema keys.
func TypePrefix() []byte {
	var buf [1]byte
	buf[0] = byteType
	return buf[:]
}

// PredicatePrefix returns the prefix for all keys belonging to this predicate except schema key.
func PredicatePrefix(predicate string) []byte {
	buf := make([]byte, 1+2+len(predicate))
	buf[0] = DefaultPrefix
	k := writeAttr(buf[1:], predicate)
	AssertTrue(len(k) == 0)
	return buf
}

// GetSplitKey takes a key baseKey and generates the key of the list split that starts at startUid.
func GetSplitKey(baseKey []byte, startUid uint64) []byte {
	keyCopy := make([]byte, len(baseKey)+8)
	copy(keyCopy, baseKey)

	p := Parse(baseKey)
	index := 1 + 2 + len(p.Attr) + 1
	keyCopy[index] = ByteSplit
	binary.BigEndian.PutUint64(keyCopy[len(baseKey):], startUid)

	return keyCopy
}

// Parse would parse the key. ParsedKey does not reuse the key slice, so the key slice can change
// without affecting the contents of ParsedKey.
func Parse(key []byte) *ParsedKey {
	p := &ParsedKey{}

	p.bytePrefix = key[0]
	sz := int(binary.BigEndian.Uint16(key[1:3]))
	k := key[3:]

	p.Attr = string(k[:sz])
	k = k[sz:]

	switch p.bytePrefix {
	case byteSchema, byteType:
		return p
	default:
	}

	p.byteType = k[0]
	k = k[1:]

	p.HasStartUid = k[0] == ByteSplit
	k = k[1:]

	switch p.byteType {
	case ByteData, ByteReverse:
		if len(k) < 8 {
			if Config.DebugMode {
				fmt.Printf("Error: Uid length < 8 for key: %q, parsed key: %+v\n", key, p)
			}
			return nil
		}
		p.Uid = binary.BigEndian.Uint64(k)

		if !p.HasStartUid {
			break
		}

		if len(k) < 16 {
			if Config.DebugMode {
				fmt.Printf("Error: StartUid length < 8 for key: %q, parsed key: %+v\n", key, p)
			}
			return nil
		}

		k = k[8:]
		p.StartUid = binary.BigEndian.Uint64(k)
	case ByteIndex:
		if !p.HasStartUid {
			p.Term = string(k)
			break
		}

		if len(k) < 8 {
			if Config.DebugMode {
				fmt.Printf("Error: StartUid length < 8 for key: %q, parsed key: %+v\n", key, p)
			}
			return nil
		}

		term := k[:len(k)-8]
		startUid := k[len(k)-8:]
		p.Term = string(term)
		p.StartUid = binary.BigEndian.Uint64(startUid)
	case ByteCount, ByteCountRev:
		if len(k) < 4 {
			if Config.DebugMode {
				fmt.Printf("Error: Count length < 4 for key: %q, parsed key: %+v\n", key, p)
			}
			return nil
		}
		p.Count = binary.BigEndian.Uint32(k)

		if !p.HasStartUid {
			break
		}

		if len(k) < 12 {
			if Config.DebugMode {
				fmt.Printf("Error: StartUid length < 8 for key: %q, parsed key: %+v\n", key, p)
			}
			return nil
		}

		k = k[4:]
		p.StartUid = binary.BigEndian.Uint64(k)
	default:
		// Some other data type.
		return nil
	}
	return p
}

// IsReservedPredicate returns true if 'pred' is in the reserved predicate list.
func IsReservedPredicate(pred string) bool {
	var m = map[string]struct{}{
		PredicateListAttr: {},
		"dgraph.type":     {},
	}
	_, ok := m[strings.ToLower(pred)]
	return ok || IsAclPredicate(pred)
}

func IsAclPredicate(pred string) bool {
	var m = map[string]struct{}{
		"dgraph.xid":        {},
		"dgraph.password":   {},
		"dgraph.user.group": {},
		"dgraph.group.acl":  {},
	}
	_, ok := m[strings.ToLower(pred)]
	return ok
}
