package main

import (
	"context"
	// "encoding/json"
	"fmt"
	"os"

	// "github.com/invopop/jsonschema"
	"github.com/romnn/gojsonschema-cli/pkg"

	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
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
	// var schemas []jsonschema.Schema
	var schemas [][]byte
	for i, schemaLocation := range schemaLocations {
		schemaJSON, err := resolve(Location{PathOrUrl: schemaLocation})
		if err != nil {
			return err
		}
		if verbose {
			fmt.Printf("##### schema[%d]:\n%s\n", i, schemaJSON)
		}
		schemas = append(schemas, schemaJSON)
		// schema, err := pkg.ParseSchema(schemaJSON)
		// if err != nil {
		// 	return err
		// }
		// schemas = append(schemas, schema)
	}

	logger.Info("merging", zap.Int("schemas", len(schemas)))

	// mergedSchema, err := pkg.MergeSchemas(schemas...)
	mergedSchemaJSON, err := pkg.MergeSchemasJSON(verbose, schemas...)
	if err != nil {
		return err
	}

	// // serialize back to pretty json and print
	// mergedSchemaJSON, err := json.MarshalIndent(mergedSchema, "", "    ")
	// if err != nil {
	// 	return err
	// }

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
