package jewl

import (
	"runtime"
	"time"

	"github.com/google/uuid"
)

type Frame struct{
    uid         uuid.UUID
    name        string
    start       time.Time
    end         time.Time
    duration    time.Duration
    args        map[string]any
    subframes   []*Frame
}

type Profiler struct{
    topFrame   *Frame
    current    *Frame
}

var G_Profiler Profiler = Profiler{
    topFrame: nil,
    current:  nil,
}

func getFrameName() string{
    pc := make([]uintptr, 10)
    runtime.Callers(2, pc)
    f := runtime.FuncForPC(pc[0])
    return f.Name()
}

func newFrame() *Frame{
    start_time := time.Now()
    frame_name := getFrameName()
    return &Frame{
        uid:    uuid.New(),
        name:   frame_name,
        start:  start_time,
        args: map[string]any{},
        subframes: []*Frame{},
    }
}

func (self *Profiler) Current() *Frame{
    return self.current
}

func (self *Profiler) StartFrame() *Frame{
    frame := newFrame()    
    if self.topFrame == nil{
        self.topFrame = frame
    }
    self.current = frame
    return frame
}

func (frame *Frame) Subframe() *Frame{
    subframe := newFrame()
    G_Profiler.current = subframe
    frame.subframes = append(frame.subframes, subframe)
    return subframe
}

func (frame *Frame) AddArg(name string, arg any){
    frame.args[name] = arg
}

func (frame *Frame) Stop(){
    frame.end = time.Now()
    frame.duration = frame.end.Sub(frame.start)
}

func (self *Profiler) Dump(outPath string) {
    //TODO: Need to find a good compact serialization format
}
