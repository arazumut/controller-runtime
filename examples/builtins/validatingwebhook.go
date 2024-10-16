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
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-v1-pod,mutating=false,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=vpod.kb.io

// podValidator validates Pods during create and update operations
type podValidator struct{}

// validate checks if a specific annotation exists and has the correct value
func (v *podValidator) validate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	log := logf.FromContext(ctx)

	// Type assertion to ensure obj is a Pod
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return nil, fmt.Errorf("expected a Pod but got a %T", obj)
	}

	log.Info("Validating Pod", "podName", pod.Name)

	// Check for the specific annotation
	key := "example-mutating-admission-webhook"
	anno, found := pod.Annotations[key]
	if !found {
		log.Info("Pod is missing required annotation", "annotationKey", key)
		return nil, fmt.Errorf("missing required annotation: %s", key)
	}

	// Validate the annotation value
	if anno != "foo" {
		log.Info("Pod has incorrect annotation value", "expected", "foo", "found", anno)
		return nil, fmt.Errorf("annotation %s did not have expected value %q", key, "foo")
	}

	log.Info("Pod passed validation", "podName", pod.Name)
	return nil, nil
}

// ValidateCreate is called during Pod creation
func (v *podValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	log := logf.FromContext(ctx)
	log.Info("Validating Pod creation")
	return v.validate(ctx, obj)
}

// ValidateUpdate is called during Pod update
func (v *podValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	log := logf.FromContext(ctx)
	log.Info("Validating Pod update")
	return v.validate(ctx, newObj)
}
