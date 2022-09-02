/*
Copyright 2022.

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

package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	logstashv1alpha1 "spacetime.k8s.elastic.co/logstash/api/v1alpha1"
)

// LogstashConfigWatcherReconciler reconciles a LogstashConfigWatcher object
type LogstashConfigWatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	configMapField = ".spec.configMap"
)

//+kubebuilder:rbac:groups=logstash.spacetime.k8s.elastic.co,resources=logstashconfigwatchers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=logstash.spacetime.k8s.elastic.co,resources=logstashconfigwatchers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=logstash.spacetime.k8s.elastic.co,resources=logstashconfigwatchers/finalizers,verbs=update
// +kubebuilder:rbac:groups=,resources=configmaps,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *LogstashConfigWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("LogstashConfigWatcher", req.NamespacedName)

	configWatcher := &logstashv1alpha1.LogstashConfigWatcher{}
	log.Info("looking for ", "namespaced name", req.NamespacedName)
	err := r.Get(context.TODO(), req.NamespacedName, configWatcher)
	log.Info("Configwatcher", "configwatcher.spec", configWatcher.Spec)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Cannot Find the config watcher")
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		log.Error(err, "something else is up")
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if configWatcher.Spec.ConfigMap != "" {
		configMapName := configWatcher.Spec.ConfigMap
		foundConfigMap := &corev1.ConfigMap{}
		err := r.Get(ctx, types.NamespacedName{Name: configMapName, Namespace: configWatcher.Namespace}, foundConfigMap)
		if err != nil {
			// If a configMap name is provided, then it must exist
			// You will likely want to create an Event for the user to understand why their reconcile is failing.
			return ctrl.Result{}, err
		}
		log.Info("My config map  looks like", "configmap", foundConfigMap.GetName)
		//log.Info("My config map  looks like", "configmap", foundConfigMap.Data)
		log.Info("My config map  looks like", "configmap", foundConfigMap.ResourceVersion)
		//// Hash the data in some way, or just use the version of the Object
		//configMapVersion = foundConfigMap.ResourceVersion
	}

	log.Info("Searching for:", "label", configWatcher.Spec.Label, "namespace", configWatcher.GetNamespace())
	listOps := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(configWatcher.Spec.Label),
		Namespace:     configWatcher.GetNamespace(),
	}
	pods := &corev1.PodList{}
	r.List(context.TODO(), pods, listOps)

	for i, item := range pods.Items {
		log.Info("Deleting pod", "number", i, "name", item.GetName())
		if err := r.Client.Delete(ctx, &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: item.GetNamespace(),
				Name:      item.GetName(),
			},
		}); err != nil {
			log.Error(err, "Error deleting pod")
		}
	}
	return ctrl.Result{}, nil
}

func (r *LogstashConfigWatcherReconciler) makeConfigWatcherReconcileRequestsForConfigMap(configMap client.Object) []reconcile.Request {
	log := log.Log.WithValues("configmap", configMap.GetName())
	configWatchers := &logstashv1alpha1.LogstashConfigWatcherList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(configMapField, configMap.GetName()),
		Namespace:     configMap.GetNamespace(),
	}
	err := r.List(context.TODO(), configWatchers, listOps)
	if err != nil {
		log.Error(err, "failed to retrieve config watchers")
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(configWatchers.Items))
	for i, item := range configWatchers.Items {
		log.Info("Make reconcile request for ", "number", i, "Item", item.GetName())
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *LogstashConfigWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &logstashv1alpha1.LogstashConfigWatcher{}, configMapField, func(rawObj client.Object) []string {
		logstashConfigWatcher := rawObj.(*logstashv1alpha1.LogstashConfigWatcher)
		if logstashConfigWatcher.Spec.ConfigMap == "" {
			return nil
		}
		return []string{logstashConfigWatcher.Spec.ConfigMap}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&logstashv1alpha1.LogstashConfigWatcher{}).
		Owns(&corev1.Pod{}).
		Watches(
			&source.Kind{Type: &corev1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(r.makeConfigWatcherReconcileRequestsForConfigMap),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}
