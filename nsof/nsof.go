package nsof

import (
	"fmt"
	"io"
	"strconv"
	"unicode/utf16"
)

type Data []byte

type Writer interface {
	WriteNSOF(*Data)
}

type Reader interface {
	ReadNSOF(*Data, *ObjectStream) Object
}

type Object interface {
	fmt.Stringer
	Reader
	Writer
}

type ObjectStream []Object

func (data Data) PeekXLong() int32 {
	var r int32

	r = int32(data[0])
	if r > 254 {
		r = ((int32(data[1])<<8+int32(data[2]))<<8+int32(data[3]))<<8 + int32(data[4])
	}
	return r
}

func (data *Data) SkipXLong() {
	if (*data)[0] < 255 {
		*data = (*data)[1:]
	} else {
		*data = (*data)[5:]
	}
}

func (data *Data) DecodeXLong() int32 {
	r := data.PeekXLong()
	data.SkipXLong()
	return r
}

func (data *Data) EncodeXLong(value int32) {
	nsof := *data
	if value < 255 {
		nsof = append(nsof, byte(value))
	} else {
		nsof = append(nsof, 255, byte(value>>24), byte(value>>16), byte(value>>8), byte(value))
	}
	*data = nsof
}

const (
	IMMEDIATE        = 0
	CHARACTER        = 1
	UNICODECHARACTER = 2
	BINARYOBJECT     = 3
	ARRAY            = 4
	PLAINARRAY       = 5
	FRAME            = 6
	SYMBOL           = 7
	STRING           = 8
	PRECEDENT        = 9
	NIL              = 10
	SMALLRECT        = 11
	LARGEBINARY      = 12
)

type Character struct {
	value uint16
}

func NewCharacter() *Character {
	return &Character{value: 0}
}

func (character *Character) ReadNSOF(data *Data, objectStream *ObjectStream) Object {
	value := data.PeekXLong()
	if value&0xf == 6 {
		character.value = uint16(value >> 4)
		data.SkipXLong()
		*objectStream = append(*objectStream, character)
	} else {
		character = nil
	}
	return character
}

func (character *Character) WriteNSOF(data *Data) {
	*data = append(*data, IMMEDIATE)
	data.EncodeXLong((int32(character.value) << 4) | 6)
}

func (character *Character) String() string {
	return fmt.Sprintf("%c", character.value)
}

type True struct {
}

func (t *True) ReadNSOF(data *Data, objectStream *ObjectStream) Object {
	value := data.PeekXLong()
	if value == 0x1a {
		data.SkipXLong()
		*objectStream = append(*objectStream, t)
	} else {
		t = nil
	}
	return t
}

func (t *True) String() string {
	return "true"
}

func (t *True) WriteNSOF(data *Data) {
	*data = append(*data, IMMEDIATE)
	data.EncodeXLong(0x1a)
}

type Nil struct {
}

func NewNil() *Nil {
	return &Nil{}
}

func (n *Nil) ReadNSOF(data *Data, objectStream *ObjectStream) Object {
	value := data.PeekXLong()
	if value == 0x2 {
		data.SkipXLong()
		*objectStream = append(*objectStream, n)
	} else {
		n = nil
	}
	return n
}

func (n *Nil) String() string {
	return "NIL"
}

func (n *Nil) WriteNSOF(data *Data) {
	data.EncodeXLong(NIL)
}

type MagicPointer struct {
	value int32
}

func (pointer *MagicPointer) ReadNSOF(data *Data, objectStream *ObjectStream) Object {
	value := data.PeekXLong()
	if value&0x3 == 3 {
		pointer.value = (value >> 4) & 0xffff
		data.SkipXLong()
		*objectStream = append(*objectStream, pointer)
	} else {
		pointer = nil
	}
	return pointer
}

func (pointer *MagicPointer) String() string {
	return fmt.Sprintf("@%d", pointer.value)
}

func (pointer *MagicPointer) WriteNSOF(data *Data) {
	*data = append(*data, IMMEDIATE)
	data.EncodeXLong((pointer.value << 4) | 6)
}

type Integer struct {
	value int32
}

func NewInteger() *Integer {
	return &Integer{}
}

func (integer *Integer) ReadNSOF(data *Data, objectStream *ObjectStream) Object {
	value := data.PeekXLong()
	if value&0x3 == 0 {
		integer.value = value >> 2
		data.SkipXLong()
		*objectStream = append(*objectStream, integer)
	} else {
		integer = nil
	}
	return integer
}

func (integer *Integer) WriteNSOF(data *Data) {
	*data = append(*data, IMMEDIATE)
	data.EncodeXLong(integer.value << 2)
}

func (integer *Integer) String() string {
	return strconv.Itoa(int(integer.value))
}

func NewImmediate(data *Data, objectStream *ObjectStream) Object {
	var r Object
	value := data.PeekXLong()
	if value&0x3 == 0 {
		r = &Integer{value >> 2}
	} else if value&0x3 == 3 {
		r = &MagicPointer{(value >> 4) & 0xffff}
	} else if value&0xf == 6 {
		r = &Character{uint16(value >> 4)}
	} else if value == 0x1a {
		r = &True{}
	} else if value == 0x2 {
		r = &Nil{}
	}
	if r != nil {
		data.SkipXLong()
		*objectStream = append(*objectStream, r)
	} else {
		panic(fmt.Sprintf("Parsing immediate %d not implemented. Data: %x\n", value, (*data)[:10]))
	}
	return r
}

type Slot struct {
	key   Object
	value Object
}

type Frame struct {
	slots []Slot
}

func NewFrame() *Frame {
	return &Frame{}
}

func (frame *Frame) ReadNSOF(data *Data, objectStream *ObjectStream) Object {
	*objectStream = append(*objectStream, frame)
	elements := data.DecodeXLong()
	frame.slots = make([]Slot, elements)
	for i, _ := range frame.slots {
		frame.slots[i] = Slot{data.DecodeObject(objectStream), nil}
	}
	for i, _ := range frame.slots {
		frame.slots[i].value = data.DecodeObject(objectStream)
	}
	return frame
}

func (frame *Frame) WriteNSOF(data *Data) {
	*data = append(*data, FRAME)
	data.EncodeXLong(int32(len(frame.slots)))
	for _, slot := range frame.slots {
		slot.key.WriteNSOF(data)
	}
	for _, slot := range frame.slots {
		slot.value.WriteNSOF(data)
	}
}

func (frame *Frame) String() string {
	return "{}"
}

func (frame *Frame) WriteTo(writer io.Writer) (n int64, err error) {
	var n0 int
	var err0 error
	n0, err0 = writer.Write([]byte{'{'})
	return int64(n0), err0
}

type Symbol struct {
	value string
}

func (symbol *Symbol) ReadNSOF(data *Data, objectStream *ObjectStream) Object {
	*objectStream = append(*objectStream, symbol)
	length := data.DecodeXLong()
	symbol.value = string((*data)[:length])
	*data = (*data)[length:]
	return symbol
}

func (symbol *Symbol) WriteNSOF(data *Data) {
	*data = append(*data, SYMBOL)
	data.EncodeXLong(int32(len(symbol.value)))
	*data = append(*data, symbol.value...)
}

func (symbol *Symbol) String() string {
	return symbol.value
}

func NewSymbol() *Symbol {
	return &Symbol{}
}

type PlainArray struct {
	objects []Object
}

func (plainArray *PlainArray) ReadNSOF(data *Data, objectStream *ObjectStream) Object {
	*objectStream = append(*objectStream, plainArray)
	for length := data.DecodeXLong(); length > 0; length-- {
		plainArray.objects = append(plainArray.objects, data.DecodeObject(objectStream))
	}
	return plainArray
}

func (plainArray *PlainArray) WriteNSOF(data *Data) {
	*data = append(*data, PLAINARRAY)
	data.EncodeXLong(int32(len(plainArray.objects)))
	for _, object := range plainArray.objects {
		object.WriteNSOF(data)
	}
}

func (plainArray *PlainArray) String() string {
	return fmt.Sprintf("%v", plainArray.objects)
}

func NewPlainArray() *PlainArray {
	return &PlainArray{}
}

type Precedent struct {
	reference int32
}

func (precedent *Precedent) ReadNSOF(data *Data, objectStream *ObjectStream) Object {
	precedent.reference = data.DecodeXLong()
	return precedent
}

func (precedent *Precedent) WriteNSOF(data *Data) {
	*data = append(*data, PRECEDENT)
	data.EncodeXLong(precedent.reference)
}

func (precedent *Precedent) String() string {
	return fmt.Sprintf("%d", precedent.reference)
}

func NewPrecedent() *Precedent {
	return &Precedent{}
}

type String struct {
	value string
}

func (str *String) ReadNSOF(data *Data, objectStream *ObjectStream) Object {
	var i int32
	*objectStream = append(*objectStream, str)
	length := data.DecodeXLong()
	chars := make([]uint16, length/2)
	for i = 0; i < length/2; i++ {
		chars[i] = uint16((*data)[i*2])*256 + uint16((*data)[i*2+1])
	}
	str.value = string(utf16.Decode(chars))
	*data = (*data)[length:]
	return str
}

func (str *String) WriteNSOF(data *Data) {
	*data = append(*data, STRING)
	data.EncodeXLong(int32(len(str.value) * 2))
	for i := 0; i < len(str.value); i++ {
		*data = append(*data, 0, str.value[i])
	}
}

func (str *String) String() string {
	return str.value
}

func NewString() *String {
	return &String{}
}

type BinaryObject struct {
	value []byte
	class Object
}

func (binaryObject *BinaryObject) ReadNSOF(data *Data, objectStream *ObjectStream) Object {
	*objectStream = append(*objectStream, binaryObject)
	length := data.DecodeXLong()
	binaryObject.class = data.DecodeObject(objectStream)
	binaryObject.value = (*data)[:length]
	*data = (*data)[length:]
	return binaryObject
}

func (binaryObject *BinaryObject) WriteNSOF(data *Data) {
	*data = append(*data, BINARYOBJECT)
	data.EncodeXLong(int32(len(binaryObject.value)))
	binaryObject.class.WriteNSOF(data)
	*data = append(*data, binaryObject.value...)
}

func (binaryObject *BinaryObject) String() string {
	return fmt.Sprintf("%x", binaryObject.value)
}

func NewBinaryObject() *BinaryObject {
	return &BinaryObject{}
}

type UnicodeCharacter struct {
	value uint16
}

func NewUnicodeCharacter() *UnicodeCharacter {
	return &UnicodeCharacter{value: 0}
}

func (character *UnicodeCharacter) ReadNSOF(data *Data, objectStream *ObjectStream) Object {
	character.value = uint16((*data)[0])<<8 + uint16((*data)[1])
	*data = (*data)[2:]
	return character
}

func (character *UnicodeCharacter) WriteNSOF(data *Data) {
	*data = append(*data, UNICODECHARACTER)
	*data = append(*data, byte(character.value>>8), byte(character.value))
}

func (character *UnicodeCharacter) String() string {
	return fmt.Sprintf("%c", character.value)
}

type Array struct {
	objects []Object
	class   Object
}

func NewArray() *Array {
	return &Array{}
}

func (array *Array) ReadNSOF(data *Data, objectStream *ObjectStream) Object {
	*objectStream = append(*objectStream, array)
	length := int(data.DecodeXLong())
	array.class = data.DecodeObject(objectStream)
	for i := 0; i < length; i++ {
		array.objects = append(array.objects, data.DecodeObject(objectStream))
	}
	return array
}

func (array *Array) WriteNSOF(data *Data) {
	*data = append(*data, ARRAY)
	data.EncodeXLong(int32(len(array.objects)))
	array.class.WriteNSOF(data)
	for _, object := range array.objects {
		object.WriteNSOF(data)
	}
}

func (array *Array) String() string {
	return fmt.Sprintf("%v", array.objects)
}

type SmallRect struct {
	left   byte
	top    byte
	right  byte
	bottom byte
}

func NewSmallRect() *SmallRect {
	return &SmallRect{}
}

func (smallRect *SmallRect) ReadNSOF(data *Data, objectStream *ObjectStream) Object {
	*objectStream = append(*objectStream, smallRect)
	smallRect.top = (*data)[0]
	smallRect.left = (*data)[1]
	smallRect.bottom = (*data)[2]
	smallRect.right = (*data)[3]
	*data = (*data)[4:]
	return smallRect
}

func (smallRect *SmallRect) WriteNSOF(data *Data) {
	*data = append(*data, ARRAY, smallRect.top, smallRect.left, smallRect.bottom, smallRect.right)
}

func (smallRect *SmallRect) String() string {
	return fmt.Sprintf("%d %d %d %d", smallRect.top, smallRect.left, smallRect.bottom, smallRect.right)
}

type LargeBinary struct {
	class               Object
	compressed          bool
	compander           string
	companderParameters string
	data                []byte
}

func NewLargeBinary() *LargeBinary {
	return &LargeBinary{}
}

func (largeBinary *LargeBinary) ReadNSOF(data *Data, objectStream *ObjectStream) Object {
	return largeBinary
}

func (largeBinary *LargeBinary) WriteNSOF(data *Data) {
}

func (largeBinary *LargeBinary) String() string {
	return fmt.Sprintf("<binary, %d bytes>", len(largeBinary.data))
}

func (data *Data) DecodeObject(stream *ObjectStream) Object {
	var object Object
	objtype := (*data)[0]
	*data = (*data)[1:]
	if objtype == IMMEDIATE {
		object = NewImmediate(data, stream)
	} else {
		switch objtype {
		case FRAME:
			object = NewFrame()
		case SYMBOL:
			object = NewSymbol()
		case PLAINARRAY:
			object = NewPlainArray()
		case PRECEDENT:
			object = NewPrecedent()
		case STRING:
			object = NewString()
		case NIL:
			object = NewNil()
		case BINARYOBJECT:
			object = NewBinaryObject()
		case CHARACTER:
			object = NewCharacter()
		case UNICODECHARACTER:
			object = NewUnicodeCharacter()
		case ARRAY:
			object = NewArray()
		case SMALLRECT:
			object = NewSmallRect()
		case LARGEBINARY:
			object = NewLargeBinary()
		default:
			panic(fmt.Sprintf("Parsing type %d not implemented. Data: %x\n", objtype, (*data)[:10]))
		}
		object.ReadNSOF(data, stream)
	}
	return object
}

func (data Data) Factory() ObjectStream {
	objects := make(ObjectStream, 0, 100)
	for len(data) > 0 {
		data.DecodeObject(&objects)
	}
	return objects
}

func (stream ObjectStream) Print() {
	for i := 0; i < len(stream); i++ {
		fmt.Printf("%d: %s\n", i, stream[i].(Object).String())
	}
}
