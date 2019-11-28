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
	"context"
	"fmt"
	tmv1beta1 "github.com/gardener/test-infra/pkg/apis/testmachinery/v1beta1"
	trerrors "github.com/gardener/test-infra/pkg/testrunner/error"
	"github.com/gardener/test-infra/pkg/util"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
)

// GetTestruns returns all testruns of a RunList as testrun array
func (rl RunList) GetTestruns() []*tmv1beta1.Testrun {
	testruns := make([]*tmv1beta1.Testrun, len(rl))
	for i, run := range rl {
		if run != nil {
			testruns[i] = run.Testrun
		}
	}
	return testruns
}

// HasErrors checks whether one run in list is erroneous.
func (rl RunList) HasErrors() bool {
	for _, run := range rl {
		if run.Error != nil {
			return true
		}
	}
	return false
}

// Errors returns all errors of all testruns in this testrun
func (rl RunList) Errors() error {
	var res *multierror.Error
	for _, run := range rl {
		if run.Error != nil {
			res = multierror.Append(res, run.Error)
		}
	}
	return util.ReturnMultiError(res)
}

func (r *Run) Exec(log logr.Logger, config *Config, prefix string) {
	ctx := context.Background()
	defer ctx.Done()
	newTR := r.Testrun.DeepCopy()

	// Remove legacy name attribute. Instead enforce usage of generateName.
	newTR.Name = ""
	newTR.GenerateName = prefix
	newTR.Namespace = config.Namespace
	err := config.Watch.Client().Create(ctx, newTR)
	if err != nil {
		log.Error(err, "unable to create testrun")
		r.Error = trerrors.NewNotCreatedError(fmt.Sprintf("cannot create testrun: %s", err.Error()))
		return
	}

	*r.Testrun = *newTR
	r.Metadata.Testrun.ID = newTR.GetName()
	log.Info(fmt.Sprintf("Testrun %s deployed", newTR.Name))

	if argoUrl, err := GetArgoURL(config.Watch.Client(), r.Testrun); err == nil {
		log.WithValues("testrun", r.Testrun.GetName()).Info(fmt.Sprintf("Argo workflow: %s", argoUrl))
	}

	testrunPhase := tmv1beta1.PhaseStatusInit
	err = config.Watch.WatchUntil(config.Timeout, r.Testrun.GetNamespace(), r.Testrun.GetName(), func(new *tmv1beta1.Testrun) (bool, error) {
		*r.Testrun = *new
		if r.Testrun.Status.State != "" {
			testrunPhase = r.Testrun.Status.Phase
			log.Info(fmt.Sprintf("Testrun %s is in %s phase. State: %s", r.Testrun.GetName(), testrunPhase, r.Testrun.Status.State))
		} else {
			log.Info(fmt.Sprintf("Testrun %s is in %s phase. Waiting ...", r.Testrun.GetName(), testrunPhase))
		}
		return util.Completed(testrunPhase), nil
	})
	if err != nil {
		r.Testrun.Status.Phase = tmv1beta1.PhaseStatusTimeout
		r.Error = trerrors.NewTimeoutError(fmt.Sprintf("maximum wait time of %d is exceeded by Testrun %s", config.Timeout, r.Testrun.GetName()))
	}

	fmt.Println(RunList{r}.RenderTable())
}
