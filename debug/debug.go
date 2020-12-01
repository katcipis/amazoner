package debug

import (
	"fmt"
	"io"
	"os"
)

func Save(path string, data io.Reader) (io.Reader, io.Closer, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	r := io.TeeReader(data, f)
	fmt.Printf("debug.Save:created file %q\n", path)
	return r, f, nil
}

func SaveString(path string, data string) (io.Closer, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if _, err := f.Write([]byte(data)); err != nil {
		f.Close() // ignore error; Write error takes precedence
		return nil, err
	}
	fmt.Printf("debug.Save:created file %q\n", path)
	return f, nil
}
