package uploadedfile

import (
    "os"
)

type UplodedFile struct {
    filename string
    path     string
    mime     string
}

func (this *UplodedFile) GetFilename() string {
    return this.filename
}

func (this *UplodedFile) GetPath() string {
    return this.path
}

func (this *UplodedFile) GetMime() string {
    return this.mime
}

func (this *UplodedFile) FileSize() (int, error) {
    f, err := os.Open(this.Filename)
    if err != nil {
        return 0, err
    }

    stats, err := f.Stat()
    if err != nil {
        return 0, err
    }

    size := stats.Size()

    return size, nil
}