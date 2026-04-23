package internal

import (
	"fmt"
	"net/url"
	"strconv"
)

const (
	MinPageSize = 1
	MaxPageSize = 100
)

func NormalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func NormalizePageSize(size int) (int, error) {
	if size <= 0 {
		return 20, nil
	}
	if size > MaxPageSize {
		return 0, fmt.Errorf("page_size must be between %d and %d", MinPageSize, MaxPageSize)
	}
	if size < MinPageSize {
		return MinPageSize, nil
	}
	return size, nil
}

func AddPagination(q url.Values, page, pageSize int) (url.Values, error) {
	page = NormalizePage(page)
	pageSize, err := NormalizePageSize(pageSize)
	if err != nil {
		return nil, err
	}
	q.Set("page", strconv.Itoa(page))
	q.Set("page_size", strconv.Itoa(pageSize))
	return q, nil
}

func Meta(page, pageSize int) (map[string]any, error) {
	pageSize, err := NormalizePageSize(pageSize)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"page":      NormalizePage(page),
		"page_size": pageSize,
	}, nil
}
