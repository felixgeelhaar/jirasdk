package issue

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateIssueLink(t *testing.T) {
	tests := []struct {
		name           string
		input          *CreateIssueLinkInput
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name: "successful link creation",
			input: &CreateIssueLinkInput{
				Type:         BlocksLinkType(),
				InwardIssue:  &IssueRef{Key: "PROJ-123"},
				OutwardIssue: &IssueRef{Key: "PROJ-456"},
			},
			responseStatus: http.StatusCreated,
			wantErr:        false,
		},
		{
			name: "link with comment",
			input: &CreateIssueLinkInput{
				Type:         DuplicatesLinkType(),
				InwardIssue:  &IssueRef{Key: "PROJ-123"},
				OutwardIssue: &IssueRef{Key: "PROJ-789"},
				Comment: &LinkComment{
					Body: "These issues are duplicates",
				},
			},
			responseStatus: http.StatusCreated,
			wantErr:        false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
			errMsg:  "create issue link input is required",
		},
		{
			name: "missing link type",
			input: &CreateIssueLinkInput{
				InwardIssue:  &IssueRef{Key: "PROJ-123"},
				OutwardIssue: &IssueRef{Key: "PROJ-456"},
			},
			wantErr: true,
			errMsg:  "link type is required",
		},
		{
			name: "missing inward issue",
			input: &CreateIssueLinkInput{
				Type:         BlocksLinkType(),
				OutwardIssue: &IssueRef{Key: "PROJ-456"},
			},
			wantErr: true,
			errMsg:  "inward issue is required",
		},
		{
			name: "missing outward issue",
			input: &CreateIssueLinkInput{
				Type:        BlocksLinkType(),
				InwardIssue: &IssueRef{Key: "PROJ-123"},
			},
			wantErr: true,
			errMsg:  "outward issue is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == nil || tt.input.Type == nil || tt.input.InwardIssue == nil || tt.input.OutwardIssue == nil {
				service := NewService(nil)
				err := service.CreateIssueLink(context.Background(), tt.input)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issueLink")

				// Verify request body
				var body CreateIssueLinkInput
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, tt.input.Type.Name, body.Type.Name)
				assert.Equal(t, tt.input.InwardIssue.Key, body.InwardIssue.Key)
				assert.Equal(t, tt.input.OutwardIssue.Key, body.OutwardIssue.Key)

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.CreateIssueLink(context.Background(), tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeleteIssueLink(t *testing.T) {
	tests := []struct {
		name           string
		linkID         string
		responseStatus int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful delete",
			linkID:         "10000",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:           "successful delete with 200 OK",
			linkID:         "10001",
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:    "empty link ID",
			linkID:  "",
			wantErr: true,
			errMsg:  "link ID is required",
		},
		{
			name:           "not found",
			linkID:         "99999",
			responseStatus: http.StatusNotFound,
			wantErr:        true,
			errMsg:         "unexpected status code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.linkID == "" {
				service := NewService(nil)
				err := service.DeleteIssueLink(context.Background(), tt.linkID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issueLink/"+tt.linkID)

				w.WriteHeader(tt.responseStatus)
			})
			defer transport.Close()

			service := NewService(transport)
			err := service.DeleteIssueLink(context.Background(), tt.linkID)

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

func TestGetIssueLinkType(t *testing.T) {
	tests := []struct {
		name           string
		linkTypeID     string
		responseStatus int
		responseBody   *IssueLinkType
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "successful get",
			linkTypeID:     "10000",
			responseStatus: http.StatusOK,
			responseBody: &IssueLinkType{
				ID:      "10000",
				Name:    "Blocks",
				Inward:  "is blocked by",
				Outward: "blocks",
			},
			wantErr: false,
		},
		{
			name:       "empty link type ID",
			linkTypeID: "",
			wantErr:    true,
			errMsg:     "link type ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.linkTypeID == "" {
				service := NewService(nil)
				_, err := service.GetIssueLinkType(context.Background(), tt.linkTypeID)
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Contains(t, r.URL.Path, "/rest/api/3/issueLinkType/"+tt.linkTypeID)

				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			})
			defer transport.Close()

			service := NewService(transport)
			linkType, err := service.GetIssueLinkType(context.Background(), tt.linkTypeID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, linkType)
				assert.Equal(t, tt.responseBody.ID, linkType.ID)
				assert.Equal(t, tt.responseBody.Name, linkType.Name)
				assert.Equal(t, tt.responseBody.Inward, linkType.Inward)
				assert.Equal(t, tt.responseBody.Outward, linkType.Outward)
			}
		})
	}
}

func TestListIssueLinkTypes(t *testing.T) {
	linkTypes := []*IssueLinkType{
		{
			ID:      "10000",
			Name:    "Blocks",
			Inward:  "is blocked by",
			Outward: "blocks",
		},
		{
			ID:      "10001",
			Name:    "Duplicate",
			Inward:  "is duplicated by",
			Outward: "duplicates",
		},
		{
			ID:      "10002",
			Name:    "Relates",
			Inward:  "relates to",
			Outward: "relates to",
		},
	}

	transport := newMockTransport(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/rest/api/3/issueLinkType")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"issueLinkTypes": linkTypes,
		})
	})
	defer transport.Close()

	service := NewService(transport)
	result, err := service.ListIssueLinkTypes(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result, 3)
	assert.Equal(t, "Blocks", result[0].Name)
	assert.Equal(t, "Duplicate", result[1].Name)
	assert.Equal(t, "Relates", result[2].Name)
}

func TestLinkTypeHelpers(t *testing.T) {
	t.Run("BlocksLinkType", func(t *testing.T) {
		linkType := BlocksLinkType()
		assert.Equal(t, "Blocks", linkType.Name)
		assert.Equal(t, "is blocked by", linkType.Inward)
		assert.Equal(t, "blocks", linkType.Outward)
	})

	t.Run("DuplicatesLinkType", func(t *testing.T) {
		linkType := DuplicatesLinkType()
		assert.Equal(t, "Duplicate", linkType.Name)
		assert.Equal(t, "is duplicated by", linkType.Inward)
		assert.Equal(t, "duplicates", linkType.Outward)
	})

	t.Run("RelatesToLinkType", func(t *testing.T) {
		linkType := RelatesToLinkType()
		assert.Equal(t, "Relates", linkType.Name)
		assert.Equal(t, "relates to", linkType.Inward)
		assert.Equal(t, "relates to", linkType.Outward)
	})

	t.Run("CausesLinkType", func(t *testing.T) {
		linkType := CausesLinkType()
		assert.Equal(t, "Causation", linkType.Name)
		assert.Equal(t, "is caused by", linkType.Inward)
		assert.Equal(t, "causes", linkType.Outward)
	})

	t.Run("ClonesLinkType", func(t *testing.T) {
		linkType := ClonesLinkType()
		assert.Equal(t, "Cloners", linkType.Name)
		assert.Equal(t, "is cloned by", linkType.Inward)
		assert.Equal(t, "clones", linkType.Outward)
	})
}
