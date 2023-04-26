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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	threatestergithubiov1alpha1 "github.com/mrtc0/threatester/api/v1alpha1"
	scenarioDomain "github.com/mrtc0/threatester/internal/domain/scenario"
)

// ScenarioReconciler reconciles a Scenario object
type ScenarioReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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

	scenario := &threatestergithubiov1alpha1.Scenario{}
	err := r.Get(ctx, req.NamespacedName, scenario)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("scenario resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
	}

	scenarioJob, err := scenarioDomain.NewScenarioJobBuilder().WithNamespace(req.Namespace).WithScenarioJobs(scenario.Spec.Templates).Build()
	if err != nil {
		log.Error(err, "failed to build scenario job")
		return ctrl.Result{}, err
	}

	err = r.Create(ctx, scenarioJob)
	if err != nil {
		log.Error(err, "failed to create scenario job")
		return ctrl.Result{}, err
	}

	err = func() error {
		for {
			job := &batchv1.Job{}
			err := r.Get(ctx, types.NamespacedName{Name: scenarioJob.Name, Namespace: scenarioJob.Namespace}, job)
			if err != nil {
				if apierrors.IsNotFound(err) {
					log.Error(err, "failed to get scenario job")
					time.Sleep(5 * time.Second)
					return nil
				}

				return fmt.Errorf("failed to get scenario job: %w", err)
			}

			// TODO: Update Scenario Status
			if len(job.Status.Conditions) == 0 {
				log.Info("scenario job is not running")
				time.Sleep(5 * time.Second)
				continue
			}

			switch jobCondition := job.Status.Conditions[0].Type; jobCondition {
			case batchv1.JobComplete:
				log.Info("scenario job is completed")
				return nil
			case batchv1.JobFailed:
				return fmt.Errorf("scenario job is failed")
			default:
				log.Info("scenario job is maybe running")
				time.Sleep(5 * time.Second)
			}
		}
	}()

	// TODO: Delete Scenario Job

	if err != nil {
		// TODO: Update Scenario Status
		return ctrl.Result{}, err
	}

	log.Info("test expectation")
	// TODO: Update Scenario Status
	// TODO: Perfoem expectation

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScenarioReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&threatestergithubiov1alpha1.Scenario{}).
		Complete(r)
}
