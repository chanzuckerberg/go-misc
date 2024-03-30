package config

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvToMap(t *testing.T) {
	r := require.New(t)
	os.Clearenv()
	err := os.Setenv("ENV1", "test1")
	r.NoError(err)
	err = os.Setenv("ENV2", "test2")
	r.NoError(err)
	defer os.Unsetenv("ENV1")
	defer os.Unsetenv("ENV2")
	m := envToMap()
	r.Contains(m, "ENV1")
	r.Contains(m, "ENV2")
}

func TestEvaluateConfigWithEnv(t *testing.T) {
	r := require.New(t)
	test := `blah={{.ENV1}}
blah2={{.ENV2}}`
	os.Setenv("ENV1", "test1")
	os.Setenv("ENV2", "test2")
	defer os.Unsetenv("ENV1")
	defer os.Unsetenv("ENV2")
	eval, err := evaluateConfigWithEnv(strings.NewReader(test))
	r.NoError(err)
	expected := `blah=test1
blah2=test2`
	b, err := io.ReadAll(eval)
	r.NoError(err)
	r.Equal(expected, string(b))
}

func TestEvaluateConfigWithMissingEnv(t *testing.T) {
	r := require.New(t)
	test := `blah={{.ENV1}}
blah2={{.ENV2}}`
	os.Setenv("ENV1", "test1")
	defer os.Unsetenv("ENV1")

	eval, err := evaluateConfigWithEnv(strings.NewReader(test))
	r.NoError(err)
	expected := `blah=test1
blah2=`
	b, err := io.ReadAll(eval)
	r.NoError(err)
	r.Equal(expected, string(b))
}

type nestedConfig struct {
	Value1 string `json:"value1"`
}

type testConfig struct {
	Value1 string       `json:"value1"`
	Value2 string       `json:"value2"`
	Nested nestedConfig `json:"nested"`
}

func TestLoadConfigurationBasic(t *testing.T) {
	r := require.New(t)

	cfg := &testConfig{}
	err := LoadConfiguration(cfg, WithConfigYamlDir[testConfig]("./testData/basic"))
	r.NoError(err)

	r.Equal("foo", cfg.Value1)
	r.Equal("bar", cfg.Value2)
	r.Equal("zap", cfg.Nested.Value1)
}

func TestLoadConfigurationOverlay(t *testing.T) {
	r := require.New(t)

	// Set the deployment stage to test so it overlays the app-config.test.yaml file
	os.Setenv("DEPLOYMENT_STAGE", "test")
	defer os.Unsetenv("DEPLOYMENT_STAGE")

	cfg := &testConfig{}
	err := LoadConfiguration(cfg, WithConfigYamlDir[testConfig]("./testData/overlay"))
	r.NoError(err)

	r.Equal("testval1", cfg.Value1)
	r.Equal("testval2", cfg.Value2)
	r.Equal("zap", cfg.Nested.Value1)
}

type validatedConfig struct {
	Value1 string `json:"value1" validate:"required"`
	Value2 string `json:"value2" validate:"required"`
	Value3 string `json:"value3"`
}

// this should fail because the values are missing from the yaml file
func TestLoadConfigurationValidatedFail(t *testing.T) {
	r := require.New(t)

	cfg := &validatedConfig{}
	err := LoadConfiguration(cfg, WithConfigYamlDir[validatedConfig]("./testData/validation"))
	r.Error(err)

	r.Contains(err.Error(), "Value1")
	r.Contains(err.Error(), "Value2")
	r.Equal("zap", cfg.Value3)
}

// this should succeed because the values are supplied by the editor function
func TestLoadConfigurationValidatedSucceedWithConfigEditor(t *testing.T) {
	r := require.New(t)

	cfg := &validatedConfig{}
	err := LoadConfiguration(
		cfg,
		WithConfigYamlDir[validatedConfig]("./testData/validation"),
		WithConfigEditorFn(func(cfg *validatedConfig) error {
			cfg.Value1 = "edited-value1"
			cfg.Value2 = "edited-value2"
			return nil
		}),
	)
	r.NoError(err)

	r.Equal("edited-value1", cfg.Value1)
	r.Equal("edited-value2", cfg.Value2)
	r.Equal("zap", cfg.Value3)
}

// this should succeed because the values are supplied by the overlay file
func TestLoadConfigurationValidatedSucceedWithOverlay(t *testing.T) {
	r := require.New(t)

	// Set the deployment stage to test so it overlays the app-config.test.yaml file
	os.Setenv("DEPLOYMENT_STAGE", "withvalues")
	defer os.Unsetenv("DEPLOYMENT_STAGE")

	cfg := &validatedConfig{}
	err := LoadConfiguration(cfg, WithConfigYamlDir[validatedConfig]("./testData/validation"))
	r.NoError(err)

	r.Equal("foo", cfg.Value1)
	r.Equal("bar", cfg.Value2)
	r.Equal("zap", cfg.Value3)
}
