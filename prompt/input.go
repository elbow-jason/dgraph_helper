package prompt

import (
	"fmt"

	"github.com/AlecAivazis/survey"
)

// InputFloat64 .
func InputFloat64(message string, defaultNum float64, validator survey.Validator) float64 {
	stringNum := fmt.Sprintf("%.2f", defaultNum)
	for {
		theSurvey := []*survey.Question{
			{
				Name: "num",
				Prompt: &survey.Input{
					Message: message,
					Default: stringNum,
				},
				Validate: validator,
			},
		}

		theAnswers := struct {
			Num float64 `survey:"num"`
		}{}

		err := survey.Ask(theSurvey, &theAnswers)
		if err != nil {
			panic(err)
		}
		return theAnswers.Num
	}
}

// InputString .
func InputString(message string, defaultAnswer string, validator survey.Validator) string {
	var questions []*survey.Question
	if defaultAnswer == "" {
		questions = []*survey.Question{
			{
				Name:     "value",
				Prompt:   &survey.Input{Message: message},
				Validate: validator,
			},
		}
	} else {
		questions = []*survey.Question{
			{
				Name:     "value",
				Prompt:   &survey.Input{Message: message, Default: defaultAnswer},
				Validate: validator,
			},
		}
	}
	answers := struct {
		Value string `survey:"value"`
	}{}
	survey.Ask(questions, &answers)
	return answers.Value
}

// InputInteger asks a question. has a default value (or not). returns an int.
func InputInteger(message string, defaultNum int, hasDefault bool, validator survey.Validator) int {
	var theSurvey []*survey.Question
	if hasDefault {
		stringNum := fmt.Sprintf("%d", defaultNum)
		theSurvey = []*survey.Question{
			{
				Name: "num",
				Prompt: &survey.Input{
					Message: message,
					Default: stringNum,
				},
				Validate: validator,
			},
		}
	} else {
		theSurvey = []*survey.Question{
			{
				Name: "num",
				Prompt: &survey.Input{
					Message: message,
				},
				Validate: validator,
			},
		}

	}

	theAnswers := struct {
		Num int `survey:"num"`
	}{}

	err := survey.Ask(theSurvey, &theAnswers)
	if err != nil {
		panic(err)
	}
	return theAnswers.Num
}

// InputYesOrNo asks a yes or no question with a default answer and returns a bool
func InputYesOrNo(message string, defaultAnswer bool) bool {
	userAnswer := ""
	prompt := &survey.Select{
		Message: message,
		Options: optionsYesOrNo(defaultAnswer),
		Default: stringYesOrNo(defaultAnswer),
	}
	survey.AskOne(prompt, &userAnswer, nil)
	return boolYesOrNo(userAnswer)
}

func boolYesOrNo(userAnswer string) bool {
	if userAnswer == "Y" || userAnswer == "y" {
		return true
	}
	return false
}

func stringYesOrNo(defaultAnswer bool) string {
	if defaultAnswer {
		return "Y"
	}
	return "N"
}

func optionsYesOrNo(defaultAnswer bool) []string {
	if defaultAnswer {
		return []string{"Y", "n"}
	}
	return []string{"N", "y"}
}

// MultiSelectInts .
func MultiSelectInts(message string, start int, count int) []string {
	chosenNumStrings := []string{}
	numStrings := make([]string, count)
	for i := 0; i < count; i++ {
		numStrings[i] = fmt.Sprintf("%d", i)
	}
	prompt := &survey.MultiSelect{
		Message:  message,
		Options:  numStrings,
		PageSize: count,
	}
	for {
		survey.AskOne(prompt, &chosenNumStrings, nil)
		if len(chosenNumStrings) > 0 {
			return chosenNumStrings
		}
	}
}
