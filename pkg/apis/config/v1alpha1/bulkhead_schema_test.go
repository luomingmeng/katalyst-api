/*
Copyright 2022 The Katalyst Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"encoding/json"
	"os"
	"testing"

	"sigs.k8s.io/yaml"
)

func TestBulkheadRDTConfigSchema(t *testing.T) {
	schema := loadBulkheadRDTConfigSchema(t)

	t.Run("default CAT ways contract", func(t *testing.T) {
		defaultCATWays := schemaProperty(t, schema, "defaultCATWays")
		assertIntOrStringSchema(t, defaultCATWays, "defaultCATWays")
		assertValidationRule(t, defaultCATWays.XKubernetesValidations, validationRule{
			Rule:    "type(self) == int ? self > 0 : self.matches('^\\s*(CBMMask|MinCBMBits|[1-9][0-9]*)(\\s*[+-]\\s*(CBMMask|MinCBMBits|[1-9][0-9]*))?\\s*$')",
			Message: "defaultCATWays must be a positive integer or a valid CAT ways expression",
		})
		assertValidationRule(t, defaultCATWays.XKubernetesValidations, validationRule{
			Rule:    "type(self) == int || !self.matches('^\\s*(CBMMask\\s*-\\s*CBMMask|MinCBMBits\\s*-\\s*(MinCBMBits|CBMMask)|[1-9][0-9]*\\s*[+-]\\s*[1-9][0-9]*)\\s*$')",
			Message: "defaultCATWays expression is statically non-positive or contains unsimplified literal arithmetic",
		})
	})

	t.Run("EnableCAT conditional contract", func(t *testing.T) {
		assertValidationRule(t, schema.XKubernetesValidations, validationRule{
			Rule:    "!has(self.enableCAT) || !self.enableCAT || has(self.defaultCATWays)",
			Message: "defaultCATWays must be specified when enableCAT is true",
		})

		enabled, disabled := true, false
		assertEnableCATContract(t, []enableCATContractCase{
			{name: "invalid: CAT enabled without default ways", enableCAT: &enabled, valid: false},
			{name: "valid: CAT enabled with positive ways", enableCAT: &enabled, hasDefaultCATWays: true, valid: true},
			{name: "valid: CAT disabled without default ways", enableCAT: &disabled, valid: true},
		})
	})

	t.Run("CLOS CAT ways contract", func(t *testing.T) {
		closCATWays := schemaProperty(t, schema, "closCATWays")
		if closCATWays.AdditionalProperties == nil {
			t.Fatal("closCATWays has no additionalProperties schema")
		}
		assertIntOrStringSchema(t, *closCATWays.AdditionalProperties, "closCATWays values")
		assertValidationRule(t, closCATWays.XKubernetesValidations, validationRule{
			Rule:    "self.all(k, type(self[k]) == int ? self[k] > 0 : self[k].matches('^\\s*(CBMMask|MinCBMBits|[1-9][0-9]*)(\\s*[+-]\\s*(CBMMask|MinCBMBits|[1-9][0-9]*))?\\s*$'))",
			Message: "all CLOS CAT ways must be positive integers or valid CAT ways expressions",
		})
		assertValidationRule(t, closCATWays.XKubernetesValidations, validationRule{
			Rule:    "self.all(k, type(self[k]) == int || !self[k].matches('^\\s*(CBMMask\\s*-\\s*CBMMask|MinCBMBits\\s*-\\s*(MinCBMBits|CBMMask)|[1-9][0-9]*\\s*[+-]\\s*[1-9][0-9]*)\\s*$'))",
			Message: "CLOS CAT ways expressions must not be statically non-positive or contain unsimplified literal arithmetic",
		})
		assertValidationRule(t, closCATWays.XKubernetesValidations, validationRule{
			Rule:    "self.all(k, k.matches('^\\S+$'))",
			Message: "CLOS CAT ways keys must not be empty or contain whitespace",
		})
	})

	t.Run("CAT policy schema", func(t *testing.T) {
		catPolicy := schemaProperty(t, schema, "catPolicy")

		defaultPlacement := schemaProperty(t, catPolicy, "defaultPlacement")
		assertEnumSchema(t, schemaProperty(t, defaultPlacement, "direction"), []string{"low", "high"}, "defaultPlacement.direction")
		allowedBitUsages := schemaProperty(t, defaultPlacement, "allowedBitUsages")
		if allowedBitUsages.Items == nil {
			t.Fatal("defaultPlacement.allowedBitUsages has no item schema")
		}
		assertEnumSchema(t, *allowedBitUsages.Items, []string{"S", "H", "X"}, "defaultPlacement.allowedBitUsages items")

		allocationGroups := schemaProperty(t, catPolicy, "allocationGroups")
		if allocationGroups.Items == nil {
			t.Fatal("allocationGroups has no item schema")
		}
		allocationGroup := *allocationGroups.Items
		closIDs := schemaProperty(t, allocationGroup, "closIDs")
		if closIDs.MinItems == nil || *closIDs.MinItems != 1 {
			t.Fatalf("allocationGroups[].closIDs minItems = %v, want 1", closIDs.MinItems)
		}
		assertRequiredSchema(t, allocationGroup, "closIDs", "allocationGroups[]")
		assertValidationRule(t, allocationGroup.XKubernetesValidations, validationRule{
			Rule:    "self.closIDs.all(id, id.matches('^\\S+$'))",
			Message: "allocation group CLOS IDs must not be empty or contain whitespace",
		})
		assertValidationRule(t, allocationGroup.XKubernetesValidations, validationRule{
			Rule:    "self.closIDs.all(id, self.closIDs.exists_one(other, other == id))",
			Message: "allocation group CLOS IDs must be unique",
		})
		if _, ok := allocationGroup.Properties["overlap"]; ok {
			t.Fatal("allocationGroups[].overlap must not be exposed")
		}

		closPlacements := schemaProperty(t, catPolicy, "closPlacements")
		assertValidationRule(t, closPlacements.XKubernetesValidations, validationRule{
			Rule:    "self.all(k, k.matches('^\\S+$'))",
			Message: "CLOS placement keys must not be empty or contain whitespace",
		})
	})
}

type customResourceDefinition struct {
	Spec struct {
		Versions []struct {
			Schema struct {
				OpenAPIV3Schema jsonSchema `json:"openAPIV3Schema"`
			} `json:"schema"`
		} `json:"versions"`
	} `json:"spec"`
}

type jsonSchema struct {
	Properties             map[string]jsonSchema `json:"properties"`
	AdditionalProperties   *jsonSchema           `json:"additionalProperties"`
	AnyOf                  []jsonSchema          `json:"anyOf"`
	Enum                   []string              `json:"enum"`
	Items                  *jsonSchema           `json:"items"`
	MinItems               *int                  `json:"minItems"`
	Required               []string              `json:"required"`
	Type                   string                `json:"type"`
	XIntOrString           bool                  `json:"x-kubernetes-int-or-string"`
	XKubernetesValidations []validationRule      `json:"x-kubernetes-validations"`
}

type validationRule struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type enableCATContractCase struct {
	name              string
	enableCAT         *bool
	hasDefaultCATWays bool
	valid             bool
}

func loadBulkheadRDTConfigSchema(t *testing.T) jsonSchema {
	t.Helper()

	data, err := os.ReadFile("../../../../config/crd/bases/config.katalyst.kubewharf.io_adminqosconfigurations.yaml")
	if err != nil {
		t.Fatalf("read AdminQoSConfiguration CRD: %v", err)
	}

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		t.Fatalf("convert AdminQoSConfiguration CRD to JSON: %v", err)
	}

	var crd customResourceDefinition
	if err := json.Unmarshal(jsonData, &crd); err != nil {
		t.Fatalf("unmarshal AdminQoSConfiguration CRD: %v", err)
	}
	if len(crd.Spec.Versions) == 0 {
		t.Fatal("AdminQoSConfiguration CRD has no versions")
	}

	return findSchemaProperty(t, crd.Spec.Versions[0].Schema.OpenAPIV3Schema, "bulkheadRDTConfig")
}

func findSchemaProperty(t *testing.T, schema jsonSchema, name string) jsonSchema {
	t.Helper()
	if property, ok := schema.Properties[name]; ok {
		return property
	}
	for _, property := range schema.Properties {
		if found := findSchemaPropertyOrNil(property, name); found != nil {
			return *found
		}
	}
	t.Fatalf("schema property %q not found", name)
	return jsonSchema{}
}

func findSchemaPropertyOrNil(schema jsonSchema, name string) *jsonSchema {
	if property, ok := schema.Properties[name]; ok {
		return &property
	}
	for _, property := range schema.Properties {
		if found := findSchemaPropertyOrNil(property, name); found != nil {
			return found
		}
	}
	return nil
}

func schemaProperty(t *testing.T, schema jsonSchema, name string) jsonSchema {
	t.Helper()
	property, ok := schema.Properties[name]
	if !ok {
		t.Fatalf("schema property %q not found", name)
	}
	return property
}

func assertValidationRule(t *testing.T, rules []validationRule, want validationRule) {
	t.Helper()
	for _, rule := range rules {
		if rule == want {
			return
		}
	}
	t.Fatalf("validation rules = %#v, want rule %#v", rules, want)
}

func assertIntOrStringSchema(t *testing.T, schema jsonSchema, name string) {
	t.Helper()
	if !schema.XIntOrString {
		t.Fatalf("%s x-kubernetes-int-or-string = false, want true", name)
	}
	types := map[string]bool{}
	for _, candidate := range schema.AnyOf {
		types[candidate.Type] = true
	}
	if !types["integer"] || !types["string"] {
		t.Fatalf("%s anyOf types = %#v, want integer and string", name, types)
	}
}

func assertEnumSchema(t *testing.T, schema jsonSchema, want []string, name string) {
	t.Helper()
	if len(schema.Enum) != len(want) {
		t.Fatalf("%s enum = %#v, want exactly %#v", name, schema.Enum, want)
	}
	got := map[string]bool{}
	for _, value := range schema.Enum {
		got[value] = true
	}
	for _, value := range want {
		if !got[value] {
			t.Fatalf("%s enum = %#v, want value %s", name, schema.Enum, value)
		}
	}
}

func assertRequiredSchema(t *testing.T, schema jsonSchema, field string, name string) {
	t.Helper()
	for _, required := range schema.Required {
		if required == field {
			return
		}
	}
	t.Fatalf("%s required = %#v, want %s", name, schema.Required, field)
}

func assertEnableCATContract(t *testing.T, cases []enableCATContractCase) {
	t.Helper()
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.enableCAT == nil || !*tt.enableCAT || tt.hasDefaultCATWays
			if got != tt.valid {
				t.Fatalf("EnableCAT contract validity = %v, want %v", got, tt.valid)
			}
		})
	}
}
