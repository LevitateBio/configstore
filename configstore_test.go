package configstore

import (
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"sync"
	"testing"
)

type testStruct struct {
	IntValue             int32            `env:"INT_VAL" default:"1"`
	BoolValue            bool             `env:"BOOL_VAL" default:"true"`
	StringValue          string           `env:"STRING_VAL" default:"default_value"`
	StringValueNoDefault string           `env:"NO_DEFAULT_VAL"`
	StringSliceValue     []string         `env:"STRING_SLICE_VAL" default:"foo,bar"`
	IntMapValue          map[string]int32 `env:"INT_MAP_VAL" default:"foo=1,bar=2"`
	SecretIntValue       int32            `env:"SECRET_INT_VAL" secret:"true" default:"3"`
}

func TestGetEnvValueString(t *testing.T) {
	s := testStruct{}
	structType := reflect.TypeOf(s)
	os.Setenv("STRING_VAL", "test_value")
	stringValField, _ := structType.FieldByName("StringValue")
	envValue := getEnvValueString(stringValField.Tag)
	assert.Equal(t, "test_value", envValue)

	os.Unsetenv("STRING_VAL")
	defaultValue := getEnvValueString(stringValField.Tag)
	assert.Equal(t, "default_value", defaultValue)

	stringValNoDefaultField, _ := structType.FieldByName("StringValueNoDefault")
	noDefaultValue := getEnvValueString(stringValNoDefaultField.Tag)
	assert.Equal(t, "", noDefaultValue)
}

func TestGetEnvValueStrings(t *testing.T) {
	s := testStruct{}
	structType := reflect.TypeOf(s)
	os.Setenv("STRING_SLICE_VAL", "test,test2")
	stringSliceField, _ := structType.FieldByName("StringSliceValue")
	envValue := getEnvValueStrings(stringSliceField.Tag)
	assert.Equal(t, []string{"test", "test2"}, envValue)

	os.Unsetenv("STRING_SLICE_VAL")
	defaultValue := getEnvValueStrings(stringSliceField.Tag)
	assert.Equal(t, []string{"foo", "bar"}, defaultValue)
}

func TestGetEnvValueBool(t *testing.T) {
	s := testStruct{}
	structType := reflect.TypeOf(s)
	os.Setenv("BOOL_VAL", "false")
	boolValField, _ := structType.FieldByName("BoolValue")
	envValue := getEnvValueBool(boolValField.Tag)
	assert.False(t, envValue)

	os.Unsetenv("BOOL_VAL")
	defaultValue := getEnvValueBool(boolValField.Tag)
	assert.True(t, defaultValue)
}

func TestGetEnvValueInt(t *testing.T) {
	s := testStruct{}
	structType := reflect.TypeOf(s)
	os.Setenv("INT_VAL", "2")
	intValField, _ := structType.FieldByName("IntValue")
	envValue := getEnvValueInt(intValField.Tag)
	assert.Equal(t, int64(2), envValue)

	os.Unsetenv("INT_VAL")
	defaultValue := getEnvValueInt(intValField.Tag)
	assert.Equal(t, int64(1), defaultValue)
}

func TestGetEnvValueIntMap(t *testing.T) {
	s := testStruct{}
	structType := reflect.TypeOf(s)
	os.Setenv("INT_MAP_VAL", "test1=5,test2=10")
	mapValueField, _ := structType.FieldByName("IntMapValue")
	mapValue := getEnvValueIntMap(mapValueField.Tag)
	assert.Equal(t, map[string]int32{"test1": 5, "test2": 10}, mapValue)

	os.Unsetenv("INT_MAP_VAL")
	defaultValue := getEnvValueIntMap(mapValueField.Tag)
	assert.Equal(t, map[string]int32{"foo": 1, "bar": 2}, defaultValue)

}

func TestIsEnvValueSecret(t *testing.T) {
	s := testStruct{}
	structType := reflect.TypeOf(s)
	secretIntField, _ := structType.FieldByName("SecretIntValue")
	intField, _ := structType.FieldByName("intValue")

	assert.True(t, isEnvValueSecret(secretIntField.Tag))
	assert.False(t, isEnvValueSecret(intField.Tag))
}

func TestFillConfigDefaults(t *testing.T) {
	s := testStruct{}
	var once sync.Once
	LoadOnce(&s, false, &once)
	expectedConfig := testStruct{
		IntValue:             1,
		BoolValue:            true,
		StringValue:          "default_value",
		StringValueNoDefault: "",
		StringSliceValue:     []string{"foo", "bar"},
		IntMapValue:          map[string]int32{"foo": 1, "bar": 2},
		SecretIntValue:       3,
	}
	assert.Equal(t, expectedConfig, s)
}

func TestFillConfigFromEnv(t *testing.T) {
	os.Setenv("INT_VAL", "2")
	os.Setenv("BOOL_VAL", "false")
	os.Setenv("STRING_VAL", "foo")
	os.Setenv("NO_DEFAULT_VAL", "bar")
	os.Setenv("STRING_SLICE_VAL", "a,b")
	os.Setenv("INT_MAP_VAL", "c=3,d=4")
	os.Setenv("SECRET_INT_VAL", "5")

	s := testStruct{}
	var once sync.Once
	LoadOnce(&s, false, &once)
	expectedConfig := testStruct{
		IntValue:             2,
		BoolValue:            false,
		StringValue:          "foo",
		StringValueNoDefault: "bar",
		StringSliceValue:     []string{"a", "b"},
		IntMapValue:          map[string]int32{"c": 3, "d": 4},
		SecretIntValue:       5,
	}
	assert.Equal(t, expectedConfig, s)
}

func TestConfigTestMode(t *testing.T) {
	s := testStruct{}
	var once sync.Once
	os.Setenv("STRING_VAL", "foo")
	s.StringValue = "bar"
	//Value of STRING_VAL env variable should be ignored when LoadOnce() is run with test_mode set to true
	LoadOnce(&s, true, &once)
	assert.Equal(t, "bar", s.StringValue)
}

func TestConfigSingleLoad(t *testing.T) {
	s := testStruct{}
	var once sync.Once
	os.Setenv("STRING_VAL", "foo")
	LoadOnce(&s, false, &once)
	os.Setenv("STRING_VAL", "bar")
	LoadOnce(&s, false, &once)
	assert.Equal(t, "foo", s.StringValue)
}
