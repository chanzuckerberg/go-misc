package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

const defaultConfigYamlDir = "./"
const defaultConfigFileBaseName = "app-config"

// ConfigEditorFn is a function that can be used to modify the configuration values after they have been loaded
type ConfigEditorFn[T any] func(cfg *T) error

type ConfigLoader[T any] struct {
	ConfigEditors             []ConfigEditorFn[T]
	ConfigYamlDir             string
	ConfigFileBaseName        string
	AdditionalConfigFileNames []string
}

// ConfigOption is a function that can be used to modify the configuration loader
type ConfigOption[T any] func(*ConfigLoader[T]) error

// WithConfigEditorFn allows setting up a callback function, which will be
// called right after loading the configs. This can be used to mutate the config,
// for example to set default values where none were set by the call to LoadConfiguration.
func WithConfigEditorFn[T any](fn ConfigEditorFn[T]) ConfigOption[T] {
	return func(c *ConfigLoader[T]) error {
		c.ConfigEditors = append(c.ConfigEditors, fn)
		return nil
	}
}

// WithConfigYamlDir allows changing the directory where the config files are located.
// The default is "./", meaning the same as where your main.go is located.
func WithConfigYamlDir[T any](dir string) ConfigOption[T] {
	return func(c *ConfigLoader[T]) error {
		c.ConfigYamlDir = dir
		return nil
	}
}

// WithConfigFileBaseName allows changing the base name of the config file to be loaded.
// The default is "app-config", so the main config file is app-config.yaml and the env config file is app-config.<env>.yaml
func WithConfigFileBaseName[T any](fileNamePrefix string) ConfigOption[T] {
	return func(c *ConfigLoader[T]) error {
		c.ConfigFileBaseName = fileNamePrefix
		return nil
	}
}

// WithOverrideConfigFile allows adding additional config files to be loaded AFTER the main config file and the env config file,
// meaning it will override any values set in the main config file and the env config file.
func WithOverrideConfigFile[T any](fileName string) ConfigOption[T] {
	return func(c *ConfigLoader[T]) error {
		c.AdditionalConfigFileNames = append(c.AdditionalConfigFileNames, fileName)
		return nil
	}
}

// LoadConfiguration loads the configuration from the app-config.yaml and app-config.<env>.yaml files
func LoadConfiguration[T any](cfg *T, opts ...ConfigOption[T]) error {
	configYamlDir := defaultConfigYamlDir
	if len(os.Getenv("CONFIG_YAML_DIRECTORY")) > 0 {
		configYamlDir = os.Getenv("CONFIG_YAML_DIRECTORY")
	}

	loader := &ConfigLoader[T]{
		ConfigEditors:      []ConfigEditorFn[T]{},
		ConfigYamlDir:      configYamlDir,
		ConfigFileBaseName: defaultConfigFileBaseName,
	}

	for _, opt := range opts {
		err := opt(loader)
		if err != nil {
			return fmt.Errorf("ConfigOption failed: %w", err)
		}
	}

	err := loader.populateConfiguration(cfg)
	if err != nil {
		return fmt.Errorf("Populating configuration failed: %w", err)
	}

	for _, fn := range loader.ConfigEditors {
		err := fn(cfg)
		if err != nil {
			return fmt.Errorf("ConfigEditorFn failed: %w", err)
		}
	}

	return validateConfiguration(cfg)
}

func (c *ConfigLoader[T]) populateConfiguration(cfg *T) error {
	configYamlDir := c.ConfigYamlDir
	path, err := filepath.Abs(configYamlDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of %s: %w", configYamlDir, err)
	}

	vpr := viper.New()

	appConfigFile := fmt.Sprintf("%s.yaml", c.ConfigFileBaseName)
	appConfigFiles := []string{appConfigFile}

	// add the env specific config file if appEnv is set
	appEnv := getAppEnv()
	if len(appEnv) > 0 {
		appEnvConfigFile := fmt.Sprintf("%s.%s.yaml", c.ConfigFileBaseName, appEnv)
		appConfigFiles = append(appConfigFiles, appEnvConfigFile)
	}

	// add additional config files
	appConfigFiles = append(appConfigFiles, c.AdditionalConfigFileNames...)

	// iterate the appConfig files to be used, read the first one and merge the rest
	for _, configFile := range appConfigFiles {
		absoluteConfigFilePath := filepath.Join(path, configFile)
		_, err := os.Stat(absoluteConfigFilePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue // if the file does not exist, skip it
			}
			return fmt.Errorf("failed to get file info for %s: %w", absoluteConfigFilePath, err)
		}

		tmp, err := evaluateConfigWithEnvToTmp(absoluteConfigFilePath)
		if len(tmp) != 0 {
			defer os.Remove(tmp)
		}
		if err != nil {
			return err
		}

		vpr.SetConfigFile(tmp)
		err = vpr.MergeInConfig()
		if err != nil {
			return fmt.Errorf("failed to merge env config: %w", err)
		}
	}

	err = vpr.Unmarshal(cfg, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc()))
	if err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return nil
}

func evaluateConfigWithEnvToTmp(configPath string) (string, error) {
	tmp, err := os.CreateTemp("./", "*.yaml")
	if err != nil {
		return "", fmt.Errorf("unable to create a temp config file: %w", err)
	}

	cfile, err := os.Open(configPath)
	if err != nil {
		return "", fmt.Errorf("unable to open %s: %w", configPath, err)
	}

	_, err = evaluateConfigWithEnv(cfile, tmp)
	if err != nil {
		return "", fmt.Errorf("unable to populate the environment: %w", err)
	}

	return tmp.Name(), nil
}

func envToMap() map[string]string {
	envMap := make(map[string]string)
	for _, v := range os.Environ() {
		s := strings.SplitN(v, "=", 2)
		if len(s) != 2 {
			continue
		}
		envMap[s[0]] = s[1]
	}
	return envMap
}

// evaluateConfigWithEnv reads a configuration reader and injects environment variables
// that exist as part of the configuration in the form a go template. For example
// {{.ENV_VAR1}} will be replace with the value of the environment variable ENV_VAR1.
// Optional support for writting the contents to other places is supported by providing
// other writers. By default, the evaluated configuartion is returned as a reader.
func evaluateConfigWithEnv(configFile io.Reader, writers ...io.Writer) (io.Reader, error) {
	envMap := envToMap()

	b, err := io.ReadAll(configFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read the config file: %w", err)
	}

	t := template.New("appConfigTemplate").Option("missingkey=zero")
	tmpl, err := t.Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("unable to parse template from: \n%s: %w", string(b), err)
	}

	populated := []byte{}
	buff := bytes.NewBuffer(populated)
	writers = append(writers, buff)
	err = tmpl.Execute(io.MultiWriter(writers...), envMap)
	if err != nil {
		return nil, fmt.Errorf("unable to execute template: %w", err)
	}
	return buff, nil
}

func getAppEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("DEPLOYMENT_STAGE")
	}
	return env
}

func validateConfiguration[T any](cfg *T) error {
	validate := validator.New()
	err := validate.Struct(*cfg)
	if err != nil {
		errSlice := &validator.ValidationErrors{}
		errors.As(err, errSlice)
		return errSlice
	}

	return nil
}
