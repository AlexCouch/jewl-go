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
    stack       []FrameIndex
    Header      map[location]FrameIndex `json:"header"`
    Frames      []*Frame                `json:"frames"`
}

func GetRecorder(config RecorderConfig) (*Recorder, error){
    rec := Recorder{
        config: config,
        Header: map[location]FrameIndex{},
        Frames: []*Frame{},
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

/**
    Get the location (the canonical name) of the current Function

    This will remove the functions leading up to this function so that the actual
    function being recorded is used
*/
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

/**
    Add a data point to the top (current) frame
*/
func (r *Recorder) AddData(name string, val any) error{
    loc := getRTFrame()
    if len(r.stack) == 0{
        return RecorderError("Recorder has no frames on the stack", loc, loc)
    }
    fidx := r.stack[len(r.stack)-1]
    frame := r.Frames[fidx]
    frame.Args[name] = val

    return nil
}

/**
    Create a new frame, and do one of the following with it:
    a. Add it to the stack if the stack is empty
    b. If the location of the top frame is the same as this location, add as a subframe
    c. If the locations are different, add it as a call
*/
func (r *Recorder) Frame(name string) error{
    //Step 1. Get the current function
    start := time.Now().Nanosecond()
    loc := getRTFrame()
    sub := &Frame{
        Location: loc,
        Name: name,
        Start: start,
        Args: map[string]any{},
        Calls: []FrameIndex{},
        Subframes: []FrameIndex{},
    }
    defer func(){
        data, err := json.Marshal(r)
        if err != nil{
            panic(err)
        }
        err = r.config.Write(data)
        if err != nil{
            panic(err)
        }
    }()
    //If there are no frames on the stack (like at the start of the main function),
    // then add a new frame to the Frames, Header, and stack
    if len(r.stack) == 0{
        println(loc + ": Stack is empty, adding the first frame")
        r.Frames = append(r.Frames, sub)
        sidx := len(r.Frames)-1
        r.Header[loc] = sidx
        r.stack = append(r.stack, sidx)
        return nil
    }
    //Otherwise, get the top of the stack, add the frame as a subframe (or call)
    //  and add this frame to the current stack
    fidx := r.stack[len(r.stack)-1]
    frame := r.Frames[fidx]
    r.Frames = append(r.Frames, sub)
    if frame.Location == loc{
        println(loc + ": New frame is not a new function: adding as a subframe")
        //This is a subframe since the locations are the same
        frame.Subframes = append(frame.Subframes, fidx)
    }else{
        println(loc + ": New frame is a new function: adding as a call")
        //The locations are different, so therefore, this is a called function
        frame.Calls = append(frame.Calls, fidx+1)
        //This is now a new function, therefore, must be in the header
        r.Header[loc] = fidx
    }

    return nil
}

/**
    Get the top of the stack, set its end time to time.Now(), set its duration, 
     and pop it off the stack
*/
func (r *Recorder) Stop() error{
    end := time.Now().Nanosecond()
    fidx := r.stack[len(r.stack)-1]
    frame := r.Frames[fidx]
    frame.End = end
    frame.Duration = end - frame.Start

    r.stack = r.stack[:len(r.stack) - 1]
    return nil
}
