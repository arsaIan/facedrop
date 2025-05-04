package utils

import (
	"strconv"

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