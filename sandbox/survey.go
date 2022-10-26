package main

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
)

func main() {
	options := []string{"Golang", "Ruby"}

	var tag string
	q := &survey.Select{
		Message: "Choose a tag",
		Options: options,
		Description: func(value string, index int) string {
			if value == "Ruby" {
				return "My favorite Language"
			}
			return ""
		},
		Default: options[0],
	}

	//err := prompt.SurveyAskOne(q, &tag)
	//if err != nil {
	//	fmt.Errorf("could not prompt: %w", err)
	//}

	survey.AskOne(q, &tag)

	fmt.Println(tag)
}
