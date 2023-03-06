package bytepack

import (
	"io"
	"reflect"
	"sync"
)

var registeredStructs = make(map[string]reflect.Type, 0)
var regLock sync.RWMutex

func Register(strct interface{}) {
	regLock.Lock()
	defer regLock.Unlock()
	rtype := reflect.TypeOf(strct)
	if rtype.Kind() == reflect.Ptr {
		rtype = rtype.Elem()
	}
	if rtype.Kind() == reflect.Struct {
		registeredStructs[rtype.PkgPath()+rtype.Name()] = rtype
	}
}

type BytePack struct {
	pool chan *Packer
}

func NewBytePack(numPackers int) *BytePack {
	bp := BytePack{
		pool: make(chan *Packer, numPackers),
	}

	for i := 0; i < numPackers; i++ {
		bp.pool <- NewPacker()
	}

	return &bp
}

func (p *BytePack) Pack(strct interface{}) ([]byte, error) {
	// get the packer from the pool
	s := <-p.pool
	bytes, err := s.Pack(strct)
	// now put the packer back into the pool
	p.pool <- s
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (p *BytePack) Unpack(data []byte, strct interface{}) error {
	// get the packer from the pool
	s := <-p.pool
	err := s.Unpack(data, strct)
	// now put the packer back into the pool
	p.pool <- s
	if err != nil {
		return err
	}
	return nil
}

func (p *BytePack) UnpackFromReader(reader BPReader, strct interface{}) error {
	// get the packer from the pool
	s := <-p.pool
	err := s.UnpackFromReader(reader, strct)
	// now put the packer back into the pool
	p.pool <- s
	if err != nil {
		return err
	}
	return nil
}

type bytePackReader struct {
	io.Reader
}

func (r bytePackReader) ReadByte() (byte, error) {
	var b [1]byte
	bs := b[:1]
	_, err := r.Read(bs)
	if err != nil {
		return 0, err
	}
	return bs[0], nil
}

func (p *BytePack) UnpackFromIOReader(reader io.Reader, strct interface{}) error {
	// get the packer from the pool
	s := <-p.pool
	bpr := bytePackReader{reader}
	err := s.UnpackFromReader(bpr, strct)
	// now put the packer back into the pool
	p.pool <- s
	if err != nil {
		return err
	}
	return nil
}
