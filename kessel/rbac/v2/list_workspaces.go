package v2

import (
	"context"
	"fmt"
	"io"
	"iter"

	v1beta2 "github.com/project-kessel/kessel-sdk-go/kessel/inventory/v1beta2"
)

func ListWorkspaces(
	ctx context.Context,
	inventory v1beta2.KesselInventoryServiceClient,
	subject *v1beta2.SubjectReference,
	relation string,
	continuationToken string,
) iter.Seq2[*v1beta2.StreamedListObjectsResponse, error] {
	return func(yield func(*v1beta2.StreamedListObjectsResponse, error) bool) {
		for {
			var pagination *v1beta2.RequestPagination
			if continuationToken != "" {
				pagination = &v1beta2.RequestPagination{
					Limit:             1000,
					ContinuationToken: &continuationToken,
				}
			}

			request := &v1beta2.StreamedListObjectsRequest{
				ObjectType: WorkspaceType(),
				Relation:   relation,
				Subject:    subject,
				Pagination: pagination,
			}

			stream, err := inventory.StreamedListObjects(ctx, request)
			if err != nil {
				yield(nil, fmt.Errorf("failed to start stream: %w", err))
				return
			}

			var lastToken string
			for {
				response, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					yield(nil, fmt.Errorf("error receiving from stream: %w", err))
					return
				}

				if !yield(response, nil) {
					return
				}

				if response.Pagination != nil {
					lastToken = response.Pagination.ContinuationToken
				}
			}

			if lastToken == "" {
				break
			}
			continuationToken = lastToken
		}
	}
}
