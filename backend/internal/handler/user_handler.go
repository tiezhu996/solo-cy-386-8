package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/marketpal/marketpal/internal/dto"
	"github.com/marketpal/marketpal/internal/middleware"
	"github.com/marketpal/marketpal/internal/service"
	"github.com/marketpal/marketpal/internal/util"
)

// UserHandler 用户 HTTP 处理器。
type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler 构造用户处理器。
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Register POST /api/v1/auth/register
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "注册失败：参数校验不通过 "+err.Error())
		return
	}
	user, err := h.svc.Register(req)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "注册成功", dto.FromUser(user))
}

// Login POST /api/v1/auth/login
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "登录失败：参数校验不通过 "+err.Error())
		return
	}
	user, token, err := h.svc.Login(req)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "登录成功", dto.LoginResponse{Token: token, User: *dto.FromUser(user)})
}

// Profile GET /api/v1/users/me
func (h *UserHandler) Profile(c *gin.Context) {
	user, err := h.svc.GetProfile(middleware.GetUserID(c))
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OK(c, dto.FromUser(user))
}

// UpdateProfile PUT /api/v1/users/me
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "资料更新失败：参数校验不通过 "+err.Error())
		return
	}
	user, err := h.svc.UpdateProfile(middleware.GetUserID(c), req)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "个人资料已更新", dto.FromUser(user))
}

// ListUsers GET /api/v1/users (admin)
func (h *UserHandler) ListUsers(c *gin.Context) {
	page := parseInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parseInt(c.DefaultQuery("page_size", "10"), 10)
	users, total, err := h.svc.ListUsers(page, pageSize)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	list := make([]dto.UserVO, 0, len(users))
	for i := range users {
		list = append(list, *dto.FromUser(&users[i]))
	}
	util.OK(c, gin.H{"list": list, "total": total, "page": page, "page_size": pageSize})
}

// UpdateRole PUT /api/v1/users/:id/role (admin)
func (h *UserHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "修改角色失败：用户 id 参数非法")
		return
	}
	var req dto.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, 40000, "修改角色失败：参数校验不通过 "+err.Error())
		return
	}
	user, err := h.svc.UpdateRole(middleware.GetUserID(c), uint(id), req.Role)
	if err != nil {
		util.AbortWithError(c, err)
		return
	}
	util.OKMessage(c, "用户角色已更新", dto.FromUser(user))
}

func parseInt(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	return def
}
