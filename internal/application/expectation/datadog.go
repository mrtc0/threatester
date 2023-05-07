package expectation

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	threatestergithubiov1alpha1 "github.com/mrtc0/threatester/api/v1alpha1"
	"github.com/mrtc0/threatester/internal/service/datadog"
)

type DatadogExpectation struct {
	datadogClient datadog.DatadogClient
	expectation   threatestergithubiov1alpha1.DatadogExpectation
}

func NewDatadogExpectation() DatadogExpectation {
	ddExpectation := DatadogExpectation{}
	ddExpectation.datadogClient = datadog.NewDatadogClient()

	return ddExpectation
}

func (e *DatadogExpectation) RunExpectation(ctx context.Context, expectation threatestergithubiov1alpha1.DatadogExpectation) (bool, error) {
	e.expectation = expectation

	if expectation.Monitor != nil {
		return e.ExpectMonitorState(ctx, expectation.Monitor.Status)
	}

	return false, fmt.Errorf("datadog expectation not found")
}

func (e *DatadogExpectation) ExpectMonitorState(ctx context.Context, expectState string) (bool, error) {
	monitorID, err := strconv.ParseInt(e.expectation.Monitor.ID, 10, 64)
	if err != nil {
		return false, err
	}

	resp, err := e.datadogClient.GetMonitor(ctx, monitorID)
	if err != nil {
		return false, err
	}

	actualState := resp.GetOverallState()
	if actualState != datadogV1.MonitorOverallStates(expectState) {
		return false, fmt.Errorf("monitor %d state is not %s, got %s", monitorID, expectState, actualState)
	}

	return true, nil
}
