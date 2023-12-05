package jewl

import (
	"bufio"
	"os"
	"path"
	"path/filepath"
	"syscall"
)

// A file config schema which simply saves the frames to a local file on the machine
//
// See also: RecorderConfig
type RecorderConfigFile struct {
	path string
    encoder Encoder
}

func (f RecorderConfigFile) Clear() error {
    if _, err := os.Stat(f.path); err != nil{
        return nil
    }
	return os.Remove(f.path)
}

func (f RecorderConfigFile) Write(data []byte) error {
    //Check if file exists, if not, make it
	if _, err := os.Stat(f.path); err != nil {
		parent, _ := path.Split(f.path)
		err = os.MkdirAll(parent, 0700)
		if err != nil {
			return err
		}
		_, err := os.Create(f.path)
		if err != nil {
			return err
		}
	}
    //Open the file
	file, err := os.OpenFile(f.path, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
    defer file.Close()

    //Lock the file
	if err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil{
        return err
    }
    defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

    //Write to the buffer and flush
    buf := bufio.NewWriter(file)
	if _, err = buf.Write(data); err != nil{
		return err
	}
    if err = buf.Flush(); err != nil{
        return err
    }
	return nil
}

func (f RecorderConfigFile) Load() ([]byte, error) {
    //Check if project file exists, if not, make it
	fi, err := os.Stat(f.path)
	if err != nil {
		parent, _ := path.Split(f.path)
		err = os.MkdirAll(parent, 0700)
		if err != nil {
            return []byte{}, err
		}
		file, err := os.Create(f.path)
		if err != nil {
            return []byte{}, err
		}
        fi, err = file.Stat()
	}
    //Open file
	file, err := os.Open(f.path)
	if err != nil {
		return []byte{}, err
	}

    defer file.Close()
    //Lock the file
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX) 
    defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

    //Read the project file
    buf := bufio.NewReader(file)
    data := make([]byte, fi.Size())
    _, err = buf.Read(data)
    if err != nil{
        return []byte{}, err 
    }
	return data, nil
}

func FileConfigWithEncoder(name string, encoder Encoder) RecorderConfigFile{
	return RecorderConfigFile{
		path: filepath.Join(JewlDir, name),
        encoder: encoder,
	}
}

func DefaultEncoder() Encoder{
    return MsgPackEncoder{}
}

func FileConfig(name string) RecorderConfigFile {
    encoder := DefaultEncoder()
	return RecorderConfigFile{
		path: filepath.Join(JewlDir, name),
        encoder: encoder,
	}
}

func (f RecorderConfigFile) Encoder() Encoder{
    return f.encoder
}
