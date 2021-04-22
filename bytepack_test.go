package bytepack

import (
    "github.com/stretchr/testify/assert"
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
    for i := 0; i < numPackers * multiplier; i++ {
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
        } ()
    }
    wg.Wait()
}
