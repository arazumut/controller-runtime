/*
Copyright 2019 The Kubernetes Authors.

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
	"math/rand"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	api "sigs.k8s.io/controller-runtime/examples/crd/pkg"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var setupLog = ctrl.Log.WithName("setup")

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile handles the logic of managing the lifecycle of ChaosPods.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("chaospod", req.NamespacedName)
	log.V(1).Info("Reconciling ChaosPod")

	// Fetch the ChaosPod resource
	var chaospod api.ChaosPod
	if err := r.Get(ctx, req.NamespacedName, &chaospod); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("ChaosPod resource not found. Ignoring since it must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get ChaosPod")
		return ctrl.Result{}, err
	}

	// Check if the associated Pod exists
	var pod corev1.Pod
	podFound := true
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Failed to get associated Pod")
			return ctrl.Result{}, err
		}
		podFound = false
	}

	if podFound {
		// Check if it's time to stop the pod
		shouldStop := chaospod.Spec.NextStop.Time.Before(time.Now())
		if !shouldStop {
			// Requeue until the NextStop time
			return ctrl.Result{RequeueAfter: chaospod.Spec.NextStop.Sub(time.Now()) + 1*time.Second}, nil
		}

		// Delete the Pod if it's time to stop
		if err := r.Delete(ctx, &pod); err != nil {
			log.Error(err, "Failed to delete Pod")
			return ctrl.Result{}, err
		}

		log.Info("Pod deleted, requeuing for next cycle")
		return ctrl.Result{Requeue: true}, nil
	}

	// Create a new Pod if not found
	podTemplate := chaospod.Spec.Template.DeepCopy()
	pod.ObjectMeta = podTemplate.ObjectMeta
	pod.Name = req.Name
	pod.Namespace = req.Namespace
	pod.Spec = podTemplate.Spec

	if err := ctrl.SetControllerReference(&chaospod, &pod, r.Scheme); err != nil {
		log.Error(err, "Failed to set pod owner reference")
		return ctrl.Result{}, err
	}

	if err := r.Create(ctx, &pod); err != nil {
		log.Error(err, "Failed to create Pod")
		return ctrl.Result{}, err
	}

	// Update the NextStop time and ChaosPod status
	chaospod.Spec.NextStop.Time = time.Now().Add(time.Duration(10*(rand.Int63n(2)+1)) * time.Second)
	chaospod.Status.LastRun = pod.CreationTimestamp
	if err := r.Update(ctx, &chaospod); err != nil {
		log.Error(err, "Failed to update ChaosPod status")
		return ctrl.Result{}, err
	}

	log.Info("Pod created successfully, requeuing for next stop")
	return ctrl.Result{}, nil
}

func main() {
	// Set up logger
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	// Create a new manager
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		setupLog.Error(err, "Unable to start manager")
		os.Exit(1)
	}

	// Add the custom scheme (ChaosPod CRD)
	if err := api.AddToScheme(mgr.GetScheme()); err != nil {
		setupLog.Error(err, "Unable to add ChaosPod scheme")
		os.Exit(1)
	}

	// Create a new controller for ChaosPod
	err = ctrl.NewControllerManagedBy(mgr).
		For(&api.ChaosPod{}).
		Owns(&corev1.Pod{}).
		Complete(&Reconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		})
	if err != nil {
		setupLog.Error(err, "Unable to create controller")
		os.Exit(1)
	}

	// Start the manager
	setupLog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Problem running manager")
		os.Exit(1)
	}
}
