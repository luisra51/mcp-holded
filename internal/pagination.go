package internal

func NormalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}
