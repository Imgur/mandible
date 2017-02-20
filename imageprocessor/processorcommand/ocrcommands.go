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

func (this *OCRResult) removeNonWords() {
	blob := this.Text

	speller, err := aspell.NewSpeller(map[string]string{
		"lang": "en_US",
	})
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
		return
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

	this.Text = strings.TrimSpace(str)
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
			result.removeNonWords()
			count := result.wordCount(result.Text)

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
	imageTif := fmt.Sprintf("%s_meme.jpg", image)
	outText := fmt.Sprintf("%s_meme", image)
	inImage := fmt.Sprintf("%s[0]", image)
	preprocessingArgs := []string{"convert", inImage, "-resize", "400%", "-fill", "black", "-fuzz", "10%", "+matte", "-matte", "-transparent", "white", imageTif}
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
	imageTif := fmt.Sprintf("%s_standard.jpg", image)
	outText := fmt.Sprintf("%s_standard", image)
	inImage := fmt.Sprintf("%s[0]", image)
	preprocessingArgs := []string{"convert", inImage, "-resize", "400%", "-type", "Grayscale", imageTif}
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
