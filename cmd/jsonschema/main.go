package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
	k8syaml "sigs.k8s.io/yaml"

	"github.com/fatih/color"
	_ "github.com/joho/godotenv/autoload"
	prettyconsole "github.com/thessem/zap-prettyconsole"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
	"golang.org/x/term"
	errors "golang.org/x/xerrors"
)

const (
	// essentially disable logging for now
	defaultLogLevel = zap.FatalLevel
)

var (
	schemaLocationFlag = cli.StringFlag{
		Name:    "schema",
		Aliases: []string{"s"},
		Usage:   "path to schema",
	}

	valuesLocationFlag = cli.StringFlag{
		Name:    "values",
		Aliases: []string{"v"},
		Usage:   "path to values",
	}

	verboseFlag = cli.BoolFlag{
		Name:  "verbose",
		Usage: "enable verbose output",
	}

	colorFlag = cli.StringFlag{
		Name:  "color",
		Usage: "configure color output [always, auto, never]",
	}

	red = color.New(color.FgRed).SprintFunc()
)

type JSONSchema struct {
	Schema string `json:"$schema,omitempty"`
}

func loadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func loadURL(u *url.URL) ([]byte, error) {
	resp, err := http.Get(u.String())
	if err != nil {
		return []byte{}, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func resolve(location string) ([]byte, error) {
	if u, err := url.Parse(location); err == nil {
		switch u.Scheme {
		case "file":
			// resolve file
			return loadFile(u.Host)
		case "http", "https":
			// resolve url
			return loadURL(u)
		default:
			// ignore
		}
	}
	// resolve file
	return loadFile(location)
}

func validate(_ context.Context, cmd *cli.Command, logger *zap.Logger) error {
	colorPreference := strings.ToLower(cmd.String(colorFlag.Name))
	var color bool
	switch colorPreference {
	case "never":
		color = false
	case "always":
		color = true
	default:
		// default: auto
		color = term.IsTerminal(int(os.Stdout.Fd()))
	}

	verbose := cmd.Bool(verboseFlag.Name)

	schemaLocation := cmd.String(schemaLocationFlag.Name)
	if schemaLocation == "" {
		return fmt.Errorf("missing schema")
	}
	schemaJSON, err := resolve(schemaLocation)
	if err != nil {
		return err
	}

	swap := false
	valuesLocation := cmd.String(valuesLocationFlag.Name)
	if valuesLocation == "" {
		var schema JSONSchema
		if err := json.Unmarshal(schemaJSON, &schema); err != nil {
			return err
		}
		valuesLocation = schema.Schema
		swap = true
	}

	if valuesLocation == "" {
		return fmt.Errorf("missing values")
	}

	valuesJSON, err := resolve(valuesLocation)
	if err != nil {
		return err
	}

	if swap {
		valuesLocation, schemaLocation = schemaLocation, valuesLocation
		valuesJSON, schemaJSON = schemaJSON, valuesJSON
	}

	logger.Debug("values", zap.String("location", valuesLocation))
	logger.Debug("schema", zap.String("location", schemaLocation))

	if verbose {
		fmt.Printf("##### schema:\n%s\n", schemaJSON)
		fmt.Printf("##### values:\n%s\n", valuesJSON)
	}

	// make sure schema is valid JSON
	var schema any
	if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		// try to parse from YAML
		if err := yaml.Unmarshal(schemaJSON, &schema); err == nil {
			// convert to JSON
			var err error
			schemaJSON, err = k8syaml.YAMLToJSON(schemaJSON)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// make sure values is valid JSON
	var values any
	if err := json.Unmarshal(valuesJSON, &values); err != nil {
		// try to parse from YAML
		if err := yaml.Unmarshal(valuesJSON, &values); err == nil {
			// convert to JSON
			var err error
			valuesJSON, err = k8syaml.YAMLToJSON(valuesJSON)
			if err != nil {
				return err
			}
		} else {
			return err
		}
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
				sb.WriteString(fmt.Sprintf("%s: %s\n", red(desc.Field()), desc.Description()))
			} else {
				sb.WriteString(fmt.Sprintf("%s: %s\n", desc.Field(), desc.Description()))
			}
		}
		return errors.New(sb.String())
	}
	return nil
}

func main() {
	var logger *zap.Logger
	if term.IsTerminal(int(os.Stdout.Fd())) {
		logger = prettyconsole.NewLogger(defaultLogLevel)
	} else {
		logger, _ = zap.NewProduction()
	}
	defer func() {
		_ = logger.Sync()
	}()

	app := cli.Command{
		Name:        "jsonschema",
		Description: "jsonschema",
		Usage:       "jsonschema",
		Flags:       []cli.Flag{},
		Commands: []*cli.Command{
			{
				Name:  "validate",
				Usage: "validate",
				Flags: []cli.Flag{
					&schemaLocationFlag,
					&valuesLocationFlag,
					&verboseFlag,
					&colorFlag,
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return validate(ctx, cmd, logger)
				},
			},
		},
	}

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
