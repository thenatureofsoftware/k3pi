/*
Copyright © 2019 The Nature of Software Nordic AB <lars@thenatureofsoftware.se>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package misc

import (
	"crypto/rand"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

func DataPipedIn() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) == 0
}

func PanicOnError(err error, message string) {
	if err != nil {
		panic(errors.Wrap(err, message))
	}
}

func ExitOnError(err error, message ...string) {
	if err != nil {
		if len(message) > 0 {
			ErrorExitWithError(errors.Wrap(err, message[0]))
		} else {
			ErrorExitWithError(err)
		}
	}
}

func Info(message string) {
	fmt.Printf("%s\n", message)
}

func ErrorExitWithMessage(message string) {
	fmt.Printf("Error: %s\n", message)
	os.Exit(1)
}

func ErrorExitWithError(err error) {
	fmt.Printf("Error: %s\n", err)
	os.Exit(1)
}

func CreateTempFileName(dir string, pattern string) string {
	dirPath, err := filepath.Abs(dir)
	PanicOnError(err, "failed to resolve abs path")

	f, err := ioutil.TempFile(dirPath, pattern)
	PanicOnError(err, "failed to create temp file")

	fn := f.Name()
	err = f.Close()
	PanicOnError(err, "failed to close temp file")

	err = os.Remove(fn)
	PanicOnError(err, "failed to remove temp file")

	return fn
}

func GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
