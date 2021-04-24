# BytePack

BytePack is a small library for serializing and deserializing structs in Go. 
It aims to be faster than gob for smaller structs,and can get between ~105 to ~90% advantage over gob, although I speculate that some more complicated structs will be slower than gob encoder. 
BytePack relies on  reflection to traverse the struct and serialize it, and requires no custom code to be written. 
However, one can implement `Packable` interface write custom `Pack()` and `Unpack()` methods to have more custom and performant serialization.

Unlike gob package, BytePack is not stream-oriented, and produces complete serialization that can be decoded by another BytePack, as long as type in known. 
Structs require no special registration with the library, as long as they are passed for encoding/packing as some srtuct type and not `interface{}`.
For decoding/unpacking an interface, its type must be registered with the library using a `Register(strct interface{})` function, similar to gob registration.

**This is not a production-grade code. There are likely to be bugs that may lead to data corruption or data loss.**

---
## Components

The library consists of two major components:
* Packer - component that handles packing and unpacking of structs. Packer is not thread safe
* BytePack - a thread safe component that uses a pool of packers for concurrency.

---
## Usage

* New BytePack:
  ```go
  numPackers := 5 // number of packers to have in the BytePack. This controls max concurrency of BytePack
  bp := bytepack.NewBytePack(numPackers)
   ```

* Packing
  
  ```go
  type foo struct {
      Score int
      Name string
  }
  
  f := foo {
      Score: 100,	
      Name: "Tester",
  }
  
  packedBytes, err := bp.Pack(f)
  if err != nil {
       panic(err)
  } 
  ```
  
* Unpacking
  ```go
  var fUnpacked foo
  
  err = bp.Unpack(packedBytes, &fUnpacked)
  if err != nil {
       panic(err)
  } 
  ```
  
  Alternatively, one may Unpack from a reader, such as ``bufio.Reader`` or ``bytes.Buffer``:
  ```go
  reader := bytes.NewBuffer(packedBytes)
  err = bp.UnpackFromReader(reader, &fUnpacked)
  if err != nil {
       panic(err)
  } 
  ```
  
* Register Struct for use as _interface{}_
  ```go
  type foo struct {
      F: string
  }
  type bar struct {
      Score int
      Foo interface{}
  }
  
  bp := bytepack.NewBytePack(5)
  bytepack.Reegister(foo{})
  
  f := foo {
      F: "test foo",
  }
  
  b := bar {
      Score: 100,	
      Foo:   f,
  }
  
  packedBytes, err := bp.Pack(b)
  if err != nil {
       panic(err)
  }
  err = bp.Unpack(packedBytes, &fUnpacked)
  if err != nil {
       panic(err)
  } 
  ```
  
* Packer can be used by itself without the BytePack. Just use `Pack` and `Unpack` methods of the packer.

---
## Overriding Pack and Unpack

BytePack allows overriding how structs are serialized and deserialized by implementing `Packable` interface.
A struct needs to have `Pack(p *Packer) error` and `Unpack(p *Packer, buf BPReader) error` methods. 
Below is an example of a simple struct implementing `Packable`:

```go
type personS struct {
    Name   string
    Age    int32
    Height float64
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
```

When encoding a struct that implements `Packable`, it is more efficient to pass a pointer to a struct: `packedBytes, err := bp.Pack(&f)`.

For packing different values, use following exported methods in Packer:

* `PackString(str string) error`
* `PackInt32(ival int32) error`
* `PackInt64(ival int64) error`
* `PackInt(ival int) error`
* `PackInt8(ival int8) error`
* `PackInt16(ival int16) error`
* `PackUint(ival uint) error`
* `PackUint8(uival uint8) error`
* `PackUint16(uival uint16) error`
* `PackUint32(uival uint32) error`
* `PackUint64(uival uint64) error`
* `PackFloat64(fval float64) error`
* `PackFloat32(fval float32) error`
* `PackBool(bval bool) error`
* `PackStruct(obj interface{}) error`
* `PackSlice(slice interface{}) error`
* `PackMap(m interface{}) error`

For unpacking the values:

* `UnpackString(buf BPReader) (string, error)`
* `UnpackInt8(buf BPReader) (int8, error)`
* `UnpackInt16(buf BPReader) (int16, error)`
* `UnpackInt32(buf BPReader) (int32, error)`
* `UnpackInt64(buf BPReader) (int64, error)`
* `UnpackInt(buf BPReader) (int, error)`
* `UnpackUint8(buf BPReader) (uint8, error)`
* `UnpackUint16(buf BPReader) (uint16, error)`
* `UnpackUint32(buf BPReader) (uint32, error)`
* `UnpackUint64(buf BPReader) (uint64, error)`
* `UnpackUint(buf BPReader) (uint, error)`
* `UnpackFloat64(buf BPReader) (float64, error)`
* `UnpackFloat32(buf BPReader) (float32, error)`
* `UnpackBool(buf BPReader) (bool, error)`
* `UnpackStruct(buf BPReader, i interface{}) error`
* `UnpackArray(arrayType reflect.Type, buf BPReader) (*reflect.Value, error)`
* `UnpackSlice(sliceType reflect.Type, buf BPReader) (*reflect.Value, error)`
* `UnpackMap(mapType reflect.Type, buf BPReader) (*reflect.Value, error)`
