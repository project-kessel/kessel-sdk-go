package v2

import (
	"context"
	"io"
	"testing"

	"google.golang.org/grpc"

	v1beta2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
)

type mockInventoryClient struct {
	v1beta2.KesselInventoryServiceClient
	responses        []*v1beta2.StreamedListObjectsResponse
	err              error
	capturedRequests []*v1beta2.StreamedListObjectsRequest
}

type mockStream struct {
	grpc.ServerStreamingClient[v1beta2.StreamedListObjectsResponse]
	responses []*v1beta2.StreamedListObjectsResponse
	index     int
}

func (m *mockStream) Recv() (*v1beta2.StreamedListObjectsResponse, error) {
	if m.index >= len(m.responses) {
		return nil, io.EOF
	}
	resp := m.responses[m.index]
	m.index++
	return resp, nil
}

func (m *mockInventoryClient) StreamedListObjects(ctx context.Context, in *v1beta2.StreamedListObjectsRequest, opts ...grpc.CallOption) (v1beta2.KesselInventoryService_StreamedListObjectsClient, error) {
	m.capturedRequests = append(m.capturedRequests, in)

	if m.err != nil {
		return nil, m.err
	}

	if in.Pagination != nil && in.Pagination.ContinuationToken != nil && *in.Pagination.ContinuationToken != "" {
		return &mockStream{
			responses: []*v1beta2.StreamedListObjectsResponse{},
			index:     0,
		}, nil
	}

	return &mockStream{
		responses: m.responses,
		index:     0,
	}, nil
}

func TestListWorkspaces(t *testing.T) {
	tests := []struct {
		name                 string
		responses            []*v1beta2.StreamedListObjectsResponse
		streamErr            error
		relation             string
		continuationToken    string
		expectedError        bool
		expectedRequestCount int
		validateRequests     func(t *testing.T, requests []*v1beta2.StreamedListObjectsRequest)
	}{
		{
			name: "builds request with correct parameters",
			responses: []*v1beta2.StreamedListObjectsResponse{
				{Pagination: &v1beta2.ResponsePagination{ContinuationToken: ""}},
			},
			relation:             "member",
			continuationToken:    "",
			expectedError:        false,
			expectedRequestCount: 1,
			validateRequests: func(t *testing.T, requests []*v1beta2.StreamedListObjectsRequest) {
				if len(requests) != 1 {
					t.Errorf("Expected 1 request, got %d", len(requests))
					return
				}
				req := requests[0]
				if req.Relation != "member" {
					t.Errorf("Expected relation 'member', got '%s'", req.Relation)
				}
				if req.ObjectType == nil {
					t.Error("Expected ObjectType to be set")
					return
				}
				if req.ObjectType.ResourceType != "workspace" {
					t.Errorf("Expected ResourceType 'workspace', got '%s'", req.ObjectType.ResourceType)
				}
				if req.ObjectType.ReporterType == nil || *req.ObjectType.ReporterType != "rbac" {
					t.Error("Expected ReporterType 'rbac'")
				}
			},
		},
		{
			name: "handles pagination with continuation token",
			responses: []*v1beta2.StreamedListObjectsResponse{
				{Pagination: &v1beta2.ResponsePagination{ContinuationToken: "next-page-token"}},
			},
			relation:             "viewer",
			continuationToken:    "",
			expectedError:        false,
			expectedRequestCount: 2,
			validateRequests: func(t *testing.T, requests []*v1beta2.StreamedListObjectsRequest) {
				if len(requests) != 2 {
					t.Errorf("Expected 2 requests (initial + continuation), got %d", len(requests))
					return
				}

				firstReq := requests[0]
				if firstReq.Pagination != nil && firstReq.Pagination.ContinuationToken != nil && *firstReq.Pagination.ContinuationToken != "" {
					t.Errorf("First request should have empty continuation token, got '%s'", *firstReq.Pagination.ContinuationToken)
				}

				secondReq := requests[1]
				if secondReq.Pagination == nil {
					t.Error("Second request should have pagination")
					return
				}
				if secondReq.Pagination.ContinuationToken == nil {
					t.Error("Second request should have continuation token")
					return
				}
				if *secondReq.Pagination.ContinuationToken != "next-page-token" {
					t.Errorf("Expected continuation token 'next-page-token', got '%s'", *secondReq.Pagination.ContinuationToken)
				}
			},
		},
		{
			name: "stops when no continuation token",
			responses: []*v1beta2.StreamedListObjectsResponse{
				{Pagination: &v1beta2.ResponsePagination{ContinuationToken: ""}},
			},
			relation:             "admin",
			continuationToken:    "",
			expectedError:        false,
			expectedRequestCount: 1,
			validateRequests: func(t *testing.T, requests []*v1beta2.StreamedListObjectsRequest) {
				if len(requests) != 1 {
					t.Errorf("Expected only 1 request, got %d", len(requests))
				}
			},
		},
		{
			name:                 "handles stream errors",
			responses:            nil,
			streamErr:            &mockStreamError{message: "stream failed"},
			relation:             "member",
			continuationToken:    "",
			expectedError:        true,
			expectedRequestCount: 1,
			validateRequests:     nil,
		},
		{
			name: "uses provided continuation token",
			responses: []*v1beta2.StreamedListObjectsResponse{
				{Pagination: &v1beta2.ResponsePagination{ContinuationToken: ""}},
			},
			relation:             "member",
			continuationToken:    "resume-from-here",
			expectedError:        false,
			expectedRequestCount: 1,
			validateRequests: func(t *testing.T, requests []*v1beta2.StreamedListObjectsRequest) {
				if len(requests) != 1 {
					t.Errorf("Expected 1 request, got %d", len(requests))
					return
				}
				req := requests[0]
				if req.Pagination == nil {
					t.Error("Request should have pagination")
					return
				}
				if req.Pagination.ContinuationToken == nil {
					t.Error("Request should have continuation token")
					return
				}
				if *req.Pagination.ContinuationToken != "resume-from-here" {
					t.Errorf("Expected continuation token 'resume-from-here', got '%s'", *req.Pagination.ContinuationToken)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockInventoryClient{
				responses: tt.responses,
				err:       tt.streamErr,
			}

			subject := PrincipalSubject("user123", "redhat")

			var iterationErr error
			for _, err := range ListWorkspaces(context.Background(), mockClient, subject, tt.relation, tt.continuationToken) {
				if err != nil {
					iterationErr = err
					break
				}
			}

			if tt.expectedError {
				if iterationErr == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if iterationErr != nil {
					t.Errorf("Unexpected error: %v", iterationErr)
				}
			}

			if tt.validateRequests != nil {
				tt.validateRequests(t, mockClient.capturedRequests)
			}

			if tt.expectedRequestCount > 0 && len(mockClient.capturedRequests) != tt.expectedRequestCount {
				t.Errorf("Expected %d requests, got %d", tt.expectedRequestCount, len(mockClient.capturedRequests))
			}
		})
	}
}

type mockStreamError struct {
	message string
}

func (e *mockStreamError) Error() string {
	return e.message
}
