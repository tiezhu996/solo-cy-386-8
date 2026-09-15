package util

// PageParams 统一分页参数。
type PageParams struct {
	Page     int
	PageSize int
}

// NormalizePage 归一化分页参数，默认 page=1、page_size=10。
func NormalizePage(page, pageSize int) PageParams {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return PageParams{Page: page, PageSize: pageSize}
}

// Offset 计算偏移量。
func (p PageParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}
