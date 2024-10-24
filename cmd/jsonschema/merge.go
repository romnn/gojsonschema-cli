package main

import (
	// "bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	// "io"
	// "net/http"
	// "net/url"
	// "os"
	// "strings"

	"dario.cat/mergo"
	"github.com/invopop/jsonschema"

	// "github.com/xeipuuv/gojsonschema"
	// "gopkg.in/yaml.v3"
	// k8syaml "sigs.k8s.io/yaml"

	// "github.com/fatih/color"
	_ "github.com/joho/godotenv/autoload"
	// prettyconsole "github.com/thessem/zap-prettyconsole"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
	// "golang.org/x/term"
	// errors "golang.org/x/xerrors"
)

var (
	mergeStrictFlag = cli.BoolFlag{
		Name:  "strict",
		Usage: "merge strict",
	}

	mergeSchemaLocationsFlag = cli.StringSliceFlag{
		Name:    "schema",
		Aliases: []string{"s"},
		Usage:   "path to schema",
	}

	mergeSchemaOutputPathFlag = cli.StringFlag{
		Name:    "out",
		Aliases: []string{"o"},
		Usage:   "output path to save the merged schema to",
	}
)

func merge(_ context.Context, cmd *cli.Command, logger *zap.Logger) error {
	verbose := cmd.Bool(verboseFlag.Name)
	schemaLocations := cmd.StringSlice(mergeSchemaLocationsFlag.Name)
	mergedSchemaOutputPath := cmd.String(mergeSchemaOutputPathFlag.Name)

	// merge schemas in order
	var mergedSchema jsonschema.Schema
	for i, schemaLocation := range schemaLocations {
		var schema jsonschema.Schema
		schemaJSON, _ := resolve(Location{PathOrUrl: schemaLocation})
		if err := json.Unmarshal(schemaJSON, &schema); err != nil {
			return err
		}

		if verbose {
			fmt.Printf("##### schema[%d]:\n%s\n", i, schemaJSON)
		}

		var transformers mergo.Transformers
		if err := mergo.Merge(
			&mergedSchema,
			schema,
			mergo.WithTransformers(transformers),
		); err != nil {
			return err
		}
	}

	// serialize back to pretty json and print
	mergedSchemaJSON, err := json.MarshalIndent(mergedSchema, "", "    ")
	if err != nil {
		return err
	}

	if mergedSchemaOutputPath != "" {
		// save to output
		if err := os.WriteFile(mergedSchemaOutputPath, mergedSchemaJSON, 0644); err != nil {
			return err
		}
		logger.Info(
			"wrote merged schema",
			zap.String("destination", mergedSchemaOutputPath),
		)
	} else {
		fmt.Println(string(mergedSchemaJSON))
	}
	return nil
}
