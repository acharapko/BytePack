package bytepack

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"net"
	"os"
	"sync"
	"testing"
	"time"
)

/*~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 * Some shred structs for testing and benchmarking
 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

type person struct {
	Name   string
	Age    int32
	Height float64
}

type personS struct {
	Name   string
	Age    int32
	Height float64
}

type barLoop struct {
	BarName string
	Foo     *fooLoop
}

type fooLoop struct {
	B       *barLoop
	FooName string
}

func (p *personS) Pack(packer *Packer) error {
	err := packer.PackFloat64(p.Height)
	if err != nil {
		return err
	}
	err = packer.PackString(p.Name)
	if err != nil {
		return err
	}
	err = packer.PackInt32(p.Age)
	if err != nil {
		return err
	}
	return nil
}

func (p *personS) Unpack(packer *Packer, buf BPReader) error {
	var err error
	p.Height, err = packer.UnpackFloat64(buf)
	if err != nil {
		return err
	}
	p.Name, err = packer.UnpackString(buf)
	if err != nil {
		return err
	}
	p.Age, err = packer.UnpackInt32(buf)
	if err != nil {
		return err
	}

	return nil
}

type person3 struct {
	Name         string
	Age          int32
	Height       float64
	Children     []person
	Spouse       person
	LuckyNumbers []int
}

type bar string

type foo struct {
	Name bar
	Bars map[int32]string
}

/*~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 * Helpers
 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

func areByteStringsSame(b1 []byte, b2 []byte) bool {
	if len(b1) != len(b2) {
		return false
	}

	for i, b := range b1 {
		if b2[i] != b {
			return false
		}
	}

	return true
}

/*~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 * Test Cases
 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

func TestPacker_EncodeReflectSimple(t *testing.T) {
	s := NewPacker()
	a := person{
		Name:   "Tester",
		Age:    30,
		Height: 5.25,
	}

	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 person
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, a.Age, a2.Age)
	assert.Equal(t, a.Height, a2.Height)
	assert.Equal(t, a.Name, a2.Name)
}

func TestPacker_EncodeReflectSimplePassedAsPointer(t *testing.T) {
	s := NewPacker()
	a := person{
		Name:   "Tester",
		Age:    30,
		Height: 5.25,
	}

	buf, _ := s.Pack(&a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 person
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, a.Age, a2.Age)
	assert.Equal(t, a.Height, a2.Height)
	assert.Equal(t, a.Name, a2.Name)
}

func TestPacker_EncodeReflectSimpleSerializable(t *testing.T) {
	s := NewPacker()
	a := personS{
		Name:   "Tester123",
		Age:    31,
		Height: 6.25,
	}

	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 personS
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, a.Age, a2.Age)
	assert.Equal(t, a.Height, a2.Height)
	assert.Equal(t, a.Name, a2.Name)
}

func TestPacker_EncodeReflectSimpleSerializablePassedAsPointer(t *testing.T) {
	s := NewPacker()
	a := personS{
		Name:   "Tester123",
		Age:    31,
		Height: 6.25,
	}

	buf, _ := s.Pack(&a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 personS
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, a.Age, a2.Age)
	assert.Equal(t, a.Height, a2.Height)
	assert.Equal(t, a.Name, a2.Name)
}

func TestPacker_EncodeSlice(t *testing.T) {
	byteSlice := make([]byte, 1024)
	rnd := rand.New(rand.NewSource(42))
	rnd.Read(byteSlice)

	s := NewPacker()
	buf, err := s.Pack(byteSlice)
	assert.NoError(t, err)
	assert.Greater(t, len(buf), 0)
	fmt.Printf("buf len = %d\n", len(buf))

	var decodedSlice []byte

	err = s.Unpack(buf, &decodedSlice)
	assert.NoError(t, err)
	assert.Equal(t, len(byteSlice), len(decodedSlice))
	assert.True(t, bytes.Equal(byteSlice, decodedSlice))
}

func TestPacker_EncodeSlice2(t *testing.T) {

	type Key []byte

	byteSlice := make([]byte, 1024)
	rnd := rand.New(rand.NewSource(42))
	rnd.Read(byteSlice)

	keySlice := Key(byteSlice)

	s := NewPacker()
	buf, err := s.Pack(keySlice)
	assert.NoError(t, err)
	assert.Greater(t, len(buf), 0)
	fmt.Printf("buf len = %d\n", len(buf))

	var decodedSlice Key

	err = s.Unpack(buf, &decodedSlice)
	assert.NoError(t, err)
	assert.Equal(t, len(keySlice), len(decodedSlice))
	assert.True(t, bytes.Equal(keySlice, decodedSlice))
}

func TestPacker_EncodeBasicType(t *testing.T) {
	rnd := rand.New(rand.NewSource(42))
	ui64 := rnd.Uint64()

	s := NewPacker()
	buf, err := s.Pack(ui64)
	assert.NoError(t, err)
	assert.Greater(t, len(buf), 0)
	fmt.Printf("buf len = %d\n", len(buf))

	var decodedui64 uint64

	err = s.Unpack(buf, &decodedui64)
	assert.NoError(t, err)
	assert.Equal(t, ui64, decodedui64)

}

func TestPacker_EncodeReflectSimpleAsInterface(t *testing.T) {
	s := NewPacker()
	a := person{
		Name:   "Tester",
		Age:    30,
		Height: 5.25,
	}

	var aiface interface{}

	aiface = a

	buf, err := s.Pack(aiface)
	assert.Nil(t, err)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 person
	err = s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, a.Age, a2.Age)
	assert.Equal(t, a.Height, a2.Height)
	assert.Equal(t, a.Name, a2.Name)
}

func TestPacker_EncodeReflectSimpleAsInterfacePointer(t *testing.T) {
	s := NewPacker()
	a := person{
		Name:   "Tester",
		Age:    30,
		Height: 5.25,
	}

	var aiface interface{}

	aiface = a

	buf, err := s.Pack(&aiface)
	assert.Nil(t, err)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 person
	err = s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, a.Age, a2.Age)
	assert.Equal(t, a.Height, a2.Height)
	assert.Equal(t, a.Name, a2.Name)
}

func TestPacker_EncodeReflectWithUnpackableType(t *testing.T) {
	s := NewPacker()
	type fooUnpackacble struct {
		Name string
		Ch   chan interface{}
	}
	a := fooUnpackacble{
		Name: "Tester",
		Ch:   make(chan interface{}, 1),
	}

	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 fooUnpackacble
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, a.Name, a2.Name)
	assert.Nil(t, a2.Ch)
}

func TestPacker_EncodeReflectWithSlice(t *testing.T) {
	s := NewPacker()

	type person2 struct {
		Name         string
		Age          int32
		Height       float64
		Children     []string
		LuckyNumbers []int64
	}

	a := person2{
		Name:         "Tester",
		Age:          30,
		Height:       5.75,
		Children:     []string{"Tester Child 1", "Tester Child 2"},
		LuckyNumbers: []int64{12, 32, 54, 87, 45, 21},
	}

	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 person2
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, a.Age, a2.Age)
	assert.Equal(t, a.Height, a2.Height)
	assert.Equal(t, a.Name, a2.Name)
	assert.Equal(t, len(a.Children), len(a2.Children))
	assert.Equal(t, a.Children[0], a2.Children[0])
	assert.Equal(t, a.Children[1], a2.Children[1])
	assert.Equal(t, len(a.LuckyNumbers), len(a2.LuckyNumbers))
	assert.Equal(t, a.LuckyNumbers[0], a2.LuckyNumbers[0])
	assert.Equal(t, a.LuckyNumbers[len(a.LuckyNumbers)-1], a2.LuckyNumbers[len(a.LuckyNumbers)-1])
}

func TestPacker_EncodeReflectWithNilSlice(t *testing.T) {
	s := NewPacker()
	type FooNil struct {
		Nums []int
		Name string
	}
	a := FooNil{
		Nums: nil,
		Name: "Tester",
	}

	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 FooNil
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, a.Name, a2.Name)
	assert.Nil(t, a2.Nums)
}

func TestPacker_EncodeReflectWithNilSliceType(t *testing.T) {
	s := NewPacker()
	type BarNil []int
	type FooNil struct {
		Name string
		Nums BarNil
	}
	a := FooNil{
		Name: "Tester",
		Nums: nil,
	}

	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 FooNil
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, a.Name, a2.Name)
	assert.Nil(t, a2.Nums)
}

func TestPacker_EncodeReflectWithSliceOfStructs(t *testing.T) {
	s := NewPacker()
	a := person3{
		Name:     "Tester",
		Age:      30,
		Height:   5.75,
		Children: []person{{Name: "Test Child 1", Age: 5, Height: 3.5}, {Name: "Test Child 2", Age: 7, Height: 3.75}},
		Spouse: person{
			Name: "Tester Spouse", Age: 28, Height: 5.25,
		},
		LuckyNumbers: []int{12, 32, 54, 87, 45, 21},
	}

	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 person3
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, a.Name, a2.Name)
	assert.Equal(t, a.Age, a2.Age)
	assert.Equal(t, a.Height, a2.Height)
	assert.Equal(t, len(a.Children), len(a2.Children))
	assert.Equal(t, len(a.LuckyNumbers), len(a2.LuckyNumbers))
	assert.Equal(t, "Test Child 1", a2.Children[0].Name)
	assert.Equal(t, int32(5), a2.Children[0].Age)
	assert.Equal(t, 3.5, a2.Children[0].Height)
	assert.Equal(t, "Test Child 2", a2.Children[1].Name)
	assert.Equal(t, int32(7), a2.Children[1].Age)
	assert.Equal(t, 3.75, a2.Children[1].Height)
	assert.Equal(t, 12, a2.LuckyNumbers[0])
	assert.Equal(t, 21, a2.LuckyNumbers[5])
}

func TestPacker_EncodeReflectWithMap(t *testing.T) {
	s := NewPacker()
	a := foo{
		Name: "Tester",
		Bars: map[int32]string{12: "test12", 123: "test123"},
	}

	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 foo
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, bar("Tester"), a2.Name)
	assert.Equal(t, len(a.Bars), len(a2.Bars))
	assert.Equal(t, a.Bars[12], a2.Bars[12])
	assert.Equal(t, a.Bars[123], a2.Bars[123])
}

func TestPacker_EncodeReflectWithMapAndPointer(t *testing.T) {
	s := NewPacker()

	type bar2 struct {
		Name string
		Bar  int
	}

	type foo2 struct {
		B    *bar2
		Bars map[int32]string
	}

	b := &bar2{
		Name: "bar",
		Bar:  100,
	}
	a := foo2{
		B:    b,
		Bars: map[int32]string{12: "test12", 123: "test123"},
	}

	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 foo2
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	assert.NotNil(t, a2.B)
	assert.Equal(t, "bar", a2.B.Name)
	assert.Equal(t, 100, a2.B.Bar)
	assert.Equal(t, len(a.Bars), len(a2.Bars))
	assert.Equal(t, a.Bars[12], a2.Bars[12])
	assert.Equal(t, a.Bars[123], a2.Bars[123])

	fmt.Printf("a2=%+v\n", a2)
}

func TestPacker_EncodeReflectWithMapAndNilPointer(t *testing.T) {
	s := NewPacker()

	type bar2 struct {
		Name string
		Bar  int
	}

	type foo2 struct {
		B    *bar2
		Bars map[int32]string
	}

	a := foo2{
		B:    nil,
		Bars: map[int32]string{12: "test12", 123: "test123"},
	}

	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 foo2
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Nil(t, a2.B)
	assert.Equal(t, len(a.Bars), len(a2.Bars))
	assert.Equal(t, a.Bars[12], a2.Bars[12])
	assert.Equal(t, a.Bars[123], a2.Bars[123])
}

func TestPacker_EncodeReflectWithPointerLoop(t *testing.T) {
	s := NewPacker()

	b := &barLoop{
		BarName: "bar",
	}

	f := fooLoop{
		B:       b,
		FooName: "foo",
	}
	b.Foo = &f

	buf, _ := s.Pack(f)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 fooLoop
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	assert.NotNil(t, a2.B)
	assert.Equal(t, "bar", a2.B.BarName)
	assert.Equal(t, "foo", a2.B.Foo.FooName)
	assert.NotNil(t, a2.B.Foo.B)
	assert.Equal(t, "bar", a2.B.Foo.B.BarName)
	assert.Equal(t, a2.B, a2.B.Foo.B)
	assert.Equal(t, a2.B.Foo, a2.B.Foo.B.Foo)

	fmt.Printf("a2=%+v\n", a2)
}

func TestPacker_EncodeReflectWithPointerLoopAndPointerEncode(t *testing.T) {
	s := NewPacker()

	b := &barLoop{
		BarName: "bar",
	}

	f := fooLoop{
		B:       b,
		FooName: "foo",
	}
	b.Foo = &f

	buf, _ := s.Pack(&f)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 fooLoop
	a2ptr := &a2
	err := s.Unpack(buf, a2ptr)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	assert.NotNil(t, a2.B)
	assert.Equal(t, "bar", a2.B.BarName)
	assert.Equal(t, "foo", a2.B.Foo.FooName)
	assert.NotNil(t, a2.B.Foo.B)
	assert.Equal(t, "bar", a2.B.Foo.B.BarName)
	a2ptrAddress := fmt.Sprintf("%p", a2ptr)
	a2BFooAddress := fmt.Sprintf("%p", a2.B.Foo)
	assert.Equal(t, a2ptrAddress, a2BFooAddress)
	assert.Equal(t, a2.B, a2.B.Foo.B)
	assert.Equal(t, a2.B.Foo, a2.B.Foo.B.Foo)

	fmt.Printf("a2=%+v\n", a2)
}

func TestPacker_EncodeReflectWithNilMap(t *testing.T) {
	s := NewPacker()

	a := foo{
		Name: "test",
		Bars: nil,
	}

	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 foo
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, bar("test"), a2.Name)
	assert.Nil(t, a2.Bars)
}

func TestPacker_EncodeReflectWithMapOfStructs(t *testing.T) {
	s := NewPacker()

	type strct struct {
		Name string
	}

	type mapOfStructs struct {
		M map[string]strct
	}

	a := mapOfStructs{
		M: map[string]strct{"key1": {Name: "name1"}, "key2": {Name: "name2"}, "key3": {Name: "name3"}},
	}

	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 mapOfStructs
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.NotNil(t, a2.M)
	assert.Equal(t, len(a.M), len(a2.M))
	assert.Equal(t, a.M["key1"], a2.M["key1"])
	assert.Equal(t, a.M["key2"], a2.M["key2"])
	assert.Equal(t, a.M["key3"], a2.M["key3"])
}

func TestPacker_EncodeReflectWithMapOfStructPointers(t *testing.T) {
	s := NewPacker()

	type strct struct {
		Name string
	}

	type mapOfStructs struct {
		M map[string]*strct
	}

	a := mapOfStructs{
		M: map[string]*strct{"key1": {Name: "name1"}, "key2": {Name: "name2"}, "key3": {Name: "name3"}},
	}

	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 mapOfStructs
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.NotNil(t, a2.M)
	assert.Equal(t, len(a.M), len(a2.M))
	assert.Equal(t, a.M["key1"], a2.M["key1"])
	assert.Equal(t, a.M["key2"], a2.M["key2"])
	assert.Equal(t, a.M["key3"], a2.M["key3"])
}

func TestPacker_EncodeReflectWithNestedStructs(t *testing.T) {
	s := NewPacker()

	type foo1 struct {
		Name string
	}

	type bar1 struct {
		Name string
		Foo  foo1
	}

	f := foo1{
		Name: "FooName",
	}

	a := bar1{
		Name: "test",
		Foo:  f,
	}

	Register(foo1{})

	fmt.Printf("original: %+v\n", a)
	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 bar1
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, a.Name, a2.Name)
	assert.NotNil(t, a2.Foo)
	fd := a2.Foo
	assert.Equal(t, f.Name, fd.Name)
}

func TestPacker_EncodeReflectWithInterface(t *testing.T) {
	s := NewPacker()

	type foo1 struct {
		Name string
	}

	type bar1 struct {
		Name string
		Foo  interface{}
	}

	f := foo1{
		Name: "FooName",
	}

	a := bar1{
		Name: "test",
		Foo:  f,
	}

	Register(foo1{})

	fmt.Printf("original: %+v\n", a)
	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 bar1
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, a.Name, a2.Name)
	assert.NotNil(t, a2.Foo)
	fd := a2.Foo.(foo1)
	assert.Equal(t, f.Name, fd.Name)
}

func TestPacker_EncodeReflectWithInterfacePointer(t *testing.T) {
	s := NewPacker()
	type foo1 struct {
		Name string
	}

	type bar1 struct {
		Name string
		Foo  interface{}
	}

	f := &foo1{
		Name: "FooName",
	}

	a := bar1{
		Name: "test",
		Foo:  f,
	}

	Register(&foo1{})

	fmt.Printf("original: %+v\n", a)
	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 bar1
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, a.Name, a2.Name)
	assert.NotNil(t, a2.Foo)
	fd := a2.Foo.(*foo1)
	assert.Equal(t, f.Name, fd.Name)
}

func TestPacker_EncodeReflectWithNilInterface(t *testing.T) {
	s := NewPacker()
	type foonil struct {
		Name string
	}

	type barnil struct {
		Name string
		Foo  interface{}
	}

	a := barnil{
		Name: "test",
		Foo:  nil,
	}

	Register(foonil{})

	fmt.Printf("original: %+v\n", a)
	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 barnil
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, a.Name, a2.Name)
	assert.Nil(t, a2.Foo)
}

func TestPacker_EncodeReflectWithByteSliceAndUUID(t *testing.T) {
	type fooByte struct {
		Id    uuid.UUID
		Bytes []byte
	}

	s := NewPacker()
	bufBytes := []byte{1, 0, 0, 6, 0, 0, 12, 25, 0, 1, 17, 69, 0, 0, 12, 25, 0, 1, 17, 70, 0, 0, 12, 25, 0, 1, 17, 69}
	id := uuid.New()
	a := fooByte{
		Id:    id,
		Bytes: bufBytes,
	}

	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 fooByte
	err := s.Unpack(buf, &a2)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}
	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, id, a2.Id)
	assert.Equal(t, len(bufBytes), len(a2.Bytes))
	assert.True(t, areByteStringsSame(bufBytes, a2.Bytes))
}

func TestPacker_EncodeReflectWithIntArray(t *testing.T) {
	s := NewPacker()
	type fa1 struct {
		Name string
		Nums [4]int
	}

	type fa2 struct {
		Name string
		Nums [4]int8
	}
	nums := [4]int{12, 25, 17, 69}
	nums8 := [4]int8{12, 25, 17, 69}
	a := fa1{
		Name: "test",
		Nums: nums,
	}

	a8 := fa2{
		Name: "test8",
		Nums: nums8,
	}

	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var decoded fa1
	err := s.Unpack(buf, &decoded)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}
	fmt.Printf("decoded=%+v\n", decoded)

	assert.Equal(t, a.Name, decoded.Name)
	assert.Equal(t, 4, len(decoded.Nums))
	assert.Equal(t, 12, decoded.Nums[0])
	assert.Equal(t, 25, decoded.Nums[1])
	assert.Equal(t, 17, decoded.Nums[2])
	assert.Equal(t, 69, decoded.Nums[3])
	s.w.Reset()
	buf8, _ := s.Pack(a8)
	fmt.Printf("buf len = %d\n", len(buf8))

	var decoded8 fa2
	err = s.Unpack(buf8, &decoded8)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		t.Fail()
	}
	fmt.Printf("decoded8=%+v\n", decoded8)

	assert.Equal(t, a8.Name, decoded8.Name)
	assert.Equal(t, 4, len(decoded8.Nums))
	assert.Equal(t, int8(12), decoded8.Nums[0])
	assert.Equal(t, int8(25), decoded8.Nums[1])
	assert.Equal(t, int8(17), decoded8.Nums[2])
	assert.Equal(t, int8(69), decoded8.Nums[3])
}

func TestPacker_EncodeReflectOverSlowTcp(t *testing.T) {
	s := NewPacker()
	type fooperson struct {
		Name      string
		Age       int
		Height    float64
		HairColor string
		Weight    int32
	}
	a := fooperson{
		Name:      "Tester",
		Age:       30,
		Height:    5.25,
		HairColor: "Brown",
		Weight:    155,
	}

	buf, _ := s.Pack(a)
	fmt.Printf("buf len = %d\n", len(buf))

	var a2 fooperson
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		fmt.Printf("start listen\n")
		l, err := net.Listen("tcp", "127.0.0.1:8001")
		if err != nil {
			fmt.Println("Error listening:", err.Error())
			os.Exit(1)
		}
		defer l.Close()
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		connReader := bufio.NewReader(conn)
		go func() {

			err := s.UnpackFromReader(connReader, &a2)
			if err != nil {
				panic(err)
			}
			// Close the connection when you're done with it.
			conn.Close()
		}()
		wg.Done()
	}()

	time.Sleep(100 * time.Millisecond)

	go func() {
		fmt.Printf("start dial\n")
		connOut, err := net.Dial("tcp", "127.0.0.1:8001")
		if err != nil {
			panic(err)
		}

		for i := 0; i < len(buf); i++ {
			connOut.Write(buf[i : i+1])
			time.Sleep(50 * time.Millisecond)
		}
		wg.Done()
	}()

	wg.Wait()

	fmt.Printf("a2=%+v\n", a2)

	assert.Equal(t, a.Age, a2.Age)
	assert.Equal(t, a.Height, a2.Height)
	assert.Equal(t, a.Name, a2.Name)
	assert.Equal(t, a.HairColor, a2.HairColor)
	assert.Equal(t, a.Weight, a2.Weight)
}
