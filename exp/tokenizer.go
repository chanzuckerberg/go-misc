package exp

import (
	"regexp"
)

func tokenizer() error {
	secret := "aasdfasfd"

	text := []string{
		"asdf@asdf",
		"asfdasf",
	}

	r, err := regexp.Compile("asdf")
	if err != nil {
		return err
	}

	for _, t := range text {
		r.ReplaceAllStringFunc
	}
}
