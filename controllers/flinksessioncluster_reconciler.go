/*
Copyright 2019 Google LLC.

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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type _ClusterReconciler struct {
	k8sClient     client.Client
	context       context.Context
	log           logr.Logger
	observedState _ObservedClusterState
	desiredState  _DesiredClusterState
}

// Compares the desired state and the observed state, if there is a difference,
// takes actions to drive the observed state towards the desired state.
func (reconciler *_ClusterReconciler) reconcile() error {
	var err error

	// Child resources of the cluster CR will be automatically reclaimed by K8S.
	if reconciler.observedState.cluster == nil {
		reconciler.log.Info("The cluster has been deleted, no action to take")
		return nil
	}

	err = reconciler.reconcileJobManagerDeployment()
	if err != nil {
		return err
	}

	err = reconciler.reconcileJobManagerService()
	if err != nil {
		return err
	}

	err = reconciler.reconcileTaskManagerDeployment()
	if err != nil {
		return err
	}

	err = reconciler.reconcileJob()
	if err != nil {
		return err
	}

	return nil
}

func (reconciler *_ClusterReconciler) reconcileJobManagerDeployment() error {
	return reconciler.reconcileDeployment(
		"JobManager",
		reconciler.desiredState.jmDeployment,
		reconciler.observedState.jmDeployment)
}

func (reconciler *_ClusterReconciler) reconcileTaskManagerDeployment() error {
	return reconciler.reconcileDeployment(
		"TaskManager",
		reconciler.desiredState.tmDeployment,
		reconciler.observedState.tmDeployment)
}

func (reconciler *_ClusterReconciler) reconcileDeployment(
	component string,
	desiredDeployment *appsv1.Deployment,
	observedDeployment *appsv1.Deployment) error {
	var log = reconciler.log.WithValues("component", component)
	var err error

	if desiredDeployment != nil && observedDeployment == nil {
		err = reconciler.createDeployment(desiredDeployment, component)
	} else {
		log.Info("Deployment already exists, no action")
		// TODO(dagang): compare and update if needed.
	}
	return err
}

func (reconciler *_ClusterReconciler) createDeployment(
	deployment *appsv1.Deployment, component string) error {
	var context = reconciler.context
	var log = reconciler.log.WithValues("component", component)
	var k8sClient = reconciler.k8sClient

	log.Info("Creating deployment", "deployment", *deployment)
	var err = k8sClient.Create(context, deployment)
	if err != nil {
		log.Error(err, "Failed to create deployment")
	} else {
		log.Info("Deployment created")
	}
	return err
}

func (reconciler *_ClusterReconciler) updateDeployment(
	deployment *appsv1.Deployment, component string) error {
	var context = reconciler.context
	var log = reconciler.log.WithValues("component", component)
	var k8sClient = reconciler.k8sClient

	log.Info("Updating deployment", "deployment", deployment)
	var err = k8sClient.Update(context, deployment)
	if err != nil {
		log.Error(err, "Failed to update deployment")
	} else {
		log.Info("Deployment updated")
	}
	return err
}

func (reconciler *_ClusterReconciler) reconcileJobManagerService() error {
	var err error
	var desiredJmService = reconciler.desiredState.jmService
	var observedJmService = reconciler.observedState.jmService
	if desiredJmService != nil && observedJmService == nil {
		err = reconciler.createService(desiredJmService, "JobManager")
	} else {
		reconciler.log.Info("JobManager service already exists, no action")
		// TODO(dagang): compare and update if needed.
	}
	return err
}

func (reconciler *_ClusterReconciler) createService(
	service *corev1.Service, component string) error {
	var context = reconciler.context
	var log = reconciler.log.WithValues("component", component)
	var k8sClient = reconciler.k8sClient

	log.Info("Creating service", "resource", *service)
	var err = k8sClient.Create(context, service)
	if err != nil {
		log.Info("Failed to create service", "error", err)
	} else {
		log.Info("Service created")
	}
	return err
}

func (reconciler *_ClusterReconciler) reconcileJob() error {
	var log = reconciler.log
	var desiredJob = reconciler.desiredState.job
	var observedJob = reconciler.observedState.job
	var err error
	if desiredJob != nil && observedJob == nil {
		err = reconciler.createJob(desiredJob)
	} else {
		log.Info("Job already exists, no action")
	}
	return err
}

func (reconciler *_ClusterReconciler) createJob(job *batchv1.Job) error {
	var context = reconciler.context
	var log = reconciler.log
	var k8sClient = reconciler.k8sClient

	log.Info("Submitting job", "resource", *job)
	var err = k8sClient.Create(context, job)
	if err != nil {
		log.Info("Failed to created job", "error", err)
	} else {
		log.Info("Job created")
	}
	return err
}