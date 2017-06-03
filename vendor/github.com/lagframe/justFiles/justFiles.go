package justFiles

import (
	"net/http"
	"os"
)

type JustFilesFilesystem struct {
    Fs http.FileSystem
}

func (fs JustFilesFilesystem) Open(name string) (http.File, error) {
    f, err := fs.Fs.Open(name)
    if err != nil {
        return nil, err
    }
    return neuteredReaddirFile{f}, nil
}

type neuteredReaddirFile struct {
    http.File
}

func (f neuteredReaddirFile) Readdir(count int) ([]os.FileInfo, error) {
    return nil, nil
}
