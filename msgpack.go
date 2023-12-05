package jewl

import (
	"github.com/vmihailenco/msgpack/v5"
)

type MsgPackEncoder struct{

}

func (e MsgPackEncoder) Encode(r *Recorder) ([]byte, error){
    return msgpack.Marshal(r)
}

func (e MsgPackEncoder) Decode(data []byte) (*Recorder, error){
    var r Recorder
    err := msgpack.Unmarshal(data, &r)
    return &r, err
}
