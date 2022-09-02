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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LogstashConfigWatcherSpec defines the desired state of LogstashConfigWatcher
type LogstashConfigWatcherSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name of configMap to watch
	ConfigMap string `json:"configMap"`

	// label(s) to restart pods with
	Label map[string]string `json:"label"`
}

// LogstashConfigWatcherStatus defines the observed state of LogstashConfigWatcher
type LogstashConfigWatcherStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LogstashConfigWatcher is the Schema for the logstashconfigwatchers API
type LogstashConfigWatcher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LogstashConfigWatcherSpec   `json:"spec,omitempty"`
	Status LogstashConfigWatcherStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LogstashConfigWatcherList contains a list of LogstashConfigWatcher
type LogstashConfigWatcherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LogstashConfigWatcher `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LogstashConfigWatcher{}, &LogstashConfigWatcherList{})
}
