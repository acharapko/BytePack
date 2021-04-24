package bytepack

import (
    "bytes"
    "encoding/gob"
    "encoding/json"
    "fmt"
    "github.com/google/uuid"
    "reflect"
    "testing"
)

func Benchmark_EncodeInt64(b *testing.B) {
    s := NewPacker()
    i64 := reflect.ValueOf(int64(5214563454864521467))
    for i := 0; i < b.N; i++ {
        err := s.encodeValue(i64)
        if err != nil {
            fmt.Printf("err=%v\n", err)
        }
        s.w.Reset()
    }
}

func Benchmark_EncodeSliceInt(b *testing.B) {
    s := NewPacker()
    intSlice := reflect.ValueOf([]int {6, 12, 45, 124, 25, 4587})
    for i := 0; i < b.N; i++ {
        err := s.encodeSlice(intSlice)
        if err != nil {
            fmt.Printf("err=%v\n", err)
        }
        s.w.Reset()
    }
}

func Benchmark_EncodeSliceString(b *testing.B) {
    s := NewPacker()
    intSlice := reflect.ValueOf([]string {"test123", "this is a test 12345", "testing string slice encoding", "test test test", "Hello world!"})
    for i := 0; i < b.N; i++ {
        err := s.encodeSlice(intSlice)
        if err != nil {
            fmt.Printf("err=%v\n", err)
        }
        s.w.Reset()
    }
}

func Benchmark_ReadIntSlice(b *testing.B) {
    bufBytes := []byte{0, 0, 0, 6, 0, 0, 12, 25, 0, 1, 17, 69, 0, 0, 12, 25, 0, 1, 17, 70, 0, 0, 12, 25, 0, 1, 17, 69, 0, 0, 12, 25, 0, 1, 17, 70, 0, 0, 12, 25, 0, 1, 17, 69, 0, 0, 12, 25, 0, 1, 17, 70}
    s := NewPacker()
    intSlice := make([]int, 0)
    for i := 0; i < b.N; i++ {
        buf := bytes.NewBuffer(bufBytes)
        _, err := s.UnpackSlice(reflect.TypeOf(intSlice), buf)
        if err != nil {
            fmt.Printf("err=%v\n", err)
        }
    }
}

func Benchmark_Serialize(b *testing.B) {
    s := NewPacker()
    a := person{
        Name:   "Tester",
        Age:    30,
        Height: 5.25,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        buf, err := s.Pack(a)
        if err != nil {
            panic(err)
        }
        var a2 person
        err = s.Unpack(buf, &a2)
        if err != nil {
            panic(err)
        }
    }
}

func Benchmark_SerializeSerializable(b *testing.B) {
    s := NewPacker()
    a := personS{
        Name:   "Tester",
        Age:    30,
        Height: 10000.25,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        buf, err := s.Pack(&a)
        if err != nil {
            panic(err)
        }
        var a2 personS
        err = s.Unpack(buf, &a2)
        if err != nil {
            panic(err)
        }
    }
}

func Benchmark_SerializeGob(b *testing.B) {
    gob.Register(person{})
    var send interface{}
    var recv interface{}

    buf := new(bytes.Buffer)
    encoder := gob.NewEncoder(buf)
    decoder := gob.NewDecoder(buf)

    send = person{
        Name:   "Tester",
        Age:    30,
        Height: 10000.25,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        err := encoder.Encode(&send)
        if err != nil {
            panic(err)
        }
        err = decoder.Decode(&recv)
        if err != nil {
            panic(err)
        }
    }
}

func Benchmark_SerializeJson(b *testing.B) {
    gob.Register(person{})
    var send interface{}

    send = person{
        Name:   "Tester",
        Age:    30,
        Height: 10000.25,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        tempbts, err := json.Marshal(send)
        if err != nil {
            panic(err)
        }
        a2 := person{}
        err = json.Unmarshal(tempbts, &a2)
        if err != nil {
            panic(err)
        }
    }
}

func Benchmark_SerializeWithSlicesAndNestedStruct(b *testing.B) {
    s := NewPacker()
    a := person3{
        Name:   "Tester",
        Age:    30,
        Height: 5.75,
        Children: []person{{Name: "Test Child 1", Age: 5, Height: 3.5}, {Name: "Test Child 2", Age: 7, Height: 3.75}, {Name: "Test Child 32", Age: 17, Height: 3.85}},
        Spouse: person{
            Name: "Tester Spouse", Age:28, Height: 5.25,
        },
        LuckyNumbers: []int{12, 32, 54, 87, 45, 21},
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        buf, err := s.Pack(a)
        if err != nil {
            panic(err)
        }
        var a2 person3
        err = s.Unpack(buf, &a2)
        if err != nil {
            panic(err)
        }
    }
}

func Benchmark_SerializeWithSlicesAndNestedStructGob(b *testing.B) {
    gob.Register(person3{})
    var send interface{}
    var recv interface{}

    buf := new(bytes.Buffer)
    encoder := gob.NewEncoder(buf)
    decoder := gob.NewDecoder(buf)

    send = person3{
        Name:   "Tester",
        Age:    30,
        Height: 5.75,
        Children: []person{{Name: "Test Child 1", Age: 5, Height: 3.5}, {Name: "Test Child 2", Age: 7, Height: 3.75}, {Name: "Test Child 32", Age: 17, Height: 3.85}},
        Spouse: person{
            Name: "Tester Spouse", Age:28, Height: 5.25,
        },
        LuckyNumbers: []int{12, 32, 54, 87, 45, 21},
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        err := encoder.Encode(&send)
        if err != nil {
            panic(err)
        }
        err = decoder.Decode(&recv)
        if err != nil {
            panic(err)
        }
    }
}

func Benchmark_SerializeWithSlicesAndNestedStructJson(b *testing.B) {
    var send interface{}

    send = person3{
        Name:   "Tester",
        Age:    30,
        Height: 5.75,
        Children: []person{{Name: "Test Child 1", Age: 5, Height: 3.5}, {Name: "Test Child 2", Age: 7, Height: 3.75}},
        Spouse: person{
            Name: "Tester Spouse", Age:28, Height: 5.25,
        },
        LuckyNumbers: []int{12, 32, 54, 87, 45, 21},
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        tempbts, err := json.Marshal(send)
        if err != nil {
            panic(err)
        }
        a2 := person{}
        err = json.Unmarshal(tempbts, &a2)
        if err != nil {
            panic(err)
        }
    }
}

func Benchmark_SerializeWithSimpleMap(b *testing.B) {
    s := NewPacker()
    a := foo{
        Name:   "Tester",
        Bars: map[int32]string{12:"test12", 123: "test123"},
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        buf, err := s.Pack(a)
        if err != nil {
            panic(err)
        }
        var a2 foo
        err = s.Unpack(buf, &a2)
        if err != nil {
            panic(err)
        }
    }
}

func Benchmark_SerializeWithSimpleMapGob(b *testing.B) {
    gob.Register(foo{})
    var send interface{}
    var recv interface{}

    buf := new(bytes.Buffer)
    encoder := gob.NewEncoder(buf)
    decoder := gob.NewDecoder(buf)

    send = foo{
        Name:   "Tester",
        Bars: map[int32]string{12:"test12", 123: "test123"},
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        err := encoder.Encode(&send)
        if err != nil {
            panic(err)
        }
        err = decoder.Decode(&recv)
        if err != nil {
            panic(err)
        }
    }
}

func Benchmark_SerializeWithByteSlice(b *testing.B) {
    type fooByte struct {
        Id     uuid.UUID
        Bytes []byte
    }

    s := NewPacker()
    bufBytes := []byte{0, 0, 0, 6, 0, 0, 12, 25, 0, 1, 17, 69, 0, 0, 12, 25, 0, 1, 17, 70, 0, 0, 12, 25, 0, 1, 17, 69,
        0, 0, 12, 25, 0, 1, 17, 70, 0, 0, 12, 25, 0, 1, 17, 69, 0, 0, 12, 25, 0, 1, 17, 70}
    id := uuid.New()
    a := fooByte{
        Id:    id,
        Bytes: bufBytes,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        buf, err := s.Pack(a)
        if err != nil {
            panic(err)
        }
        var a2 fooByte
        err = s.Unpack(buf, &a2)
        if err != nil {
            panic(err)
        }
    }
}

func Benchmark_SerializeWithByteSliceGob(b *testing.B) {
    type fooByte struct {
        Id     uuid.UUID
        Bytes []byte
    }

    gob.Register(fooByte{})
    var send interface{}
    var recv interface{}

    buf := new(bytes.Buffer)
    encoder := gob.NewEncoder(buf)
    decoder := gob.NewDecoder(buf)

    id := uuid.New()
    bufBytes := []byte{0, 0, 0, 6, 0, 0, 12, 25, 0, 1, 17, 69, 0, 0, 12, 25, 0, 1, 17, 70, 0, 0, 12, 25, 0, 1, 17, 69,
        0, 0, 12, 25, 0, 1, 17, 70, 0, 0, 12, 25, 0, 1, 17, 69, 0, 0, 12, 25, 0, 1, 17, 70}
    send = fooByte{
        Id:    id,
        Bytes: bufBytes,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        err := encoder.Encode(&send)
        if err != nil {
            panic(err)
        }
        err = decoder.Decode(&recv)
        if err != nil {
            panic(err)
        }
    }
}

func Benchmark_EncodeReflectWithInterface(b *testing.B) {
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

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        buf, err := s.Pack(a)
        if err != nil {
            panic(err)
        }
        var a2 bar1
        err = s.Unpack(buf, &a2)
        if err != nil {
            panic(err)
        }
    }
}

func Benchmark_EncodeReflectWithInterfaceGob(b *testing.B) {
    var send interface{}
    var recv interface{}

    buf := new(bytes.Buffer)
    encoder := gob.NewEncoder(buf)
    decoder := gob.NewDecoder(buf)

    type foo1 struct {
        Name string
    }

    type bar1 struct {
        Name string
        Foo  interface{}
    }
    gob.Register(foo1{})
    gob.Register(bar1{})

    f := foo1{
        Name: "FooName",
    }

    send = bar1{
        Name: "test",
        Foo:  f,
    }
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        err := encoder.Encode(&send)
        if err != nil {
            panic(err)
        }
        err = decoder.Decode(&recv)
        if err != nil {
            panic(err)
        }
    }
}

func Benchmark_EncodeReflectWithUint8Array(b *testing.B) {
    s := NewPacker()
    type fa1 struct {
        Nums [16]byte
        Name string
    }

    nums := [16]byte{12, 25, 17, 69, 12, 0, 48, 125, 12, 25, 17, 69, 12, 0, 48, 125}
    a := fa1{
        Nums: nums,
        Name: "test",
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        buf, err := s.Pack(a)
        if err != nil {
            panic(err)
        }
        var a2 fa1
        err = s.Unpack(buf, &a2)
        if err != nil {
            panic(err)
        }
    }
}

func Benchmark_EncodeReflectWithUint8ArrayGob(b *testing.B) {
    type fa1 struct {
        Nums [16]byte
        Name string
    }
    gob.Register(fa1{})
    var send interface{}
    var recv interface{}

    buf := new(bytes.Buffer)
    encoder := gob.NewEncoder(buf)
    decoder := gob.NewDecoder(buf)

    nums := [16]byte{12, 25, 17, 69, 12, 0, 48, 125, 12, 25, 17, 69, 12, 0, 48, 125}
    send = fa1{
        Nums: nums,
        Name: "test",
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        err := encoder.Encode(&send)
        if err != nil {
            panic(err)
        }
        err = decoder.Decode(&recv)
        if err != nil {
            panic(err)
        }
    }
}

func Benchmark_AcceptMsgInternalSerializer(b *testing.B) {
    s := NewPacker()

    type readOp struct {
        TableUUID uuid.UUID
        ReadMode  int
        SkipFirst bool
        SkipLast  bool
        StartKey  []byte
        EndKey    []byte
    }

    Register(readOp{})
    rOp := readOp{
        TableUUID: uuid.New(),
        ReadMode:  2,
        SkipFirst: true,
        SkipLast:  false,
        StartKey:  []byte("some longer test key for testing byte slice serialization speed"),
        EndKey:    nil,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        buf, err := s.Pack(rOp)
        if err != nil {
            panic(err)
        }
        var a2 readOp
        err = s.Unpack(buf, &a2)
        if err != nil {
            panic(err)
        }
    }
}

func Benchmark_AcceptMsgGob(b *testing.B) {
    type readOp struct {
        TableUUID uuid.UUID
        ReadMode  int
        SkipFirst bool
        SkipLast  bool
        StartKey  []byte
        EndKey    []byte
    }
    gob.Register(readOp{})
    var send interface{}
    var recv interface{}

    buf := new(bytes.Buffer)
    encoder := gob.NewEncoder(buf)
    decoder := gob.NewDecoder(buf)

    send = readOp{
        TableUUID: uuid.New(),
        ReadMode:  2,
        SkipFirst: true,
        SkipLast:  false,
        StartKey:  []byte("some longer test key for testing byte slice serialization speed"),
        EndKey:    nil,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        err := encoder.Encode(&send)
        if err != nil {
            panic(err)
        }
        err = decoder.Decode(&recv)
        if err != nil {
            panic(err)
        }
    }
}