package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/project-kessel/kessel-sdk-go/kessel/auth"
)

const workspaceEndpoint = "/api/rbac/v2/workspaces/"

type Workspace struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type FetchWorkspaceOptions struct {
	// Optionally specify an http.Client or use http.DefaultClient
	HttpClient *http.Client
	Auth       auth.AuthRequest
}

type workspaceAPIResponse struct {
	Data []Workspace `json:"data"`
}

func fetchWorkspace(ctx context.Context, rbacBaseEndpoint string, orgId string, workspaceType string, options FetchWorkspaceOptions) (*Workspace, error) {
	httpClient := options.HttpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	url := strings.TrimRight(rbacBaseEndpoint, "/") + workspaceEndpoint

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	query := request.URL.Query()
	query.Set("type", workspaceType)
	request.URL.RawQuery = query.Encode()

	request.Header.Set("x-rh-rbac-org-id", orgId)

	if options.Auth != nil {
		err = options.Auth.ConfigureRequest(ctx, request)
		if err != nil {
			return nil, err
		}
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching %s workspace - http status %s", workspaceType, response.Status)
	}

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var workspaceResponse workspaceAPIResponse

	err = json.Unmarshal(body, &workspaceResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	if len(workspaceResponse.Data) != 1 {
		return nil, fmt.Errorf("unexpected number of %s workspaces: %d. %v", workspaceType, len(workspaceResponse.Data), workspaceResponse.Data)
	}

	return &workspaceResponse.Data[0], nil
}

func FetchRootWorkspace(ctx context.Context, rbacBaseEndpoint string, orgId string, options FetchWorkspaceOptions) (*Workspace, error) {
	return fetchWorkspace(ctx, rbacBaseEndpoint, orgId, "root", options)
}

func FetchDefaultWorkspace(ctx context.Context, rbacBaseEndpoint string, orgId string, options FetchWorkspaceOptions) (*Workspace, error) {
	return fetchWorkspace(ctx, rbacBaseEndpoint, orgId, "default", options)
}
