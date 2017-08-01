package prompt

import (
	"os/user"
	"path/filepath"

	"github.com/AlecAivazis/survey"
)

// DefaultDir  is the default directory for storing dgraph data and config

// InstallDirectory Prompts the user to enter the directory to keep dgraph data and config
// default value is "/var/lib/dgraph"
func InstallDirectory(defaultDir string) string {
	message := "The directory to use for dgraph data and config?"
	theSurvey := []*survey.Question{
		{
			Name: "install_dir",
			Prompt: &survey.Input{
				Message: message,
				Default: defaultDir,
			},
			Validate: survey.Required,
		},
	}

	theAnswers := struct {
		InstallDir string `survey:"install_dir"`
	}{}

	err := survey.Ask(theSurvey, &theAnswers)
	if err != nil {
		panic(err)
	}

	// installDir := defaultString(theAnswers.InstallDir, defaultDir)

	if theAnswers.InstallDir[:2] == "~/" {
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		dir := usr.HomeDir
		return filepath.Join(dir, theAnswers.InstallDir[2:])
	}

	return theAnswers.InstallDir
}
