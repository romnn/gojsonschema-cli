## JSON schema valiation CLI

```bash
go install 'github.com/romnn/gojsonschema-cli/cmd/jsonschema@latest'
```

#### Validate a JSON schema

This is similar to ajv's `compile`, in that we validate the JSON schema itself.
This can be useful when you first want to ensure that your schemas are well-formed.

```bash
jsonschema validate -s ./schemas/my-schema.json

# assuming my-schema.json uses draft 07, this would be equal:
jsonschema validate -s "http://json-schema.org/draft-07/schema#" -v ./schemas/my-schema.json
```

#### Validate a file against a schema

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

### Development

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
