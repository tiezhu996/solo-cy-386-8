package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/constants"
	"github.com/marketpal/marketpal/internal/util"
)

// UploadHandler 商品图片上传处理器（本地存储，MinIO 未启用）。
type UploadHandler struct {
	uploadDir string
	publicURL string
}

// NewUploadHandler 构造上传处理器。
func NewUploadHandler(uploadDir, publicURL string) *UploadHandler {
	return &UploadHandler{uploadDir: uploadDir, publicURL: publicURL}
}

// Upload POST /api/v1/upload
func (h *UploadHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		util.Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "图片上传失败：未找到文件字段 file")
		return
	}
	defer file.Close()
	if header.Size > 5<<20 {
		util.Fail(c, http.StatusBadRequest, constants.CodeFileTooLarge, "图片上传失败：文件不能超过 5MB")
		return
	}
	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		util.Fail(c, http.StatusBadRequest, constants.CodeUnsupportedMedia, "图片上传失败：仅支持 jpg/jpeg/png/webp 格式")
		return
	}
	if err := os.MkdirAll(h.uploadDir, 0o755); err != nil {
		util.Fail(c, http.StatusInternalServerError, constants.CodeInternalError, "图片上传失败：无法创建上传目录")
		return
	}
	name := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(h.uploadDir, name)
	out, err := os.Create(dst)
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, constants.CodeInternalError, "图片上传失败：无法保存文件")
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		util.Fail(c, http.StatusInternalServerError, constants.CodeInternalError, "图片上传失败：写入文件出错")
		return
	}
	util.OK(c, gin.H{"url": h.publicURL + "/" + name})
}
