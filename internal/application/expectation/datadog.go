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

func NewDatadogExpectation(
	ddExpecation threatestergithubiov1alpha1.DatadogExpectation,
) DatadogExpectation {
	expectation := DatadogExpectation{}
	expectation.datadogClient = datadog.NewDatadogClient()
	expectation.expectation = ddExpecation

	return expectation
}

func (e DatadogExpectation) RunExpectation(ctx context.Context) (bool, error) {
	if e.expectation.Monitor != nil {
		return e.ExpectMonitorState(ctx, e.expectation.Monitor.Status)
	}

	return false, fmt.Errorf("no expectation found")
}

func (e DatadogExpectation) ExpectMonitorState(ctx context.Context, expectState string) (bool, error) {
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
