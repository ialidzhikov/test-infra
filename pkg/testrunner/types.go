// Copyright 2019 Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testrunner

import (
	argov1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/gardener/test-infra/pkg/testrunner/componentdescriptor"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SummaryType defines the type of a test result or summary
type SummaryType string

// Summary types can be testrun or teststep
const (
	SummaryTypeTestrun  SummaryType = "testrun"
	SummaryTypeTeststep SummaryType = "teststep"
)

// TestrunParameters are the parameters which describe the test that is executed by the testrunner.
type TestrunParameters struct {
	// Path to the kubeconfig where the gardener is running.
	GardenKubeconfigPath string
	TestrunName          string
	TestrunChartPath     string

	ProjectName             string
	ShootName               string
	Landscape               string
	Cloudprovider           string
	Cloudprofile            string
	SecretBinding           string
	Region                  string
	Zone                    string
	K8sVersion              string
	MachineType             string
	AutoscalerMin           string
	AutoscalerMax           string
	FloatingPoolName        string
	ComponentDescriptorPath string
}

// TestrunConfig are configuration of the evironment like the testmachinery cluster or S3 store
// where the testrunner executes the testrun.
type TestrunConfig struct {
	// Path to the kubeconfig where the testmachinery is running.
	TmKubeconfigPath string

	// Max wait time for a testrun to finish.
	Timeout *int64

	// outputFilePath is the path where the testresult is written to.
	OutputFile string

	// config name of the elastic search to store the test results.
	ESConfigName string

	// Endpint of the s3 storage of the testmachinery.
	S3Endpoint string

	// Path to the error directory of concourse to put the notify.cfg in.
	ConcourseOnErrorDir string
}

// Metadata is the common metadata of all ouputs and summaries.
type Metadata struct {
	// Landscape describes the current dev,staging,canary,office or live.
	Landscape         string `json:"landscape"`
	CloudProvider     string `json:"cloudprovider"`
	KubernetesVersion string `json:"kubernetes_version"`

	// BOM describes the current component_descriptor of the direct landscape-setup components.
	BOM       []*componentdescriptor.Component `json:"bom"`
	TestrunID string                           `json:"testrun_id"`
}

// StepExportMetadata is the metadata of one step of a testrun.
type StepExportMetadata struct {
	Metadata
	TestDefName string           `json:"testdefinition"`
	Phase       argov1.NodePhase `json:"phase,omitempty"`
	StartTime   *metav1.Time     `json:"startTime,omitempty"`
	Duration    int64            `json:"duration,omitempty"`
}

// TestrunSummary is the result of the overall testrun.
type TestrunSummary struct {
	Metadata  *Metadata        `json:"tm_meta"`
	Type      SummaryType      `json:"type"`
	Phase     argov1.NodePhase `json:"phase,omitempty"`
	StartTime *metav1.Time     `json:"startTime,omitempty"`
	Duration  int64            `json:"duration,omitempty"`
	TestsRun  int              `json:"testsRun,omitempty"`
}

// StepSummary is the result of a specific step.
type StepSummary struct {
	Metadata  *Metadata        `json:"tm_meta"`
	Type      SummaryType      `json:"type"`
	Name      string           `json:"name,omitempty"`
	Phase     argov1.NodePhase `json:"phase,omitempty"`
	StartTime *metav1.Time     `json:"startTime,omitempty"`
	Duration  int64            `json:"duration,omitempty"`
}

// notificationConfig is the configuration that is used by concourse to send notifications.
type notificationCfg struct {
	Email email `yaml:"email"`
}

type email struct {
	Subject    string   `yaml:"subject"`
	Recipients []string `yaml:"recipients"`
	MailBody   string   `yaml:"mail_body"`
}
