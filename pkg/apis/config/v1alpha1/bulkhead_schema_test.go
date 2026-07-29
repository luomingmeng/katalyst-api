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
		minimum := schemaProperty(t, schema, "defaultCATWays").Minimum
		if minimum == nil || *minimum != 1 {
			t.Fatalf("defaultCATWays minimum = %v, want 1", minimum)
		}

		assertMinimumContract(t, *minimum, []minimumContractCase{
			{name: "invalid: zero default CAT ways", value: 0, valid: false},
			{name: "valid: positive default ways without explicit enablement", value: 1, valid: true},
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
		assertValidationRule(t, closCATWays.XKubernetesValidations, validationRule{
			Rule:    "self.all(k, self[k] > 0)",
			Message: "all CLOS CAT ways must be greater than 0",
		})

		assertClosCATWaysContract(t, []closCATWaysContractCase{
			{name: "invalid: zero CLOS CAT ways", values: []int64{0}, valid: false},
			{name: "invalid: negative CLOS CAT ways", values: []int64{-1}, valid: false},
			{name: "valid: positive CLOS CAT ways", values: []int64{1, 2}, valid: true},
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
	Minimum                *float64              `json:"minimum"`
	XKubernetesValidations []validationRule      `json:"x-kubernetes-validations"`
}

type validationRule struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type minimumContractCase struct {
	name  string
	value float64
	valid bool
}

type enableCATContractCase struct {
	name              string
	enableCAT         *bool
	hasDefaultCATWays bool
	valid             bool
}

type closCATWaysContractCase struct {
	name   string
	values []int64
	valid  bool
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

func assertMinimumContract(t *testing.T, minimum float64, cases []minimumContractCase) {
	t.Helper()
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value >= minimum; got != tt.valid {
				t.Fatalf("value %v validity = %v, want %v for minimum %v", tt.value, got, tt.valid, minimum)
			}
		})
	}
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

func assertClosCATWaysContract(t *testing.T, cases []closCATWaysContractCase) {
	t.Helper()
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := true
			for _, value := range tt.values {
				if value <= 0 {
					got = false
					break
				}
			}
			if got != tt.valid {
				t.Fatalf("CLOS CAT ways contract validity = %v, want %v", got, tt.valid)
			}
		})
	}
}
