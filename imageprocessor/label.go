package imageprocessor

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"log"

	"github.com/Imgur/mandible/uploadedfile"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
)

type LabelModel struct {
	graph     *tf.Graph
	modelFile string
	labels    []string
}

func NewLabelModel(modelDir string) (*LabelModel, error) {
	if modelDir == "" {
		return nil, errors.New("Invalid location to save label model")
	}

	// Load the serialized GraphDef from a file.
	modelFile, labelsFile, err := modelFiles(modelDir)
	if err != nil {
		return nil, err
	}
	model, err := ioutil.ReadFile(modelFile)
	if err != nil {
		return nil, err
	}

	// Construct an in-memory graph from the serialized form.
	graph := tf.NewGraph()
	if err := graph.Import(model, ""); err != nil {
		return nil, err
	}

	// Found the best match. Read the string from labelsFile, which
	// contains one line per label.
	file, err := os.Open(labelsFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var labels []string
	for scanner.Scan() {
		labels = append(labels, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &LabelModel{
		graph:     graph,
		modelFile: modelFile,
		labels:    labels,
	}, nil
}

func (m *LabelModel) Process(image *uploadedfile.UploadedFile) error {
	session, err := tf.NewSession(m.graph, nil)
	if err != nil {
		return fmt.Errorf("Error creating tf session: %s", err)
	}
	defer session.Close()

	// Run inference on *imageFile.
	// For multiple images, session.Run() can be called in a loop (and
	// concurrently). Alternatively, images can be batched since the model
	// accepts batches of image data as input.
	tensor, err := makeTensorFromImage(image)
	if err != nil {
		return err
	}
	output, err := session.Run(
		map[tf.Output]*tf.Tensor{
			m.graph.Operation("input").Output(0): tensor,
		},
		[]tf.Output{
			m.graph.Operation("output").Output(0),
		},
		nil)
	if err != nil {
		return err
	}
	// output[0].Value() is a vector containing probabilities of
	// labels for each image in the "batch". The batch size was 1.
	// Find the most probably label index.
	probabilities := output[0].Value().([][]float32)[0]
	topLabels := m.printBestLabel(probabilities)
	image.SetLabels(topLabels)

	return nil
}

func (this *LabelModel) String() string {
	return "LabelModel runner"
}

type Pair struct {
	Label       string
	Probability float32
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Probability < p[j].Probability }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (m *LabelModel) printBestLabel(probabilities []float32) map[string]float32 {
	labelsToProb := make(PairList, len(m.labels))
	for i, _ := range m.labels {
		labelsToProb[i] = Pair{m.labels[i], probabilities[i]}
	}
	sort.Sort(sort.Reverse(labelsToProb))

	topLabelsToProb := make(map[string]float32, 5)
	for i := 0; i < 5; i++ {
		topLabelsToProb[labelsToProb[i].Label] = labelsToProb[i].Probability
	}

	return topLabelsToProb
}

// Convert the image in filename to a Tensor suitable as input to the Inception model.
func makeTensorFromImage(image *uploadedfile.UploadedFile) (*tf.Tensor, error) {
	bytes, err := ioutil.ReadFile(image.GetPath())
	if err != nil {
		return nil, err
	}
	// DecodeJpeg uses a scalar String-valued tensor as input.
	tensor, err := tf.NewTensor(string(bytes))
	if err != nil {
		return nil, err
	}
	// Construct a graph to normalize the image
	graph, input, output, err := constructGraphToNormalizeImage(image.GetMime())
	if err != nil {
		return nil, err
	}
	// Execute that graph to normalize this one image
	session, err := tf.NewSession(graph, nil)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	normalized, err := session.Run(
		map[tf.Output]*tf.Tensor{input: tensor},
		[]tf.Output{output},
		nil)
	if err != nil {
		return nil, err
	}
	return normalized[0], nil
}

// The inception model takes as input the image described by a Tensor in a very
// specific normalized format (a particular image size, shape of the input tensor,
// normalized pixel values etc.).
//
// This function constructs a graph of TensorFlow operations which takes as
// input a JPEG-encoded string and returns a tensor suitable as input to the
// inception model.
func constructGraphToNormalizeImage(format string) (graph *tf.Graph, input, output tf.Output, err error) {
	// Some constants specific to the pre-trained model at:
	// https://storage.googleapis.com/download.tensorflow.org/models/inception5h.zip
	//
	// - The model was trained after with images scaled to 224x224 pixels.
	// - The colors, represented as R, G, B in 1-byte each were converted to
	//   float using (value - Mean)/Scale.
	const (
		H, W  = 224, 224
		Mean  = float32(117)
		Scale = float32(1)
	)
	// - input is a String-Tensor, where the string the JPEG-encoded image.
	// - The inception model takes a 4D tensor of shape
	//   [BatchSize, Height, Width, Colors=3], where each pixel is
	//   represented as a triplet of floats
	// - Apply normalization on each pixel and use ExpandDims to make
	//   this single image be a "batch" of size 1 for ResizeBilinear.
	s := op.NewScope()
	input = op.Placeholder(s, tf.String)
	var decode tf.Output
	switch format {
	case "image/jpg", "image/jpeg":
		decode = op.DecodeJpeg(s, input, op.DecodeJpegChannels(3))
	case "image/png":
		decode = op.DecodePng(s, input, op.DecodePngChannels(3))
	default:
		err = errors.New("Image type not suppored")
		return
	}

	output = op.Div(s,
		op.Sub(s,
			op.ResizeBilinear(s,
				op.ExpandDims(s,
					op.Cast(s, decode, tf.Float),
					op.Const(s.SubScope("make_batch"), int32(0))),
				op.Const(s.SubScope("size"), []int32{H, W})),
			op.Const(s.SubScope("mean"), Mean)),
		op.Const(s.SubScope("scale"), Scale))
	graph, err = s.Finalize()

	return graph, input, output, err
}

func modelFiles(dir string) (modelfile, labelsfile string, err error) {
	const URL = "https://storage.googleapis.com/download.tensorflow.org/models/inception5h.zip"
	var (
		model   = filepath.Join(dir, "tensorflow_inception_graph.pb")
		labels  = filepath.Join(dir, "imagenet_comp_graph_label_strings.txt")
		zipfile = filepath.Join(dir, "inception5h.zip")
	)
	if filesExist(model, labels) == nil {
		return model, labels, nil
	}
	log.Println("Did not find model in", dir, "downloading from", URL)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", err
	}
	if err := download(URL, zipfile); err != nil {
		return "", "", fmt.Errorf("failed to download %v - %v", URL, err)
	}
	if err := unzip(dir, zipfile); err != nil {
		return "", "", fmt.Errorf("failed to extract contents from model archive: %v", err)
	}
	os.Remove(zipfile)
	return model, labels, filesExist(model, labels)
}

func filesExist(files ...string) error {
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			return fmt.Errorf("unable to stat %s: %v", f, err)
		}
	}
	return nil
}

func download(URL, filename string) error {
	resp, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	return err
}

func unzip(dir, zipfile string) error {
	r, err := zip.OpenReader(zipfile)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		src, err := f.Open()
		if err != nil {
			return err
		}
		log.Println("Extracting", f.Name)
		dst, err := os.OpenFile(filepath.Join(dir, f.Name), os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		if _, err := io.Copy(dst, src); err != nil {
			return err
		}
		dst.Close()
	}
	return nil
}
