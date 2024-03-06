package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

const defaultConfigYamlDir = "./"

// ConfigEditorFn is a function that can be used to modify the configuration values after they have been loaded
type ConfigEditorFn[T any] func(cfg *T) error

type ConfigLoader[T any] struct {
	ConfigEditors []ConfigEditorFn[T]
	ConfigYamlDir string
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

func WithConfigYamlDir[T any](dir string) ConfigOption[T] {
	return func(c *ConfigLoader[T]) error {
		c.ConfigYamlDir = dir
		return nil
	}
}

type ValidationError struct {
	FailedField string `json:"failed_field"` // the field that failed to be validated
	Tag         string `json:"tag"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Message     string `json:"message"` // a description of the error that occured
}

func (e ValidationError) Error() string {
	return e.Message
}

// LoadConfiguration loads the configuration from the app-config.yaml and app-config.<env>.yaml files
func LoadConfiguration[T any](cfg *T, opts ...ConfigOption[T]) error {
	configYamlDir := defaultConfigYamlDir
	if len(os.Getenv("CONFIG_YAML_DIRECTORY")) > 0 {
		fmt.Println("CONFIG_YAML_DIRECTORY", os.Getenv("CONFIG_YAML_DIRECTORY"))
		configYamlDir = os.Getenv("CONFIG_YAML_DIRECTORY")
	}

	loader := &ConfigLoader[T]{
		ConfigEditors: []ConfigEditorFn[T]{},
		ConfigYamlDir: configYamlDir,
	}

	for _, opt := range opts {
		err := opt(loader)
		if err != nil {
			return fmt.Errorf("ConfigOption failed: %w", err)
		}
	}

	// TODO: apply defaults?

	loader.populateConfiguration(cfg)

	for _, fn := range loader.ConfigEditors {
		err := fn(cfg)
		if err != nil {
			return fmt.Errorf("ConfigEditorFn failed: %w", err)
		}
	}

	return validateConfiguration(cfg)
}

func validateConfiguration[T any](cfg *T) []ValidationError {
	var errs []*model.ValidationError

	err := validate.Struct(cfg)
	if err != nil {
		errSlice := &validator.ValidationErrors{}
		errors.As(err, errSlice)
		for _, err := range *errSlice {
			var element model.ValidationError
			field, _ := reflect.ValueOf(cfg).Type().FieldByName(err.Field())
			element.FailedField = field.Tag.Get("json")
			if element.FailedField == "" {
				element.FailedField = field.Tag.Get("query")
			}
			element.Tag = err.Tag()
			element.Value = err.Param()
			element.Type = err.Kind().String()

			if _, ok := translatorMessages[element.Tag]; ok {
				element.Message = err.Translate(translator)
			} else {
				element.Message = fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", element.FailedField, element.Tag)
			}

			errs = append(errs, &element)
		}
	}
	return errs
}

func (c *ConfigLoader[T]) populateConfiguration(cfg *T) error {
	configYamlDir := c.ConfigYamlDir
	path, err := filepath.Abs(configYamlDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of %s: %w", configYamlDir, err)
	}

	vpr := viper.New()
	appConfigFile := filepath.Join(path, "app-config.yaml")
	if _, err := os.Stat(appConfigFile); err == nil {
		tmp, err := evaluateConfigWithEnvToTmp(appConfigFile)
		if len(tmp) != 0 {
			defer os.Remove(tmp)
		}
		if err != nil {
			return err
		}

		vpr.SetConfigFile(tmp)
		err = vpr.ReadInConfig()
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	envConfigFilename := fmt.Sprintf("app-config.%s.yaml", getAppEnv())
	appEnvConfigFile := filepath.Join(path, envConfigFilename)
	if _, err := os.Stat(appEnvConfigFile); err == nil {
		tmp, err := evaluateConfigWithEnvToTmp(appEnvConfigFile)
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

	t := template.New("appConfigTemplate")
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
