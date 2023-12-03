package jewl 

type Encoder interface{
    Encode(r *Recorder) ([]byte, error)
    Decode(data []byte) (*Recorder, error)
}
