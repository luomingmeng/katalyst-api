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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQRMPluginConfigRDTAndBulkheadRDTConfig(t *testing.T) {
	disableRDT := true
	enableCPUList := true
	enableCAT := true
	defaultCATWays := int64(4)
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
					ClosCATWays:    map[string]int64{"reclaim": 2},
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
	if bulkheadRDTConfig.DefaultCATWays == nil || *bulkheadRDTConfig.DefaultCATWays != 4 {
		t.Fatalf("DefaultCATWays = %v, want 4", bulkheadRDTConfig.DefaultCATWays)
	}
	if bulkheadRDTConfig.ClosCATWays["reclaim"] != 2 {
		t.Fatalf("ClosCATWays[reclaim] = %d, want 2", bulkheadRDTConfig.ClosCATWays["reclaim"])
	}
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
