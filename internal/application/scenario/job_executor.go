package scenario

//go:generate moq -out job_executor_mock.go . ScenarioJobExecutor

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const RetryInterval = 5 * time.Second

type ScenarioJobExecutor interface {
	Execute(ctx context.Context, namespacedName types.NamespacedName) error
}

type scenarioJobExecutor struct {
	client.Client
}

func NewScenarioJobExecutor(client client.Client) ScenarioJobExecutor {
	return &scenarioJobExecutor{
		Client: client,
	}
}

func (e *scenarioJobExecutor) Execute(ctx context.Context, namespacedName types.NamespacedName) error {
	err := func() error {
		for {
			job := &batchv1.Job{}
			err := e.Get(ctx, namespacedName, job)
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
				err = e.Delete(ctx, job)
				if err != nil {
					return err
				}

				return nil
			case batchv1.JobFailed:
				err = e.Delete(ctx, job)
				if err != nil {
					return err
				}

				return fmt.Errorf("scenario job is failed")
			default:
				time.Sleep(RetryInterval)
			}
		}
	}()

	return err
}
