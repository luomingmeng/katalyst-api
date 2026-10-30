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

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"
)

func TestQRMPluginConfigRDTAndBulkheadRDTConfig(t *testing.T) {
	disableRDT := true
	enableCPUList := true
	enableCAT := true
	defaultCATWays := intstr.FromString("MaxCATWays")
	exclusiveClosIDs := []string{"clos-a", "peer-b"}
	config := QRMPluginConfig{
		RDTConfig: &RDTConfig{
			DisableRDT: &disableRDT,
		},
		CPUPluginConfig: &CPUPluginConfig{
			BulkheadConfig: &BulkheadConfig{
				BulkheadRDTConfig: &BulkheadRDTConfig{
					EnableCPUList:  &enableCPUList,
					EnableCAT:      &enableCAT,
					DefaultCATWays: &defaultCATWays,
					ClosCATWays: map[string]intstr.IntOrString{
						"share-00": intstr.FromString("MaxCATWays-MinCATWays"),
						"share-01": intstr.FromInt(2),
					},
					CATPolicy: &CATPolicy{
						DefaultPlacement: &CATPlacementPolicy{
							AllowedBitUsages: []CATBitUsage{CATBitUsageAll},
							Direction:        CATAllocationDirectionLow,
						},
						ClosPlacements: map[string]CATPlacementPolicy{
							"share-01": {
								AllowedBitUsages: []CATBitUsage{CATBitUsageSoftware, CATBitUsageExclusive},
								Direction:        CATAllocationDirectionHigh,
							},
						},
						ExclusiveClosIDs: &exclusiveClosIDs,
					},
				},
			},
		},
	}

	raw, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}

	var got QRMPluginConfig
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if got.RDTConfig == nil || got.RDTConfig.DisableRDT == nil || !*got.RDTConfig.DisableRDT {
		t.Fatal("DisableRDT was not preserved")
	}
	bulkheadRDTConfig := got.CPUPluginConfig.BulkheadConfig.BulkheadRDTConfig
	if bulkheadRDTConfig == nil || bulkheadRDTConfig.EnableCPUList == nil || !*bulkheadRDTConfig.EnableCPUList {
		t.Fatal("EnableCPUList was not preserved")
	}
	if bulkheadRDTConfig.EnableCAT == nil || !*bulkheadRDTConfig.EnableCAT {
		t.Fatal("EnableCAT was not preserved")
	}
	if bulkheadRDTConfig.DefaultCATWays == nil || bulkheadRDTConfig.DefaultCATWays.String() != "MaxCATWays" {
		t.Fatalf("DefaultCATWays = %v, want MaxCATWays", bulkheadRDTConfig.DefaultCATWays)
	}
	share00CATWays := bulkheadRDTConfig.ClosCATWays["share-00"]
	if got := share00CATWays.String(); got != "MaxCATWays-MinCATWays" {
		t.Fatalf("ClosCATWays[share-00] = %s, want MaxCATWays-MinCATWays", got)
	}
	share01CATWays := bulkheadRDTConfig.ClosCATWays["share-01"]
	if got := share01CATWays.IntValue(); got != 2 {
		t.Fatalf("ClosCATWays[share-01] = %d, want 2", got)
	}
	if bulkheadRDTConfig.CATPolicy == nil || bulkheadRDTConfig.CATPolicy.ExclusiveClosIDs == nil {
		t.Fatalf("CATPolicy exclusiveClosIDs = %#v, want pointer", bulkheadRDTConfig.CATPolicy)
	}
	if got := *bulkheadRDTConfig.CATPolicy.ExclusiveClosIDs; len(got) != 2 || got[0] != "clos-a" || got[1] != "peer-b" {
		t.Fatalf("ExclusiveClosIDs = %#v, want clos-a and peer-b", got)
	}
	if got := bulkheadRDTConfig.CATPolicy.DefaultPlacement.Direction; got != CATAllocationDirectionLow {
		t.Fatalf("DefaultPlacement.Direction = %s, want %s", got, CATAllocationDirectionLow)
	}
	if got := bulkheadRDTConfig.CATPolicy.ClosPlacements["share-01"].Direction; got != CATAllocationDirectionHigh {
		t.Fatalf("ClosPlacements[share-01].Direction = %s, want %s", got, CATAllocationDirectionHigh)
	}

	copied := got.DeepCopy()
	copiedPolicy := copied.CPUPluginConfig.BulkheadConfig.BulkheadRDTConfig.CATPolicy
	copiedPolicy.DefaultPlacement.AllowedBitUsages[0] = CATBitUsageExclusive
	share01Placement := copiedPolicy.ClosPlacements["share-01"]
	share01Placement.AllowedBitUsages[0] = CATBitUsageExclusive
	copiedPolicy.ClosPlacements["share-01"] = share01Placement
	(*copiedPolicy.ExclusiveClosIDs)[0] = "mutated"
	require.Equal(t, CATBitUsageAll, bulkheadRDTConfig.CATPolicy.DefaultPlacement.AllowedBitUsages[0])
	require.Equal(t, CATBitUsageSoftware, bulkheadRDTConfig.CATPolicy.ClosPlacements["share-01"].AllowedBitUsages[0])
	require.Equal(t, "clos-a", (*bulkheadRDTConfig.CATPolicy.ExclusiveClosIDs)[0])
}

func TestCPUProvisionConfigFillDefaultSharePoolRoundTrip(t *testing.T) {
	enabled := true
	in := &CPUProvisionConfig{
		FillDefaultSharePoolWithNonReclaimCPUs: &enabled,
	}
	data, err := json.Marshal(in)
	require.NoError(t, err)
	require.JSONEq(t, `{"fillDefaultSharePoolWithNonReclaimCPUs":true}`, string(data))
	out := &CPUProvisionConfig{}
	require.NoError(t, json.Unmarshal(data, out))
	require.NotNil(t, out.FillDefaultSharePoolWithNonReclaimCPUs)
	require.True(t, *out.FillDefaultSharePoolWithNonReclaimCPUs)
	copied := in.DeepCopy()
	require.NotSame(t, in.FillDefaultSharePoolWithNonReclaimCPUs, copied.FillDefaultSharePoolWithNonReclaimCPUs)
	require.True(t, *copied.FillDefaultSharePoolWithNonReclaimCPUs)
}

func TestControlKnobConstraintsEnforceOptionalRoundTrip(t *testing.T) {
	t.Run("nil enforce is omitted (omitempty) and stays nil", func(t *testing.T) {
		in := &ControlKnobConstraints{Name: "reclaimed-cores-cpu-size"}
		data, err := json.Marshal(in)
		require.NoError(t, err)
		require.JSONEq(t, `{"name":"reclaimed-cores-cpu-size"}`, string(data))
		out := &ControlKnobConstraints{}
		require.NoError(t, json.Unmarshal(data, out))
		require.Nil(t, out.Enforce)
	})

	t.Run("false and true are preserved and round-tripped", func(t *testing.T) {
		shadow := false
		enforced := true
		for _, tc := range []*bool{&shadow, &enforced} {
			in := &ControlKnobConstraints{Name: "reclaimed-cores-cpu-quota", Enforce: tc}
			data, err := json.Marshal(in)
			require.NoError(t, err)
			require.Contains(t, string(data), `"enforce":`)
			out := &ControlKnobConstraints{}
			require.NoError(t, json.Unmarshal(data, out))
			require.NotNil(t, out.Enforce)
			require.Equal(t, *tc, *out.Enforce)
		}
	})

	t.Run("deepcopy copies the enforce pointer deeply", func(t *testing.T) {
		enforced := true
		in := &ControlKnobConstraints{Name: "reclaimed-cores-cpu-size", Enforce: &enforced}
		copied := in.DeepCopy()
		require.NotNil(t, copied.Enforce)
		require.NotSame(t, in.Enforce, copied.Enforce)
		require.True(t, *copied.Enforce)
	})
}

func TestControlKnobConstraintsTargetReferenceOptionalRoundTrip(t *testing.T) {
	t.Run("both omitted are omitted in JSON (legacy entry)", func(t *testing.T) {
		in := &ControlKnobConstraints{Name: "non-reclaimed-cpu-requirement"}
		data, err := json.Marshal(in)
		require.NoError(t, err)
		require.JSONEq(t, `{"name":"non-reclaimed-cpu-requirement"}`, string(data))
		out := &ControlKnobConstraints{}
		require.NoError(t, json.Unmarshal(data, out))
		require.Empty(t, out.TargetPolicy)
		require.Empty(t, out.ReferencePolicy)
	})

	t.Run("both set are preserved and round-tripped", func(t *testing.T) {
		gap := 24.0
		in := &ControlKnobConstraints{
			Name:                "non-reclaimed-cpu-requirement",
			TargetPolicy:        CPUProvisionPolicyRama,
			ReferencePolicy:     CPUProvisionPolicyCanonical,
			RestrictConstraints: RestrictConstraints{MaxUpperGap: &gap},
		}
		data, err := json.Marshal(in)
		require.NoError(t, err)
		require.Contains(t, string(data), `"targetPolicy":"rama"`)
		require.Contains(t, string(data), `"referencePolicy":"canonical"`)
		out := &ControlKnobConstraints{}
		require.NoError(t, json.Unmarshal(data, out))
		require.Equal(t, CPUProvisionPolicyRama, out.TargetPolicy)
		require.Equal(t, CPUProvisionPolicyCanonical, out.ReferencePolicy)
	})

	t.Run("same knob may appear for multiple policy pairs", func(t *testing.T) {
		in := &CPUProvisionConfig{
			Constraints: &[]ControlKnobConstraints{
				{Name: "non-reclaimed-cpu-requirement", TargetPolicy: CPUProvisionPolicyRama, ReferencePolicy: CPUProvisionPolicyCanonical},
				{Name: "non-reclaimed-cpu-requirement", TargetPolicy: "dynamic-quota", ReferencePolicy: CPUProvisionPolicyCanonical},
			},
		}
		data, err := json.Marshal(in)
		require.NoError(t, err)
		out := &CPUProvisionConfig{}
		require.NoError(t, json.Unmarshal(data, out))
		require.NotNil(t, out.Constraints)
		require.Len(t, *out.Constraints, 2)
		require.Equal(t, "dynamic-quota", string((*out.Constraints)[1].TargetPolicy))
	})
}

func TestMemoryGuardConfigCriticalWatermarkSource(t *testing.T) {
	t.Run("JSON round trip and optional omission", func(t *testing.T) {
		source := CriticalWatermarkSourceHigh
		in := &MemoryGuardConfig{CriticalWatermarkSource: &source}

		data, err := json.Marshal(in)
		require.NoError(t, err)
		require.JSONEq(t, `{"criticalWatermarkSource":"high"}`, string(data))

		out := &MemoryGuardConfig{}
		require.NoError(t, json.Unmarshal(data, out))
		require.NotNil(t, out.CriticalWatermarkSource)
		require.Equal(t, CriticalWatermarkSourceHigh, *out.CriticalWatermarkSource)

		data, err = json.Marshal(&MemoryGuardConfig{})
		require.NoError(t, err)
		require.JSONEq(t, `{}`, string(data))
	})

	t.Run("deep copy owns an independent pointer", func(t *testing.T) {
		source := CriticalWatermarkSourceLow
		in := &MemoryGuardConfig{CriticalWatermarkSource: &source}

		copied := in.DeepCopy()
		require.NotNil(t, copied.CriticalWatermarkSource)
		require.NotSame(t, in.CriticalWatermarkSource, copied.CriticalWatermarkSource)
		require.Equal(t, CriticalWatermarkSourceLow, *copied.CriticalWatermarkSource)

		*copied.CriticalWatermarkSource = CriticalWatermarkSourceHigh
		require.Equal(t, CriticalWatermarkSourceLow, *in.CriticalWatermarkSource)
	})

	t.Run("CRD schema exposes the low and high enum", func(t *testing.T) {
		data, err := os.ReadFile("../../../../config/crd/bases/config.katalyst.kubewharf.io_adminqosconfigurations.yaml")
		require.NoError(t, err)

		jsonData, err := yaml.YAMLToJSON(data)
		require.NoError(t, err)

		var crd struct {
			Spec struct {
				Versions []struct {
					Schema struct {
						OpenAPIV3Schema adminQoSSchema `json:"openAPIV3Schema"`
					} `json:"schema"`
				} `json:"versions"`
			} `json:"spec"`
		}
		require.NoError(t, json.Unmarshal(jsonData, &crd))
		require.NotEmpty(t, crd.Spec.Versions)

		schema := crd.Spec.Versions[0].Schema.OpenAPIV3Schema
		schema = adminQoSSchemaProperty(t, schema, "spec")
		schema = adminQoSSchemaProperty(t, schema, "config")
		schema = adminQoSSchemaProperty(t, schema, "advisorConfig")
		schema = adminQoSSchemaProperty(t, schema, "memoryAdvisorConfig")
		schema = adminQoSSchemaProperty(t, schema, "memoryGuardConfig")
		criticalWatermarkSource := adminQoSSchemaProperty(t, schema, "criticalWatermarkSource")

		require.Equal(t, "string", criticalWatermarkSource.Type)
		require.Equal(t, []string{"low", "high"}, criticalWatermarkSource.Enum)
	})
}

func TestReclaimedMemorySafeZoneGateConfig(t *testing.T) {
	t.Run("JSON round trip and optional omission", func(t *testing.T) {
		enabled := true
		threshold := 0.5
		in := &MemoryPressureEvictionConfig{
			EnableReclaimedMemorySafeZoneGate:  &enabled,
			ReclaimedMemoryUsageRatioThreshold: &threshold,
		}

		data, err := json.Marshal(in)
		require.NoError(t, err)
		require.Contains(t, string(data), `"enableReclaimedMemorySafeZoneGate":true`)
		require.Contains(t, string(data), `"reclaimedMemoryUsageRatioThreshold":0.5`)

		out := &MemoryPressureEvictionConfig{}
		require.NoError(t, json.Unmarshal(data, out))
		require.NotNil(t, out.EnableReclaimedMemorySafeZoneGate)
		require.True(t, *out.EnableReclaimedMemorySafeZoneGate)
		require.NotNil(t, out.ReclaimedMemoryUsageRatioThreshold)
		require.Equal(t, 0.5, *out.ReclaimedMemoryUsageRatioThreshold)

		// Explicit false round-trips as non-nil false and is preserved.
		disabled := false
		falseIn := &MemoryPressureEvictionConfig{EnableReclaimedMemorySafeZoneGate: &disabled}
		falseData, err := json.Marshal(falseIn)
		require.NoError(t, err)
		require.Contains(t, string(falseData), `"enableReclaimedMemorySafeZoneGate":false`)
		falseOut := &MemoryPressureEvictionConfig{}
		require.NoError(t, json.Unmarshal(falseData, falseOut))
		require.NotNil(t, falseOut.EnableReclaimedMemorySafeZoneGate)
		require.False(t, *falseOut.EnableReclaimedMemorySafeZoneGate)

		data, err = json.Marshal(&MemoryPressureEvictionConfig{})
		require.NoError(t, err)
		require.NotContains(t, string(data), "enableReclaimedMemorySafeZoneGate")
		require.NotContains(t, string(data), "reclaimedMemoryUsageRatioThreshold")

		// Omitted objects keep both pointers nil after round-trip, so omission
		// stays distinguishable from explicit false / zero values.
		legacy := `{"enableNumaLevelEviction":true}`
		legacyOut := &MemoryPressureEvictionConfig{}
		require.NoError(t, json.Unmarshal([]byte(legacy), legacyOut))
		require.Nil(t, legacyOut.EnableReclaimedMemorySafeZoneGate)
		require.Nil(t, legacyOut.ReclaimedMemoryUsageRatioThreshold)
	})

	t.Run("deep copy owns independent pointers", func(t *testing.T) {
		enabled := true
		threshold := 0.25
		in := &MemoryPressureEvictionConfig{
			EnableReclaimedMemorySafeZoneGate:  &enabled,
			ReclaimedMemoryUsageRatioThreshold: &threshold,
		}

		copied := in.DeepCopy()
		require.NotSame(t, in.EnableReclaimedMemorySafeZoneGate, copied.EnableReclaimedMemorySafeZoneGate)
		require.NotSame(t, in.ReclaimedMemoryUsageRatioThreshold, copied.ReclaimedMemoryUsageRatioThreshold)

		*copied.EnableReclaimedMemorySafeZoneGate = false
		*copied.ReclaimedMemoryUsageRatioThreshold = 0.75
		require.True(t, *in.EnableReclaimedMemorySafeZoneGate)
		require.Equal(t, 0.25, *in.ReclaimedMemoryUsageRatioThreshold)
	})

	t.Run("CRD schema exposes bool flag and 0..1 threshold", func(t *testing.T) {
		data, err := os.ReadFile("../../../../config/crd/bases/config.katalyst.kubewharf.io_adminqosconfigurations.yaml")
		require.NoError(t, err)

		jsonData, err := yaml.YAMLToJSON(data)
		require.NoError(t, err)

		var crd struct {
			Spec struct {
				Versions []struct {
					Schema struct {
						OpenAPIV3Schema adminQoSSchema `json:"openAPIV3Schema"`
					} `json:"schema"`
				} `json:"versions"`
			} `json:"spec"`
		}
		require.NoError(t, json.Unmarshal(jsonData, &crd))
		require.NotEmpty(t, crd.Spec.Versions)

		schema := crd.Spec.Versions[0].Schema.OpenAPIV3Schema
		schema = adminQoSSchemaProperty(t, schema, "spec")
		schema = adminQoSSchemaProperty(t, schema, "config")
		schema = adminQoSSchemaProperty(t, schema, "evictionConfig")
		schema = adminQoSSchemaProperty(t, schema, "memoryPressureEvictionConfig")

		flag := adminQoSSchemaProperty(t, schema, "enableReclaimedMemorySafeZoneGate")
		require.Equal(t, "boolean", flag.Type)
		require.Empty(t, flag.Enum)

		threshold := adminQoSSchemaProperty(t, schema, "reclaimedMemoryUsageRatioThreshold")
		require.Equal(t, "number", threshold.Type)
		require.NotNil(t, threshold.Minimum)
		require.Equal(t, 0.0, *threshold.Minimum)
		require.NotNil(t, threshold.Maximum)
		require.Equal(t, 1.0, *threshold.Maximum)
	})
}

type adminQoSSchema struct {
	Properties map[string]adminQoSSchema `json:"properties"`
	Type       string                    `json:"type"`
	Enum       []string                  `json:"enum"`
	Minimum    *float64                  `json:"minimum"`
	Maximum    *float64                  `json:"maximum"`
	Items      *adminQoSSchema           `json:"items"`
	Required   []string                  `json:"required"`
	MinItems   *int                      `json:"minItems"`
}

func adminQoSSchemaProperty(t *testing.T, schema adminQoSSchema, name string) adminQoSSchema {
	t.Helper()
	property, ok := schema.Properties[name]
	require.Truef(t, ok, "schema property %q not found", name)
	return property
}

// TestCPUProvisionConfigConstraintsTriStateRoundTrip pins the pointer-level
// tri-state contract on CPUProvisionConfig.Constraints: nil (omitted) means
// fall back to Core flag defaults; &empty ("constraints: []") explicitly
// disables bounds; &non-empty is a whole-list override that is never merged
// with Core defaults. "constraints: null" is not part of the write contract.
func TestCPUProvisionConfigConstraintsTriStateRoundTrip(t *testing.T) {
	newEntry := func(name string, gap float64) ControlKnobConstraints {
		return ControlKnobConstraints{
			Name:                ControlKnobName(name),
			RestrictConstraints: RestrictConstraints{MaxUpperGap: &gap},
		}
	}

	t.Run("nil is omitted in JSON and stays nil across JSON/YAML/unstructured", func(t *testing.T) {
		in := &CPUProvisionConfig{}

		jsonData, err := json.Marshal(in)
		require.NoError(t, err)
		require.JSONEq(t, `{}`, string(jsonData))

		out := &CPUProvisionConfig{}
		require.NoError(t, json.Unmarshal(jsonData, out))
		require.Nil(t, out.Constraints)

		// Explicit "constraints: null" must also decode to nil; it is not a
		// distinct state we surface.
		nullOut := &CPUProvisionConfig{}
		require.NoError(t, json.Unmarshal([]byte(`{"constraints":null}`), nullOut))
		require.Nil(t, nullOut.Constraints)

		// YAML: a nil pointer renders without the key.
		yamlData, err := yaml.Marshal(in)
		require.NoError(t, err)
		require.NotContains(t, string(yamlData), "constraints:")

		// Typed-unstructured round-trip preserves nil.
		unstr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(in)
		require.NoError(t, err)
		_, hasConstraints := unstr["constraints"]
		require.False(t, hasConstraints, "nil Constraints must not surface in unstructured map")
		back := &CPUProvisionConfig{}
		require.NoError(t, runtime.DefaultUnstructuredConverter.FromUnstructured(unstr, back))
		require.Nil(t, back.Constraints)
	})

	t.Run("&empty renders as constraints: [] and decodes to a non-nil empty slice", func(t *testing.T) {
		in := &CPUProvisionConfig{Constraints: &[]ControlKnobConstraints{}}

		jsonData, err := json.Marshal(in)
		require.NoError(t, err)
		require.JSONEq(t, `{"constraints":[]}`, string(jsonData))

		out := &CPUProvisionConfig{}
		require.NoError(t, json.Unmarshal(jsonData, out))
		require.NotNil(t, out.Constraints, "constraints:[] must decode to a non-nil pointer")
		require.Len(t, *out.Constraints, 0)

		yamlData, err := yaml.Marshal(in)
		require.NoError(t, err)
		require.Contains(t, string(yamlData), "constraints: []")

		unstr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(in)
		require.NoError(t, err)
		list, ok := unstr["constraints"].([]interface{})
		require.True(t, ok, "constraints must be a list in unstructured map")
		require.Len(t, list, 0)
		back := &CPUProvisionConfig{}
		require.NoError(t, runtime.DefaultUnstructuredConverter.FromUnstructured(unstr, back))
		require.NotNil(t, back.Constraints)
		require.Len(t, *back.Constraints, 0)
	})

	t.Run("&non-empty round-trips entries and DeepCopy owns an independent slice", func(t *testing.T) {
		gap := 12.0
		enforced := true
		in := &CPUProvisionConfig{Constraints: &[]ControlKnobConstraints{
			{Name: "reclaimed-cores-cpu-size", RestrictConstraints: RestrictConstraints{MaxUpperGap: &gap}, Enforce: &enforced},
			newEntry("non-reclaimed-cpu-requirement", 24.0),
		}}

		jsonData, err := json.Marshal(in)
		require.NoError(t, err)
		require.Contains(t, string(jsonData), `"constraints":[`)
		require.Contains(t, string(jsonData), `"name":"reclaimed-cores-cpu-size"`)
		require.Contains(t, string(jsonData), `"enforce":true`)

		out := &CPUProvisionConfig{}
		require.NoError(t, json.Unmarshal(jsonData, out))
		require.NotNil(t, out.Constraints)
		require.Len(t, *out.Constraints, 2)
		require.Equal(t, ControlKnobName("reclaimed-cores-cpu-size"), (*out.Constraints)[0].Name)
		require.NotNil(t, (*out.Constraints)[0].MaxUpperGap)
		require.Equal(t, gap, *(*out.Constraints)[0].MaxUpperGap)
		require.NotNil(t, (*out.Constraints)[0].Enforce)
		require.True(t, *(*out.Constraints)[0].Enforce)

		// DeepCopy must not alias the backing slice, the entry struct, or the
		// gap/enforce pointers carried by each entry.
		copied := in.DeepCopy()
		require.NotNil(t, copied.Constraints)
		require.NotSame(t, in.Constraints, copied.Constraints)
		require.NotSame(t, &(*in.Constraints)[0], &(*copied.Constraints)[0])
		require.NotSame(t, (*in.Constraints)[0].MaxUpperGap, (*copied.Constraints)[0].MaxUpperGap)
		require.NotSame(t, (*in.Constraints)[0].Enforce, (*copied.Constraints)[0].Enforce)

		*(*copied.Constraints)[0].MaxUpperGap = 999.0
		*(*copied.Constraints)[0].Enforce = false
		require.Equal(t, 12.0, *(*in.Constraints)[0].MaxUpperGap, "DeepCopy must isolate mutated gaps")
		require.True(t, *(*in.Constraints)[0].Enforce, "DeepCopy must isolate mutated enforce flags")

		// YAML + unstructured paths also preserve a non-empty list.
		yamlData, err := yaml.Marshal(in)
		require.NoError(t, err)
		require.Contains(t, string(yamlData), "constraints:")
		require.Contains(t, string(yamlData), "reclaimed-cores-cpu-size")

		unstr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(in)
		require.NoError(t, err)
		list, ok := unstr["constraints"].([]interface{})
		require.True(t, ok)
		require.Len(t, list, 2)
		back := &CPUProvisionConfig{}
		require.NoError(t, runtime.DefaultUnstructuredConverter.FromUnstructured(unstr, back))
		require.NotNil(t, back.Constraints)
		require.Len(t, *back.Constraints, 2)
	})

	t.Run("legacy non-empty wire format (bare array) decodes into the pointer field", func(t *testing.T) {
		// AQC manifests written against the previous non-pointer field emit
		// "constraints":[{"name":...}] directly. The new pointer field must
		// keep accepting that shape as a non-nil, non-empty override,
		// including entries that carry the enforce/target/reference policy
		// knobs introduced alongside.
		legacy := `{"constraints":[{"name":"reclaimed-cores-cpu-size","maxUpperGap":8,"enforce":true,"targetPolicy":"rama","referencePolicy":"canonical"}]}`
		out := &CPUProvisionConfig{}
		require.NoError(t, json.Unmarshal([]byte(legacy), out))
		require.NotNil(t, out.Constraints)
		require.Len(t, *out.Constraints, 1)
		entry := (*out.Constraints)[0]
		require.Equal(t, ControlKnobName("reclaimed-cores-cpu-size"), entry.Name)
		require.NotNil(t, entry.MaxUpperGap)
		require.Equal(t, 8.0, *entry.MaxUpperGap)
		require.NotNil(t, entry.Enforce)
		require.True(t, *entry.Enforce)
		require.Equal(t, CPUProvisionPolicyRama, entry.TargetPolicy)
		require.Equal(t, CPUProvisionPolicyCanonical, entry.ReferencePolicy)
	})

	t.Run("CRD exposes constraints as an optional array (not required, no minItems)", func(t *testing.T) {
		data, err := os.ReadFile("../../../../config/crd/bases/config.katalyst.kubewharf.io_adminqosconfigurations.yaml")
		require.NoError(t, err)

		jsonData, err := yaml.YAMLToJSON(data)
		require.NoError(t, err)

		var crd struct {
			Spec struct {
				Versions []struct {
					Schema struct {
						OpenAPIV3Schema adminQoSSchema `json:"openAPIV3Schema"`
					} `json:"schema"`
				} `json:"versions"`
			} `json:"spec"`
		}
		require.NoError(t, json.Unmarshal(jsonData, &crd))
		require.NotEmpty(t, crd.Spec.Versions)

		schema := crd.Spec.Versions[0].Schema.OpenAPIV3Schema
		schema = adminQoSSchemaProperty(t, schema, "spec")
		schema = adminQoSSchemaProperty(t, schema, "config")
		schema = adminQoSSchemaProperty(t, schema, "advisorConfig")
		schema = adminQoSSchemaProperty(t, schema, "cpuAdvisorConfig")
		provision := adminQoSSchemaProperty(t, schema, "cpuProvisionConfig")

		constraints, ok := provision.Properties["constraints"]
		require.True(t, ok, "constraints must still appear in the CRD schema")
		require.Equal(t, "array", constraints.Type)
		require.NotNil(t, constraints.Items)
		require.Nil(t, constraints.MinItems, "constraints must remain optional (no minItems)")

		// The entry schema must expose the enrichments introduced alongside the
		// tri-state pointer (enforce, targetPolicy, referencePolicy).
		items := *constraints.Items
		require.Contains(t, items.Properties, "enforce")
		require.Contains(t, items.Properties, "targetPolicy")
		require.Contains(t, items.Properties, "referencePolicy")
		require.Contains(t, items.Properties, "name")

		// cpuProvisionConfig itself must not list constraints as required.
		for _, name := range provision.Required {
			require.NotEqual(t, "constraints", name, "constraints must not be required")
		}
	})
}
