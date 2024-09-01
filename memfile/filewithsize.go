package memfile

import "os"

type FileWithSize struct {
	file *os.File
}

func NewFileWithSize(file *os.File) *FileWithSize {
	return &FileWithSize{file: file}
}

func (fws *FileWithSize) Size() int64 {
	stat, _ := fws.file.Stat()
	return stat.Size()
}

func (fws *FileWithSize) Close() error {
	return fws.file.Close()
}

func (fws *FileWithSize) ReadAt(b []byte, off int64) (int, error) {
	return fws.file.ReadAt(b, off)
}

func (fws *FileWithSize) Sync() error {
	return fws.file.Sync()
}

func (fws *FileWithSize) WriteAt(p []byte, off int64) (int, error) {
	return fws.file.WriteAt(p, off)
}
