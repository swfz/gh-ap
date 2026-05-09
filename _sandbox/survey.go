package main

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"strconv"
)

func main() {
	var qs = []*survey.Question{
		{
			Name:      "name",
			Prompt:    &survey.Input{Message: "What is your name?"},
			Validate:  survey.Required,
			Transform: survey.Title,
		},
		{
			Name: "color",
			Prompt: &survey.Select{
				Message: "Choose a color:",
				Options: []string{"red", "blue", "green"},
				Default: "red",
			},
		},
		{
			Name:   "age",
			Prompt: &survey.Input{Message: "How old are you?"},
			Validate: func(v interface{}) error {
				//value := reflect.ValueOf(v)
				//return nil
				strValue := v.(string)
				_, err := strconv.ParseFloat(strValue, 64)
				if err != nil {
					return errors.New("Value is Int or Float")
				}
				return nil
			},
		},
	}

	answers := struct {
		Name          string  // survey will match the question and field names
		FavoriteColor string  `survey:"color"` // or you can tag fields to match a specific name
		Age           float64 // if the types don't match, survey will convert it
	}{}

	// perform the questions
	err := survey.Ask(qs, &answers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("%s chose %s.", answers.Name, answers.FavoriteColor)

	fmt.Printf("%+v\n", answers)
}
