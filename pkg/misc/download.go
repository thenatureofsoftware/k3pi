package misc

import (
    "bufio"
    "crypto/sha256"
    "fmt"
    "github.com/dustin/go-humanize"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strings"
)

type FileDownload struct {
    Filename, CheckSumFilename, Url, CheckSumUrl string
}

// WriteCounter counts the number of bytes written to it. It implements to the io.Writer
// interface and we can pass this into io.TeeReader() which will report progress on each
// write cycle.
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
    // Clear the line by using a character return to go back to the start and remove
    // the remaining characters by filling it with spaces
    fmt.Printf("\r%s", strings.Repeat(" ", 35))

    // Return again and print current status of download
    // We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
    fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory. We pass an io.TeeReader
// into Copy() to report progress on the download.
func DownloadFile(filepath string, url string) error {

    // Create the file, but give it a tmp file extension, this means we won't overwrite a
    // file until it's downloaded, but we'll remove the tmp extension once downloaded.
    out, err := os.Create(filepath + ".tmp")
    if err != nil {
        return err
    }
    defer out.Close()
    defer os.RemoveAll(filepath + ".tmp")

    // Get the data
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    if resp.StatusCode != 200 {
        return fmt.Errorf("%s - %s", url, resp.Status)
    }
    defer resp.Body.Close()

    // Create our progress reporter and pass it to be used alongside our writer
    counter := &WriteCounter{}
    _, err = io.Copy(out, io.TeeReader(resp.Body, counter))
    if err != nil {
        return err
    }

    // The progress use the same line so print a new line once it's finished downloading
    fmt.Print("\n")

    err = os.Rename(filepath+".tmp", filepath)
    if err != nil {
        return err
    }

    return nil
}

func DownloadAndVerify(download FileDownload) error {

    err := DownloadFile(download.Filename, download.Url)
    if err != nil {
        return err
    }

    err = DownloadFile(download.CheckSumFilename, download.CheckSumUrl)
    if err != nil {
        return err
    }

    checksum, err := ioutil.ReadFile(download.CheckSumFilename)
    if err != nil {
        return err
    }

    validCheckSum := strings.Split(string(checksum), " ")[0]

    if calcSHA256, _ := CalculateSHA256(download.Filename); validCheckSum != calcSHA256  {
        return fmt.Errorf("calculated sha256 checksum don't match: %s\n", calcSHA256)
    }

    return nil
}

func CalculateSHA256(filename string) (string, error) {
    f, err := os.Open(filename)
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