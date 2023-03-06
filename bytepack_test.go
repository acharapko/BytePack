package bytepack

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"io"
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

func TestBytePack_PackMultiple(t *testing.T) {
	numPackers := 5
	multiplier := 1000
	bp := NewBytePack(numPackers)
	rnd := rand.New(rand.NewSource(123))
	var wg sync.WaitGroup
	wg.Add(numPackers * multiplier)
	for i := 0; i < numPackers*multiplier; i++ {
		go func() {
			j := rnd.Intn(1000)
			p := person{
				Name:   "Test" + strconv.Itoa(j),
				Age:    int32(j),
				Height: float64(j) + 0.75,
			}
			packedBytes, err := bp.Pack(p)
			assert.Nil(t, err)
			assert.True(t, len(packedBytes) > 0)
			var p2 person
			err = bp.Unpack(packedBytes, &p2)
			assert.Nil(t, err)
			assert.Equal(t, p.Age, p2.Age)
			assert.Equal(t, p.Height, p2.Height)
			assert.Equal(t, p.Name, p2.Name)
			wg.Done()
		}()
	}
	wg.Wait()
}

func Test_SerializeWithByteSliceAndUnpackingFromReader(t *testing.T) {

	bp := NewBytePack(1)

	type fooByte struct {
		Id    uuid.UUID
		Bytes []byte
	}

	bufBytes := []byte{0, 0, 0, 6, 0, 0, 12, 25, 0, 1, 17, 69, 0, 0, 12, 25, 0, 1, 17, 70, 0, 0, 12, 25, 0, 1, 17, 69,
		0, 0, 12, 25, 0, 1, 17, 70, 0, 0, 12, 25, 0, 1, 17, 69, 0, 0, 12, 25, 0, 1, 17, 70}
	id := uuid.New()
	a := fooByte{
		Id:    id,
		Bytes: bufBytes,
	}

	buf, err := bp.Pack(a)
	assert.NoError(t, err)
	var a2 fooByte
	var reader = bytes.NewBuffer(buf)
	err = bp.UnpackFromReader(reader, &a2)
	assert.NoError(t, err)
	assert.Equal(t, a.Id, a2.Id)
	assert.Equal(t, len(a.Bytes), len(a2.Bytes))
}

func Test_ByteSlice(t *testing.T) {

	bp := NewBytePack(1)

	type fooByte struct {
		Id    uuid.UUID
		Bytes []byte
	}

	bufBytes := []byte{0, 0, 0, 6, 0, 0, 12, 25, 0, 1, 17, 69, 0, 0, 12, 25, 0, 1, 17, 70, 0, 0, 12, 25, 0, 1, 17, 69,
		0, 0, 12, 25, 0, 1, 17, 70, 0, 0, 12, 25, 0, 1, 17, 69, 0, 0, 12, 25, 0, 1, 17, 70}
	id := uuid.New()
	a := fooByte{
		Id:    id,
		Bytes: bufBytes,
	}

	buf, err := bp.Pack(a)
	assert.NoError(t, err)
	var a2 fooByte
	err = bp.Unpack(buf, &a2)
	assert.NoError(t, err)
	assert.Equal(t, a.Id, a2.Id)
	assert.Equal(t, len(a.Bytes), len(a2.Bytes))
}

func Benchmark_SerializeWithByteSliceAndUnpackingFromReader(b *testing.B) {

	bp := NewBytePack(1)

	type fooByte struct {
		Id    uuid.UUID
		Bytes []byte
	}

	bufBytes := []byte{0, 0, 0, 6, 0, 0, 12, 25, 0, 1, 17, 69, 0, 0, 12, 25, 0, 1, 17, 70, 0, 0, 12, 25, 0, 1, 17, 69,
		0, 0, 12, 25, 0, 1, 17, 70, 0, 0, 12, 25, 0, 1, 17, 69, 0, 0, 12, 25, 0, 1, 17, 70}
	id := uuid.New()
	a := fooByte{
		Id:    id,
		Bytes: bufBytes,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, err := bp.Pack(a)
		if err != nil {
			panic(err)
		}
		var a2 fooByte
		var reader = bytes.NewBuffer(buf)
		err = bp.UnpackFromReader(reader, &a2)
		if err != nil {
			panic(err)
		}
	}
}

func Benchmark_SerializeWithByteSliceAndUnpackingFromIOReader(b *testing.B) {

	bp := NewBytePack(1)

	type fooByte struct {
		Id    uuid.UUID
		Bytes []byte
	}

	bufBytes := []byte{0, 0, 0, 6, 0, 0, 12, 25, 0, 1, 17, 69, 0, 0, 12, 25, 0, 1, 17, 70, 0, 0, 12, 25, 0, 1, 17, 69,
		0, 0, 12, 25, 0, 1, 17, 70, 0, 0, 12, 25, 0, 1, 17, 69, 0, 0, 12, 25, 0, 1, 17, 70}
	id := uuid.New()
	a := fooByte{
		Id:    id,
		Bytes: bufBytes,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, err := bp.Pack(a)
		if err != nil {
			panic(err)
		}
		var a2 fooByte
		var ioreader io.Reader = bytes.NewBuffer(buf)
		err = bp.UnpackFromIOReader(ioreader, &a2)
		if err != nil {
			panic(err)
		}
	}
}
