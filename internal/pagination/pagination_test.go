package pagination

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPageInfoHasNextPage(t *testing.T) {
	tests := []struct {
		name     string
		pageInfo PageInfo
		expected bool
	}{
		{
			name: "has next page",
			pageInfo: PageInfo{
				StartAt:    0,
				MaxResults: 50,
				Total:      100,
				IsLast:     false,
			},
			expected: true,
		},
		{
			name: "last page",
			pageInfo: PageInfo{
				StartAt:    50,
				MaxResults: 50,
				Total:      100,
				IsLast:     false,
			},
			expected: false,
		},
		{
			name: "marked as last",
			pageInfo: PageInfo{
				StartAt:    0,
				MaxResults: 50,
				Total:      100,
				IsLast:     true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pageInfo.HasNextPage()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPageInfoNextStartAt(t *testing.T) {
	tests := []struct {
		name     string
		pageInfo PageInfo
		expected int
	}{
		{
			name: "first page",
			pageInfo: PageInfo{
				StartAt:    0,
				MaxResults: 50,
			},
			expected: 50,
		},
		{
			name: "second page",
			pageInfo: PageInfo{
				StartAt:    50,
				MaxResults: 50,
			},
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pageInfo.NextStartAt()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOptionsApplyToURL(t *testing.T) {
	tests := []struct {
		name           string
		options        Options
		expectedParams map[string]string
	}{
		{
			name:    "default options",
			options: Options{},
			expectedParams: map[string]string{
				"maxResults": "50",
			},
		},
		{
			name: "custom options",
			options: Options{
				StartAt:    100,
				MaxResults: 25,
			},
			expectedParams: map[string]string{
				"startAt":    "100",
				"maxResults": "25",
			},
		},
		{
			name: "exceeds max",
			options: Options{
				MaxResults: 200,
			},
			expectedParams: map[string]string{
				"maxResults": "100",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse("https://example.com/api")
			require.NoError(t, err)

			tt.options.ApplyToURL(u)

			for key, expected := range tt.expectedParams {
				actual := u.Query().Get(key)
				assert.Equal(t, expected, actual, "parameter %s", key)
			}
		})
	}
}

func TestOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		options Options
		wantErr bool
	}{
		{
			name:    "valid options",
			options: Options{StartAt: 0, MaxResults: 50},
			wantErr: false,
		},
		{
			name:    "negative startAt",
			options: Options{StartAt: -1},
			wantErr: true,
		},
		{
			name:    "negative maxResults",
			options: Options{MaxResults: -1},
			wantErr: true,
		},
		{
			name:    "exceeds max",
			options: Options{MaxResults: 200},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIterator(t *testing.T) {
	// Mock data
	allItems := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	pageSize := 3

	fetchPage := func(startAt int) ([]string, PageInfo, error) {
		end := startAt + pageSize
		if end > len(allItems) {
			end = len(allItems)
		}

		items := allItems[startAt:end]
		isLast := end >= len(allItems)

		return items, PageInfo{
			StartAt:    startAt,
			MaxResults: pageSize,
			Total:      len(allItems),
			IsLast:     isLast,
		}, nil
	}

	iterator := NewIterator(fetchPage)

	var collected []string
	for iterator.Next() {
		collected = append(collected, iterator.Item())
		iterator.Advance()
	}

	assert.Equal(t, allItems, collected)
}
