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
	"bufio"
	"crypto/sha256"
	"fmt"
	"github.com/TheNatureOfSoftware/k3pi/pkg/model"
	"github.com/dustin/go-humanize"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc WriteCounter) PrintProgress() {
	fmt.Printf("\r%s", strings.Repeat(" ", 35))
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

func DownloadFile(resourceDir string, filename string, url string) error {

	absPath := resourceDir + string(os.PathSeparator) + filename
	out, err := os.Create(absPath + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()
	defer os.RemoveAll(absPath + ".tmp")

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s - %s", url, resp.Status)
	}
	defer resp.Body.Close()

	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	fmt.Print("\n")

	err = os.Rename(absPath+".tmp", absPath)
	if err != nil {
		return err
	}

	return nil
}

func DownloadAndVerify(resourceDir string, download *model.RemoteAsset) error {

	err := DownloadFile(resourceDir, download.Filename, download.FileUrl)
	if err != nil {
		return err
	}

	err = DownloadFile(resourceDir, download.CheckSumFilename, download.CheckSumUrl)
	if err != nil {
		return err
	}

	checksum, err := ioutil.ReadFile(resourceDir + string(os.PathSeparator) + download.CheckSumFilename)
	if err != nil {
		return err
	}
	allValidCheckSums := string(checksum)

	calcSHA256, err := CalculateSHA256(resourceDir, download.Filename)
	if err != nil {
		return fmt.Errorf("failed to calculate check sum: %v", err)
	}

	if !strings.Contains(allValidCheckSums, calcSHA256) {
		return fmt.Errorf("%s check sum is not valid for %s", calcSHA256, download.Filename)
	}

	return nil
}

func CalculateSHA256(resourceDir string, filename string) (string, error) {
	f, err := os.Open(resourceDir + string(os.PathSeparator) + filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	input := bufio.NewReader(f)

	hash := sha256.New()
	if _, err := io.Copy(hash, input); err != nil {
		log.Fatal(err)
	}
	sum := hash.Sum(nil)

	return fmt.Sprintf("%x", sum), nil
}
