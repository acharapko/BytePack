package bytepack

import (
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
        registeredStructs[rtype.PkgPath() + rtype.Name()] = rtype
    }
}



type BytePack struct {
    pool chan *Packer
}

func NewBytePack(numPackers int) *BytePack {
    bp := BytePack{
        pool: make(chan *Packer, numPackers),
    }

    for i:=0; i < numPackers; i++ {
        bp.pool <- NewPacker()
    }

    return &bp
}

func (p *BytePack) Pack(strct interface{}) ([]byte, error) {
    // get the packer from the pool
    s := <- p.pool
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
    s := <- p.pool
    err := s.Unpack(data, strct)
    // now put the packer back into the pool
    p.pool <- s
    if err != nil {
        return err
    }
    return nil
}