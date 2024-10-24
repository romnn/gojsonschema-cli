package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
	k8syaml "sigs.k8s.io/yaml"

	"github.com/fatih/color"
	_ "github.com/joho/godotenv/autoload"
	prettyconsole "github.com/thessem/zap-prettyconsole"
	"github.com/urfave/cli/v3"
	"go.uber.org/zap"
	"golang.org/x/term"
)

const (
	defaultLogLevel = zap.InfoLevel
)

var (
	verboseFlag = cli.BoolFlag{
		Name:  "verbose",
		Usage: "enable verbose output",
	}

	colorFlag = cli.StringFlag{
		Name:  "color",
		Usage: "configure color output [always, auto, never]",
	}

	stdinFlag = cli.BoolFlag{
		Name:  "stdin",
		Usage: "read values to be validated from standard input",
	}

	red = color.New(color.FgRed).SprintFunc()
)

type JSONSchema struct {
	Schema string `json:"$schema,omitempty"`
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func readURL(u *url.URL) ([]byte, error) {
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

type Location struct {
	PathOrUrl string
	Stdin     bool
}

func (location Location) String() string {
	if location.Stdin {
		return "STDIN"
	}
	return location.PathOrUrl
}

func (location Location) Valid() bool {
	return location.Stdin || location.PathOrUrl != ""
}

func resolve(location Location) ([]byte, error) {
	if location.Stdin {
		return io.ReadAll(os.Stdin)
	}
	if u, err := url.Parse(location.PathOrUrl); err == nil {
		switch u.Scheme {
		case "file":
			// resolve file
			return readFile(u.Host)
		case "http", "https":
			// resolve url
			return readURL(u)
		default:
			// ignore
		}
	}
	// resolve file
	return readFile(location.PathOrUrl)
}

func toJSON(data []byte) ([]byte, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		// try to parse from YAML
		if err := yaml.Unmarshal(data, &v); err == nil {
			// convert to JSON
			var err error
			data, err = k8syaml.YAMLToJSON(data)
			if err != nil {
				return []byte{}, err
			}
		} else {
			// invalid JSON and invalid YAML
			return []byte{}, err
		}
	}
	// valid JSON
	return data, nil
}

func isColor(cmd *cli.Command) bool {
	colorPreference := strings.ToLower(cmd.String(colorFlag.Name))
	switch colorPreference {
	case "never":
		return false
	case "always":
		return true
	default:
		// default: auto
		return term.IsTerminal(int(os.Stdout.Fd()))
	}
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
					&validateSchemaLocationFlag,
					&validateValuesLocationFlag,
					&stdinFlag,
					&verboseFlag,
					&colorFlag,
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return validate(ctx, cmd, logger)
				},
			},
			{
				Name:  "merge",
				Usage: "merge",
				Flags: []cli.Flag{
					&mergeSchemaLocationsFlag,
					&mergeSchemaOutputPathFlag,
					&mergeStrictFlag,
					&verboseFlag,
					&colorFlag,
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return merge(ctx, cmd, logger)
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
