package pkg

import (
	"encoding/json"
	"testing"

	// "github.com/go-test/deep"
	"github.com/google/go-cmp/cmp"
	// "github.com/google/go-cmp/cmp/cmpopts"
	"github.com/invopop/jsonschema"
	// "github.com/romnn/deepequal"
	"github.com/stretchr/testify/assert"
	// orderedmap "github.com/wk8/go-ordered-map/v2"
)

func DeepEqual(
	t *testing.T,
	have *jsonschema.Schema,
	want *jsonschema.Schema,
) (bool, string) {
	opts := cmp.Options{
		// cmp.Comparer(func(a, b Properties) bool {
		// 	// return (math.IsNaN(x) && math.IsNaN(y)) || x == y
		// }),
		// cmpopts.IgnoreUnexported(),
		// protocmp.Transform(),
	}
	haveJSON := marshalSchema(t, have)
	wantJSON := marshalSchema(t, want)

	var haveCmp, wantCmp interface{}
	if err := json.Unmarshal(haveJSON, &haveCmp); err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if err := json.Unmarshal(wantJSON, &wantCmp); err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	equal := cmp.Equal(wantCmp, haveCmp, opts)
	diff := cmp.Diff(wantCmp, haveCmp, opts)
	return equal, diff
}

func marshalSchema(t *testing.T, schema *jsonschema.Schema) []byte {
	schemaJSON, err := json.MarshalIndent(schema, "", "    ")
	if err != nil {
		t.Errorf("failed to marshal schema: %v", err)
	}
	return schemaJSON
}

// func mergeJSONSchemas(t *testing.T, rawSchemas ...[]byte) jsonschema.Schema {
// 	var schemas []jsonschema.Schema
// 	for _, schema := range rawSchemas {
// 		schema, err := ParseSchema(schema)
// 		if err != nil {
// 			t.Fatalf("failed to parse schema: %v", err)
// 		}
// 		schemas = append(schemas, schema)
// 	}
// 	mergedSchemas, err := MergeSchemas(schemas...)
// 	if err != nil {
// 		t.Fatalf("failed to merge schemas: %v", err)
// 	}
// 	return mergedSchemas
// }

func Test(t *testing.T) {
	assert.Equal(t, "a", "a")
}

func TestSimpleMerge(t *testing.T) {
	t.Parallel()
	schema1 := jsonschema.Schema{
		ID:                   "https://luup.io/chart/signoz/values.schema.json",
		Version:              "http://json-schema.org/draft-07/schema#",
		AdditionalProperties: jsonschema.FalseSchema,
		Description:          "Signoz schema",
		Properties: NewProperties(
			NewProp(
				"appVersions",
				&jsonschema.Schema{Ref: "file://../../schemas/versions.json"},
			),
			NewProp("signoz", &jsonschema.Schema{
				Properties: NewProperties(
					NewProp("frontend", &jsonschema.Schema{
						Properties: NewProperties(
							NewProp("name", &jsonschema.Schema{Type: "string"}),
							NewProp(
								"replicaCount",
								&jsonschema.Schema{Type: "integer"},
							),
						),
						Type: "object",
					}),
				),
				Type: "object",
			}),
		),
		Type: "object",
	}

	resourcesProperties := NewProperties(
		NewProp("limits", &jsonschema.Schema{
			Properties: NewProperties(
				NewProp("cpu", &jsonschema.Schema{Type: "string"}),
				NewProp("memory", &jsonschema.Schema{Type: "string"}),
				NewProp("ephemeral-storage", &jsonschema.Schema{Type: "string"}),
			),
		}),
		NewProp("requests", &jsonschema.Schema{
			Properties: NewProperties(
				NewProp("cpu", &jsonschema.Schema{Type: "string"}),
				NewProp("memory", &jsonschema.Schema{Type: "string"}),
				NewProp("ephemeral-storage", &jsonschema.Schema{Type: "string"}),
			),
		}),
	)

	schema2 := jsonschema.Schema{
		ID:                   "https://luup.io/chart/signoz/values.schema.json",
		Version:              "http://json-schema.org/draft-07/schema#",
		AdditionalProperties: jsonschema.FalseSchema,
		Description:          "Signoz schema",
		Properties: NewProperties(
			// app versions is missing
			NewProp("signoz", &jsonschema.Schema{
				Properties: NewProperties(
					NewProp("frontend", &jsonschema.Schema{
						Properties: NewProperties(
							NewProp("resources", &jsonschema.Schema{
								Ref:        "file://../../schemas/resources.json",
								Properties: resourcesProperties,
							}),
						),
						Type: "object",
					}),
				),
				Type: "object",
			}),
		),
		Type: "object",
	}

	expected := jsonschema.Schema{
		ID:                   "https://luup.io/chart/signoz/values.schema.json",
		Version:              "http://json-schema.org/draft-07/schema#",
		AdditionalProperties: jsonschema.FalseSchema,
		Description:          "Signoz schema",
		Properties: NewProperties(
			NewProp(
				"appVersions",
				&jsonschema.Schema{Ref: "file://../../schemas/versions.json"},
			),
			NewProp("signoz", &jsonschema.Schema{
				Properties: NewProperties(
					NewProp("frontend", &jsonschema.Schema{
						Properties: NewProperties(
							NewProp("name", &jsonschema.Schema{Type: "string"}),
							NewProp(
								"replicaCount",
								&jsonschema.Schema{Type: "integer"},
							),
							NewProp("resources", &jsonschema.Schema{
								Ref:        "file://../../schemas/resources.json",
								Properties: resourcesProperties,
							}),
						),
						Type: "object",
					}),
				),
				Type: "object",
			}),
		),
		Type: "object",
	}

	// merge using golang native types
	// merged, err := MergeSchemas(schema1, schema2)
	// if err != nil {
	// 	t.Fatalf("failed to merge schemas: %v", err)
	// }

	schema1JSON := marshalSchema(t, &schema1)
	schema2JSON := marshalSchema(t, &schema2)
	merged, err := MergeSchemasJSON(schema1JSON, schema2JSON)
	if err != nil {
		t.Fatalf("failed to merge schemas: %v", err)
	}

	t.Logf("expected: %s\n", string(marshalSchema(t, &expected)))
	t.Logf("merged: %s\n", string(marshalSchema(t, merged)))

	if equal, diff := DeepEqual(t, merged, &expected); !equal {
		t.Errorf("%s", diff)
	}
	// assert.True(t, false, "TODO")
}
