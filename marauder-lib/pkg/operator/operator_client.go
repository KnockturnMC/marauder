package operator

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/models/networkmodel"
	"github.com/knockturnmc/marauder/marauder-lib/pkg/utils"
)

// None is a sample implementation for the Client#DoHTTPRequest methods configurator.
func None(_ *http.Request) error {
	return nil
}

// The Client is responsible for interacting with the controller from the operator side.
type Client interface {
	// DoHTTPRequest runs a http request against the given path on the operator.
	DoHTTPRequest(
		ctx context.Context,
		method string,
		path string,
		body io.Reader,
		configurator func(r *http.Request) error,
	) (*http.Response, error)

	// ExecuteLifecycleAction executes a lifecycle action o the specific server on the operator
	ExecuteLifecycleAction(
		ctx context.Context,
		serverUUID uuid.UUID,
		action networkmodel.LifecycleAction,
	) error

	// ScheduleCacheClear schedules the clearing of the caches on the operator for any cachable item older than the passed age.
	ScheduleCacheClear(ctx context.Context, age time.Duration) error
}

// HTTPClient implements the Client interface by using the operators rest API.
type HTTPClient struct {
	*http.Client
	OperatorURL string
}

func (c HTTPClient) DoHTTPRequest(
	ctx context.Context,
	method string,
	path string,
	body io.Reader,
	configurator func(r *http.Request) error,
) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s/v1%s", c.OperatorURL, path), body)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request for operator: %w", err)
	}

	if err := configurator(request); err != nil {
		return nil, fmt.Errorf("failed to execute configurator: %w", err)
	}

	response, err := c.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}

	return response, nil
}

func (c HTTPClient) ExecuteLifecycleAction(
	ctx context.Context,
	serverUUID uuid.UUID,
	action networkmodel.LifecycleAction,
) error {
	response, err := c.DoHTTPRequest(
		ctx,
		http.MethodPost,
		fmt.Sprintf("/server/%s/%s", serverUUID.String(), action),
		&bytes.Buffer{},
		None,
	)
	if err != nil {
		return fmt.Errorf("failed to create http request for lifecycle actions: %w", err)
	}

	defer func() { _ = response.Body.Close() }()

	if err := utils.IsOkayStatusCodeOrErrorWithBody(response); err != nil {
		return fmt.Errorf("failed to execute lifecycle action: %w", err)
	}

	return nil
}

func (c HTTPClient) ScheduleCacheClear(ctx context.Context, age time.Duration) error {
	response, err := utils.PerformHTTPRequest(
		ctx,
		c.Client,
		http.MethodPost,
		fmt.Sprintf("%s/v1/cron/cache/clear?age=%s", c.OperatorURL, age.String()),
		"application/json",
		&bytes.Buffer{},
	)
	if err != nil {
		return fmt.Errorf("failed to perform http request: %w", err)
	}

	defer func() { _ = response.Body.Close() }()

	if !utils.IsOkayStatusCode(response.StatusCode) {
		return utils.ErrStatusCodeUnrecoverable
	}

	return nil
}
