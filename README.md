# JSON schema valiation CLI

![Build status](https://github.com/romnn/gojsonschema-cli/actions/workflows/build.yaml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/romnn/gojsonschema-cli)](https://goreportcard.com/report/github.com/romnn/gojsonschema-cli)

A golang CLI wrapper for [xeipuuv/gojsonschema](https://github.com/xeipuuv/gojsonschema).

##### Features:

- Simple to install (static go binary)
- Validate JSON or YAML files against a JSON schema
- Check JSON schemas for correctness
- Supports local and remote schemas

**Note**: Only schemas up to `draft07` are officially supported.

```bash
go install 'github.com/romnn/gojsonschema-cli/cmd/jsonschema@latest'
```

### Validate a JSON schema

This is similar to ajv's `compile`, in that we validate the JSON schema itself.
This can be useful when you first want to ensure that your schemas are well-formed.

```bash
jsonschema validate -s ./schemas/my-schema.json

# assuming my-schema.json uses draft 07, this would be equal:
jsonschema validate -s "http://json-schema.org/draft-07/schema#" -v ./schemas/my-schema.json
```

### Validate a file against a schema

```bash
jsonschema validate -s ./schemas/my-schema.json -v ./my-data.json
```

Also, you can validate against remote schemas:

```bash
export REMOTE="https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master/master-standalone-strict"
jsonschema validate -s "${REMOTE}/horizontalpodautoscaler.json" -v ./my-autoscaler.json
```

Also, you can validate YAML files, such as helm values files:

```bash
jsonschema validate -s ./values.schema.json -v ./values.prod.yaml
```

## Development

To use the provided tasks in `taskfile.yaml`, install [task](https://taskfile.dev/).

```bash
# view all available tasks
task --list-all
# install development tools
task dependencies:install
```

After setup, you can use the following tasks during development:

```bash
task tidy
task run:race
task run:race -- validate -s ./schema.json -v ./values.json
task build:race
task test
task lint
task format
```

## Acknowledgements

- This CLI uses [xeipuuv/gojsonschema](https://github.com/xeipuuv/gojsonschema) under the hood.
- This CLI intentionallyh behaves very similar to the internal JSON schema validation of [`helm lint`](https://github.com/helm/helm/blob/main/pkg/chartutil/jsonschema.go).
- This CLI has a similar goal as [neilpa/yajsv](https://github.com/neilpa/yajsv), which does not support remote schemas, schema checking, and other features.

## License

The project is licensed under the same license as [xeipuuv/gojsonschema](https://github.com/xeipuuv/gojsonschema).
