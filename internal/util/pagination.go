package util

// CalcTotalPages menghitung jumlah halaman dari total item dan limit per halaman.
func CalcTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	return int((total + int64(limit) - 1) / int64(limit))
}
