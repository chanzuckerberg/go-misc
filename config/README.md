# config

A utility for loading configuration YAML files.

## Design

The `config` package provides a way to load configuration from YAML files. By default it will load a file named `app-config.yaml` in the root directory where your main package resides. Then it will load a file named `app-config.<env>.yaml` and use those values to override the previous configuration. The environment is determined by the value of the `APP_ENV` environment variable. If `APP_ENV` is not set, it will fall back to the value of the `DEPLOYEMENT_STAGE` environment variable. If neither are set, no override file will be loaded. The `config` package also supports loading configuration from a custom directory.

### Custom Configuration Directory

There are two ways to specify a custom directory to load configuration files from. The first is to set the `CONFIG_YAML_DIRECTORY` environment variable. The second is to use the `WithConfigYamlDir` option when calling `LoadConfiguration`.

### Configuration Structs

The `config` package uses the `mapstructure` package to load configuration into a struct. This allows for the use of tags to map configuration fields to struct fields. For example, the following YAML file:
```yaml
auth:
  enable: true
```
can be loaded into the following struct:
```go
type AuthConfiguration struct {
  Enable    *bool  `mapstructure:"enable"`
}

type Configuration struct {
  Auth     AuthConfiguration     `mapstructure:"auth"`
}
```

### Environment Variables

The `config` package supports templated injection of environment variables in your YAML files to avoid storing sensitive values in the YAMLs. For example, the following YAML file:
```yaml
database:
  driver: postgres
  data_source_name: host={{.PLATFORM_DATABASE_HOST}} user={{.PLATFORM_DATABASE_USER}} password={{.PLATFORM_DATABASE_PASSWORD}} port={{.PLATFORM_DATABASE_PORT}} dbname={{.PLATFORM_DATABASE_NAME}}
```
will have the environment variables `PLATFORM_DATABASE_HOST`, `PLATFORM_DATABASE_USER`, `PLATFORM_DATABASE_PASSWORD`, `PLATFORM_DATABASE_PORT`, and `PLATFORM_DATABASE_NAME` injected into the `data_source_name` field prior to loading into your configuration struct.

## Usage

Example:
```go
package main

import (
  "fmt"

  "github.com/chanzuckerberg/go-misc/config"
)

type AuthConfiguration struct {
	Enable    *bool  `mapstructure:"enable"`
}

type ApiConfiguration struct {
	Port                uint   `mapstructure:"port"`
	LogLevel            string `mapstructure:"log_level"`
}

type DBDriver string

func (d *DBDriver) String() string {
	return string(*d)
}

const (
	Sqlite   DBDriver = "sqlite"
	Postgres DBDriver = "postgres"
)

type DatabaseConfiguration struct {
	Driver         DBDriver `mapstructure:"driver"`
	DataSourceName string   `mapstructure:"data_source_name"`
}

type Configuration struct {
	Auth     AuthConfiguration     `mapstructure:"auth"`
	Api      ApiConfiguration      `mapstructure:"api"`
	Database DatabaseConfiguration `mapstructure:"database"`
}

func main() {
  cfg := &Configuration{}

  configOpts := []config.ConfigOption[Configuration]{
    config.WithConfigYamlDir[Configuration]("./configs"),
    config.WithConfigEditorFn(func(cfg *Configuration) error {
      // default to having auth enabled
      if cfg.Auth.Enable == nil {
        enable := true
        cfg.Auth.Enable = &enable
      }
      return nil
    }),
  }

  err := config.LoadConfiguration(cfg, configOpts...)
  if err != nil {
    panic(fmt.Sprintf("Failed to load app configuration: %s", err.Error()))
  }
}
```
