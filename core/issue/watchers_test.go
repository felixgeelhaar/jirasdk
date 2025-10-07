package issue

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWatchers(t *testing.T) {
	tests := []struct {
		name           string
		issueKeyOrID   string
		responseStatus int
		responseBody   *Watchers
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get",
			issueKeyOrID:   "PROJ-123",
			responseStatus: http.StatusOK,
			responseBody: &Watchers{
				IsWatching: true,
				WatchCount: 2,
				Watchers: []*User{
					{AccountID: "user1", DisplayName: "User One"},
					{AccountID: "user2", DisplayName: "User Two"},
				},
			},
			wantErr: false,
		},
		{
			name:         "empty issue key",
			issueKeyOrID: "",
			wantErr:      true,
			errMsg:       "issue key or ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKeyOrID == "" {
				service := NewService(nil)
				_, err := service.GetWatchers(context.Background(), tt.issueKeyOrID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/"+tt.issueKeyOrID+"/watchers")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			watchers, err := service.GetWatchers(context.Background(), tt.issueKeyOrID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, watchers)
				assert.Equal(t, tt.responseBody.IsWatching, watchers.IsWatching)
				assert.Equal(t, tt.responseBody.WatchCount, watchers.WatchCount)
				assert.Equal(t, len(tt.responseBody.Watchers), len(watchers.Watchers))
			}
		})
	}
}

func TestAddWatcher(t *testing.T) {
	tests := []struct {
		name           string
		issueKeyOrID   string
		accountID      string
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful add",
			issueKeyOrID:   "PROJ-123",
			accountID:      "user123",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:         "empty issue key",
			issueKeyOrID: "",
			accountID:    "user123",
			wantErr:      true,
			errMsg:       "issue key or ID is required",
		},
		{
			name:         "empty account ID",
			issueKeyOrID: "PROJ-123",
			accountID:    "",
			wantErr:      true,
			errMsg:       "account ID is required",
		},
		{
			name:           "unexpected status code",
			issueKeyOrID:   "PROJ-123",
			accountID:      "user123",
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
			errMsg:         "unexpected status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKeyOrID == "" || tt.accountID == "" {
				service := NewService(nil)
				err := service.AddWatcher(context.Background(), tt.issueKeyOrID, tt.accountID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/"+tt.issueKeyOrID+"/watchers")

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.AddWatcher(context.Background(), tt.issueKeyOrID, tt.accountID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRemoveWatcher(t *testing.T) {
	tests := []struct {
		name           string
		issueKeyOrID   string
		accountID      string
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful remove",
			issueKeyOrID:   "PROJ-123",
			accountID:      "user123",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:         "empty issue key",
			issueKeyOrID: "",
			accountID:    "user123",
			wantErr:      true,
			errMsg:       "issue key or ID is required",
		},
		{
			name:         "empty account ID",
			issueKeyOrID: "PROJ-123",
			accountID:    "",
			wantErr:      true,
			errMsg:       "account ID is required",
		},
		{
			name:           "unexpected status code",
			issueKeyOrID:   "PROJ-123",
			accountID:      "user123",
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
			errMsg:         "unexpected status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKeyOrID == "" || tt.accountID == "" {
				service := NewService(nil)
				err := service.RemoveWatcher(context.Background(), tt.issueKeyOrID, tt.accountID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/"+tt.issueKeyOrID+"/watchers")
				assert.Contains(t, r.URL.RawQuery, "accountId="+tt.accountID)

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.RemoveWatcher(context.Background(), tt.issueKeyOrID, tt.accountID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetVotes(t *testing.T) {
	tests := []struct {
		name           string
		issueKeyOrID   string
		responseStatus int
		responseBody   *Votes
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get",
			issueKeyOrID:   "PROJ-123",
			responseStatus: http.StatusOK,
			responseBody: &Votes{
				Votes:    5,
				HasVoted: true,
				Voters: []*User{
					{AccountID: "user1", DisplayName: "User One"},
					{AccountID: "user2", DisplayName: "User Two"},
				},
			},
			wantErr: false,
		},
		{
			name:         "empty issue key",
			issueKeyOrID: "",
			wantErr:      true,
			errMsg:       "issue key or ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKeyOrID == "" {
				service := NewService(nil)
				_, err := service.GetVotes(context.Background(), tt.issueKeyOrID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/"+tt.issueKeyOrID+"/votes")

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			votes, err := service.GetVotes(context.Background(), tt.issueKeyOrID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, votes)
				assert.Equal(t, tt.responseBody.Votes, votes.Votes)
				assert.Equal(t, tt.responseBody.HasVoted, votes.HasVoted)
				assert.Equal(t, len(tt.responseBody.Voters), len(votes.Voters))
			}
		})
	}
}

func TestAddVote(t *testing.T) {
	tests := []struct {
		name           string
		issueKeyOrID   string
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful add",
			issueKeyOrID:   "PROJ-123",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:         "empty issue key",
			issueKeyOrID: "",
			wantErr:      true,
			errMsg:       "issue key or ID is required",
		},
		{
			name:           "unexpected status code",
			issueKeyOrID:   "PROJ-123",
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
			errMsg:         "unexpected status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKeyOrID == "" {
				service := NewService(nil)
				err := service.AddVote(context.Background(), tt.issueKeyOrID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/"+tt.issueKeyOrID+"/votes")

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.AddVote(context.Background(), tt.issueKeyOrID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRemoveVote(t *testing.T) {
	tests := []struct {
		name           string
		issueKeyOrID   string
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful remove",
			issueKeyOrID:   "PROJ-123",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:         "empty issue key",
			issueKeyOrID: "",
			wantErr:      true,
			errMsg:       "issue key or ID is required",
		},
		{
			name:           "unexpected status code",
			issueKeyOrID:   "PROJ-123",
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
			errMsg:         "unexpected status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issueKeyOrID == "" {
				service := NewService(nil)
				err := service.RemoveVote(context.Background(), tt.issueKeyOrID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issue/"+tt.issueKeyOrID+"/votes")

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.RemoveVote(context.Background(), tt.issueKeyOrID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
