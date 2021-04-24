package bytepack

import (
    "bytes"
    "encoding/binary"
    "errors"
    "fmt"
    "io"
    "math"
    "reflect"
    "strconv"
)

var intSize = strconv.IntSize / 8 // intSize in bytes
var packableType = reflect.TypeOf((*Packable)(nil)).Elem()


type BPReader interface {
    io.ByteReader
    io.Reader
}

// Packable
/*
   Packable interface allows struct to implement own Pack and Unpack methods for data serialization and deserialization.
*/
type Packable interface {
    Pack(s *Packer) error
    Unpack(s *Packer, buf BPReader) error
}

type Packer struct {
    w *bytes.Buffer
}

func NewPacker() *Packer {
    return &Packer{
        w: new(bytes.Buffer),
    }
}

/*~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *                                                  Packing
 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

func (s *Packer) Pack(obj interface{}) ([]byte, error) {
    err := s.encode(obj)
    if err != nil {
        return nil, err
    }
    retBytes := make([]byte, s.w.Len())
    copy(retBytes, s.w.Bytes()) // make a copy of a slice, so we can reuse the buffer right away without overwriting
    s.w.Reset()
    return retBytes, nil
}

func (s *Packer) encode(obj interface{}) error {
    switch obj.(type) {
    case Packable:
        return obj.(Packable).Pack(s)
    }

    t := reflect.TypeOf(obj)
    if t.Kind() == reflect.Struct {
        v := reflect.ValueOf(obj)
        p := reflect.New(t)
        piface := p.Interface()
        switch piface.(type) {
        case Packable:
            p.Elem().Set(v)
            return piface.(Packable).Pack(s)
        }
        err := s.encodeStruct(v)
        if err != nil {
            return err
        }
    } else if t.Kind() == reflect.Ptr {
        v := reflect.ValueOf(obj)
        //err := s.encodePointer(v)
        if v.IsNil() {
            return errors.New("cannot encode nil struct")
        }
        if v.Elem().Kind() == reflect.Struct {
            err := s.encodeStruct(v.Elem())
            if err != nil {
                return err
            }
        } else if v.Elem().Kind() == reflect.Interface {
            return s.encode(v.Elem().Interface())
        } else {
            return errors.New(fmt.Sprintf("cannot encode non-struct. Got: %v", v.Elem().Kind()))
        }
    }

    return nil
}

/*-----------------------------------
  Reflection-based default encoding
 -----------------------------------*/

func (s *Packer) encodeStruct(v reflect.Value) error {
    for i := 0; i < v.NumField(); i++ {
        err := s.encodeValue(v.Field(i))
        if err != nil {
            return err
        }
    }
    return nil
}

func (s *Packer) encodeValue(val reflect.Value) error {
    switch val.Kind() {
    case reflect.Struct:
        err := s.encodeStruct(val)
        //err := s.encode(val.Interface())
        if err != nil {
            return err
        }
    case reflect.String:
        err := s.PackString(val.String())
        if err != nil {
            return err
        }
    case reflect.Int8:
        err := s.PackInt8(int8(val.Int()))
        if err != nil {
            return err
        }
    case reflect.Int16:
        err := s.PackInt16(int16(val.Int()))
        if err != nil {
            return err
        }
    case reflect.Int32:
        err := s.PackInt32(int32(val.Int()))
        if err != nil {
            return err
        }
    case reflect.Int:
        err := s.PackInt(int(val.Int()))
        if err != nil {
            return err
        }
    case reflect.Int64:
        err := s.PackInt64(val.Int())
        if err != nil {
            return err
        }
    case reflect.Uint8:
        err := s.PackUint8(uint8(val.Uint()))
        if err != nil {
            return err
        }
    case reflect.Uint16:
        err := s.PackUint16(uint16(val.Uint()))
        if err != nil {
            return err
        }
    case reflect.Uint32:
        err := s.PackUint32(uint32(val.Uint()))
        if err != nil {
            return err
        }
    case reflect.Uint:
        err := s.PackUint(uint(val.Uint()))
        if err != nil {
            return err
        }
    case reflect.Uint64:
        err := s.PackUint64(val.Uint())
        if err != nil {
            return err
        }
    case reflect.Float32:
        err := s.PackFloat32(float32(val.Float()))
        if err != nil {
            return err
        }
    case reflect.Float64:
        err := s.PackFloat64(val.Float())
        if err != nil {
            return err
        }
    case reflect.Bool:
        err := s.PackBool(val.Bool())
        if err != nil {
            return err
        }
    case reflect.Slice:
        err := s.encodeSlice(val)
        if err != nil {
            return err
        }
    case reflect.Array:
        err := s.encodeArray(val)
        if err != nil {
            return err
        }
    case reflect.Map:
        err := s.encodeMap(val)
        if err != nil {
            return err
        }
    case reflect.Ptr:
        err := s.encodePointer(val)
        if err != nil {
            return err
        }
    case reflect.Interface:
        err := s.encodeInterface(val)
        if err != nil {
            return err
        }
    }
    return nil
}

func (s *Packer) encodeValueWithType(val reflect.Value) error {
    switch val.Kind() {
    case reflect.Ptr:
        dataType := val.Elem().Type()
        err := s.PackBool(true) // true for pointer type
        if err != nil {
            return err
        }
        err = s.PackString(dataType.PkgPath() + dataType.Name())
        if err != nil {
            return err
        }
        err = s.encodeValue(val.Elem())
        if err != nil {
            return err
        }
    case reflect.Struct:
        dataType := val.Type()
        //log.Debugf("dt := %v", dataType)
        err := s.PackBool(false) // false for non-pointers
        if err != nil {
            return err
        }
        err = s.PackString(dataType.PkgPath() + dataType.Name())
        if err != nil {
            return err
        }
        //err = s.encodeStruct(val)
        err = s.encode(val.Interface())
        if err != nil {
            return err
        }
    }
    return nil
}

func (s *Packer) encodePointer(ptr reflect.Value) error {
    if ptr.IsNil() {
        err := s.PackBool(false)
        if err != nil {
            return err
        }
    } else {
        err := s.PackBool(true)
        if err != nil {
            return err
        }
        err = s.encodeValue(ptr.Elem())
        if err != nil {
            return err
        }
    }
    return nil
}

func (s *Packer) encodeInterface(iface reflect.Value) error {
    if iface.IsNil() {
        err := s.PackBool(false)
        if err != nil {
            return err
        }
    } else {
        err := s.PackBool(true)
        if err != nil {
            return err
        }
        err = s.encodeValueWithType(iface.Elem())
        if err != nil {
            return err
        }
        //log.Debugf("iface elem kind: %v", iface.Elem().Kind())
    }
    return nil
}

func (s *Packer) encodeMap(m reflect.Value) error {
    if m.IsNil() {
        return s.PackBool(true)
    }
    err := s.PackBool(false)
    // write down the number of kv-pairs
    mapLen := m.Len()
    err = s.PackInt32(int32(mapLen))
    if err != nil {
        return err
    }

    for _, key := range m.MapKeys() {
        //write key
        err = s.encodeValue(key)
        if err != nil {
            return err
        }
        //write value
        val := m.MapIndex(key)
        err = s.encodeValue(val)
        if err != nil {
            return err
        }
    }
    return nil
}

func (s *Packer) encodeArray(arrayValue reflect.Value) error {
    // when dealing with slices, first write the number of elements
    arrayLen := arrayValue.Len()
    arrayKind := reflect.TypeOf(arrayValue.Interface()).Elem().Kind()
    switch arrayKind {
    case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Bool, reflect.Float32, reflect.Float64:
        err := binary.Write(s.w, binary.BigEndian, arrayValue.Interface())
        if err != nil {
            return err
        }
    case reflect.Uint8:
        for i := 0; i < arrayLen; i++ {
            err := s.w.WriteByte(byte(arrayValue.Index(i).Uint()))
            if err != nil {
                return err
            }
        }
        /*_, err := s.w.Write(arrayValue.
        if err != nil {
            return err
        }*/
    case reflect.Int:
        err := s.writeIntSliceOrArray(arrayValue)
        if err != nil {
            return err
        }
    case reflect.String:
        for i := 0; i < arrayLen; i++ {
            err := s.PackString(arrayValue.Index(i).String())
            if err != nil {
                return err
            }
        }
    case reflect.Struct:
        for i := 0; i < arrayLen; i++ {
            //err := s.encodeStruct(arrayValue.Index(i))
            err := s.encode(arrayValue.Index(i).Interface())
            if err != nil {
                return err
            }
        }
    case reflect.Map:
        for i := 0; i < arrayLen; i++ {
            err := s.encodeMap(arrayValue.Index(i))
            if err != nil {
                return err
            }
        }
    case reflect.Ptr:
        for i := 0; i < arrayLen; i++ {
            err := s.encodePointer(arrayValue.Index(i))
            if err != nil {
                return err
            }
        }
    }
    return nil
}

func (s *Packer) encodeSlice(sliceValue reflect.Value) error {
    if sliceValue.IsNil() {
        return s.PackBool(true)
    }
    err := s.PackBool(false)
    if err != nil {
        return err
    }
    // when dealing with slices, first write the number of elements
    sliceLen := sliceValue.Len()
    err = s.PackInt32(int32(sliceLen))
    if err != nil {
        return err
    }
    //valueField.Slice()
    sliceKind := reflect.TypeOf(sliceValue.Interface()).Elem().Kind()
    switch sliceKind {
    case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Bool, reflect.Float32, reflect.Float64:
        err = binary.Write(s.w, binary.BigEndian, sliceValue.Interface())
        if err != nil {
            return err
        }
    case reflect.Uint8:
        _, err = s.w.Write(sliceValue.Bytes())
        if err != nil {
            return err
        }
    case reflect.Int64:
        err = s.writeInt64SliceOrArray(sliceValue)
        if err != nil {
            return err
        }
    case reflect.Int:
        err = s.writeIntSliceOrArray(sliceValue)
        if err != nil {
            return err
        }
    case reflect.String:
        for i := 0; i < sliceLen; i++ {
            err = s.PackString(sliceValue.Index(i).String())
            if err != nil {
                return err
            }
        }
    case reflect.Struct:
        for i := 0; i < sliceLen; i++ {
            err = s.encodeStruct(sliceValue.Index(i))
            //err := s.encode(sliceValue.Index(i).Interface())
            if err != nil {
                return err
            }
        }
    case reflect.Map:
        for i := 0; i < sliceLen; i++ {
            err = s.encodeMap(sliceValue.Index(i))
            if err != nil {
                return err
            }
        }
    case reflect.Ptr:
        for i := 0; i < sliceLen; i++ {
            err = s.encodePointer(sliceValue.Index(i))
            if err != nil {
                return err
            }
        }
    }
    return nil
}

/*-----------------------------------
  helpers of Packable encoders
 -----------------------------------*/

func (s *Packer) PackString(str string) error {
    err := s.PackInt32(int32(len(str)))
    if err != nil {
        return err
    }
    _, err = s.w.WriteString(str)
    return err
}

func (s *Packer) writeInt64SliceOrArray(arrayValue reflect.Value) error {
    arrayLen := arrayValue.Len()
    bs := make([]byte, 8 * arrayLen)
    for i := 0; i < arrayLen; i++ {
        uival := uint64(arrayValue.Index(i).Int())
        bs[8*i] = byte(uival >> 56)
        bs[8*i+1] = byte(uival >> 48)
        bs[8*i+2] = byte(uival >> 40)
        bs[8*i+3] = byte(uival >> 32)
        bs[8*i+4] = byte(uival >> 24)
        bs[8*i+5] = byte(uival >> 16)
        bs[8*i+6] = byte(uival >> 8)
        bs[8*i+7] = byte(uival)
    }
    _, err := s.w.Write(bs)
    return err
}

func (s *Packer) writeIntSliceOrArray(arrayValue reflect.Value) error {
    arrayLen := arrayValue.Len()
    if intSize == 8 {
        bs := make([]byte, 8 * arrayLen)
        for i := 0; i < arrayLen; i++ {
            uival := uint64(arrayValue.Index(i).Int())
            bs[8*i] = byte(uival >> 56)
            bs[8*i+1] = byte(uival >> 48)
            bs[8*i+2] = byte(uival >> 40)
            bs[8*i+3] = byte(uival >> 32)
            bs[8*i+4] = byte(uival >> 24)
            bs[8*i+5] = byte(uival >> 16)
            bs[8*i+6] = byte(uival >> 8)
            bs[8*i+7] = byte(uival)
        }
        _, err := s.w.Write(bs)
        return err
    } else if intSize == 4 {
        bs := make([]byte, 4 * arrayLen)
        for i := 0; i < arrayLen; i++ {
            uival := uint32(arrayValue.Index(i).Int())
            bs[8*i] = byte(uival >> 24)
            bs[8*i+1] = byte(uival >> 16)
            bs[8*i+2] = byte(uival >> 8)
            bs[8*i+3] = byte(uival)
        }
        _, err := s.w.Write(bs)
        return err
    }
    return errors.New("unknown int size")
}

func (s *Packer) PackInt32(ival int32) error {
    uival := uint32(ival)
    err := s.w.WriteByte(byte(uival >> 24))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival >> 16))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival >> 8))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival))
    return err
}

func (s *Packer) PackInt64(ival int64) error {
    uival := uint64(ival)
    err := s.w.WriteByte(byte(uival >> 56))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival >> 48))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival >> 40))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival >> 32))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival >> 24))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival >> 16))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival >> 8))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival))
    return err
}

func (s *Packer) PackInt(ival int) error {
    if intSize == 8 {
        return s.PackInt64(int64(ival))
    } else if intSize == 4 {
        return s.PackInt32(int32(ival))
    }
    return errors.New("unknown int size")
}

func (s *Packer) PackInt8(ival int8) error {
    return s.w.WriteByte(uint8(ival))
}

func (s *Packer) PackInt16(ival int16) error {
    uival := uint16(ival)
    err := s.w.WriteByte(byte(uival >> 8))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival))
    return err
}

func (s *Packer) PackUint(ival uint) error {
    return binary.Write(s.w, binary.BigEndian, ival)
}

func (s *Packer) PackUint8(uival uint8) error {
    return s.w.WriteByte(uival)
}

func (s *Packer) PackUint16(uival uint16) error {
    err := s.w.WriteByte(byte(uival >> 8))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival))
    return err
}

func (s *Packer) PackUint32(uival uint32) error {
    err := s.w.WriteByte(byte(uival >> 24))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival >> 16))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival >> 8))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival))
    return err
}

func (s *Packer) PackUint64(uival uint64) error {
    err := s.w.WriteByte(byte(uival >> 56))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival >> 48))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival >> 40))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival >> 32))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival >> 24))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival >> 16))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival >> 8))
    if err != nil {
        return err
    }
    err = s.w.WriteByte(byte(uival))
    return err
}

func (s *Packer) PackFloat64(fval float64) error {
    return s.PackUint64(math.Float64bits(fval))
}

func (s *Packer) PackFloat32(fval float32) error {
    return s.PackUint32(math.Float32bits(fval))
}

func (s *Packer) PackBool(bval bool) error {
    //return binary.Write(s.w, binary.BigEndian, bval)
    if bval {
        return s.w.WriteByte(1)
    } else {
        return s.w.WriteByte(0)
    }
}

func (s *Packer) PackStruct(obj interface{}) error {
    return s.encode(obj)
}

func (s *Packer) PackSlice(slice interface{}) error {
    sliceVal := reflect.ValueOf(slice)
    if sliceVal.Type().Kind() == reflect.Slice {
        err := s.encodeSlice(sliceVal)
        if err != nil {
            return err
        }
    } else {
        return errors.New("not a slice")
    }
    return nil
}

func (s *Packer) PackMap(m interface{}) error {
    mapVal := reflect.ValueOf(m)
    if mapVal.Type().Kind() == reflect.Map {
        err := s.encodeMap(mapVal)
        if err != nil {
            return err
        }
    } else {
        return errors.New("not a slice")
    }
    return nil
}


/*~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
 *                                                  Unpacking
 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

func (s *Packer) Unpack(data []byte, obj interface{}) error {
    buf := bytes.NewBuffer(data)
    switch obj.(type) {
    case Packable:
        return obj.(Packable).Unpack(s, buf)
    default:
        v := reflect.ValueOf(obj)
        if v.Kind() != reflect.Ptr {
            return errors.New("must pass a pointer to an object")
        }
        if v.Elem().Kind() == reflect.Struct {
            // so we found struct
            err := s.readStruct(buf, v.Elem())
            if err != nil {
                return err
            }

            return nil
        }
    }

    return nil
}

func (s *Packer) UnpackFromReader(buf BPReader, obj interface{}) error {
    switch obj.(type) {
    case Packable:
        return obj.(Packable).Unpack(s, buf)
    default:
        v := reflect.ValueOf(obj)
        if v.Kind() != reflect.Ptr {
            return errors.New("must pass a pointer to an object")
        }
        if v.Elem().Kind() == reflect.Struct {
            // so we found struct
            err := s.readStruct(buf, v.Elem())
            if err != nil {
                return err
            }

            return nil
        }
    }

    return nil
}

func (s *Packer) readStruct(buf BPReader, objVal reflect.Value) error {
    numFields := objVal.NumField()
    for i := 0; i < numFields; i++ {
        f := objVal.Field(i)
        ft := objVal.Field(i).Type()
        // we could have used the readBasicValues method here:
        // ---------------------
        /*val, err := s.readBasicValues(ft.Type, buf)
        if err != nil {
            return err
        }
        f.Set(val)*/
        // ---------------------
        //but it seems to have worse perf than rewriting it specific to the struct

        switch ft.Kind() {
        case reflect.Struct:
            //initializeStruct(ft.Type, f)
            st := reflect.New(ft)
            err := s.readStruct(buf, st.Elem())
            if err != nil {
                return err
            }
            f.Set(st.Elem())
        case reflect.Ptr:
            val, err := s.readPointer(ft, buf)
            if err != nil {
                return err
            }
            if !val.IsNil() {
                f.Set(val)
            }
        case reflect.String:
            str, err := s.UnpackString(buf)
            if err != nil {
                return err
            }
            f.SetString(str)
        case reflect.Int:
            if intSize == 8 {
                intVal, err := s.UnpackInt64(buf)
                if err != nil {
                    return err
                }
                f.SetInt(intVal)
            } else if intSize == 4 {
                intVal, err := s.UnpackInt32(buf)
                if err != nil {
                    return err
                }
                f.SetInt(int64(intVal))
            }
        case reflect.Int32:
            intVal, err := s.UnpackInt32(buf)
            if err != nil {
                return err
            }
            f.SetInt(int64(intVal))
        case reflect.Int64:
            intVal, err := s.UnpackInt64(buf)
            if err != nil {
                return err
            }
            f.SetInt(intVal)
        case reflect.Float64:
            floatVal, err := s.UnpackFloat64(buf)
            if err != nil {
                return err
            }
            f.SetFloat(floatVal)
        case reflect.Float32:
            floatVal, err := s.UnpackFloat32(buf)
            if err != nil {
                return err
            }
            f.SetFloat(float64(floatVal))
        case reflect.Bool:
            var boolVal bool
            err := binary.Read(buf, binary.BigEndian, &boolVal)
            if err != nil {
                return err
            }
            f.SetBool(boolVal)
        case reflect.Int8:
            intVal, err := s.UnpackInt8(buf)
            if err != nil {
                return err
            }
            f.SetInt(int64(intVal))
        case reflect.Int16:
            intVal, err := s.UnpackInt16(buf)
            if err != nil {
                return err
            }
            f.SetInt(int64(intVal))
        case reflect.Uint8:
            intVal, err := s.UnpackUint8(buf)
            if err != nil {
                return err
            }
            f.SetUint(uint64(intVal))
        case reflect.Uint16:
            intVal, err := s.UnpackUint16(buf)
            if err != nil {
                return err
            }
            f.SetUint(uint64(intVal))
        case reflect.Uint32:
            intVal, err := s.UnpackUint32(buf)
            if err != nil {
                return err
            }
            f.SetUint(uint64(intVal))
        case reflect.Uint64:
            intVal, err := s.UnpackUint64(buf)
            if err != nil {
                return err
            }
            f.SetUint(intVal)
        case reflect.Slice:
            sliceVal, err := s.UnpackSlice(ft, buf)
            if err != nil {
                return err
            }
            if sliceVal != nil {
                f.Set(*sliceVal)
            }
        case reflect.Array:
            arrayVal, err := s.UnpackArray(ft, buf)
            if err != nil {
                return err
            }
            f.Set(*arrayVal)
        case reflect.Map:
            decodedMap := reflect.MakeMap(ft)
            exists, err := s.readMap(ft, buf, decodedMap)
            if err != nil {
                return err
            }
            if exists {
                f.Set(decodedMap)
            }
        case reflect.Interface:
            val, err := s.readInterface(buf)
            if err != nil {
                return err
            }
            if val != nil {
                f.Set(*val)
            }
        case reflect.Chan:
            // do nothing with the chan and leave it nil
        default:
            return errors.New(fmt.Sprintf("decoding unsupported type %v", ft.Kind()))
        }
    }

    return nil
}

func (s *Packer) readMap(mapType reflect.Type, buf BPReader, readMap reflect.Value) (bool, error) {
	// read nil flag
	isNil, err := s.UnpackBool(buf)
	if isNil {
		return false, nil
	}

	numEntries, err := s.UnpackInt32(buf)
	if err != nil {
		return false, err
	}
	var mapKey reflect.Value
	var mapValue reflect.Value
	for i := 0; i < int(numEntries); i++ {
		// decode key
		mapKey, err = s.readBasicValues(mapType.Key(), buf)
		if err != nil {
			return false, err
		}
		//decode value
		mapValue, err = s.readBasicValues(mapType.Elem(), buf)
		if err != nil {
			return false, err
		}
		// set to map
		readMap.SetMapIndex(mapKey, mapValue)
	}
	return true, nil
}

func (s *Packer) readPointer(ptrType reflect.Type, buf BPReader) (reflect.Value, error) {
    // first read ptr nil flag
    notNil, err := s.UnpackBool(buf)
    if err != nil {
        return reflect.New(ptrType).Elem(), err
    }
    if notNil {
        val, err := s.readBasicValues(ptrType.Elem(), buf)
        if err != nil {
            return reflect.New(ptrType).Elem(), err
        }
        return val.Addr(), err
    }
    return reflect.New(ptrType).Elem(), nil
}

func (s *Packer) readInterface(buf BPReader) (*reflect.Value, error) {
    // first read interface nil flag
    notNil, err := s.UnpackBool(buf)
    if err != nil {
        return nil, err
    }
    if notNil {
        // read if pointer
        isPointer, err := s.UnpackBool(buf)
        if err != nil {
            return nil, err
        }
        // read the type
        typeStr, err := s.UnpackString(buf)
        if err != nil {
            return nil, err
        }
        if ifaceType, exists := registeredStructs[typeStr]; exists {
            val, err := s.readBasicValues(ifaceType, buf)
            if err != nil {
                return nil, err
            }
            if isPointer {
                val = val.Addr()
            }
            return &val, err
        } else {
            return nil, errors.New(fmt.Sprintf("type %s is not registered", typeStr))
        }
    }
    return nil, nil
}

func (s *Packer) readBasicValues(valType reflect.Type, buf BPReader) (reflect.Value, error) {
    val := reflect.New(valType)
    switch valType.Kind() {
    case reflect.Int:
        if intSize == 8 {
            intVal, err := s.UnpackInt64(buf)
            if err != nil {
                return val, err
            }
            val = reflect.ValueOf(intVal)
        } else if intSize == 4 {
            intVal, err := s.UnpackInt32(buf)
            if err != nil {
                return val, err
            }
            val = reflect.ValueOf(intVal)
        }
    case reflect.Int8:
        intVal, err := s.UnpackInt8(buf)
        if err != nil {
            return val, err
        }
        val = reflect.ValueOf(intVal)
    case reflect.Int16:
        intVal, err := s.UnpackInt16(buf)
        if err != nil {
            return val, err
        }
        val = reflect.ValueOf(intVal)
    case reflect.Uint:
        if intSize == 8 {
            intVal, err := s.UnpackUint64(buf)
            if err != nil {
                return val, err
            }
            val = reflect.ValueOf(intVal)
        } else if intSize == 4 {
            intVal, err := s.UnpackUint32(buf)
            if err != nil {
                return val, err
            }
            val = reflect.ValueOf(intVal)
        }
    case reflect.Uint8:
        intVal, err := s.UnpackInt8(buf)
        if err != nil {
            return val, err
        }
        val = reflect.ValueOf(intVal)
    case reflect.Uint16:
        intVal, err := s.UnpackInt16(buf)
        if err != nil {
            return val, err
        }
        val = reflect.ValueOf(intVal)
    case reflect.Uint32:
        intVal, err := s.UnpackInt32(buf)
        if err != nil {
            return val, err
        }
        val = reflect.ValueOf(intVal)
    case reflect.Uint64:
        intVal, err := s.UnpackInt64(buf)
        if err != nil {
            return val, err
        }
        val = reflect.ValueOf(intVal)
    case reflect.Int32:
        intVal, err := s.UnpackInt32(buf)
        if err != nil {
            return val, err
        }
        val = reflect.ValueOf(intVal)
    case reflect.Int64:
        intVal, err := s.UnpackInt64(buf)
        if err != nil {
            return val, err
        }
        val = reflect.ValueOf(intVal)
    case reflect.Float64:
        floatVal, err := s.UnpackFloat64(buf)
        if err != nil {
            return val, err
        }
        val = reflect.ValueOf(floatVal)
    case reflect.Float32:
        floatVal, err := s.UnpackFloat32(buf)
        if err != nil {
            return val, err
        }
        val = reflect.ValueOf(floatVal)
    case reflect.Bool:
        var boolVal bool
        err := binary.Read(buf, binary.BigEndian, &boolVal)
        if err != nil {
            return val, err
        }
        val = reflect.ValueOf(boolVal)
    case reflect.String:
        strVal, err := s.UnpackString(buf)
        if err != nil {
            return val, err
        }
        val = reflect.ValueOf(strVal)
    case reflect.Struct:
        err := s.readStruct(buf, val.Elem())
        if err != nil {
            return val, err
        }
        val = val.Elem()
    case reflect.Slice:
        sliceVal, err := s.UnpackSlice(valType, buf)
        if err != nil {
            return val, err
        }
        val = *sliceVal
        //log.Debugf("slice type = %v", ft.Type)
    case reflect.Map:
        decodedMap := reflect.MakeMap(valType)
        _, err := s.readMap(valType, buf, decodedMap)
        if err != nil {
            return val, err
        }
        val = decodedMap
    case reflect.Ptr:
        var err error
        val, err = s.readPointer(valType, buf)
        if err != nil {
            return val, err
        }
    case reflect.Chan:
        // do nothing with the chan and leave it nil
    }
    return val, nil
}

/*-----------------------------------
  helpers of Packable decoders
 -----------------------------------*/

func (s *Packer) UnpackString(buf BPReader) (string, error) {
    strLen, err := s.UnpackInt32(buf)
    if err != nil {
        return "", err
    }

    strBuf := make([]byte, strLen)
    _, err = buf.Read(strBuf)
    if err != nil {
        return "", err
    }
    return string(strBuf), nil
}

func (s *Packer) UnpackInt8(buf BPReader) (int8, error) {
    var err error
    b1, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    return int8(b1), nil
}

func (s *Packer) UnpackInt16(buf BPReader) (int16, error) {
    var i uint16
    var err error
    b1, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b2, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    i = uint16(b2) | uint16(b1)<<8
    return int16(i), nil
}

func (s *Packer) UnpackInt32(buf BPReader) (int32, error) {
    var i uint32
    var err error
    b1, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b2, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b3, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b4, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    i = uint32(b4) | uint32(b3)<<8 | uint32(b2)<<16 | uint32(b1)<<24
    return int32(i), nil
}

func (s *Packer) UnpackInt64(buf BPReader) (int64, error) {
    var i uint64
    var err error
    b1, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b2, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b3, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b4, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b5, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b6, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b7, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b8, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    i = uint64(b8) | uint64(b7)<<8 | uint64(b6)<<16 | uint64(b5)<<24 | uint64(b4)<<32 | uint64(b3)<<40 | uint64(b2)<<48 | uint64(b1)<<56
    return int64(i), nil
}

func (s *Packer) UnpackInt(buf BPReader) (int, error) {
    if intSize == 8 {
        i, err := s.UnpackInt64(buf)
        return int(i), err
    } else if intSize == 4 {
        i, err := s.UnpackInt32(buf)
        return int(i), err
    }
    return 0, errors.New("int must be 4 or 8 bytes depending on the system")
}

func (s *Packer) UnpackUint8(buf BPReader) (uint8, error) {
    return buf.ReadByte()
}

func (s *Packer) UnpackUint16(buf BPReader) (uint16, error) {
    var i uint16
    var err error
    b1, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b2, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    i = uint16(b2) | uint16(b1)<<8
    return i, nil
}

func (s *Packer) UnpackUint32(buf BPReader) (uint32, error) {
    var i uint32
    var err error
    b1, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b2, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b3, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b4, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    i = uint32(b4) | uint32(b3)<<8 | uint32(b2)<<16 | uint32(b1)<<24
    return i, nil
}

func (s *Packer) UnpackUint64(buf BPReader) (uint64, error) {
    var i uint64
    var err error
    b1, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b2, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b3, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b4, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b5, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b6, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b7, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    b8, err := buf.ReadByte()
    if err != nil {
        return 0, err
    }
    i = uint64(b8) | uint64(b7)<<8 | uint64(b6)<<16 | uint64(b5)<<24 | uint64(b4)<<32 | uint64(b3)<<40 | uint64(b2)<<48 | uint64(b1)<<56
    return i, nil
}

func (s *Packer) UnpackUint(buf BPReader) (uint, error) {
    if intSize == 8 {
        i, err := s.UnpackUint64(buf)
        return uint(i), err
    } else if intSize == 4 {
        i, err := s.UnpackUint32(buf)
        return uint(i), err
    }
    return 0, errors.New("int must be 4 or 8 bytes depending on the system")
}

func (s *Packer) UnpackFloat64(buf BPReader) (float64, error) {
    bits, err := s.UnpackUint64(buf)
    if err != nil {
        return 0, err
    }
    f := math.Float64frombits(bits)
    return f, err
}

func (s *Packer) UnpackFloat32(buf BPReader) (float32, error) {
    bits, err := s.UnpackUint32(buf)
    if err != nil {
        return 0, err
    }
    f := math.Float32frombits(bits)
    return f, err
}

func (s *Packer) UnpackBool(buf BPReader) (bool, error) {
    var err error
    b1, err := buf.ReadByte()
    if err != nil {
        return false, err
    }
    return b1 != 0, nil
}

func (s *Packer) UnpackStruct(buf BPReader, i interface{}) error {
    iVal := reflect.ValueOf(i)
    if iVal.Kind() == reflect.Ptr {
        // expect a pointer to a struct
        iVal = iVal.Elem()
    } else {
        return errors.New("expect a pointer to a struct")
    }
    if iVal.Kind() == reflect.Struct {
        err := s.UnpackFromReader(buf, i)
        if err != nil {
            return err
        }
    } else {
        return errors.New("not a struct")
    }
    return nil
}

func (s *Packer) UnpackArray(arrayType reflect.Type, buf BPReader) (*reflect.Value, error) {
	// first find out how many items are in the slice
	arrayKind := arrayType.Elem().Kind()
	var err error
	switch arrayKind {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Bool, reflect.Float32, reflect.Float64:
		arrayValue := reflect.New(arrayType)
		err = binary.Read(buf, binary.BigEndian, arrayValue.Interface())
		if err != nil {
			return nil, err
		}
		a := arrayValue.Elem()
		return &a, nil
	case reflect.Uint8:
		arrayValue := reflect.New(arrayType).Elem()
		numEntries := arrayValue.Len()
		for i := 0; i < numEntries; i++ {
			bval, err := s.UnpackUint8(buf)
			if err != nil {
				return nil, err
			}
			arrayValue.Index(i).SetUint(uint64(bval))

		}
		return &arrayValue, nil
	case reflect.Int:
		arrayValue := reflect.New(arrayType).Elem()
		numEntries := arrayValue.Len()
		if intSize == 8 {
			var ival int64
			for i := 0; i < numEntries; i++ {
				ival, err = s.UnpackInt64(buf)
				if err != nil {
					return nil, err
				}
				arrayValue.Index(i).SetInt(ival)
			}
		} else if intSize == 4 {
			var ival int32
			for i := 0; i < numEntries; i++ {
				ival, err = s.UnpackInt32(buf)
				if err != nil {
					return nil, err
				}
				arrayValue.Index(i).SetInt(int64(ival))
			}
		}
		return &arrayValue, nil
	case reflect.String:
		arrayValue := reflect.New(arrayType).Elem()
		numEntries := arrayValue.Len()
		for i := 0; i < numEntries; i++ {
			str, err := s.UnpackString(buf)
			if err != nil {
				return nil, err
			}
			arrayValue.Index(i).SetString(str)
		}
		return &arrayValue, nil
	case reflect.Struct:
		arrayValue := reflect.New(arrayType).Elem()
		numEntries := arrayValue.Len()
		for i := 0; i < numEntries; i++ {
			err = s.readStruct(buf, arrayValue.Index(i))
			if err != nil {
				return nil, err
			}
		}
		return &arrayValue, nil
	default:
		return nil, errors.New("unknown slice type")
	}
}

func (s *Packer) UnpackSlice(sliceType reflect.Type, buf BPReader) (*reflect.Value, error) {
	// read nil flag
	isNil, err := s.UnpackBool(buf)
	if isNil {
		return nil, nil
	}
	// first find out how many items are in the slice
	numEntries, err := s.UnpackInt32(buf)
	if err != nil {
		return nil, err
	}
	sliceKind := sliceType.Elem().Kind()
	switch sliceKind {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Bool, reflect.Float32, reflect.Float64:
		sliceValue := reflect.MakeSlice(sliceType, int(numEntries), int(numEntries))
		err = binary.Read(buf, binary.BigEndian, sliceValue.Interface())
		if err != nil {
			return nil, err
		}
		return &sliceValue, nil
	case reflect.Uint8:
		b := make([]byte, int(numEntries), int(numEntries))
		_, err = buf.Read(b)
		if err != nil {
			return nil, err
		}
		reflectedB := reflect.ValueOf(b)
		return &reflectedB, nil
	case reflect.Int:
		intSlice := make([]int, numEntries, numEntries)
		if intSize == 8 {
			var ival int64
			for i := 0; i < int(numEntries); i++ {
				ival, err = s.UnpackInt64(buf)
				if err != nil {
					return nil, err
				}
				intSlice[i] = int(ival)
			}
		} else if intSize == 4 {
			var ival int32
			for i := 0; i < int(numEntries); i++ {
				ival, err = s.UnpackInt32(buf)
				if err != nil {
					return nil, err
				}
				intSlice[i] = int(ival)
			}
		}
		sliceVal := reflect.ValueOf(intSlice)
		return &sliceVal, nil
	case reflect.String:
		strSlice := make([]string, numEntries, numEntries)
		for i := 0; i < int(numEntries); i++ {
			strSlice[i], err = s.UnpackString(buf)
			if err != nil {
				return nil, err
			}
		}
		sliceVal := reflect.ValueOf(strSlice)
		return &sliceVal, nil
	case reflect.Struct:
		sliceValue := reflect.MakeSlice(sliceType, int(numEntries), int(numEntries))
		for i := 0; i < int(numEntries); i++ {
			err = s.readStruct(buf, sliceValue.Index(i))
			if err != nil {
				return nil, err
			}
		}
		return &sliceValue, nil
	default:
		return nil, errors.New("unknown slice type")
	}
}

func (s *Packer) UnpackMap(mapType reflect.Type, buf BPReader) (*reflect.Value, error) {
	decodedMap := reflect.MakeMap(mapType)
	_, err := s.readMap(mapType, buf, decodedMap)
	if err != nil {
		return nil, err
	}
	return &decodedMap, nil
}
