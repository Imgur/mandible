package imageprocessor

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/Imgur/mandible/uploadedfile"
)

func TestStandardOCR(t *testing.T) {
	image, err := getUploadedFileObject()
	if err != nil {
		t.Fatalf("Could not initialize standard OCR test")
	}
	defer image.Clean()

	ocrStratagy := StandardOCRStratagy()
	ocrStratagy.Process(image)

	if image.GetOCRText() != "hello" {
		t.Fatalf("Did not get proper standard OCR text back %s != hello", image.GetOCRText())
	}
}

func getUploadedFileObject() (*uploadedfile.UploadedFile, error) {
	filename, err := copyTestImage("testdata/ocrtestimage.png")
	if err != nil {
		return nil, err
	}

	image, err := uploadedfile.NewUploadedFile("ocrtestimage.png", filename, nil)
	if err != nil {
		return nil, errors.New("Could not initialize standard OCR test")
	}

	return image, nil
}

func copyTestImage(filename string) (string, error) {
	uploadFile, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer uploadFile.Close()

	tmpFile, err := ioutil.TempFile(os.TempDir(), "image")
	if err != nil {
		return "", errors.New("Unable to write to /tmp")
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, uploadFile)
	if err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}
