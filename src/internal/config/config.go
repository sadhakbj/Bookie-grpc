// Package config provides shared application configuration helpers.
package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Load populates a configuration struct from its env tags.
func Load(target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Pointer || targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("configuration target must be a pointer to a struct")
	}

	targetValue = targetValue.Elem()
	targetType := targetValue.Type()
	for fieldIndex := 0; fieldIndex < targetValue.NumField(); fieldIndex++ {
		field := targetType.Field(fieldIndex)
		envName, defaultValue, err := parseEnvTag(field.Tag.Get("env"))
		if err != nil {
			return fmt.Errorf("field %s: %w", field.Name, err)
		}
		if envName == "" {
			continue
		}
		value := os.Getenv(envName)
		if value == "" {
			value = defaultValue
		}
		if targetValue.Field(fieldIndex).Kind() != reflect.String {
			return fmt.Errorf("field %s must be a string", field.Name)
		}
		targetValue.Field(fieldIndex).SetString(value)
	}
	return nil
}

func parseEnvTag(tag string) (string, string, error) {
	parts := strings.Split(tag, ",")
	if len(parts) == 0 || parts[0] == "" {
		return "", "", nil
	}
	defaultValue := ""
	for _, option := range parts[1:] {
		key, value, ok := strings.Cut(option, "=")
		if !ok || key != "default" {
			return "", "", fmt.Errorf("invalid env tag option %q", option)
		}
		defaultValue = value
	}
	return parts[0], defaultValue, nil
}

// Validate validates a configuration struct with the shared application rules.
func Validate(value any) error {
	validate := validator.New()
	if err := validate.RegisterValidation("port", func(field validator.FieldLevel) bool {
		port, err := strconv.Atoi(field.Field().String())
		return err == nil && port >= 1 && port <= 65535
	}); err != nil {
		return err
	}
	return validate.Struct(value)
}
