package jewl

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"runtime"
	"strconv"
	"time"

)

type location = string
type gid = int

// A frame recorder which keeps track of where in the program it is,
//
//	and allows for creating new frames, which are saved via the config
//
// The stack is the current call stack which is used for adding frames as either
//
//	function call frames or subframes within the same function
//
// # The Header is simply a reference table of locations and indices of new function Frames
//
// The Frames slice is simply all the frames, regardless of locations.
//
//	A new function frame is referred to in the Header
//	A subframe is referred to in a frame's Subframes slice
//
// See also: Frame
type Recorder struct {
	config RecorderConfig
	cache  RecorderCache
	stack  map[gid][]FrameIndex
	Header map[location][]FrameIndex
	Frames []*Frame
}

const JewlDir string = ".jewl"

func NewRecorder(config RecorderConfig) (*Recorder, error) {
	rec := Recorder{
		cache: RecorderCache{
			path: path.Join(JewlDir, "cache"),
		},
		config: config,
        stack: map[gid][]FrameIndex{},
		Header: map[string][]FrameIndex{},
		Frames: []*Frame{},
	}
	err := rec.cache.Clear()
	if err != nil {
		return nil, err
	}
    err = config.Clear()
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// Creates a new recorder using the given config, and load the current stack cache,
//
//	and frames saved from other instances
func GetRecorder(config RecorderConfig) (*Recorder, error) {
	rec := Recorder{
		cache: RecorderCache{
			path: ".jewl/cache",
		},
		config: config,
	}
    _, gid, err := getRTFrame()
    if err != nil{
        return nil, err
    }
	err = rec.loadState(gid)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *Recorder) Close() error {
	return r.cache.Clear()
}

/**
	Get the location (the canonical name) of the current Function

	This will remove the functions leading up to this function so that the actual
	function being recorded is used
*/
func getRTFrame() (string, int, error) {
	pc := make([]uintptr, 10)
    buf := make([]byte, 32)
    nn := runtime.Stack(buf, false)
    buf = buf[:nn]
    buf, ok := bytes.CutPrefix(buf, []byte("goroutine "))
    if !ok{
        return "", 0, errors.New("invalid runtime.Stack output")
    }
    i := bytes.IndexByte(buf, ' ')
    if i < 0{
        return "", 0, errors.New("invalid runtime.Stack output")
    }
	n := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
    gid, err := strconv.Atoi(string(buf[:i]))
    return frame.Function, gid, err
}

type recorderError struct {
	message string
	loc     string
	name    string
    cause   error
}

// Create a general Recorder Error with a message, location, and frame name
func RecorderError(message string, loc string, name string, err error) recorderError {
	return recorderError{
		message: message,
		loc:     loc,
		name:    name,
        cause:   err,
	}
}

func (e recorderError) Error() string {
    return fmt.Sprintf("%s @ %s for frame %s\n    Caused by: %s", e.message, e.loc, e.name, e.cause.Error())
}

/*
*

	Add a data point to the top (current) frame
*/
func (r *Recorder) AddData(name string, val any) error {
	loc, gid, err := getRTFrame()
    if err != nil{
        return err
    }
	if len(r.stack) == 0 {
		return RecorderError("Recorder has no frames on the stack", loc, loc, nil)
	}
    frame := r.getTopOfStack(gid)
	frame.Args[name] = val

	r.saveState()

	return nil
}

func (r *Recorder) getFromStack(goid gid) (*Frame, error){
	stack, ok := r.stack[goid]
    if !ok{
        r.stack[goid] = []FrameIndex{}
        return nil, nil
    }
    fidx := stack[len(stack)-1]
    frame := r.Frames[fidx]
    return frame, nil
}

func (r *Recorder) saveState() {
    _, gid, err := getRTFrame()
    if err != nil{
        panic(err)
    }
    top := r.getTopOfStack(gid)
    if top != nil{
        println("Saving with top frame:" + top.Name)
    }
	data, err := r.config.Encoder().Encode(r)
	if err != nil {
		panic(err)
	}
	err = r.config.Write(data)
	if err != nil {
		panic(err)
	}
	err = r.cache.Save(r.stack, r.config.Encoder())
	if err != nil {
		panic(err)
	}
}

func (r *Recorder) loadState(goid gid) error {
	data, err := r.config.Load()
	if err != nil {
        top := r.getTopOfStack(goid)
		return RecorderError("Attempted to load current frame data from save state but failed", top.Location, top.Name, err)
	}
	if len(data) == 0 {
		// Early return because there's nothing to append the next frames onto
		r.Frames = []*Frame{}
		r.Header = map[string][]FrameIndex{}
		return nil
	}
    rr, err := r.config.Encoder().Decode(data)
	if err != nil {
        if len(r.stack) > 0{
            top := r.getTopOfStack(goid)
            return RecorderError("Attempted to load current frame data from save state but failed", top.Location, top.Name, err)
        }
        return RecorderError("Failed to load current frame data while stack is empty", "loadState", "nil", err)
	}
	r.Frames = rr.Frames
	r.Header = rr.Header

	stack, err := r.cache.Load()
	if err != nil {
		return err
	}
	r.stack = stack
    println(fmt.Sprint(r.stack))
	return nil
}

func (r *Recorder) getTopOfStack(goid gid) *Frame{
    if len(r.stack) == 0{
        return nil
    }
    println("Recorder.getTopOfStack v:stack: " + fmt.Sprint(r.stack))
    println("Recorder.getTopOfStack v:Frames: " + fmt.Sprint(r.Frames))
    stack, ok := r.stack[goid]
    if !ok{
        r.stack[goid] = []FrameIndex{}
        stack = r.stack[goid]
    }
    if len(stack) == 0{
        return nil
    }
    fidx := stack[len(stack)-1]
    println("Recorder.getTopOfStack v:fidx: " + fmt.Sprint(fidx))
	frame := r.Frames[fidx]
    return frame
}

func (r *Recorder) addNewFrame(goid gid, frame *Frame){
    r.Frames = append(r.Frames, frame)
    r.Header[frame.Location] = append(r.Header[frame.Location], 0)
    r.pushToStack(0, goid)
}

func (r *Recorder) addFrame(newFrame *Frame, goid gid){
    //If there are no frames on the stack, just add it as a new frame and no subframes
    frame := r.getTopOfStack(goid)
    if frame == nil{
        r.addNewFrame(goid, newFrame)
        return
    }
    println("Recorder.addFrame v:topOfStack: " + fmt.Sprint(frame.Name))
    sidx := r.pushFrame(goid, newFrame)
    println("Recorder.addFrame e:frame.Location == newFrame.Location: " + fmt.Sprint(frame.Location == newFrame.Location))
    if frame.Location == newFrame.Location {
		// This is a subframe since the locations are the same
		frame.Subframes = append(frame.Subframes, sidx)
	} else {
		// The locations are different, so therefore, this is a called function
		frame.Calls = append(frame.Calls, sidx)
		// This is now a new function, therefore, must be in the header
		r.Header[newFrame.Location] = append(r.Header[newFrame.Location], sidx)
	}
}

func (r *Recorder) pushToStack(fidx FrameIndex, goid gid) {
    println("Recorder.pushToStack v:stack: " + fmt.Sprint(r.stack))
    stack, ok := r.stack[goid]
    if !ok{
        r.stack[goid] = []FrameIndex{}
        stack = r.stack[goid]
    }
    stack = append(stack, fidx)
    r.stack[goid] = stack
    println("Recorder.pushToStack v:stack: " + fmt.Sprint(r.stack))
}

func (r *Recorder) pushFrame(goid gid, frame *Frame) FrameIndex{
    r.Frames = append(r.Frames, frame)
    sidx := len(r.Frames) - 1
    r.pushToStack(sidx, goid)
    return sidx
}
/*
*

	Create a new frame, and do one of the following with it:
	a. Add it to the stack if the stack is empty
	b. If the location of the top frame is the same as this location, add as a subframe
	c. If the locations are different, add it as a call
*/
func (r *Recorder) Frame(name string) error {
	// Step 1. Get the current function
	start := time.Now().Nanosecond()
	loc, gid, err := getRTFrame()
    if err != nil{
        return err
    }
	sub := &Frame{
		Location:  loc,
		Name:      name,
		Start:     start,
		Args:      map[string]any{},
		Calls:     []FrameIndex{},
		Subframes: []FrameIndex{},
	}
    println("Recorder.Frame: Adding a new frame @ " + fmt.Sprint(loc))
	defer r.saveState()
    r.addFrame(sub, gid)

	return nil
}

/*
*

	Get the top of the stack, set its end time to time.Now(), set its duration,
	 and pop it off the stack
*/
func (r *Recorder) Stop() error {
	end := time.Now().Nanosecond()
    println("Stopping frame")
    _, gid, err := getRTFrame()
    if err != nil{
        return err
    }
	defer r.saveState()
	err = r.loadState(gid)
	if err != nil {
		return err
	}

    frame := r.getTopOfStack(gid)
    if frame == nil{
        return errors.New("Current stack is empty for gid " + fmt.Sprint(gid))
    }
	frame.End = end
	frame.Duration = end - frame.Start

    r.popStack(gid)
	return nil
}

func (r *Recorder) popStack(goid gid) {
    stack, ok := r.stack[goid]
    if !ok{
        return 
    }
    stack_len := len(stack)
    println("Recorder.popStack v:stack_len: " + fmt.Sprint(stack_len))
    stack = stack[:stack_len]
    r.stack[goid] = stack
    println("Recorder.popStack v:stack: " + fmt.Sprint(r.stack))

}
