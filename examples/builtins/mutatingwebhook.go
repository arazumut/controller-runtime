/*
Copyright 2018 The Kubernetes Authors.

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

package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io

// podAnnotator is the webhook used to annotate Pods with custom metadata
type podAnnotator struct{}

// Create handles Pod creation mutation requests
func (a *podAnnotator) Create(ctx context.Context, obj runtime.Object) error {
	return a.Default(ctx, obj) // Default mutation logic is reused
}

// Update handles Pod update mutation requests
func (a *podAnnotator) Update(ctx context.Context, obj runtime.Object) error {
	return a.Default(ctx, obj) // Default mutation logic is reused
}

// Delete handles Pod deletion, nothing to do here for mutations
func (a *podAnnotator) Delete(ctx context.Context, obj runtime.Object) error {
	return nil
}

// Default implements the mutation logic for the Pod webhook
func (a *podAnnotator) Default(ctx context.Context, obj runtime.Object) error {
	log := logf.FromContext(ctx)

	// Cast the object to a Pod to ensure type correctness
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return fmt.Errorf("expected a Pod but got a %T", obj)
	}

	// Ensure the Pod has annotations
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}

	// Add or modify the custom annotation
	pod.Annotations["example-mutating-admission-webhook"] = "foo"
	log.Info("Annotated Pod with example-mutating-admission-webhook")

	return nil
}
