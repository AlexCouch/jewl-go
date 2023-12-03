package jewl

import "os"

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
	if _, err := os.Stat(f.path); err != nil {
		_, err := os.Create(f.path)
		if err != nil {
			return err
		}
	}
	err := os.WriteFile(f.path, data, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (f RecorderConfigFile) Load() ([]byte, error) {
	if _, err := os.Stat(f.path); err != nil {
		_, err := os.Create(f.path)
		if err != nil {
			return []byte{}, err
		}
	}
	data, err := os.ReadFile(f.path)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func FileConfigWithEncoder(path string, encoder Encoder) RecorderConfigFile{
	return RecorderConfigFile{
		path: path,
        encoder: encoder,
	}
}

func DefaultEncoder() Encoder{
    return MsgPackEncoder{}
}

func FileConfig(path string) RecorderConfigFile {
    encoder := DefaultEncoder()
	return RecorderConfigFile{
		path: path,
        encoder: encoder,
	}
}

func (f RecorderConfigFile) Encoder() Encoder{
    return f.encoder
}
