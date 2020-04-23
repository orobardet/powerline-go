package main

import (
	"fmt"
	"github.com/joho/godotenv"
	pwl "github.com/justjanne/powerline-go/powerline"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

const (
	MelvisFromSettings = iota
	MelvisFromMdkEnv
	MelvisFromEnv
)

type StackbuilderSettings struct {
	StackName       string `yaml:"stack_name"`
	StackNameSource uint8
	PfName          string `yaml:"pf"`
	PfNameSource    uint8
	SBVersion       string `yaml:"sb_version"`
}

func MelvisReadStackBuilderSettings(config *StackbuilderSettings, path string) (err error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return
	}
	fileContent, err := ioutil.ReadFile(absolutePath)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(fileContent, config)
	if err != nil {
		return
	}

	config.StackNameSource = MelvisFromSettings
	config.PfNameSource = MelvisFromSettings

	return
}

func MelvisReadMdkEnv(config *StackbuilderSettings, path string) (err error) {
	Envs, err := godotenv.Read(path)
	if err != nil {
		return
	}

	if value, exists := Envs["PF"]; exists {
		config.PfName = value
		config.PfNameSource = MelvisFromMdkEnv
	}
	if value, exists := Envs["STACK_NAME"]; exists {
		config.StackName = value
		config.StackNameSource = MelvisFromSettings
	}

	return
}

func MelvisReadEnv(config *StackbuilderSettings) (err error) {
	if value, exists := os.LookupEnv("PF"); exists {
		config.PfName = value
		config.PfNameSource = MelvisFromEnv
	}
	if value, exists := os.LookupEnv("STACK_NAME"); exists {
		config.StackName = value
		config.StackNameSource = MelvisFromEnv
	}

	return
}

func MelvisAnnotateSource(text string, from uint8) string {
	sourceIcon := ""
	switch from {
	case MelvisFromSettings:
		sourceIcon = ""
	case MelvisFromMdkEnv:
		sourceIcon = "\u1D39"
	case MelvisFromEnv:
		sourceIcon = "\u1D31"
	}

	return fmt.Sprintf("%s%s", text, sourceIcon)
}

func segmentMelvis(p *powerline) []pwl.Segment {
	cwd, err := os.Getwd()
	segments := []pwl.Segment{}
	if err != nil {
		return segments
	}
	stackSettings := &StackbuilderSettings{}
	MelvisReadStackBuilderSettings(stackSettings, path.Join(cwd, "/settings.yml"))
	MelvisReadMdkEnv(stackSettings, path.Join(cwd, "/.mdk.env"))
	MelvisReadEnv(stackSettings)

	if stackSettings.StackName == "" && stackSettings.PfName == "" && stackSettings.SBVersion == "" {
		return segments
	}

	if stackSettings.StackName == "" {
		stackSettings.StackName = "??"
	}
	if stackSettings.PfName == "" {
		stackSettings.PfName = "??"
	}

	// Add Melvis icon
	content := "\u22C0\u22C0"

	// Stackbuilder version
	if stackSettings.SBVersion != "" {
		content = fmt.Sprintf("%s v%s", content, stackSettings.SBVersion)
	}

	// Stack@PF
	stackName := stackSettings.StackName
	if stackName == "" {
		stackName = "??"
	}
	pfName := stackSettings.PfName
	if pfName == "" {
		pfName = "??"
	}
	content = fmt.Sprintf("%s %s \u2192 %s",
		content,
		MelvisAnnotateSource(stackName, stackSettings.StackNameSource),
		MelvisAnnotateSource(pfName, stackSettings.PfNameSource),
	)

	// Add the segment
	segments = append(segments, pwl.Segment{
		Name:           "melvis",
		Content:        content,
		Foreground:     p.theme.MelvisFg,
		Background:     p.theme.MelvisBg,
		HideSeparators: false,
	})

	return segments
}
