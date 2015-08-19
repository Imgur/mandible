package processorcommand

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/trustmaster/go-aspell"
)

type OCRResult struct {
	Type string
	Text string
}

func newOCRResult(ocrType string, result string) *OCRResult {
	return &OCRResult{
		ocrType,
		result,
	}
}

func (this *OCRResult) removeNonWords(blob string) string {
	speller, err := aspell.NewSpeller(map[string]string{
		"lang": "en_US",
	})
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return ""
	}
	defer speller.Delete()

	singleCharWords := regexp.MustCompile("(a|i)")
	numberRegex := regexp.MustCompile("\\d{3,}")
	wordRegexp := regexp.MustCompile("\\b(\\w+)\\b")
	words := wordRegexp.FindAllString(blob, -1)

	str := ""

	for _, word := range words {
		if numberRegex.MatchString(word) {
			str += " " + word
		} else if len(word) == 1 {
			if singleCharWords.MatchString(word) {
				str += " " + word
			}
		} else if speller.Check(word) {
			str += " " + word
		}
	}

	return strings.TrimSpace(str)
}

func (this *OCRResult) getBestString() string {
	cleanString := this.removeNonWords(this.Text)
	ocrWordCount := this.wordCount(this.Text)
	cleanWordCount := this.wordCount(cleanString)

	percentKept := float64(cleanWordCount) / float64(ocrWordCount)

	if percentKept > .90 {
		return this.Text
	}

	return cleanString
}

func (this *OCRResult) wordCount(blob string) int {
	word_regexp := regexp.MustCompile("\\b(\\w+)\\b")
	words := word_regexp.FindAllString(blob, -1)

	// don't let single char words count towards the overal word count. Gets thrown off by poor OCR results
	count := 0
	for _, word := range words {
		if len(word) > 1 {
			count++
		}
	}

	return count
}

type MultiOCRCommand []OCRCommand

func (this MultiOCRCommand) Run(image string) (*OCRResult, error) {
	results := make(chan *OCRResult, len(this))
	errs := make(chan error, len(this))

	for _, command := range this {
		go func(c OCRCommand) {
			k, err := c.Run(image)
			if err != nil {
				errs <- err
				return
			}
			results <- k
		}(command)
	}

	max := -1
	var best *OCRResult

	for i := 0; i < len(this); i++ {
		select {
		case result := <-results:
			blob := result.getBestString()
			count := result.wordCount(blob)

			if count > max {
				best = result
				max = count
			}

		case err := <-errs:
			return nil, err
		}
	}

	// Return the average, same as before.
	return best, nil
}

type OCRCommand interface {
	Run(image string) (*OCRResult, error)
}

type MemeOCR struct {
	name string
}

func NewMemeOCR() *MemeOCR {
	return &MemeOCR{
		"MemeOCR",
	}
}

func (this *MemeOCR) Run(image string) (*OCRResult, error) {
	imageTif := fmt.Sprintf("%s_meme.tif", image)
	outText := fmt.Sprintf("%s_meme", image)
	preprocessingArgs := []string{image, "-resize", "400%", "-fill", "black", "-fuzz", "10%", "+opaque", "#FFFFFF", imageTif}
	tesseractArgs := []string{"-l", "meme", imageTif, outText}

	err := runProcessorCommand(GM_COMMAND, preprocessingArgs)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Meme preprocessing command failed with error = %v", err))
	}
	defer os.Remove(imageTif)

	err = runProcessorCommand("tesseract", tesseractArgs)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Meme tesseract command failed with error = %v", err))
	}
	defer os.Remove(outText + ".txt")

	text, err := ioutil.ReadFile(outText + ".txt")
	if err != nil {
		return nil, err
	}

	result := strings.ToLower(strings.TrimSpace(string(text[:])))

	return newOCRResult(this.name, result), nil
}

type StandardOCR struct {
	name string
}

func NewStandardOCR() *StandardOCR {
	return &StandardOCR{
		"StandardOCR",
	}
}

func (this *StandardOCR) Run(image string) (*OCRResult, error) {
	imageTif := fmt.Sprintf("%s_standard.tif", image)
	outText := fmt.Sprintf("%s_standard", image)

	preprocessingArgs := []string{image, "-resize", "400%", "-type", "Grayscale", imageTif}
	tesseractArgs := []string{"-l", "eng", imageTif, outText}

	err := runProcessorCommand(GM_COMMAND, preprocessingArgs)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Standard preprocessing command failed with error = %v", err))
	}
	defer os.Remove(imageTif)

	err = runProcessorCommand("tesseract", tesseractArgs)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Standard tesseract command failed with error = %v", err))
	}
	defer os.Remove(outText + ".txt")

	text, err := ioutil.ReadFile(outText + ".txt")
	if err != nil {
		return nil, err
	}

	result := strings.ToLower(strings.TrimSpace(string(text[:])))

	return newOCRResult(this.name, result), nil
}
