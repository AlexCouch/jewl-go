package jewl

type FrameIndex = int
type SubframeIndex = int

type Frame interface{
    GetName() string
    AddArg(name string, val any)
    GetStart() int
    GetEnd() int
    SetEnd(end int)
    SetDuration(dur int)
    GetDuration() int
}

type Subframe struct{
    name        string
    args        map[string]any
    start       int
    end         int
    duration    int
    calls       []FrameIndex
    subframes   []SubframeIndex
}

func (f *Subframe) GetName() string{
    return f.name
}

func (f *Subframe) AddArg(name string, val any){
    f.args[name] = val
}

func (f *Subframe) GetStart() int{
    return f.start
}

func (f *Subframe) GetEnd() int{
    return f.end
}
func (f *Subframe) SetEnd(end int) {
    f.end = end
}

func (f *Subframe) GetDuration() int{
    return f.duration
}

func (f *Subframe) SetDuration(dur int){
    f.duration = dur
}

type Function struct{
    location    string
    Subframe
}

func (f *Function) GetName() string{
    return f.name
}

func (f *Function) AddArg(name string, val any){
    f.args[name] = val
}

func (f *Function) GetStart() int{
    return f.start
}

func (f *Function) GetEnd() int{
    return f.end
}
func (f *Function) SetEnd(end int) {
    f.end = end
}

func (f *Function) GetDuration() int{
    return f.duration
}

func (f *Function) SetDuration(dur int){
    f.duration = dur
}
