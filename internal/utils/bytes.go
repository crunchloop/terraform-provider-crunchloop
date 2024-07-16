package utils

func BytesToMegabytes(bytes int64) int32 {
	return int32(bytes / 1024 / 1024)
}

func BytesToGigabytes(bytes int64) int32 {
	return int32(bytes / 1024 / 1024 / 1024)
}
