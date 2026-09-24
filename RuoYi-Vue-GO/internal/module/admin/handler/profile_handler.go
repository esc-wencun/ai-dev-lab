// 个人中心 / 注册 / 通用文件 HTTP 端点（对位 SysProfileController + SysRegisterController + CommonController）。
package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"ruoyi-vue-go/internal/middleware"
	"ruoyi-vue-go/internal/module/admin/service"
	"ruoyi-vue-go/pkg/response"
	"ruoyi-vue-go/pkg/utils/fileutil"
)

// ProfileHandler 个人中心端点集合。
type ProfileHandler struct {
	profile *service.ProfileService
	login   *LoginHandler // 复用 /register 场景的登录态约束（无）
	upload  fileUploadDir
}

// fileUploadDir 上传根目录决策（profile 根目录由 fileutil 持有）。
type fileUploadDir struct{}

// NewProfileHandler 构造。
func NewProfileHandler(p *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{profile: p}
}

// Profile GET /system/user/profile（登录即可）。
func (h *ProfileHandler) Profile(c *gin.Context) {
	lu := middleware.GetLoginUser(c)
	if lu == nil {
		response.Unauthorized(c).JSON()
		return
	}
	res, err := h.profile.GetProfile(c.Request.Context(), lu)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	// data=user + 顶层 roleGroup/postGroup（对位 AjaxResult.success(user) + put）；
	// json.RawMessage 保证 user 以内联 JSON 对象输出（[]byte 会被 base64 化——踩坑）
	raw := json.RawMessage(res.UserJSON)
	if len(raw) == 0 {
		raw = json.RawMessage("{}")
	}
	response.Ok(c).Data(raw).
		Put("roleGroup", res.RoleGroup).
		Put("postGroup", res.PostGroup).JSON()
}

// UpdateProfile PUT /system/user/profile（仅四字段 + 唯一性校验）。
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	lu := middleware.GetLoginUser(c)
	if lu == nil {
		response.Unauthorized(c).JSON()
		return
	}
	var body struct {
		NickName    string `json:"nickName"`
		Email       string `json:"email"`
		Phonenumber string `json:"phonenumber"`
		Sex         string `json:"sex"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.profile.UpdateProfile(c.Request.Context(), lu, &service.UpdateProfileInput{
		NickName: body.NickName, Email: body.Email, Phonenumber: body.Phonenumber, Sex: body.Sex,
	}); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	// 会话刷新（对位 tokenService.setLoginUser）
	if err := h.profile.RefreshSessionUser(c.Request.Context(), lu); err != nil {
		response.Error(c, "修改个人信息异常，请联系管理员").JSON()
		return
	}
	response.Ok(c).JSON()
}

// UpdatePwd PUT /system/user/profile/updatePwd。
func (h *ProfileHandler) UpdatePwd(c *gin.Context) {
	lu := middleware.GetLoginUser(c)
	if lu == nil {
		response.Unauthorized(c).JSON()
		return
	}
	var body struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	if err := h.profile.UpdatePwd(c.Request.Context(), lu, body.OldPassword, body.NewPassword); err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	if err := h.profile.RefreshSessionUser(c.Request.Context(), lu); err != nil {
		response.Error(c, "修改密码异常，请联系管理员").JSON()
		return
	}
	response.Ok(c).JSON()
}

// Avatar POST /system/user/profile/avatar（multipart 字段名 avatarfile）。
func (h *ProfileHandler) Avatar(c *gin.Context) {
	lu := middleware.GetLoginUser(c)
	if lu == nil {
		response.Unauthorized(c).JSON()
		return
	}
	file, err := c.FormFile("avatarfile")
	if err != nil || file == nil || file.Size == 0 {
		response.Error(c, "上传图片异常，请联系管理员").JSON()
		return
	}
	src, err := file.Open()
	if err != nil {
		response.Error(c, "上传图片异常，请联系管理员").JSON()
		return
	}
	defer src.Close()
	url, err := fileutil.Upload(fileutil.AvatarPath(), file.Filename, file.Size, src, fileutil.ImageExtension)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	if err := h.profile.UpdateAvatar(c.Request.Context(), lu, url); err != nil {
		response.Error(c, "上传图片异常，请联系管理员").JSON()
		return
	}
	response.Ok(c).Put("imgUrl", url).JSON()
}

// Register POST /register（匿名；开关 sys.account.registerUser 控制）。
func (h *ProfileHandler) Register(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Code     string `json:"code"`
		UUID     string `json:"uuid"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, "参数异常").JSON()
		return
	}
	msg := h.profile.Register(c.Request.Context(), &service.RegisterInput{
		Username: body.Username, Password: body.Password, Code: body.Code, UUID: body.UUID,
		IP: c.ClientIP(),
	})
	if msg != "" {
		response.Error(c, msg).JSON()
		return
	}
	response.Ok(c).JSON()
}

// Upload POST /common/upload（单文件，multipart 字段名 file）。
func (h *ProfileHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil || file == nil {
		response.Error(c, "上传文件异常").JSON()
		return
	}
	src, err := file.Open()
	if err != nil {
		response.Error(c, "上传文件异常").JSON()
		return
	}
	defer src.Close()
	url, err := fileutil.Upload(fileutil.UploadPath(), file.Filename, file.Size, src, fileutil.DefaultExtension)
	if err != nil {
		response.Error(c, err.Error()).JSON()
		return
	}
	name := url
	if idx := strings.LastIndex(url, "/"); idx >= 0 {
		name = url[idx+1:]
	}
	response.Ok(c).
		Put("url", serverURL(c)+url).
		Put("fileName", url).
		Put("newFileName", name).
		Put("originalFilename", file.Filename).
		JSON()
}

// Uploads POST /common/uploads（多文件，multipart 字段名 files；返回逗号拼接）。
func (h *ProfileHandler) Uploads(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil || form == nil {
		response.Error(c, "上传文件异常").JSON()
		return
	}
	files := form.File["files"]
	var urls, fileNames, newNames, originals []string
	for _, file := range files {
		src, err := file.Open()
		if err != nil {
			response.Error(c, "上传文件异常").JSON()
			return
		}
		url, err := fileutil.Upload(fileutil.UploadPath(), file.Filename, file.Size, src, fileutil.DefaultExtension)
		src.Close()
		if err != nil {
			response.Error(c, err.Error()).JSON()
			return
		}
		name := url
		if idx := strings.LastIndex(url, "/"); idx >= 0 {
			name = url[idx+1:]
		}
		urls = append(urls, serverURL(c)+url)
		fileNames = append(fileNames, url)
		newNames = append(newNames, name)
		originals = append(originals, file.Filename)
	}
	response.Ok(c).
		Put("urls", strings.Join(urls, ",")).
		Put("fileNames", strings.Join(fileNames, ",")).
		Put("newFileNames", strings.Join(newNames, ",")).
		Put("originalFilenames", strings.Join(originals, ",")).
		JSON()
}

// Download GET /common/download?fileName=&delete=（对位 fileDownload）。
func (h *ProfileHandler) Download(c *gin.Context) {
	fileName := c.Query("fileName")
	deleteFlag := c.Query("delete") == "true"
	if !fileutil.CheckAllowDownload(fileName) {
		// Java 侧异常只打日志、响应为空 body，前端 blob 校验失败走错误分支
		c.Status(http.StatusOK)
		return
	}
	local := fileutil.DownloadPath() + fileName
	data, err := os.ReadFile(local)
	if err != nil {
		c.Status(http.StatusOK)
		return
	}
	// realFileName = 时间戳 + 第一个 _ 之后的部分（对位 Java fileDownload）
	real := fileName
	if idx := strings.Index(fileName, "_"); idx >= 0 {
		real = fmtInt(timeNowMillis()) + fileName[idx:]
	}
	writeAttachment(c, data, real)
	if deleteFlag {
		_ = os.Remove(local)
	}
}

// DownloadResource GET /common/download/resource?resource=（对位 resourceDownload）。
func (h *ProfileHandler) DownloadResource(c *gin.Context) {
	resource := c.Query("resource")
	if !fileutil.CheckAllowDownload(resource) {
		c.Status(http.StatusOK)
		return
	}
	local, err := fileutil.ResolveLocal(resource)
	if err != nil {
		c.Status(http.StatusOK)
		return
	}
	data, err := os.ReadFile(local)
	if err != nil {
		c.Status(http.StatusOK)
		return
	}
	name := local
	if idx := strings.LastIndexAny(local, "/\\"); idx >= 0 {
		name = local[idx+1:]
	}
	writeAttachment(c, data, name)
}
