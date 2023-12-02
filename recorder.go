package jewl

import (
	//"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

//A configuration interface which allows for creating different configuration
// schemas such as emitting frames to a server or a file
type RecorderConfig interface{
    //Load the current state of the recorder
    Load() ([]byte, error)
    //Write the current state of the recorder
    Write([]byte) error
    //Clears the current project state, if applicable
    Clear() error
}

//A stack cache to save the state of the stack so the recorder can keep track of
// where in the code the program is.
//
// This makes it easier for different functions to add new frames to the hierarchy
type RecorderCache struct{
    path string
}

func (r *RecorderCache) Clear() error{
    if _, err := os.Stat(JewlDir); err == nil{
        err = os.RemoveAll(JewlDir)
        if err != nil{
            return err
        }

    }
    return nil
}

//Save the stack to the given cache path
func (r *RecorderCache) Save(stack []int) error{
    if _, err := os.Stat(r.path); err != nil{
        parent, _ := path.Split(r.path)
        err = os.MkdirAll(parent, 0700)
        if err != nil{
            return err
        }
        _, err = os.Create(r.path)
        if err != nil{
            return err
        }
    }
    data, err := msgpack.Marshal(stack)
    if err != nil{
        return err
    }
    //data, err := json.Marshal(stack)
    //if err != nil{
    //    return err
    //}
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
    err = msgpack.Unmarshal(data, &stack)
    //err = json.Unmarshal(data, &stack)
    //if err != nil{
    //    return stack, err
    //}
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
    Header      map[location]FrameIndex `msgpack:"header"`
    Frames      []*Frame                `msgpack:"frames"`
}

const JewlDir string = ".jewl"

func NewRecorder(config RecorderConfig) (*Recorder, error){
    rec := Recorder{
        cache: RecorderCache{
            path: path.Join(JewlDir, "cache"),
        },
        config: config,
        Header: map[string]int{},
        Frames: []*Frame{},
    }
    err := rec.cache.Clear()
    if err != nil{
        return nil, err
    }
    err = config.Clear()
    if err != nil{
        return nil, err
    }
    return &rec, nil
}

//Creates a new recorder using the given config, and load the current stack cache,
//  and frames saved from other instances
func GetRecorder(config RecorderConfig) (*Recorder, error){
    rec := Recorder{
        cache: RecorderCache{
            path: ".jewl/cache",
        },
        config: config,
    }
    err := rec.loadState()
    if err != nil{
        return nil, err
    }
    return &rec, nil
}

func (r *Recorder) Close() error{
    return r.cache.Clear()
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

    r.saveState()

    return nil
}

func (r *Recorder) saveState() {
    data, err := msgpack.Marshal(r)
    //data, err := json.MarshalIndent(r, "", "    ")
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
}

func (r *Recorder) loadState() error {
    data, err := r.config.Load()
    if err != nil{
        top := r.stack[len(r.stack)-1]
        frame := r.Frames[top]
        return RecorderError("Attempted to load current frame data from save state but failed", frame.Location, frame.Name)
    }
    if len(data) == 0{
        //Early return because there's nothing to append the next frames onto
        r.Frames = []*Frame{}
        r.Header = map[string]int{}
        return nil
    }
    var rr Recorder
    //err = json.Unmarshal(data, &rr)
    err = msgpack.Unmarshal(data, &rr)
    if err != nil{
        top := r.stack[len(r.stack)-1]
        frame := r.Frames[top]
        return RecorderError("Attempted to load current frame data from save state but failed", frame.Location, frame.Name)
    }
    r.Frames = rr.Frames
    r.Header = rr.Header

    stack, err := r.cache.Load()
    if err != nil{
        return err
    }
    r.stack = stack
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
    defer r.saveState()
    //If there are no frames on the stack (like at the start of the main function),
    // then add a new frame to the Frames, Header, and stack
    if len(r.stack) == 0{
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
    sidx := len(r.Frames)-1
    r.stack = append(r.stack, sidx)
    if frame.Location == loc{
        //This is a subframe since the locations are the same
        frame.Subframes = append(frame.Subframes, sidx)
    }else{
        //The locations are different, so therefore, this is a called function
        frame.Calls = append(frame.Calls, sidx)
        //This is now a new function, therefore, must be in the header
        r.Header[loc] = sidx
    }

    return nil
}

/**
    Get the top of the stack, set its end time to time.Now(), set its duration, 
     and pop it off the stack
*/
func (r *Recorder) Stop() error{
    end := time.Now().Nanosecond()
    defer r.saveState()
    err := r.loadState()
    if err != nil{
        return err
    }

    fidx := r.stack[len(r.stack)-1]
    frame := r.Frames[fidx]
    frame.End = end
    frame.Duration = end - frame.Start

    r.stack = r.stack[:len(r.stack) - 1]
    return nil
}
