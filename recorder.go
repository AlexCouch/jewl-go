package jewl

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

//A configuration interface which allows for creating different configuration 
// schemas such as emitting frames to a server or a file
type RecorderConfig interface{
    //Load the current state of the recorder
    Load() ([]byte, error)
    //Write the current state of the recorder
    Write([]byte) error
}

//A stack cache to save the state of the stack so the recorder can keep track of
// where in the code the program is.
//
// This makes it easier for different functions to add new frames to the hierarchy
type RecorderCache struct{
    path string
}

//Save the stack to the given cache path
func (r *RecorderCache) Save(stack []int) error{
    if _, err := os.Stat(r.path); err != nil{
        _, err := os.Create(r.path)
        if err != nil{
            return err
        }
    }
    data, err := json.Marshal(stack)
    if err != nil{
        return err
    }
    err = os.WriteFile(r.path, data, 0666)
    if err != nil{
        return err
    }
    return nil
}

//Load the cache from the given cache path
func (r *RecorderCache) Load() ([]int, error){
    var stack []int
    if _, err := os.Stat(r.path); err != nil{
        _, err := os.Create(r.path)
        if err != nil{
            return stack, err
        }
    }
    data, err := os.ReadFile(r.path)
    if err != nil{
        return stack, err
    }
    if len(data) == 0{
        return stack, nil
    }
    err = json.Unmarshal(data, &stack)
    if err != nil{
        return stack, err
    }
    return stack, nil
}

type location = string

//A frame recorder which keeps track of where in the program it is,
//  and allows for creating new frames, which are saved via the config
//
//The stack is the current call stack which is used for adding frames as either
//  function call frames or subframes within the same function
//
//The Header is simply a reference table of locations and indices of new function Frames
//
//The Frames slice is simply all the frames, regardless of locations. 
//  A new function frame is referred to in the Header
//  A subframe is referred to in a frame's Subframes slice
//
//See also: Frame
type Recorder struct{
    config      RecorderConfig
    cache       RecorderCache
    stack       []FrameIndex
    Header      map[location]FrameIndex `json:"header"`
    Frames      []*Frame                `json:"frames"`
}

//Creates a new recorder using the given config, and load the current stack cache,
//  and frames saved from other instances
func GetRecorder(config RecorderConfig) (*Recorder, error){
    rec := Recorder{
        cache: RecorderCache{
            path: "cache.json",
        },
        config: config,
    }
    //Load the stack
    stack, err := rec.cache.Load()
    if err != nil{
        return nil, err
    }
    rec.stack = stack

    //Load the current data via the config
    data, err := rec.config.Load()
    if err != nil{
        
        return nil, err
    }
    if len(data) == 0{
        //Early return because there's nothing to append the next frames onto
        //We are starting fresh
        rec.Frames = []*Frame{}
        rec.Header = map[string]FrameIndex{}
        return &rec, nil
    }
    err = json.Unmarshal(data, &rec)
    if err != nil{
        return nil, err
    }
    println(rec.Frames)
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

//Create a general Recorder Error with a message, location, and frame name
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
        err = r.cache.Save(r.stack)
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
    r.stack = append(r.stack, fidx+1)
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
    err := r.cache.Save(r.stack)
    if err != nil{
        return err
    }
    return nil
}
