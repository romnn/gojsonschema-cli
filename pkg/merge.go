package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"dario.cat/mergo"
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astjson"
)

type Properties = *orderedmap.OrderedMap[string, *jsonschema.Schema]

type Property = orderedmap.Pair[string, *jsonschema.Schema]

func NewProp(key string, value *jsonschema.Schema) Property {
	return Property{Key: key, Value: value}
}

func NewProperties(props ...Property) Properties {
	return orderedmap.New[string, *jsonschema.Schema](
		orderedmap.WithInitialData(props...),
	)
}

type strictTransformers struct {
}

func (t strictTransformers) Transformer(
	typ reflect.Type,
) func(dst, src reflect.Value) error {
	// if typ == reflect.TypeOf(jsonschema.Schema{}) {
	// orderedmap.OrderedMap[string, *jsonschema.Schema]

	// fmt.Printf("called transformer for %s\n", typ.Name())
	if typ == reflect.TypeOf(NewProperties()) {
		// props, ok :=
		fmt.Printf("found properties to merge with length=%d\n", 0)

		return func(dst, src reflect.Value) error {
			fmt.Printf("found properties to merge with length=%d\n", 0)
			if dst.CanSet() {
				// isZero := dst.MethodByName("IsZero")
				// result := isZero.Call([]reflect.Value{})
				// if result[0].Bool() {
				// 	dst.Set(src)
				// }
				dst.Set(src)
			}
			return nil
		}
	}
	return nil
}

func MergeSchemasJSON(schemas ...[]byte) (*jsonschema.Schema, error) {
	debug := false
	js := &astjson.JSON{}
	var out bytes.Buffer

	// initial root node
	if err := js.ParseObject([]byte(`{}`)); err != nil {
		return nil, err
	}

	for _, schema := range schemas {
		node, err := js.AppendObject(schema)
		if err != nil {
			return nil, err
		}
		js.MergeNodes(js.RootNode, node)

		if debug {
			out.Reset()
			if err := js.PrintNode(js.Nodes[js.RootNode], &out); err != nil {
				return nil, err
			}
			fmt.Printf("POST MERGE:\n%s\n", out.String())
		}
	}

	out.Reset()
	if err := js.PrintNode(js.Nodes[js.RootNode], &out); err != nil {
		return nil, err
	}
	var mergedSchema jsonschema.Schema
	if err := json.Unmarshal(out.Bytes(), &mergedSchema); err != nil {
		return nil, err
	}
	return &mergedSchema, nil
}

func MergeSchemas(schemas ...jsonschema.Schema) (*jsonschema.Schema, error) {
	var mergedSchema jsonschema.Schema
	for _, schema := range schemas {
		if err := mergo.Merge(
			&mergedSchema,
			schema,
			// mergo.WithOverride,
			mergo.WithTransformers(strictTransformers{}),
		); err != nil {
			return nil, err
		}
	}
	return &mergedSchema, nil
}

func ParseSchema(value []byte) (jsonschema.Schema, error) {
	var schema jsonschema.Schema
	err := json.Unmarshal(value, &schema)
	return schema, err
}
