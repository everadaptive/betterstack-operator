/*
Copyright 2023.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type MonitorConfig struct {
	SSLExpiration       int                  `json:"sslExpiration,omitempty"`
	DomainExpiration    int                  `json:"domainExpiration,omitempty"`
	PolicyID            string               `json:"policyId,omitempty"`
	URL                 string               `json:"url,omitempty"`
	MonitorType         string               `json:"monitorType,omitempty"`
	RequiredKeyword     string               `json:"requiredKeyword,omitempty"`
	ExpectedStatusCodes *[]int               `json:"expectedStatus_codes,omitempty"`
	Call                bool                 `json:"call,omitempty"`
	SMS                 bool                 `json:"sms,omitempty"`
	Email               bool                 `json:"email,omitempty"`
	Push                bool                 `json:"push,omitempty"`
	TeamWait            int                  `json:"teamWait,omitempty"`
	Paused              bool                 `json:"paused,omitempty"`
	PausedAt            string               `json:"pausedAt,omitempty"`
	FollowRedirects     bool                 `json:"followRedirects,omitempty"`
	Port                string               `json:"port,omitempty"`
	Regions             *[]string            `json:"regions,omitempty"`
	MonitorGroupID      int                  `json:"monitorGroupId,omitempty"`
	PronounceableName   string               `json:"pronounceableName,omitempty"`
	RecoveryPeriod      int                  `json:"recoveryPeriod,omitempty"`
	VerifySSL           bool                 `json:"verifySSL,omitempty"`
	CheckFrequency      int                  `json:"checkFrequency,omitempty"`
	ConfirmationPeriod  int                  `json:"confirmationPeriod,omitempty"`
	HTTPMethod          string               `json:"httpMethod,omitempty"`
	RequestTimeout      int                  `json:"requestTimeout,omitempty"`
	RequestBody         string               `json:"requestBody,omitempty"`
	RequestHeaders      *[]map[string]string `json:"requestHeaders,omitempty"`
	AuthUsername        string               `json:"authUsername,omitempty"`
	AuthPassword        string               `json:"authPassword,omitempty"`
	MaintenanceFrom     string               `json:"maintenanceFrom,omitempty"`
	MaintenanceTo       string               `json:"maintenanceTo,omitempty"`
	MaintenanceTimezone string               `json:"maintenanceTimezone,omitempty"`
	RememberCookies     bool                 `json:"rememberCookies,omitempty"`
	LastCheckedAt       string               `json:"lastCheckedAt,omitempty"`
	Status              string               `json:"status,omitempty"`
	CreatedAt           string               `json:"createdAt,omitempty"`
	UpdatedAt           string               `json:"updatedAt,omitempty"`
}

// IngressMonitorSpec defines the desired state of IngressMonitor
type IngressMonitorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of IngressMonitor. Edit ingressmonitor_types.go to remove/update
	IngressName   string        `json:"ingressName,omitempty"`
	MonitorConfig MonitorConfig `json:"monitorConfig,omitempty"`
}

type MonitorStatus struct {
	ID          string `json:"id"`
	Name        string `json:"pronounceableName"`
	Paused      bool   `json:"paused"`
	MonitorType string `json:"monitorType"`
}

type MonitorGroup struct {
	ID   string `json:"id"`
	Name string `json:"pronounceableName"`
}

// IngressMonitorStatus defines the observed state of IngressMonitor
type IngressMonitorStatus struct {
	Monitors     []MonitorStatus `json:"monitors"`
	MonitorGroup MonitorGroup    `json:"monitorGroup"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// IngressMonitor is the Schema for the ingressmonitors API
type IngressMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IngressMonitorSpec   `json:"spec,omitempty"`
	Status IngressMonitorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IngressMonitorList contains a list of IngressMonitor
type IngressMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IngressMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IngressMonitor{}, &IngressMonitorList{})
}
