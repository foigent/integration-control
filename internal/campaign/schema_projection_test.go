package campaign

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestCommittedSchemasMatchDeterministicProjection(t *testing.T) {
	directory := filepath.Join("..", "..", "contracts", "v1")
	if err := CheckProjectedSchemas(directory); err != nil {
		t.Fatal(err)
	}
	projected, err := ProjectedSchemas()
	if err != nil {
		t.Fatal(err)
	}
	if len(projected) != 5 {
		t.Fatalf("expected exactly five projected schemas, got %d", len(projected))
	}
	for name, body := range projected {
		if err := VerifyProjectedSchemaDocument(body); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
}

func TestTopLevelTypedFieldsMatchProjectedSchemas(t *testing.T) {
	projected, err := ProjectedSchemas()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		typeOf reflect.Type
	}{
		{ManifestSchemaFile, reflect.TypeOf(Manifest{})},
		{AuthorizationSchemaFile, reflect.TypeOf(Authorization{})},
		{SignatureSchemaFile, reflect.TypeOf(AuthorizationSignature{})},
		{TrustSetSchemaFile, reflect.TypeOf(AuthorizationTrustSet{})},
		{ReleaseSchemaFile, reflect.TypeOf(ReleaseManifest{})},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var schema map[string]any
			if err := json.Unmarshal(projected[testCase.name], &schema); err != nil {
				t.Fatal(err)
			}
			typedFields := jsonFields(testCase.typeOf)
			projectedFields := mapKeys(schema["properties"])
			required := stringItems(schema["required"])
			if !reflect.DeepEqual(typedFields, projectedFields) || !reflect.DeepEqual(typedFields, required) {
				t.Fatalf("typed/schema field drift: typed=%v projected=%v required=%v", typedFields, projectedFields, required)
			}
		})
	}
}

func TestAuthorizationSchemaOwnsCommonAcknowledgementEnvelope(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", "contracts", "v1", AuthorizationSchemaFile))
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(body, &schema); err != nil {
		t.Fatal(err)
	}
	definitions, ok := schema["$defs"].(map[string]any)
	if !ok {
		t.Fatal("authorization schema has no definitions")
	}
	ack, ok := definitions["component_acknowledgement"].(map[string]any)
	if !ok {
		t.Fatal("authorization schema does not define the common acknowledgement envelope")
	}
	projected := mapKeys(ack["properties"])
	required := stringItems(ack["required"])
	typed := jsonFields(reflect.TypeOf(ComponentAcknowledgement{}))
	if !reflect.DeepEqual(typed, projected) || !reflect.DeepEqual(typed, required) {
		t.Fatalf("acknowledgement typed/schema field drift: typed=%v projected=%v required=%v", typed, projected, required)
	}
}

func TestManifestProjectionPinsFixedPOCContractIdentities(t *testing.T) {
	projected, err := ProjectedSchemas()
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(projected[ManifestSchemaFile], &schema); err != nil {
		t.Fatal(err)
	}
	properties := schema["properties"].(map[string]any)
	source := properties["source_binding"].(map[string]any)["properties"].(map[string]any)
	requestSet := properties["request_set"].(map[string]any)["properties"].(map[string]any)
	components := properties["components"].(map[string]any)["prefixItems"].([]any)
	uniComponent := components[0].(map[string]any)["properties"].(map[string]any)
	uniContract := uniComponent["contract"].(map[string]any)["properties"].(map[string]any)
	if source["schema_version"].(map[string]any)["const"] != SourceBindingSchemaVersion {
		t.Fatal("projected manifest widens the Source binding identity")
	}
	if requestSet["schema_version"].(map[string]any)["const"] != RequestSetSchemaVersion {
		t.Fatal("projected manifest widens the request-set identity")
	}
	if uniContract["id"].(map[string]any)["const"] != UniRouterComponentContractID {
		t.Fatal("projected manifest widens the UniRouter component contract identity")
	}
}

func TestNestedTypedFieldsMatchProjectedSchemas(t *testing.T) {
	projected, err := ProjectedSchemas()
	if err != nil {
		t.Fatal(err)
	}
	type pathPart struct {
		key   string
		index int
		array bool
	}
	key := func(name string) pathPart { return pathPart{key: name} }
	index := func(value int) pathPart { return pathPart{index: value, array: true} }
	cases := []struct {
		name   string
		schema string
		path   []pathPart
		typeOf reflect.Type
	}{
		{"manifest artifact reference", ManifestSchemaFile, []pathPart{key("$defs"), key("artifact_reference")}, reflect.TypeOf(ArtifactReference{})},
		{"manifest contract reference", ManifestSchemaFile, []pathPart{key("$defs"), key("contract_reference")}, reflect.TypeOf(ContractReference{})},
		{"manifest window", ManifestSchemaFile, []pathPart{key("$defs"), key("authorization_window")}, reflect.TypeOf(AuthorizationWindow{})},
		{"manifest request set", ManifestSchemaFile, []pathPart{key("properties"), key("request_set")}, reflect.TypeOf(RequestSetReference{})},
		{"manifest component", ManifestSchemaFile, []pathPart{key("properties"), key("components"), key("prefixItems"), index(0)}, reflect.TypeOf(ComponentInput{})},
		{"manifest source signing", ManifestSchemaFile, []pathPart{key("properties"), key("source_evidence_signing")}, reflect.TypeOf(SourceEvidenceSigning{})},
		{"manifest authorization signing", ManifestSchemaFile, []pathPart{key("properties"), key("authorization_signing")}, reflect.TypeOf(AuthorizationSigning{})},
		{"authorization ack reference", AuthorizationSchemaFile, []pathPart{key("$defs"), key("acknowledgement_reference")}, reflect.TypeOf(ComponentAcknowledgementReference{})},
		{"authorization acknowledgement", AuthorizationSchemaFile, []pathPart{key("$defs"), key("component_acknowledgement")}, reflect.TypeOf(ComponentAcknowledgement{})},
		{"trust-set key", TrustSetSchemaFile, []pathPart{key("properties"), key("keys"), key("items")}, reflect.TypeOf(AuthorizationTrustKey{})},
		{"release artifact", ReleaseSchemaFile, []pathPart{key("properties"), key("artifacts"), key("items")}, reflect.TypeOf(ReleaseArtifact{})},
		{"release gate", ReleaseSchemaFile, []pathPart{key("properties"), key("gates"), key("items")}, reflect.TypeOf(ReleaseGate{})},
		{"release effects", ReleaseSchemaFile, []pathPart{key("properties"), key("effects")}, reflect.TypeOf(ReleaseEffects{})},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var root any
			if err := json.Unmarshal(projected[testCase.schema], &root); err != nil {
				t.Fatal(err)
			}
			node := root
			for _, part := range testCase.path {
				if part.array {
					values, ok := node.([]any)
					if !ok || part.index >= len(values) {
						t.Fatalf("schema path array step is missing: %#v", part)
					}
					node = values[part.index]
					continue
				}
				mapping, ok := node.(map[string]any)
				if !ok {
					t.Fatalf("schema path object step is missing: %s", part.key)
				}
				node, ok = mapping[part.key]
				if !ok {
					t.Fatalf("schema path key is missing: %s", part.key)
				}
			}
			object, ok := node.(map[string]any)
			if !ok {
				t.Fatal("projected schema node is not an object")
			}
			typed := jsonFields(testCase.typeOf)
			properties := mapKeys(object["properties"])
			required := stringItems(object["required"])
			if !reflect.DeepEqual(typed, properties) || !reflect.DeepEqual(typed, required) {
				t.Fatalf("nested typed/schema drift: typed=%v projected=%v required=%v", typed, properties, required)
			}
		})
	}
}

func jsonFields(value reflect.Type) []string {
	fields := make([]string, 0, value.NumField())
	for index := 0; index < value.NumField(); index++ {
		fields = append(fields, value.Field(index).Tag.Get("json"))
	}
	sort.Strings(fields)
	return fields
}

func mapKeys(value any) []string {
	mapping, _ := value.(map[string]any)
	keys := make([]string, 0, len(mapping))
	for key := range mapping {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func stringItems(value any) []string {
	items, _ := value.([]any)
	strings := make([]string, 0, len(items))
	for _, item := range items {
		strings = append(strings, item.(string))
	}
	sort.Strings(strings)
	return strings
}
