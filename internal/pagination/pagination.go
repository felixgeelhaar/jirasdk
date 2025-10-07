// Package pagination provides pagination utilities for Jira API responses.
package pagination

import (
	"fmt"
	"net/url"
	"strconv"
)

// PageInfo contains pagination metadata.
type PageInfo struct {
	// StartAt is the index of the first item returned
	StartAt int `json:"startAt"`

	// MaxResults is the maximum number of items that could be returned
	MaxResults int `json:"maxResults"`

	// Total is the total number of items available
	Total int `json:"total"`

	// IsLast indicates if this is the last page
	IsLast bool `json:"isLast,omitempty"`
}

// HasNextPage returns true if there are more pages to fetch.
func (p *PageInfo) HasNextPage() bool {
	if p.IsLast {
		return false
	}
	return p.StartAt+p.MaxResults < p.Total
}

// NextStartAt returns the startAt value for the next page.
func (p *PageInfo) NextStartAt() int {
	return p.StartAt + p.MaxResults
}

// Options contains pagination options for list requests.
type Options struct {
	// StartAt is the index of the first item to return (0-based)
	StartAt int

	// MaxResults is the maximum number of items to return per page
	MaxResults int
}

// DefaultMaxResults is the default page size.
const DefaultMaxResults = 50

// MaxMaxResults is the maximum allowed page size.
const MaxMaxResults = 100

// ApplyToURL adds pagination parameters to a URL.
func (o *Options) ApplyToURL(u *url.URL) {
	q := u.Query()

	if o.StartAt > 0 {
		q.Set("startAt", strconv.Itoa(o.StartAt))
	}

	maxResults := o.MaxResults
	if maxResults <= 0 {
		maxResults = DefaultMaxResults
	}
	if maxResults > MaxMaxResults {
		maxResults = MaxMaxResults
	}

	q.Set("maxResults", strconv.Itoa(maxResults))
	u.RawQuery = q.Encode()
}

// Validate validates pagination options.
func (o *Options) Validate() error {
	if o.StartAt < 0 {
		return fmt.Errorf("startAt must be non-negative")
	}

	if o.MaxResults < 0 {
		return fmt.Errorf("maxResults must be non-negative")
	}

	if o.MaxResults > MaxMaxResults {
		return fmt.Errorf("maxResults cannot exceed %d", MaxMaxResults)
	}

	return nil
}

// Iterator provides an iterator over paginated results.
type Iterator[T any] struct {
	// fetchPage is called to fetch the next page of results
	fetchPage func(startAt int) (items []T, pageInfo PageInfo, err error)

	// current page data
	items     []T
	pageInfo  PageInfo
	position  int
	fetchNext bool
}

// NewIterator creates a new pagination iterator.
//
// The fetchPage function should fetch a page of results starting at the given index.
func NewIterator[T any](fetchPage func(startAt int) ([]T, PageInfo, error)) *Iterator[T] {
	return &Iterator[T]{
		fetchPage: fetchPage,
		fetchNext: true,
	}
}

// Next advances the iterator and returns true if there's a next item.
func (it *Iterator[T]) Next() bool {
	// Check if we need to fetch the next page
	if it.fetchNext {
		startAt := 0
		if it.pageInfo.StartAt > 0 || it.pageInfo.MaxResults > 0 {
			startAt = it.pageInfo.NextStartAt()
		}

		items, pageInfo, err := it.fetchPage(startAt)
		if err != nil {
			return false
		}

		it.items = items
		it.pageInfo = pageInfo
		it.position = 0
		it.fetchNext = false

		// If no items returned, we're done
		if len(items) == 0 {
			return false
		}
	}

	// Check if there are more items in the current page
	if it.position >= len(it.items) {
		// Check if there are more pages
		if !it.pageInfo.HasNextPage() {
			return false
		}

		// Fetch next page
		it.fetchNext = true
		return it.Next()
	}

	return true
}

// Item returns the current item.
func (it *Iterator[T]) Item() T {
	return it.items[it.position]
}

// Advance moves to the next item.
func (it *Iterator[T]) Advance() {
	it.position++
}

// Err returns any error encountered during iteration.
func (it *Iterator[T]) Err() error {
	// TODO: Store and return errors from fetchPage
	return nil
}

// PagedResponse is a generic paginated response.
type PagedResponse[T any] struct {
	Items    []T      `json:"values,omitempty"` // Jira uses "values" for most endpoints
	PageInfo PageInfo `json:"-"`                // Embedded pagination info
	StartAt  int      `json:"startAt"`
	MaxResults int    `json:"maxResults"`
	Total    int      `json:"total"`
	IsLast   bool     `json:"isLast,omitempty"`
}

// UnmarshalPageInfo extracts PageInfo from a PagedResponse.
func (r *PagedResponse[T]) UnmarshalPageInfo() PageInfo {
	return PageInfo{
		StartAt:    r.StartAt,
		MaxResults: r.MaxResults,
		Total:      r.Total,
		IsLast:     r.IsLast,
	}
}
