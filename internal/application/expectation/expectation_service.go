package expectation

//go:generate moq -out expectation_service_mock.go . ExpectationService

import (
	"context"
	"fmt"

	threatestergithubiov1alpha1 "github.com/mrtc0/threatester/api/v1alpha1"
)

type ExpectationService interface {
	RunExpectation(ctx context.Context) (bool, error)
	SetExpectations(expectations []threatestergithubiov1alpha1.Expectation)
}

type expectationService struct {
	Expectations       []threatestergithubiov1alpha1.Expectation
	datadogExpectation DatadogExpectation
}

func NewExpectationService() ExpectationService {
	return &expectationService{
		datadogExpectation: NewDatadogExpectation(),
	}
}

func (e *expectationService) RunExpectation(ctx context.Context) (bool, error) {
	for _, expect := range e.Expectations {
		if expect.Datadog != nil {
			return e.datadogExpectation.RunExpectation(ctx, *expect.Datadog)
		}
	}

	return false, fmt.Errorf("expectation not found")
}

func (e *expectationService) SetExpectations(expectations []threatestergithubiov1alpha1.Expectation) {
	e.Expectations = expectations
}
