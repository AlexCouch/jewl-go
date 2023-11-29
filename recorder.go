package jewl

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

type RecorderConfig interface{
    Load() ([]byte, error)
    Write([]byte) error
}

type location = string

type Recorder struct{
    config      RecorderConfig
    stack       map[location]*Function
    Header      map[location]FrameIndex `json:"header"`
    Frames      []*Function             `json:"frames"`
    Subframes   []*Subframe             `json:"subframes"`
}

func GetRecorder(config RecorderConfig) (*Recorder, error){
    rec := Recorder{
        config: config,
        Header: map[location]FrameIndex{},
        Frames: []*Function{},
        Subframes: []*Subframe{},
    }
    data, err := rec.config.Load()
    if err != nil{
        return nil, err
    }
    var frame Frame
    if len(data) == 0{
        //Early return because there's nothing to append the next frames onto
        //We are starting fresh
        return &rec, nil
    }
    err = json.Unmarshal(data, &frame)
    if err != nil{
        return nil, err
    }
    return &rec, nil
}

func getRTFrame() string{
    pc := make([]uintptr, 10)
    n := runtime.Callers(3, pc)
    frames := runtime.CallersFrames(pc[:n])
    frame, _ := frames.Next()
    return frame.Function
}

type recorderError struct{
    message string 
    loc     string
    name    string
}

func RecorderError(message string, loc string, name string) recorderError{
    return recorderError{
        message: message,
        loc: loc,
        name: name,
    }
}

func (e recorderError) Error() string{
    return fmt.Sprintf("%s @ %s for frame %s", e.message, e.loc, e.name)
}

func (r *Recorder) Frame(name string) (*Subframe, error){
    loc := getRTFrame()
    frame, ok := r.stack[loc]
    if !ok{
        return nil, RecorderError(
            "Frame is currently not on the stack",
            loc, name,
        )
    }
    sub := &Subframe{
        name: name,
        start: time.Now().Nanosecond(),
        args: map[string]any{},
        calls: []FrameIndex{},
        subframes: []SubframeIndex{},
    }
    sidx := len(r.Subframes)
    r.Subframes = append(r.Subframes, sub)
    frame.subframes = append(frame.subframes, sidx)

    return sub, nil
}

func (r *Recorder) Function() (*Function, error){
    loc := getRTFrame()
    _, ok := r.stack[loc]
    if ok{
        return nil, RecorderError("Function is already on the stack", loc, "")
    }
    frame := &Function{
        location: loc,
        Subframe: Subframe{
            name: loc,
            start: time.Now().Nanosecond(),
            args: map[string]any{},
            calls: []FrameIndex{},
            subframes: []SubframeIndex{},
        },
    }
    r.stack[loc] = frame
    fidx := len(r.Frames)
    r.Frames = append(r.Frames, frame)
    r.Header[loc] = fidx
    return frame, nil
}

func (r *Recorder) Stop(frame Frame) error{
    end := time.Now().Nanosecond()
    frame.SetEnd(end)
    frame.SetDuration(end - frame.GetStart())

    switch v := frame.(type){
    case Function:
        delete(r.stack, v.name)
    case Subframe:
        loc := getRTFrame()
        frame, ok := r.stack[loc]
        if !ok{
            return RecorderError("Function not on the stack", loc, frame.name)
        }
    }
    return nil
}
