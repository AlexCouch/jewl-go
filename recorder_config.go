package jewl

// A configuration interface which allows for creating different configuration
// schemas such as emitting frames to a server or a file
type RecorderConfig interface {
	// Load the current state of the recorder
	Load() ([]byte, error)
	// Write the current state of the recorder
	Write([]byte) error
	// Clears the current project state, if applicable
	Clear() error
    Encoder() Encoder
}
