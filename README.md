# configstore

[![Build Status](https://drone.levitate.bio/api/badges/LevitateBio/configstore/status.svg)](https://drone.levitate.bio/LevitateBio/configstore)

This is an opinionated library for handling the management of internal configuration in web applications. 
This library was designed around the following observations.

* If your service is deployed as a Docker container it's very convenient to load configuration in through env variables
* The service may need the configuration values across the codebase, so it should be managed as a singleton
* A huge range of potential deployment problems can be identified if you print the entire loaded config to logs when the process starts
* Sometimes there are secrets in your configuration you don't want in your logs, so you need a way to censor secrets
* One of the deployment problems is failing to set a secret, so you want your PrettyPrinted config table to make it clear when a secret value isnt set at all
* It should be easy to override env variable definitions in unit tests

# Installation

just run `go get github.com/levitatebio/configstore`

# Usage

To use this library you should define a struct containing your config values, with struct tags defining default values and env variables, like this:

```go
type MyConfig struct {
	IntValue             int32            `env:"INT_VAL" default:"1"`
	BoolValue            bool             `env:"BOOL_VAL" default:"true"`
	StringValue          string           `env:"STRING_VAL" default:"default_value"`
	StringSliceValue     []string         `env:"STRING_SLICE_VAL" default:"foo,bar"`
	IntMapValue          map[string]int32 `env:"INT_MAP_VAL" default:"foo=1,bar=2"`
	SecretIntValue       int32            `env:"SECRET_INT_VAL" secret:"true" default:"3"`
}
```

Ideally, you want to manage this struct as a singleton, like this:

```go

var (
	once sync.Once
	TestMode bool
	config MyConfig
)

func GetConfig() *MyConfig{
	configstore.LoadOnce(&config, TestMode, &once)
	return &config
}
```

You can then retrieve config values anywhere in your application like this:

```go
config := GetConfig()
myInt := config.IntValue
```

You should ideally load the config and print it out as early as possible in the execution of your application:

```go
config := GetConfig()
configstore.Print(config)
```

This will result in a table being printed like below:

```
OPTION                 ENV VAR            SETTING
IntValue               INT_VAL            2
BoolValue              BOOL_VAL           false
StringValue            STRING_VAL         foo
StringSliceValue       STRING_SLICE_VAL   [a b]
IntMapValue            INT_MAP_VAL        map[c:3 d:4]
SecretIntValue         SECRET_INT_VAL     ********
```

