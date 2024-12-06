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

// Code generated by informer-gen. DO NOT EDIT.

package externalversions

import (
	"fmt"

	v1alpha1 "github.com/kubewharf/katalyst-api/pkg/apis/autoscaling/v1alpha1"
	configv1alpha1 "github.com/kubewharf/katalyst-api/pkg/apis/config/v1alpha1"
	nodev1alpha1 "github.com/kubewharf/katalyst-api/pkg/apis/node/v1alpha1"
	overcommitv1alpha1 "github.com/kubewharf/katalyst-api/pkg/apis/overcommit/v1alpha1"
	recommendationv1alpha1 "github.com/kubewharf/katalyst-api/pkg/apis/recommendation/v1alpha1"
	tidev1alpha1 "github.com/kubewharf/katalyst-api/pkg/apis/tide/v1alpha1"
	workloadv1alpha1 "github.com/kubewharf/katalyst-api/pkg/apis/workload/v1alpha1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	cache "k8s.io/client-go/tools/cache"
)

// GenericInformer is type of SharedIndexInformer which will locate and delegate to other
// sharedInformers based on type
type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

type genericInformer struct {
	informer cache.SharedIndexInformer
	resource schema.GroupResource
}

// Informer returns the SharedIndexInformer.
func (f *genericInformer) Informer() cache.SharedIndexInformer {
	return f.informer
}

// Lister returns the GenericLister.
func (f *genericInformer) Lister() cache.GenericLister {
	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
}

// ForResource gives generic access to a shared informer of the matching type
// TODO extend this to unknown resources with a client pool
func (f *sharedInformerFactory) ForResource(resource schema.GroupVersionResource) (GenericInformer, error) {
	switch resource {
	// Group=autoscaling.katalyst.kubewharf.io, Version=v1alpha1
	case v1alpha1.SchemeGroupVersion.WithResource("katalystverticalpodautoscalers"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().KatalystVerticalPodAutoscalers().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("verticalpodautoscalerrecommendations"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Autoscaling().V1alpha1().VerticalPodAutoscalerRecommendations().Informer()}, nil

		// Group=config.katalyst.kubewharf.io, Version=v1alpha1
	case configv1alpha1.SchemeGroupVersion.WithResource("customnodeconfigs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Config().V1alpha1().CustomNodeConfigs().Informer()}, nil
	case configv1alpha1.SchemeGroupVersion.WithResource("katalystcustomconfigs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Config().V1alpha1().KatalystCustomConfigs().Informer()}, nil

		// Group=node.katalyst.kubewharf.io, Version=v1alpha1
	case nodev1alpha1.SchemeGroupVersion.WithResource("customnoderesources"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Node().V1alpha1().CustomNodeResources().Informer()}, nil
	case nodev1alpha1.SchemeGroupVersion.WithResource("nodeprofiledescriptors"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Node().V1alpha1().NodeProfileDescriptors().Informer()}, nil

		// Group=overcommit.katalyst.kubewharf.io, Version=v1alpha1
	case overcommitv1alpha1.SchemeGroupVersion.WithResource("nodeovercommitconfigs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Overcommit().V1alpha1().NodeOvercommitConfigs().Informer()}, nil

		// Group=recommendation.katalyst.kubewharf.io, Version=v1alpha1
	case recommendationv1alpha1.SchemeGroupVersion.WithResource("resourcerecommends"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Recommendation().V1alpha1().ResourceRecommends().Informer()}, nil

		// Group=tide.katalyst.kubewharf.io, Version=v1alpha1
	case tidev1alpha1.SchemeGroupVersion.WithResource("tidenodepools"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Tide().V1alpha1().TideNodePools().Informer()}, nil

		// Group=workload.katalyst.kubewharf.io, Version=v1alpha1
	case workloadv1alpha1.SchemeGroupVersion.WithResource("serviceprofiledescriptors"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Workload().V1alpha1().ServiceProfileDescriptors().Informer()}, nil

	}

	return nil, fmt.Errorf("no informer found for %v", resource)
}
