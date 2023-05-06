package datadog

import (
	"context"

	dd "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	ddv1 "github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

type DatadogClient interface {
	GetMonitor(ctx context.Context, monitorID int64) (*ddv1.Monitor, error)
}

type datadogClient struct {
	client *dd.APIClient
}

func NewDatadogClient() DatadogClient {
	configuration := dd.NewConfiguration()
	client := dd.NewAPIClient(configuration)

	return datadogClient{client: client}
}

func (d datadogClient) GetMonitor(ctx context.Context, monitorID int64) (*ddv1.Monitor, error) {
	ddCtx := dd.NewDefaultContext(ctx)
	api := ddv1.NewMonitorsApi(d.client)

	resp, _, err := api.GetMonitor(ddCtx, monitorID, *ddv1.NewGetMonitorOptionalParameters())
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
