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
	"fmt"
	"io/ioutil"
	"time"

	"github.com/gardener/gardener/pkg/chartrenderer"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tmv1beta1 "github.com/gardener/test-infra/pkg/apis/testmachinery/v1beta1"
	tmclientset "github.com/gardener/test-infra/pkg/client/testmachinery/clientset/versioned"
	"github.com/gardener/test-infra/pkg/testrunner/componentdescriptor"
	"github.com/gardener/test-infra/pkg/util"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func addBOMLocationsToTestrun(tr *tmv1beta1.Testrun, componenets []*componentdescriptor.Component) {
	if tr == nil || componenets == nil {
		return
	}
	for _, component := range componenets {
		tr.Spec.TestLocations = append(tr.Spec.TestLocations, tmv1beta1.TestLocation{
			Type:     tmv1beta1.LocationTypeGit,
			Repo:     fmt.Sprintf("https://%s", component.Name),
			Revision: component.Version,
		})
	}
}

func runTestrun(tmClient *tmclientset.Clientset, tr *tmv1beta1.Testrun, parameters *TestrunParameters) (*tmv1beta1.Testrun, error) {
	// TODO: Remove legacy name attribute. Instead enforce usage of generateName.
	tr.Name = ""
	tr.GenerateName = parameters.TestrunName
	tr, err := tmClient.Testmachinery().Testruns(namespace).Create(tr)
	if err != nil {
		return nil, fmt.Errorf("Cannot create testrun: %s", err.Error())
	}
	log.Infof("Testrun %s deployed", tr.Name)

	testrunPhase := tmv1beta1.PhaseStatusInit
	interval := time.Duration(pollIntervalSeconds) * time.Second
	timeout := time.Duration(maxWaitTimeSeconds) * time.Second
	err = wait.PollImmediate(interval, timeout, func() (bool, error) {
		tr, err = tmClient.Testmachinery().Testruns(namespace).Get(tr.Name, metav1.GetOptions{})
		if err != nil {
			log.Errorf("Cannot get testrun: %s", err.Error())
		}

		if tr.Status.Phase != "" {
			testrunPhase = tr.Status.Phase
		}
		if tr.Status.State != "" {
			log.Infof("Testrun %s is in %s phase. State: %s", tr.Name, testrunPhase, tr.Status.State)
		} else {
			log.Infof("Testrun %s is in %s phase. Waiting ...", tr.Name, testrunPhase)
		}
		return util.Completed(testrunPhase), nil
	})
	if err != nil {
		return nil, fmt.Errorf("Maximum wait time of %d is exceeded by Testrun %s", maxWaitTimeSeconds, parameters.TestrunName)
	}

	return tr, nil
}

func renderChart(config *TestrunConfig, parameters *TestrunParameters) (*chartrenderer.RenderedChart, error) {
	log.Infof("Render chart from %s", parameters.TestrunChartPath)

	tmClusterClient, err := kubernetes.NewClientFromFile(config.TmKubeconfigPath, nil, client.Options{})
	if err != nil {
		return nil, fmt.Errorf("couldn't create k8s client from kubeconfig filepath %s: %v", config.TmKubeconfigPath, err)
	}
	tmChartRenderer, err := chartrenderer.New(tmClusterClient)
	if err != nil {
		return nil, fmt.Errorf("Cannot create chartrenderer for gardener  %s", err.Error())
	}

	gardenKubeconfig, err := ioutil.ReadFile(parameters.GardenKubeconfigPath)
	if err != nil {
		log.Fatalf("Cannot read gardener kubeconfig %s, Error: %s", parameters.GardenKubeconfigPath, err.Error())
	}

	return tmChartRenderer.Render(parameters.TestrunChartPath, parameters.TestrunName, namespace, map[string]interface{}{
		"shoot": map[string]interface{}{
			"name":             fmt.Sprintf("%s-%s", parameters.ShootName, util.RandomString(5)),
			"projectNamespace": fmt.Sprintf("garden-%s", parameters.ProjectName),
			"cloudprovider":    parameters.Cloudprovider,
			"cloudprofile":     parameters.Cloudprofile,
			"secretBinding":    parameters.SecretBinding,
			"region":           parameters.Region,
			"zone":             parameters.Zone,
			"k8sVersion":       parameters.K8sVersion,
			"machinetype":      parameters.MachineType,
			"autoscalerMin":    parameters.AutoscalerMin,
			"autoscalerMax":    parameters.AutoscalerMax,
			"floatingPoolName": parameters.FloatingPoolName,
		},
		"kubeconfigs": map[string]interface{}{
			"gardener": string(gardenKubeconfig),
		},
	})
}
