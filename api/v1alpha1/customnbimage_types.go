/*
Copyright 2021, 2022 The Meteor Authors.

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

// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
const (
	CNBiPhasePending            = "Pending"
	CNBiPhaseFailed             = "Failed"
	CNBiPhaseCreatingRepository = "CreatingRepository"
	CNBiPhaseResolving          = "Resolving"
	CNBiPhaseBuilding           = "Building"
	CNBiPhaseOk                 = "Ready"
)

// CustomNBImageRuntimeSpec defines a Runtime Environment, aka 'the Python version used'
type CustomNBImageRuntimeSpec struct {
	// PythonVersion is the version of Python to use
	PythonVersion string `json:"pythonVersion,omitempty"`
	// OSName is the Name of the Operating System to use
	OSName string `json:"osName,omitempty"`
	// OSVersion is the Version of the Operating System to use
	OSVersion string `json:"osVersion,omitempty"`
}

// CustomNBImageSpec defines the desired state of CustomNBImage
type CustomNBImageSpec struct {
	// Name that should be shown in the UI
	DisplayName string `json:"displayName"`
	// Description that should be shown in the UI
	Description string `json:"description,omitempty"`
	// Creator is the name of the user who created the CustomNBImage
	Creator string `json:"creator"`
	// RuntimeEnvironment is the runtime environment to use for the Custome Notebook Image
	RuntimeEnvironment CustomNBImageRuntimeSpec `json:"runtimeEnvironment,omitempty"`
	// PackageVersion is a set of Packages including their Version Specifiers
	PackageVersion []string `json:"packageVersions,omitempty"`
	// An existing image to validate/import, or to use as the base for builds
	BaseImage string `json:"baseImage,omitempty"`
	// Git repository containing source artifacts and build configuration
	RepoURL string `json:"repoURL,omitempty"`
	// Branch or tag or commit reference within the repository.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Branch Reference",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:text"}
	Ref string `json:"ref,omitempty"`
	// Time to live after the resource was created.
	//+optional
	TTL int64 `json:"ttl,omitempty"`
	// List of pipelines to initiate for this meteor
	//+kubebuilder:default={jupyterhub,jupyterbook}
	Pipelines []string `json:"pipelines"`
}

// CustomNBImageStatus defines the observed state of CustomNBImage
type CustomNBImageStatus struct {
	// Current condition of the Shower.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Phase",xDescriptors={"urn:alm:descriptor:io.kubernetes.phase'"}
	//+optional
	Phase string `json:"phase,omitempty"`
	// Current service state of Meteor.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Conditions",xDescriptors={"urn:alm:descriptor:io.kubernetes.conditions"}
	//+optional
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// Stores results from pipelines. Empty if neither pipeline has completed.
	//+optional
	Pipelines []PipelineResult `json:"pipelines,omitempty"`
}

// Aggregate phase from conditions
func (cnbi *CustomNBImage) AggregatePhase() string {
	if len(cnbi.Status.Conditions) == 0 {
		return CNBiPhasePending
	}

	for _, c := range cnbi.Status.Conditions {
		if c.Status == metav1.ConditionFalse {
			return CNBiPhaseFailed
		}

		// Claim ready only if pipelineruns have completed
		if strings.HasPrefix(c.Type, "PipelineRun") {
			switch c.Reason {
			case "Succeeded", "Completed":
				continue
			}
			// TODO distinguish between preparing and building (depending on the pipelinerun name containing 'prepare' or 'build'?)
			return CNBiPhaseBuilding
		}

		if c.Reason != "Ready" {
			return CNBiPhaseBuilding
		}
	}
	return CNBiPhaseOk
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Phase"

// CustomNBImage is the Schema for the customnbimages API
type CustomNBImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomNBImageSpec   `json:"spec,omitempty"`
	Status CustomNBImageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CustomNBImageList contains a list of CustomNBImage
type CustomNBImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomNBImage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomNBImage{}, &CustomNBImageList{})
}

// IsReady returns true the Ready condition status is True
func (status CustomNBImageStatus) IsReady() bool {
	for _, condition := range status.Conditions {
		if condition.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
}
