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

package controller

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	threatestergithubiov1alpha1 "github.com/mrtc0/threatester/api/v1alpha1"
	"github.com/mrtc0/threatester/internal/application/expectation"
	scenarioApplication "github.com/mrtc0/threatester/internal/application/scenario"
)

const (
	scenarioFinalizer = "threatester.github.io/finalizers"

	typeAvailableScenario   = "Available"
	typeSucceededScenario   = "Succeeded"
	typeFailedScenario      = "Failed"
	typeProgressingScenario = "Progressing"
	typeDegradedScenario    = "Degraded"
)

// ScenarioReconciler reconciles a Scenario object
type ScenarioReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	ExpectationService  expectation.ExpectationService
	ScenarioJobExecutor scenarioApplication.ScenarioJobExecutor
}

//+kubebuilder:rbac:groups=threatester.github.io,resources=scenarios,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=threatester.github.io,resources=scenarios/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=threatester.github.io,resources=scenarios/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Scenario object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ScenarioReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// check if the CRD for kind Scenario exists
	scenario := &threatestergithubiov1alpha1.Scenario{}
	err := r.Get(ctx, req.NamespacedName, scenario)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("scenario resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}

		log.Error(err, "failed to get scenario")
		return ctrl.Result{}, err
	}

	// update status as Unknown when no status are avaiable
	if scenario.Status.Conditions == nil || len(scenario.Status.Conditions) == 0 {
		err = r.updateScenarioStatus(ctx, req, metav1.Condition{Type: typeProgressingScenario, Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting Reconciling"})
		if err != nil {
			log.Error(err, "failed to update scenario status")
			return ctrl.Result{}, err
		}
	}

	err = r.Get(ctx, req.NamespacedName, scenario)
	if err != nil {
		log.Error(err, "failed to fe-fetch scenario")
		return ctrl.Result{}, err
	}

	// Add Finalaizer
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers
	if !controllerutil.ContainsFinalizer(scenario, scenarioFinalizer) {
		log.Info("Adding Finalizer")
		if ok := controllerutil.AddFinalizer(scenario, scenarioFinalizer); !ok {
			log.Error(err, "Failed to add finalizer into the custom resource")
			return ctrl.Result{Requeue: true}, err
		}

		if err = r.Update(ctx, scenario); err != nil {
			log.Error(err, "Failed to update custom resource with finalizer")
			return ctrl.Result{}, err
		}

		if err := r.Get(ctx, req.NamespacedName, scenario); err != nil {
			log.Error(err, "failed to fe-fetch scenario")
			return ctrl.Result{}, err
		}
	}

	// TODO: refactor
	isScenarioMarkedToBeDeleted := scenario.GetDeletionTimestamp() != nil
	if isScenarioMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(scenario, scenarioFinalizer) {
			log.Info("Performing finalizer operation for scenario")

			err = r.updateScenarioStatus(ctx, req, metav1.Condition{Type: typeAvailableScenario, Status: metav1.ConditionUnknown, Reason: "Finalizing", Message: fmt.Sprintf("Performing finalizer operation for %s", scenario.Name)})
			if err != nil {
				log.Error(err, "failed to update scenario status")
				return ctrl.Result{}, err
			}

			// TODO: remove all scneario jobs

			err = r.updateScenarioStatus(ctx, req, metav1.Condition{Type: typeAvailableScenario, Status: metav1.ConditionTrue, Reason: "Finalizing", Message: fmt.Sprintf("Successfully finalizer operation %s", scenario.Name)})
			if err != nil {
				log.Error(err, "failed to update scenario status")
				return ctrl.Result{}, err
			}

			log.Info("Removing finalizer for scenario after successfully perform the operations")
			err := r.Get(ctx, req.NamespacedName, scenario)
			if err != nil {
				log.Error(err, "failed to fe-fetch scenario")
				return ctrl.Result{}, err
			}

			if ok := controllerutil.RemoveFinalizer(scenario, scenarioFinalizer); !ok {
				log.Error(err, "failed to remove finalizer for schenario")
				return ctrl.Result{Requeue: true}, err
			}

			if err := r.Update(ctx, scenario); err != nil {
				log.Error(err, "failed to remove finalizer for schenario")
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	if scenario.Status.Status == typeSucceededScenario {
		log.Info(fmt.Sprintf("scenario %s/%s is already succeeded. skip.", scenario.Namespace, scenario.Name))
		return ctrl.Result{}, nil
	}

	scenarioJob, err := scenarioApplication.NewScenarioJobBuilder().WithNamespace(req.Namespace).WithScenarioJobs(scenario.Spec.Templates).Build()
	if err != nil {
		log.Error(err, "failed to build scenario job")
		return ctrl.Result{}, err
	}

	found := &batchv1.Job{}
	err = r.Get(ctx, types.NamespacedName{Name: scenarioJob.Name, Namespace: scenarioJob.Namespace}, found)

	if err == nil {
		log.Info("scenario job already exists. skip.")
		return ctrl.Result{}, nil
	}

	err = r.ScenarioJobExecutor.Execute(ctx, *scenarioJob)
	defer r.ScenarioJobExecutor.DeleteScenarioJob(ctx, *scenarioJob)

	if err != nil {
		log.Error(err, "failed to execute scenario job")
		err := r.updateScenarioStatus(ctx, req, metav1.Condition{Type: typeFailedScenario, Status: metav1.ConditionFalse, Reason: "Failed", Message: err.Error()})
		if err != nil {
			log.Error(err, "failed update scenario status")
		}

		return ctrl.Result{}, err
	}

	log.Info("Perform sceario expectation")

	// TODO: sleep timeout seconds

	r.ExpectationService.SetExpectations(scenario.Spec.Expectations)

	// TODO: Run all expectations and get the results
	passed, err := r.ExpectationService.RunExpectation(ctx)
	if err != nil {
		log.Error(err, "failed to run expectation")
		err := r.updateScenarioStatus(ctx, req, metav1.Condition{Type: typeFailedScenario, Status: metav1.ConditionTrue, Reason: "Failed", Message: err.Error()})
		if err != nil {
			log.Error(err, "failed update scenario status")
		}
		return ctrl.Result{}, err
	}

	log.Info("scenario expectation is success")

	duration, _ := time.ParseDuration("1m") // TODO
	result := threatestergithubiov1alpha1.ExpectationResult{
		Passed:   passed,
		Duration: duration, // TODO
		SucceededExpectations: []threatestergithubiov1alpha1.Expectation{
			scenario.Spec.Expectations[0],
		}, // TODO
		FailedExpectations: []threatestergithubiov1alpha1.FailedExpectation{}, // TODO

	}
	_, err = r.updateScenarioExpectationResultStatus(ctx, req, result)
	if err != nil {
		log.Error(err, "failed update scenario expectation result")
		return ctrl.Result{}, err
	}

	err = r.updateScenarioStatus(ctx, req, metav1.Condition{Type: typeSucceededScenario, Status: metav1.ConditionTrue, Reason: "Success", Message: "Successfully run scenario expectations"})
	if err != nil {
		log.Error(err, "failed update scenario status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ScenarioReconciler) updateScenarioStatus(ctx context.Context, req reconcile.Request, condition metav1.Condition) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		scenario := &threatestergithubiov1alpha1.Scenario{}

		if err := r.Get(ctx, req.NamespacedName, scenario); err != nil {
			return err
		}

		scenario.Status.Status = condition.Type
		meta.SetStatusCondition(&scenario.Status.Conditions, condition)

		return r.Status().Update(ctx, scenario)
	})

	return err
}

func (r *ScenarioReconciler) updateScenarioExpectationResultStatus(ctx context.Context, req reconcile.Request, expectationResult threatestergithubiov1alpha1.ExpectationResult) (*threatestergithubiov1alpha1.Scenario, error) {
	scenario := &threatestergithubiov1alpha1.Scenario{}

	if err := r.Get(ctx, req.NamespacedName, scenario); err != nil {
		return nil, err
	}

	scenario.Status.Result = expectationResult
	if err := r.Status().Update(ctx, scenario); err != nil {
		return nil, err
	}

	return scenario, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScenarioReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&threatestergithubiov1alpha1.Scenario{}).
		Complete(r)
}
