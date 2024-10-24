package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"

	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
	// errors "golang.org/x/xerrors"
)

var (
	validateSchemaLocationFlag = cli.StringFlag{
		Name:    "schema",
		Aliases: []string{"s"},
		Usage:   "path to schema",
	}

	validateValuesLocationFlag = cli.StringFlag{
		Name:    "values",
		Aliases: []string{"v"},
		Usage:   "path to values",
	}
)

func validate(_ context.Context, cmd *cli.Command, logger *zap.Logger) error {
	color := isColor(cmd)
	verbose := cmd.Bool(verboseFlag.Name)
	stdin := cmd.Bool(stdinFlag.Name)

	schemaLocation := Location{PathOrUrl: cmd.String(validateSchemaLocationFlag.Name)}
	valuesLocation := Location{PathOrUrl: cmd.String(validateValuesLocationFlag.Name)}

	// choose what to read from stdin
	if stdin {
		if valuesLocation.PathOrUrl == "" && schemaLocation.PathOrUrl == "" {
			// read values and hope they have a $schema definition
			valuesLocation.Stdin = true
		} else if valuesLocation.PathOrUrl == "" {
			valuesLocation.Stdin = true
		} else if schemaLocation.PathOrUrl == "" {
			schemaLocation.Stdin = true
		}
	}

	// sanity check
	if valuesLocation.Stdin && valuesLocation.PathOrUrl != "" {
		return fmt.Errorf(
			"attempt to read values from stdin and %q",
			valuesLocation.PathOrUrl,
		)
	}
	if schemaLocation.Stdin && schemaLocation.PathOrUrl != "" {
		return fmt.Errorf(
			"attempt to read schema from stdin and %q",
			schemaLocation.PathOrUrl,
		)
	}

	if !valuesLocation.Valid() && !schemaLocation.Valid() {
		return fmt.Errorf("missing schema and values to validate")
	} else if !schemaLocation.Valid() {
		// try to infer schema from values file $schema
		var schema JSONSchema
		valuesJSON, _ := resolve(valuesLocation)
		if err := json.Unmarshal(valuesJSON, &schema); err != nil {
			return err
		}
		schemaLocation = Location{PathOrUrl: schema.Schema}
	} else if !valuesLocation.Valid() {
		// use schema as values and use inner schema as schema
		var schema JSONSchema
		schemaJSON, _ := resolve(schemaLocation)
		if err := json.Unmarshal(schemaJSON, &schema); err != nil {
			return err
		}
		valuesLocation = schemaLocation
		schemaLocation = Location{PathOrUrl: schema.Schema}
	}

	logger.Info(
		"validating",
		zap.String("values", valuesLocation.String()),
		zap.String("schema", schemaLocation.String()),
	)

	if !schemaLocation.Valid() {
		return fmt.Errorf("missing schema")
	}
	schemaJSON, err := resolve(schemaLocation)
	if err != nil {
		return err
	}

	if !valuesLocation.Valid() {
		return fmt.Errorf("missing values")
	}
	valuesJSON, err := resolve(valuesLocation)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Printf("##### schema:\n%s\n", schemaJSON)
		fmt.Printf("##### values:\n%s\n", valuesJSON)
	}

	// make sure schema is valid JSON
	schemaJSON, err = toJSON(schemaJSON)
	if err != nil {
		return err
	}

	// make sure values are valid JSON
	valuesJSON, err = toJSON(valuesJSON)
	if err != nil {
		return err
	}

	if bytes.Equal(valuesJSON, []byte("null")) {
		valuesJSON = []byte("{}")
	}

	schemaLoader := gojsonschema.NewBytesLoader(schemaJSON)
	valuesLoader := gojsonschema.NewBytesLoader(valuesJSON)

	result, err := gojsonschema.Validate(schemaLoader, valuesLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		var sb strings.Builder
		for _, desc := range result.Errors() {
			logger.Error(desc.Description(), zap.String("field", desc.Field()))
			if color {
				sb.WriteString(
					fmt.Sprintf("%s: %s\n", red(desc.Field()), desc.Description()),
				)
			} else {
				sb.WriteString(fmt.Sprintf("%s: %s\n", desc.Field(), desc.Description()))
			}
		}
		return fmt.Errorf("%s", sb.String())
		// return errors.New(sb.String())
	} else {
		logger.Info("passed")
	}
	return nil
}
