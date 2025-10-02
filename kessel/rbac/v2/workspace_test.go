package v2

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/project-kessel/kessel-sdk-go/kessel/auth"
)

func TestFetchDefaultWorkspace(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler func(w http.ResponseWriter, r *http.Request)
		expectedError bool
		expectedWS    *Workspace
		rbacEndpoint  string
		orgId         string
		validateReq   func(t *testing.T, r *http.Request)
	}{
		{
			name: "successful default workspace fetch",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				response := workspaceAPIResponse{
					Data: []Workspace{
						{
							Id:          "default-ws-123",
							Name:        "Default Workspace",
							Type:        "default",
							Description: "Organization default workspace",
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			},
			expectedError: false,
			expectedWS: &Workspace{
				Id:          "default-ws-123",
				Name:        "Default Workspace",
				Type:        "default",
				Description: "Organization default workspace",
			},
			rbacEndpoint: "",
			orgId:        "org123",
			validateReq: func(t *testing.T, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				if !strings.Contains(r.URL.Path, workspaceEndpoint) {
					t.Errorf("Expected URL to contain %s, got %s", workspaceEndpoint, r.URL.Path)
				}
				if r.URL.Query().Get("type") != "default" {
					t.Errorf("Expected type=default in query, got %s", r.URL.Query().Get("type"))
				}
				if r.Header.Get("x-rh-rbac-org-id") != "org123" {
					t.Errorf("Expected org ID header org123, got %s", r.Header.Get("x-rh-rbac-org-id"))
				}
			},
		},
		{
			name: "server error response",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			},
			expectedError: true,
			rbacEndpoint:  "",
			orgId:         "org123",
		},
		{
			name: "empty workspace response",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				response := workspaceAPIResponse{Data: []Workspace{}}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			},
			expectedError: true,
			rbacEndpoint:  "",
			orgId:         "org123",
		},
		{
			name: "multiple workspaces returned",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				response := workspaceAPIResponse{
					Data: []Workspace{
						{Id: "ws1", Name: "WS1", Type: "default"},
						{Id: "ws2", Name: "WS2", Type: "default"},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			},
			expectedError: true,
			rbacEndpoint:  "",
			orgId:         "org123",
		},
		{
			name: "rbac endpoint with trailing slash",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				response := workspaceAPIResponse{
					Data: []Workspace{{Id: "ws1", Name: "WS1", Type: "default"}},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			},
			expectedError: false,
			rbacEndpoint:  "http://example.com/",
			orgId:         "org123",
			expectedWS:    &Workspace{Id: "ws1", Name: "WS1", Type: "default"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.validateReq != nil {
					tt.validateReq(t, r)
				}
				tt.serverHandler(w, r)
			}))
			defer server.Close()

			endpoint := tt.rbacEndpoint
			if endpoint == "" {
				endpoint = server.URL
			} else {
				endpoint = strings.Replace(endpoint, "http://example.com", server.URL, 1)
			}

			result, err := FetchDefaultWorkspace(context.Background(), endpoint, tt.orgId, FetchWorkspaceOptions{
				HttpClient: http.DefaultClient,
			})

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected workspace but got nil")
				} else if tt.expectedWS != nil {
					if result.Id != tt.expectedWS.Id {
						t.Errorf("Expected workspace ID %s, got %s", tt.expectedWS.Id, result.Id)
					}
					if result.Name != tt.expectedWS.Name {
						t.Errorf("Expected workspace name %s, got %s", tt.expectedWS.Name, result.Name)
					}
					if result.Type != tt.expectedWS.Type {
						t.Errorf("Expected workspace type %s, got %s", tt.expectedWS.Type, result.Type)
					}
				}
			}
		})
	}
}

func TestFetchRootWorkspace(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler func(w http.ResponseWriter, r *http.Request)
		expectedError bool
		expectedWS    *Workspace
		validateReq   func(t *testing.T, r *http.Request)
	}{
		{
			name: "successful root workspace fetch",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				response := workspaceAPIResponse{
					Data: []Workspace{
						{
							Id:          "root-ws-456",
							Name:        "Root Workspace",
							Type:        "root",
							Description: "Organization root workspace",
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			},
			expectedError: false,
			expectedWS: &Workspace{
				Id:          "root-ws-456",
				Name:        "Root Workspace",
				Type:        "root",
				Description: "Organization root workspace",
			},
			validateReq: func(t *testing.T, r *http.Request) {
				if r.URL.Query().Get("type") != "root" {
					t.Errorf("Expected type=root in query, got %s", r.URL.Query().Get("type"))
				}
			},
		},
		{
			name: "unauthorized error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.validateReq != nil {
					tt.validateReq(t, r)
				}
				tt.serverHandler(w, r)
			}))
			defer server.Close()

			result, err := FetchRootWorkspace(context.Background(), server.URL, "org123", FetchWorkspaceOptions{
				HttpClient: http.DefaultClient,
			})

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected workspace but got nil")
				} else if tt.expectedWS != nil {
					if result.Id != tt.expectedWS.Id {
						t.Errorf("Expected workspace ID %s, got %s", tt.expectedWS.Id, result.Id)
					}
					if result.Type != tt.expectedWS.Type {
						t.Errorf("Expected workspace type %s, got %s", tt.expectedWS.Type, result.Type)
					}
				}
			}
		})
	}
}

func TestFetchWorkspace_WithAuth(t *testing.T) {
	tests := []struct {
		name          string
		setupAuth     func() auth.AuthRequest
		serverHandler func(w http.ResponseWriter, r *http.Request)
		expectedError bool
		validateAuth  func(t *testing.T, r *http.Request)
	}{
		{
			name: "successful request with auth",
			setupAuth: func() auth.AuthRequest {
				return &mockAuthRequest{token: "test-bearer-token"}
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				response := workspaceAPIResponse{
					Data: []Workspace{{Id: "ws1", Name: "WS1", Type: "default"}},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			},
			expectedError: false,
			validateAuth: func(t *testing.T, r *http.Request) {
				authHeader := r.Header.Get("authorization")
				if authHeader != "Bearer test-bearer-token" {
					t.Errorf("Expected authorization header 'Bearer test-bearer-token', got '%s'", authHeader)
				}
			},
		},
		{
			name: "auth request fails",
			setupAuth: func() auth.AuthRequest {
				return &mockAuthRequest{shouldFail: true}
			},
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				t.Error("Server handler should not be called when auth fails")
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.validateAuth != nil {
					tt.validateAuth(t, r)
				}
				tt.serverHandler(w, r)
			}))
			defer server.Close()

			_, err := FetchDefaultWorkspace(context.Background(), server.URL, "org123", FetchWorkspaceOptions{
				HttpClient: http.DefaultClient,
				Auth:       tt.setupAuth(),
			})

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestFetchWorkspace_DefaultHttpClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := workspaceAPIResponse{
			Data: []Workspace{{Id: "ws1", Name: "WS1", Type: "default"}},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Test with nil HttpClient (should use default)
	_, err := FetchDefaultWorkspace(context.Background(), server.URL, "org123", FetchWorkspaceOptions{
		HttpClient: nil, // This should use http.DefaultClient
	})

	if err != nil {
		t.Errorf("Unexpected error with nil HttpClient: %v", err)
	}
}

func TestFetchWorkspace_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json response"))
	}))
	defer server.Close()

	_, err := FetchDefaultWorkspace(context.Background(), server.URL, "org123", FetchWorkspaceOptions{
		HttpClient: http.DefaultClient,
	})

	if err == nil {
		t.Error("Expected error due to invalid JSON response")
	}
	if !strings.Contains(err.Error(), "unmarshalling") {
		t.Errorf("Expected unmarshalling error, got: %v", err)
	}
}

func TestFetchWorkspace_URLConstruction(t *testing.T) {
	tests := []struct {
		name         string
		rbacEndpoint string
		expectedPath string
	}{
		{
			name:         "endpoint without trailing slash",
			rbacEndpoint: "http://example.com",
			expectedPath: workspaceEndpoint,
		},
		{
			name:         "endpoint with trailing slash",
			rbacEndpoint: "http://example.com/",
			expectedPath: workspaceEndpoint,
		},
		{
			name:         "endpoint with multiple trailing slashes",
			rbacEndpoint: "http://example.com///",
			expectedPath: workspaceEndpoint,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.URL.Path, tt.expectedPath) {
					t.Errorf("Expected URL path to contain %s, got %s", tt.expectedPath, r.URL.Path)
				}
				response := workspaceAPIResponse{
					Data: []Workspace{{Id: "ws1", Name: "WS1", Type: "default"}},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			endpoint := strings.Replace(tt.rbacEndpoint, "http://example.com", server.URL, 1)
			_, err := FetchDefaultWorkspace(context.Background(), endpoint, "org123", FetchWorkspaceOptions{
				HttpClient: http.DefaultClient,
			})

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Mock auth request for testing
type mockAuthRequest struct {
	token      string
	shouldFail bool
}

func (m *mockAuthRequest) ConfigureRequest(ctx context.Context, request *http.Request) error {
	if m.shouldFail {
		return &mockAuthError{message: "auth failed"}
	}
	request.Header.Set("authorization", "Bearer "+m.token)
	return nil
}

type mockAuthError struct {
	message string
}

func (e *mockAuthError) Error() string {
	return e.message
}
