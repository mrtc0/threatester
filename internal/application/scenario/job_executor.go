package scenario

//go:generate moq -rm -out job_executor_mock.go . ScenarioJobExecutor

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const RetryInterval = 5 * time.Second

var (
	defaultDeletePropagation = metav1.DeletePropagationBackground
)

type ScenarioJobExecutor interface {
	Execute(ctx context.Context, scenarioJob batchv1.Job) error
	DeleteScenarioJob(ctx context.Context, scenarioJob batchv1.Job) error
}

type scenarioJobExecutor struct {
	client.Client
}

func NewScenarioJobExecutor(client client.Client) ScenarioJobExecutor {
	return &scenarioJobExecutor{
		Client: client,
	}
}

func (e *scenarioJobExecutor) Execute(ctx context.Context, scenarioJob batchv1.Job) error {
	err := e.Create(ctx, &scenarioJob)
	if err != nil {
		return err
	}

	err = func() error {
		for {
			job := &batchv1.Job{}
			err := e.Get(ctx, types.NamespacedName{Name: scenarioJob.Name, Namespace: scenarioJob.Namespace}, job)
			if err != nil {
				if apierrors.IsNotFound(err) {
					return err
				}

				return fmt.Errorf("failed to get scenario job: %w", err)
			}

			if len(job.Status.Conditions) == 0 {
				time.Sleep(RetryInterval)
				continue
			}

			switch jobCondition := job.Status.Conditions[0].Type; jobCondition {
			case batchv1.JobComplete:
				return nil
			case batchv1.JobFailed:
				return fmt.Errorf("scenario job is failed")
			default:
				time.Sleep(RetryInterval)
			}
		}
	}()

	return err
}

func (e *scenarioJobExecutor) DeleteScenarioJob(ctx context.Context, job batchv1.Job) error {
	log := log.FromContext(ctx)

	err := e.Client.Delete(ctx, &job, &client.DeleteOptions{PropagationPolicy: &defaultDeletePropagation})
	if err != nil {
		log.Error(err, fmt.Sprintf("failed to delete scenario job %s/%s", job.Namespace, job.Name))
		return fmt.Errorf("failed to delete scenario job: %w", err)
	}

	log.Info(fmt.Sprintf("scenario job %s/%s deleted", job.Namespace, job.Name))
	return nil
}
