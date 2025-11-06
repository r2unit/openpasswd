package toml

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type Table map[string]interface{}

func DecodeFile(filePath string, v interface{}) (interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	data := make(map[string]Table)
	currentTable := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentTable = strings.Trim(line, "[]")
			if data[currentTable] == nil {
				data[currentTable] = make(Table)
			}
			continue
		}

		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			value = strings.Trim(value, "\"")

			if currentTable != "" {
				if data[currentTable] == nil {
					data[currentTable] = make(Table)
				}
				data[currentTable][key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if err := mapToStruct(data, v); err != nil {
		return nil, err
	}

	return v, nil
}

func mapToStruct(data map[string]Table, v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("v must be a pointer")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("v must be a pointer to struct")
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		tag := fieldType.Tag.Get("toml")
		if tag == "" {
			tag = strings.ToLower(fieldType.Name)
		}

		if table, ok := data[tag]; ok {
			if field.Kind() == reflect.Struct {
				if err := setStructFields(field, table); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func setStructFields(val reflect.Value, table Table) error {
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		tag := fieldType.Tag.Get("toml")
		if tag == "" {
			continue
		}

		if value, ok := table[tag]; ok {
			if field.CanSet() {
				if str, ok := value.(string); ok {
					field.SetString(str)
				}
			}
		}
	}

	return nil
}
