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

type BulkheadConfig struct {
	// Enable controls whether core bulkhead is enabled.
	// +optional
	Enable *bool `json:"enable,omitempty"`
	// EnableBulkheadCpusetTopology controls whether the core bulkhead cpuset
	// topology plugin is enabled when Enable is true.
	// +optional
	EnableBulkheadCpusetTopology *bool `json:"enableBulkheadCpusetTopology,omitempty"`
	// EnableBulkheadCpusetMems controls whether the core bulkhead cpuset_mems
	// plugin is enabled when Enable is true. The plugin writes cpuset.mems for
	// reclaim NUMA buckets independently from cpuset topology.
	// +optional
	EnableBulkheadCpusetMems *bool `json:"enableBulkheadCpusetMems,omitempty"`
	// EnableBulkheadWorkqueue controls whether the core bulkhead workqueue
	// plugin is enabled when Enable is true.
	// +optional
	EnableBulkheadWorkqueue *bool `json:"enableBulkheadWorkqueue,omitempty"`
	// EnableBulkheadSystemService controls whether the core bulkhead
	// system_service plugin is enabled when Enable is true.
	// +optional
	EnableBulkheadSystemService *bool `json:"enableBulkheadSystemService,omitempty"`
	// NonReclaimPoolMinSize is the minimum CPU count kept in the non-reclaim
	// pool for bulkhead cpuset topology. When the current non-reclaim pool is
	// smaller than this value, CPUs are padded from reclaim effective candidates.
	// The padded CPUs never overlap with reserve pool CPUs.
	// +optional
	// +kubebuilder:validation:Minimum=0
	NonReclaimPoolMinSize *int64 `json:"nonReclaimPoolMinSize,omitempty"`
	// BulkheadRDTConfig is the dynamic config for RDT bulkhead plugins.
	// +optional
	BulkheadRDTConfig *BulkheadRDTConfig `json:"bulkheadRDTConfig,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="!has(self.enableCAT) || !self.enableCAT || has(self.defaultCATWays)",message="defaultCATWays must be specified when enableCAT is true"
type BulkheadRDTConfig struct {
	// EnableCPUList controls whether the RDT CPU list bulkhead plugin is enabled.
	// +optional
	EnableCPUList *bool `json:"enableCPUList,omitempty"`
	// EnableCAT controls whether the RDT CAT bulkhead plugin is enabled.
	// +optional
	EnableCAT *bool `json:"enableCAT,omitempty"`
	// DefaultCATWays is the default CAT way count for non-root CLOS.
	// +optional
	// +kubebuilder:validation:Minimum=1
	DefaultCATWays *int64 `json:"defaultCATWays,omitempty"`
	// ClosCATWays maps pool names or CLOS IDs to CAT way counts.
	// +kubebuilder:validation:XValidation:rule="self.all(k, self[k] > 0)",message="all CLOS CAT ways must be greater than 0"
	// +optional
	ClosCATWays map[string]int64 `json:"closCATWays,omitempty"`
}
