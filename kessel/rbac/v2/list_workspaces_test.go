package v2

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		consistency          *v1beta2.Consistency
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
			consistency:          nil,
			expectedError:        false,
			expectedRequestCount: 1,
			validateRequests: func(t *testing.T, requests []*v1beta2.StreamedListObjectsRequest) {
				require.Len(t, requests, 1)
				req := requests[0]
				assert.Equal(t, "member", req.Relation)
				assert.Equal(t, "workspace", req.ObjectType.ResourceType)
				assert.Equal(t, "rbac", *req.ObjectType.ReporterType)
			},
		},
		{
			name: "handles pagination with continuation token",
			responses: []*v1beta2.StreamedListObjectsResponse{
				{Pagination: &v1beta2.ResponsePagination{ContinuationToken: "next-page-token"}},
			},
			relation:             "viewer",
			continuationToken:    "",
			consistency:          nil,
			expectedError:        false,
			expectedRequestCount: 2,
			validateRequests: func(t *testing.T, requests []*v1beta2.StreamedListObjectsRequest) {
				require.Len(t, requests, 2)

				firstReq := requests[0]
				if firstReq.Pagination != nil && firstReq.Pagination.ContinuationToken != nil {
					assert.Empty(t, *firstReq.Pagination.ContinuationToken)
				}

				secondReq := requests[1]
				require.NotNil(t, secondReq.Pagination)
				require.NotNil(t, secondReq.Pagination.ContinuationToken)
				assert.Equal(t, "next-page-token", *secondReq.Pagination.ContinuationToken)
			},
		},
		{
			name: "stops when no continuation token",
			responses: []*v1beta2.StreamedListObjectsResponse{
				{Pagination: &v1beta2.ResponsePagination{ContinuationToken: ""}},
			},
			relation:             "admin",
			continuationToken:    "",
			consistency:          nil,
			expectedError:        false,
			expectedRequestCount: 1,
			validateRequests: func(t *testing.T, requests []*v1beta2.StreamedListObjectsRequest) {
				assert.Len(t, requests, 1)
			},
		},
		{
			name:                 "handles stream errors",
			responses:            nil,
			streamErr:            &mockStreamError{message: "stream failed"},
			relation:             "member",
			continuationToken:    "",
			consistency:          nil,
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
			consistency:          nil,
			expectedError:        false,
			expectedRequestCount: 1,
			validateRequests: func(t *testing.T, requests []*v1beta2.StreamedListObjectsRequest) {
				require.Len(t, requests, 1)
				req := requests[0]
				require.NotNil(t, req.Pagination)
				require.NotNil(t, req.Pagination.ContinuationToken)
				assert.Equal(t, "resume-from-here", *req.Pagination.ContinuationToken)
			},
		},
		{
			name: "passes consistency to every request",
			responses: []*v1beta2.StreamedListObjectsResponse{
				{Pagination: &v1beta2.ResponsePagination{ContinuationToken: ""}},
			},
			relation:             "member",
			continuationToken:    "",
			consistency:          &v1beta2.Consistency{Requirement: &v1beta2.Consistency_MinimizeLatency{MinimizeLatency: true}},
			expectedError:        false,
			expectedRequestCount: 1,
			validateRequests: func(t *testing.T, requests []*v1beta2.StreamedListObjectsRequest) {
				require.Len(t, requests, 1)
				req := requests[0]
				require.NotNil(t, req.Consistency)
				ml, ok := req.Consistency.Requirement.(*v1beta2.Consistency_MinimizeLatency)
				require.True(t, ok)
				assert.True(t, ml.MinimizeLatency)
			},
		},
		{
			name: "nil consistency is fine",
			responses: []*v1beta2.StreamedListObjectsResponse{
				{Pagination: &v1beta2.ResponsePagination{ContinuationToken: ""}},
			},
			relation:             "member",
			continuationToken:    "",
			consistency:          nil,
			expectedError:        false,
			expectedRequestCount: 1,
			validateRequests: func(t *testing.T, requests []*v1beta2.StreamedListObjectsRequest) {
				require.Len(t, requests, 1)
				assert.Nil(t, requests[0].Consistency)
			},
		},
		{
			name: "consistency preserved across paginated pages",
			responses: []*v1beta2.StreamedListObjectsResponse{
				{Pagination: &v1beta2.ResponsePagination{ContinuationToken: "page-2-token"}},
			},
			relation:             "view",
			continuationToken:    "",
			consistency:          &v1beta2.Consistency{Requirement: &v1beta2.Consistency_MinimizeLatency{MinimizeLatency: true}},
			expectedError:        false,
			expectedRequestCount: 2,
			validateRequests: func(t *testing.T, requests []*v1beta2.StreamedListObjectsRequest) {
				require.Len(t, requests, 2)
				for i, req := range requests {
					require.NotNil(t, req.Consistency, "request %d should have consistency set", i)
					ml, ok := req.Consistency.Requirement.(*v1beta2.Consistency_MinimizeLatency)
					require.True(t, ok, "request %d should have MinimizeLatency requirement", i)
					assert.True(t, ml.MinimizeLatency, "request %d MinimizeLatency should be true", i)
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
			for _, err := range ListWorkspaces(context.Background(), mockClient, subject, tt.relation, tt.continuationToken, tt.consistency) {
				if err != nil {
					iterationErr = err
					break
				}
			}

			if tt.expectedError {
				assert.Error(t, iterationErr)
			} else {
				assert.NoError(t, iterationErr)
			}

			if tt.validateRequests != nil {
				tt.validateRequests(t, mockClient.capturedRequests)
			}

			if tt.expectedRequestCount > 0 {
				assert.Len(t, mockClient.capturedRequests, tt.expectedRequestCount)
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
