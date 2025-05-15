package utils

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func StrToUint(str string) (uint, error) {
	num, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(num), nil
}

func AddZipDownloadHeader(ctx *gin.Context, filename string, zipData []byte) {
	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", filename)
	ctx.Header("Content-Type", "application/zip")
	ctx.Header("Content-Length", strconv.Itoa(len(zipData)))
}

func CleanFilename(filename string) string {
	// Remove spaces from filename
	// Remove any special characters and keep only alphanumeric characters
	filename = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || (r == '.') {
			return r
		}
		return -1
	}, filename)
	filename = strings.ReplaceAll(filename, " ", "")
	return filename
}
