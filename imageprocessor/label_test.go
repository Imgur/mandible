package imageprocessor

import (
	"fmt"
	"testing"

	"github.com/Imgur/mandible/uploadedfile"
	"github.com/hashicorp/go-uuid"
)

func TestLabel(t *testing.T) {
	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}

	image, model := getTestFixtures(t, id)
	defer image.Clean()

	err = model.Process(image)
	if err != nil {
		t.Fatal(err)
	}

	labels := image.GetLabels()
	fmt.Println(labels)
	if len(labels) != 5 {
		t.Fatalf("Expected 5 labels, got %d", len(labels))
	}

	if _, ok := labels["Labrador retriever"]; !ok {
		t.Fatalf("Expected Labrador retriever to be in labels list, got: %#v", labels)
	}
}

func BenchmarkLabelModel(b *testing.B) {
	id, err := uuid.GenerateUUID()
	if err != nil {
		b.Fatal(err)
	}

	image, model := getTestFixtures(b, id)
	defer image.Clean()

	err = model.Process(image)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := model.Process(image)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func getTestFixtures(t testing.TB, modelDir string) (*uploadedfile.UploadedFile, *LabelModel) {
	image, err := getUploadedFileObject("dog.jpg")
	if err != nil {
		t.Fatalf("Could not initialize label test")
	}

	model, err := NewLabelModel("/tmp/"+modelDir, 5)
	if err != nil {
		t.Fatalf("Error starting model: %s", err)
	}

	return image, model
}
