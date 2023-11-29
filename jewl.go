package jewl

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/google/uuid"
)

type OldFrame struct{
    Uid         uuid.UUID       `json:"uuid"`
    Name        string          `json:"name"`
    Start       time.Time       `json:"start"`
    End         time.Time       `json:"end"`
    Duration    time.Duration   `json:"duration"`
    Args        map[string]any  `json:"args"`
    Subframes   []*OldFrame        `json:"subframes"`
}

type Profiler struct{
    topFrame   *OldFrame
    current    *OldFrame
}

var G_Profiler Profiler = Profiler{
    topFrame: nil,
    current:  nil,
}

func getFrameName() string{
    pc := make([]uintptr, 10)
    n := runtime.Callers(3, pc)
    frames := runtime.CallersFrames(pc[:n])
    frame, _ := frames.Next()
    return frame.Function
}

func newFrame(name string) *OldFrame{
    start_time := time.Now()
    return &OldFrame{
        Uid:    uuid.New(),
        Name:   name,
        Start:  start_time,
        Args: map[string]any{},
        Subframes: []*OldFrame{},
    }
}

func (self *Profiler) Current() *OldFrame{
    return self.current
}

func (self *Profiler) StartFrame(name string) *OldFrame{
    //name := getFrameName()
    frame := newFrame(name)    
    if self.topFrame == nil{
        self.topFrame = frame
    }else{
        self.topFrame.Subframes = append(self.topFrame.Subframes, frame)
    }
    self.current = frame
    return frame
}

func (frame *OldFrame) Subframe(name string) *OldFrame{
    subframe := newFrame(name)
    G_Profiler.current = subframe
    frame.Subframes = append(frame.Subframes, subframe)
    return subframe
}

func (frame *OldFrame) AddArg(name string, arg any){
    frame.Args[name] = arg
}

func (frame *OldFrame) Stop(){
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
