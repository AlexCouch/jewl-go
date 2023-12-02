package jewl

//This is simply a type alias to make reading the code easier
type FrameIndex = int

//A frame contains all the information needed to analyze the current place in the program
//
//A subframe is a frame which shares the same `location` but has a different `name`
//A call is a frame which has a different `location` while this frame instance 
// is on the top of the stack
//
//When a frame is on the top of the stack, and a new function creates a frame,
// that funtion's frame is added here as a `Call`
type Frame struct{
    Location    string          `msgpack:"location"`
    Name        string          `msgpack:"name"`
    Args        map[string]any  `msgpack:"args"`
    Start       int             `msgpack:"start"`
    End         int             `msgpack:"end"`
    Duration    int             `msgpack:"duration"`
    Calls       []FrameIndex    `msgpack:"calls"`
    Subframes   []FrameIndex    `msgpack:"subframes"`
}
