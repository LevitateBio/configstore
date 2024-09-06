package configstore

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
)

// LoadOnce config from the execution environment
func LoadOnce(c interface{}, testMode bool, once *sync.Once) {
	if testMode {
		zap.L().Info("WARNING: running in test mode, configuration not loaded from env")
	} else {
		once.Do(func() { fillConfig(c) })
	}
}

// Print will pretty print the contents of the configuration object. Any struct values with a 'secret=true' struct
// tag will be obscured if set
func Print(c interface{}) {
	var (
		minWidth int  = 0
		tabWidth int  = 0
		padding  int  = 3
		padChar  byte = ' '
		flags    uint = 0
	)
	writer := tabwriter.NewWriter(os.Stdout, minWidth, tabWidth, padding, padChar, flags)

	fmt.Fprint(writer, "OPTION\tENV VAR\tSETTING\n")

	structType := reflect.ValueOf(c).Elem().Type()
	structValue := reflect.ValueOf(c).Elem()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		var stringValue string
		if isEnvValueSecret(field.Tag) {

			// It is useful to be able to distinguish between an unset password and a set password
			if structValue.Field(i).String() == "" {
				stringValue = ""
			} else {
				stringValue = "********"
			}

		} else {
			switch field.Type.Kind() {
			case reflect.String:
				stringValue = structValue.Field(i).String()
			case reflect.Int32:
				stringValue = strconv.Itoa(int(structValue.Field(i).Int()))
			case reflect.Bool:
				stringValue = strconv.FormatBool(structValue.Field(i).Bool())
			case reflect.Slice:
				stringValue = fmt.Sprintf("%v", structValue.Field(i).Interface().([]string))
			case reflect.Map:
				stringValue = fmt.Sprintf("%v", structValue.Field(i).Interface().(map[string]int32))
			default:
				panic("GetConfig currently only supports string, int32, bool and map")
			}
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\n", field.Name, field.Tag.Get("env"), stringValue)
	}
	writer.Flush()
}

// fillConfig loads the environment
func fillConfig(c interface{}) {
	structType := reflect.ValueOf(c).Elem().Type()
	structValue := reflect.ValueOf(c).Elem()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		switch field.Type.Kind() {
		case reflect.String:
			structValue.Field(i).SetString(getEnvValueString(field.Tag))
		case reflect.Int32:
			structValue.Field(i).SetInt(getEnvValueInt(field.Tag))
		case reflect.Bool:
			structValue.Field(i).SetBool(getEnvValueBool(field.Tag))
		case reflect.Slice:
			structValue.Field(i).Set(reflect.ValueOf(getEnvValueStrings(field.Tag)))
		case reflect.Map:
			structValue.Field(i).Set(reflect.ValueOf(getEnvValueIntMap(field.Tag)))
		default:
			panic("GetConfig currently only supports string, string slice, int32, bool and map")
		}
	}
}

func getEnvValueString(fieldTag reflect.StructTag) string {

	envVar := fieldTag.Get("env")
	defaultValue := fieldTag.Get("default")
	var value string
	value, ok := os.LookupEnv(envVar)
	if !ok {
		value = defaultValue
	}
	return value
}

// isEnvValueSecret returns true if the struct has a tag "secret=true". The value is not case sensitive
func isEnvValueSecret(fieldTag reflect.StructTag) bool {
	return strings.ToLower(fieldTag.Get("secret")) == "true"
}

func getEnvValueStrings(fieldTag reflect.StructTag) []string {
	stringValue := getEnvValueString(fieldTag)
	if stringValue == "" {
		return []string{}
	} else {
		return strings.Split(stringValue, ",")
	}
}

// This method panics if it encounters parsing errors
func getEnvValueBool(fieldTag reflect.StructTag) bool {
	valueString := getEnvValueString(fieldTag)
	result, err := strconv.ParseBool(valueString)
	if err != nil {
		panic(fmt.Sprintf("value for %s could not be parsed as a bool", fieldTag.Get("env")))
	}
	return result
}

// This method panics if it encounters parsing errors
func getEnvValueInt(fieldTag reflect.StructTag) int64 {
	valueString := getEnvValueString(fieldTag)
	result, err := strconv.Atoi(valueString)
	if err != nil {
		panic(fmt.Sprintf("value for %s could not be parsed as an int32", fieldTag.Get("env")))
	}
	return int64(result)
}

func getEnvValueIntMap(fieldTag reflect.StructTag) map[string]int32 {
	valueStrings := getEnvValueStrings(fieldTag)
	valueMap := map[string]int32{}
	for _, entryString := range valueStrings {
		pair := strings.Split(entryString, "=")

		key := pair[0]
		value, err := strconv.Atoi(pair[1])
		if err != nil {
			panic(fmt.Sprintf("Value for %s could not be parsed into a map[string]int32", fieldTag.Get("env")))
		}
		valueMap[key] = int32(value)
	}
	return valueMap
}
