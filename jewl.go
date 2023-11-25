package jewl

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/google/uuid"
)

type Frame struct{
    Uid         uuid.UUID       `json:"uuid"`
    Name        string          `json:"name"`
    Start       time.Time       `json:"start"`
    End         time.Time       `json:"end"`
    Duration    time.Duration   `json:"duration"`
    Args        map[string]any  `json:"args"`
    Subframes   []*Frame        `json:"subframes"`
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
        Uid:    uuid.New(),
        Name:   frame_name,
        Start:  start_time,
        Args: map[string]any{},
        Subframes: []*Frame{},
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
    frame.Subframes = append(frame.Subframes, subframe)
    return subframe
}

func (frame *Frame) AddArg(name string, arg any){
    frame.Args[name] = arg
}

func (frame *Frame) Stop(){
    frame.End = time.Now()
    frame.Duration = frame.End.Sub(frame.Start)
}

func (self *Profiler) Dump(outPath string) error {
    //TODO: Need to find a good compact serialization format
    f, err := os.OpenFile(outPath, os.O_CREATE|os.O_RDWR, 0666)
    defer f.Close()
    if err != nil{
        return err
    }
    fmt.Println(f)
    data, err := json.Marshal(self.topFrame)
    if err != nil{
        return err
    }
    count, err := f.Write(data)
    if err != nil{
        return err
    }
    if count < len(data){
        for count != 0{
            count, err = f.Write(data)
            if err != nil{
                return err
            }
        }
    }
    return nil
}
