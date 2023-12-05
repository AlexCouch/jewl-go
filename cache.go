package jewl

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"syscall"

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
	if _, err := os.Stat(r.path); err == nil {
		err = os.RemoveAll(r.path)
		if err != nil {
			return err
		}

	}
	return nil
}

// Save the stack to the given cache path
func (r *RecorderCache) Save(stack []int, encoder Encoder) error {
    //Check if the file exists, and if not, make it
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
    //Open the file
    file, err := os.OpenFile(r.path, os.O_WRONLY, 0666)
    if err != nil{
        return err
    }
    defer file.Close()

    //Lock the file
    if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil{
        return err
    }
    defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

    //Marshal the recorder
    data, err := msgpack.Marshal(stack)
	if err != nil {
		return err
	}

    //Buffer the data and flush it
    buf := bufio.NewWriter(file)
    _, err = buf.Write(data)
	if err != nil {
		return err
	}
    if err := buf.Flush(); err != nil{
        return err
    }
	return nil
}

// Load the cache from the given cache path
func (r *RecorderCache) Load() (stack []int, err error) {
    //Check if the cache file exists, and if not, make it
    fi, err := os.Stat(r.path)
	if err != nil {
		file, err := os.Create(r.path)
		if err != nil {
			return stack, err
		}
        fi, err = file.Stat()
	}
    file, err := os.OpenFile(r.path, os.O_RDONLY, 0666)
    if err != nil{
        return
    }
    defer file.Close()
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX) 
    defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

    buf := bufio.NewReader(file)
    data := make([]byte, fi.Size())
    if _, err = buf.Read(data); !errors.Is(err, io.EOF){
        return 
    }
    println("RecorderCache.Load v:data: " + fmt.Sprint(data))
	if len(data) == 0 {
		return stack, nil
	}
	if err = msgpack.Unmarshal(data, &stack); err != nil{
        return
    }
	return stack, nil
}
