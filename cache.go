package jewl

import (
	"os"
	"path"

	"github.com/vmihailenco/msgpack/v5"
)

// A stack cache to save the state of the stack so the recorder can keep track of
// where in the code the program is.
//
// This makes it easier for different functions to add new frames to the hierarchy
type RecorderCache struct {
	path string
}

func (r *RecorderCache) Clear() error {
	if _, err := os.Stat(JewlDir); err == nil {
		err = os.RemoveAll(JewlDir)
		if err != nil {
			return err
		}

	}
	return nil
}

// Save the stack to the given cache path
func (r *RecorderCache) Save(stack []int, encoder Encoder) error {
	if _, err := os.Stat(r.path); err != nil {
		parent, _ := path.Split(r.path)
		err = os.MkdirAll(parent, 0700)
		if err != nil {
			return err
		}
		_, err = os.Create(r.path)
		if err != nil {
			return err
		}
	}
    data, err := msgpack.Marshal(stack)
	if err != nil {
		return err
	}
	err = os.WriteFile(r.path, data, 0666)
	if err != nil {
		return err
	}
	return nil
}

// Load the cache from the given cache path
func (r *RecorderCache) Load() ([]int, error) {
	var stack []int
	if _, err := os.Stat(r.path); err != nil {
		_, err := os.Create(r.path)
		if err != nil {
			return stack, err
		}
	}
	data, err := os.ReadFile(r.path)
	if err != nil {
		return stack, err
	}
	if len(data) == 0 {
		return stack, nil
	}
	err = msgpack.Unmarshal(data, &stack)
	//err = json.Unmarshal(data, &stack)
	//if err != nil{
	//    return stack, err
	//}
	return stack, nil
}
