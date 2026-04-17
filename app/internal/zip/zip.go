package zip

import (
	"archive/zip"
	"io"
	"time"

	"github.com/xuri/excelize/v2"
)

type Writer struct {
	zipWriter *zip.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		zipWriter: zip.NewWriter(w),
	}
}

func (w *Writer) AddFile(reader io.Reader, path string) error {
	header := &zip.FileHeader{
		Name:     path,
		Modified: time.Now(),
		Method:   zip.Deflate,
	}

	fileWriter, err := w.zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(fileWriter, reader)
	return err
}

func (w *Writer) AddExcelize(xl *excelize.File, path string) error {
	header := &zip.FileHeader{
		Name:     path,
		Modified: time.Now(),
		Method:   zip.Deflate,
	}

	fileWriter, err := w.zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = xl.WriteTo(fileWriter)
	return err
}

func (w *Writer) AddDirectory(path string) error {
	if path[len(path)-1] != '/' {
		path += "/"
	}

	header := &zip.FileHeader{
		Name:     path,
		Modified: time.Now(),
		Method:   zip.Store,
	}

	_, err := w.zipWriter.CreateHeader(header)
	return err
}

func (w *Writer) Close() error {
	return w.zipWriter.Close()
}
