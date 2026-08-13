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

import "k8s.io/apimachinery/pkg/util/intstr"

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
	// DefaultCATWays is the default CAT way count or expression for non-root CLOS.
	// Supported expression variables are MaxCATWays and MinCATWays.
	// +optional
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:XValidation:rule="type(self) == int ? self > 0 : self.matches('^\\s*(MaxCATWays|MinCATWays|[1-9][0-9]*)(\\s*[+-]\\s*(MaxCATWays|MinCATWays|[1-9][0-9]*))?\\s*$')",message="defaultCATWays must be a positive integer or a valid CAT ways expression"
	// +kubebuilder:validation:XValidation:rule="type(self) == int || !self.matches('^\\s*(MaxCATWays\\s*-\\s*MaxCATWays|MinCATWays\\s*-\\s*(MinCATWays|MaxCATWays)|[1-9][0-9]*\\s*[+-]\\s*[1-9][0-9]*)\\s*$')",message="defaultCATWays expression is statically non-positive or contains unsimplified literal arithmetic"
	DefaultCATWays *intstr.IntOrString `json:"defaultCATWays,omitempty"`
	// ClosCATWays maps pool names or CLOS IDs to CAT way counts or expressions.
	// +kubebuilder:validation:XValidation:rule="self.all(k, type(self[k]) == int ? self[k] > 0 : self[k].matches('^\\s*(MaxCATWays|MinCATWays|[1-9][0-9]*)(\\s*[+-]\\s*(MaxCATWays|MinCATWays|[1-9][0-9]*))?\\s*$'))",message="all CLOS CAT ways must be positive integers or valid CAT ways expressions"
	// +kubebuilder:validation:XValidation:rule="self.all(k, type(self[k]) == int || !self[k].matches('^\\s*(MaxCATWays\\s*-\\s*MaxCATWays|MinCATWays\\s*-\\s*(MinCATWays|MaxCATWays)|[1-9][0-9]*\\s*[+-]\\s*[1-9][0-9]*)\\s*$'))",message="CLOS CAT ways expressions must not be statically non-positive or contain unsimplified literal arithmetic"
	// +kubebuilder:validation:XValidation:rule="self.all(k, k.matches('^\\S+$'))",message="CLOS CAT ways keys must not be empty or contain whitespace"
	// +optional
	ClosCATWays map[string]intstr.IntOrString `json:"closCATWays,omitempty"`
	// CATPolicy controls CAT bit placement and deterministic allocation groups.
	// +optional
	CATPolicy *CATPolicy `json:"catPolicy,omitempty"`
}

// CATWaysExpressionVariable is a supported symbolic operand in CAT way expressions.
type CATWaysExpressionVariable string

const (
	CATWaysExpressionVariableMaxCATWays CATWaysExpressionVariable = "MaxCATWays"
	CATWaysExpressionVariableMinCATWays CATWaysExpressionVariable = "MinCATWays"
)

// CATPolicy controls CAT bit placement and deterministic allocation groups.
type CATPolicy struct {
	// DefaultPlacement is used by CLOS IDs without a CLOS-specific placement
	// or allocation group placement.
	// +optional
	DefaultPlacement *CATPlacementPolicy `json:"defaultPlacement,omitempty"`
	// ClosPlacements maps pool names or CLOS IDs to per-CLOS placement policy.
	// It also applies to CLOS IDs that appear in allocationGroups and overrides
	// the corresponding group default placement for non-empty fields.
	// +kubebuilder:validation:XValidation:rule="self.all(k, k.matches('^\\S+$'))",message="CLOS placement keys must not be empty or contain whitespace"
	// +optional
	ClosPlacements map[string]CATPlacementPolicy `json:"closPlacements,omitempty"`
	// AllocationGroups pack ordered CLOS IDs into non-overlapping CAT masks.
	// +optional
	AllocationGroups []CATAllocationGroup `json:"allocationGroups,omitempty"`
}

// CATPlacementPolicy controls where CAT masks are selected for one CLOS.
type CATPlacementPolicy struct {
	// AllowedBitUsages restricts candidate CAT bits by resctrl bit_usage class.
	// Empty means all ways supported by the domain are available.
	// +optional
	AllowedBitUsages []CATBitUsage `json:"allowedBitUsages,omitempty"`
	// Direction controls deterministic contiguous mask selection.
	// Empty means low.
	// +optional
	Direction CATAllocationDirection `json:"direction,omitempty"`
}

// CATAllocationGroup packs ordered CLOS IDs into non-overlapping CAT masks.
// +kubebuilder:validation:XValidation:rule="self.closIDs.all(id, id.matches('^\\S+$'))",message="allocation group CLOS IDs must not be empty or contain whitespace"
// +kubebuilder:validation:XValidation:rule="self.closIDs.all(id, self.closIDs.exists_one(other, other == id))",message="allocation group CLOS IDs must be unique"
type CATAllocationGroup struct {
	// Name identifies this allocation group in errors and diagnostics.
	// +optional
	Name string `json:"name,omitempty"`
	// ClosIDs is the ordered list of pool names or CLOS IDs to pack.
	// +kubebuilder:validation:MinItems=1
	ClosIDs []string `json:"closIDs"`
	// AllowedBitUsages is the default bit_usage constraint for CLOS IDs in this
	// group. catPolicy.closPlacements may override it for individual group members.
	// +optional
	AllowedBitUsages []CATBitUsage `json:"allowedBitUsages,omitempty"`
	// Direction controls deterministic packing direction.
	// +optional
	Direction CATAllocationDirection `json:"direction,omitempty"`
}

// CATBitUsage is one resctrl info/L3/bit_usage class.
// +kubebuilder:validation:Enum=S;H;X
type CATBitUsage string

const (
	CATBitUsageSoftware  CATBitUsage = "S"
	CATBitUsageHardware  CATBitUsage = "H"
	CATBitUsageExclusive CATBitUsage = "X"
)

// CATAllocationDirection controls deterministic contiguous mask selection.
// +kubebuilder:validation:Enum=low;high
type CATAllocationDirection string

const (
	CATAllocationDirectionLow  CATAllocationDirection = "low"
	CATAllocationDirectionHigh CATAllocationDirection = "high"
)
