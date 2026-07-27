// Copyright 2026 The Katalyst Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"encoding/json"
	"testing"
)

func TestCPUAdvisorConfigDisableDedicatedCoresOverlapReclaimedCores(t *testing.T) {
	t.Parallel()

	var config CPUAdvisorConfig
	if err := json.Unmarshal([]byte(`{"disableDedicatedCoresOverlapReclaimedCores":true}`), &config); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}

	if config.DisableDedicatedCoresOverlapReclaimedCores == nil {
		t.Fatal("disable dedicated overlap flag must be decoded")
	}
	if !*config.DisableDedicatedCoresOverlapReclaimedCores {
		t.Fatal("disable dedicated overlap flag must preserve its value")
	}

	copied := config.DeepCopy()
	*copied.DisableDedicatedCoresOverlapReclaimedCores = false
	if !*config.DisableDedicatedCoresOverlapReclaimedCores {
		t.Fatal("deep copy must not alias disable dedicated overlap flag")
	}
}
