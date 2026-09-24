// Package router 路由注册中心（对位 Python server.py 的 controller_list / Java 各 Controller）。
// 业务模块路由统一在此注册。
package router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"ruoyi-vue-go/docs"

	"ruoyi-vue-go/internal/cache"
	"ruoyi-vue-go/internal/middleware"
	"ruoyi-vue-go/internal/module/admin/dao"
	"ruoyi-vue-go/internal/module/admin/handler"
	"ruoyi-vue-go/internal/module/admin/service"
	"ruoyi-vue-go/internal/security"
)

// Deps 路由装配依赖（main 构造后注入）。
type Deps struct {
	DB    *gorm.DB
	Cache *cache.RedisCache
	Token *security.TokenService
}

// New 创建 gin 引擎并注册全部路由。
func New(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 横切中间件：认证（token 解析 → 会话入 context）
	r.Use(middleware.Auth(d.Token))

	// API 文档（对位 Java springdoc：/v3/api-docs + swagger-ui；前端 系统工具→接口文档 iframe 加载）
	// URL 用相对路径 "doc.json"：iframe src 带 /dev-api 前缀时随代理前缀自适应
	r.GET("/swagger-ui/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("doc.json"), ginSwagger.DefaultModelsExpandDepth(0)))
	r.GET("/swagger-ui.html", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger-ui/index.html")
	})
	r.GET("/v3/api-docs", func(c *gin.Context) {
		c.Header("Content-Type", "application/json; charset=utf-8")
		c.String(http.StatusOK, docs.SwaggerInfo.ReadDoc())
	})

	// ---------- 登录闭环（2.0.0） ----------
	loginDao := dao.NewLoginDao(d.DB)
	loginSvc := service.NewLoginService(loginDao, d.Cache)
	menuSvc := service.NewMenuService(loginDao)
	loginH := handler.NewLoginHandler(loginSvc, menuSvc, d.Token, d.Cache, loginDao)

	r.GET("/", loginH.Index)                    // 欢迎语（免认证，白名单）
	r.GET("/captchaImage", loginH.CaptchaImage) // 验证码（免认证，白名单）
	r.POST("/login", loginH.Login)              // 登录（免认证，白名单）

	// 受保护端点（Auth 中间件解析会话；无会话由 handler 返回 401 信封）
	r.GET("/getInfo", loginH.GetInfo)
	r.GET("/getRouters", loginH.GetRouters)
	r.GET("/getPlatformInfo", loginH.GetPlatformInfo) // 平台标识（登录即可；前端监控页降级依据）
	r.POST("/logout", loginH.Logout)
	r.POST("/unlockscreen", loginH.Unlockscreen)

	// ---------- 个人中心 + 注册 + 通用文件（3.0.0） ----------
	profileDao := dao.NewProfileDao(d.DB)
	profileSvc := service.NewProfileService(profileDao, loginDao, d.Cache, loginSvc)
	profileSvc.SetToken(d.Token)
	profileH := handler.NewProfileHandler(profileSvc)

	r.POST("/register", profileH.Register) // 注册（匿名，白名单；开关控制）
	pro := r.Group("/system/user/profile")
	{
		pro.GET("", profileH.Profile)
		pro.PUT("", profileH.UpdateProfile)
		pro.PUT("/updatePwd", profileH.UpdatePwd)
		pro.POST("/avatar", profileH.Avatar)
	}
	common := r.Group("/common")
	{
		common.POST("/upload", profileH.Upload)
		common.POST("/uploads", profileH.Uploads)
		common.GET("/download", profileH.Download)
		common.GET("/download/resource", profileH.DownloadResource)
	}

	// ---------- 部门 + 岗位（4.0.0） ----------
	deptSvc := service.NewDeptService(dao.NewDeptDao(d.DB))
	postSvc := service.NewPostService(dao.NewPostDao(d.DB))
	deptH := handler.NewDeptHandler(deptSvc)
	postH := handler.NewPostHandler(postSvc)

	dept := r.Group("/system/dept")
	{
		dept.GET("/list", deptH.List)
		dept.GET("/list/exclude/:deptId", deptH.ListExclude)
		dept.GET("/:deptId", deptH.GetInfo)
		dept.POST("", deptH.Add)
		dept.PUT("", deptH.Edit)
		dept.DELETE("/:deptId", deptH.Remove)
	}
	post := r.Group("/system/post")
	{
		post.GET("/list", postH.List)
		post.GET("/optionselect", postH.Optionselect)
		post.GET("/:postId", postH.GetInfo)
		post.POST("", postH.Add)
		post.PUT("", postH.Edit)
		post.DELETE("/:postIds", postH.Remove)
	}

	// ---------- 用户管理（5.0.0） ----------
	userSvc := service.NewUserService(dao.NewUserDAO(d.DB), dao.NewDeptDao(d.DB), d.Cache, d.DB)
	userH := handler.NewUserHandler(userSvc)

	user := r.Group("/system/user")
	{
		user.GET("/list", userH.List)
		user.POST("/export", userH.Export)
		user.POST("/importData", userH.ImportData)
		user.POST("/importTemplate", userH.ImportTemplate)
		user.GET("/deptTree", userH.DeptTree)
		user.GET("/authRole/:userId", userH.AuthRolePage)
		user.PUT("/authRole", userH.InsertAuthRole)
		user.PUT("/resetPwd", userH.ResetPwd)
		user.PUT("/changeStatus", userH.ChangeStatus)
		user.GET("/:userId", userH.GetInfo)
		user.POST("", userH.Add)
		user.PUT("", userH.Edit)
		user.DELETE("/:userIds", userH.Remove)
	}
	// GET /system/user/（空 userId，对位 Java @GetMapping({"/", "/{userId}"})）
	user.GET("/", userH.GetInfo)

	// ---------- 角色菜单（6.0.0） ----------
	menuDAO := dao.NewMenuDAO(d.DB)
	roleSvc := service.NewRoleService(dao.NewRoleDAO(d.DB), menuDAO)
	menuMgmtSvc := service.NewMenuServiceMgmt(menuDAO)
	roleH := handler.NewRoleHandler(roleSvc)
	menuH := handler.NewMenuHandler(menuMgmtSvc)

	role := r.Group("/system/role")
	{
		role.GET("/list", roleH.List)
		role.POST("/export", roleH.Export)
		role.GET("/optionselect", roleH.Optionselect)
		role.GET("/authUser/allocatedList", roleH.AllocatedList)
		role.GET("/authUser/unallocatedList", roleH.UnallocatedList)
		role.PUT("/authUser/cancel", roleH.CancelAuthUser)
		role.PUT("/authUser/cancelAll", roleH.CancelAuthUserAll)
		role.PUT("/authUser/selectAll", roleH.SelectAuthUserAll)
		role.PUT("/dataScope", roleH.DataScope)
		role.PUT("/changeStatus", roleH.ChangeStatus)
		role.GET("/:roleId", roleH.GetInfo)
		role.POST("", roleH.Add)
		role.PUT("", roleH.Edit)
		role.DELETE("/:roleIds", roleH.Remove)
	}
	menu := r.Group("/system/menu")
	{
		menu.GET("/list", menuH.List)
		menu.GET("/treeselect", menuH.TreeSelect)
		menu.GET("/roleMenuTreeselect/:roleId", menuH.RoleMenuTreeSelect)
		menu.GET("/:menuId", menuH.GetInfo)
		menu.POST("", menuH.Add)
		menu.PUT("", menuH.Edit)
		menu.DELETE("/:menuId", menuH.Remove)
	}

	// ---------- 字典参数（7.0.0） ----------
	dictTypeSvc := service.NewDictTypeService(dao.NewDictTypeDAO(d.DB), dao.NewDictDataDAO(d.DB), d.Cache)
	dictDataSvc := service.NewDictDataService(dao.NewDictDataDAO(d.DB), d.Cache)
	configSvc := service.NewConfigService(dao.NewConfigDAO(d.DB), d.Cache)
	dictTypeH := handler.NewDictTypeHandler(dictTypeSvc)
	dictDataH := handler.NewDictDataHandler(dictDataSvc)
	configH := handler.NewConfigHandler(configSvc)

	dictType := r.Group("/system/dict/type")
	{
		dictType.GET("/list", dictTypeH.List)
		dictType.GET("/optionselect", dictTypeH.Optionselect)
		dictType.GET("/:dictId", dictTypeH.GetInfo)
		dictType.POST("", dictTypeH.Add)
		dictType.PUT("", dictTypeH.Edit)
		dictType.DELETE("/refreshCache", dictTypeH.RefreshCache)
		dictType.DELETE("/:dictIds", dictTypeH.Remove)
	}
	dictData := r.Group("/system/dict/data")
	{
		dictData.GET("/list", dictDataH.List)
		dictData.GET("/type/:dictType", dictDataH.TypeByDictType)
		dictData.GET("/:dictCode", dictDataH.GetInfo)
		dictData.POST("", dictDataH.Add)
		dictData.PUT("", dictDataH.Edit)
		dictData.DELETE("/:dictCodes", dictDataH.Remove)
	}
	config := r.Group("/system/config")
	{
		config.GET("/list", configH.List)
		config.GET("/configKey/:configKey", configH.ConfigKey)
		config.GET("/:configId", configH.GetInfo)
		config.POST("", configH.Add)
		config.PUT("", configH.Edit)
		config.DELETE("/refreshCache", configH.RefreshCache)
		config.DELETE("/:configIds", configH.Remove)
	}

	// ---------- 通知公告（8.0.0） + 监控日志（9.0.0） ----------
	noticeDAO := dao.NewNoticeDAO(d.DB)
	noticeH := handler.NewNoticeHandler(noticeDAO)
	operLogH := handler.NewOperLogHandler(dao.NewOperLogDAO(d.DB))
	logininforH := handler.NewLogininforHandler(dao.NewLogininforDAO(d.DB), d.Cache)
	onlineH := handler.NewOnlineHandler(d.Token, d.Cache)
	cacheH := handler.NewCacheHandler(d.Cache)

	notice := r.Group("/system/notice")
	{
		notice.GET("/list", noticeH.List)
		notice.GET("/listTop", noticeH.ListTop)
		notice.POST("/markRead", noticeH.MarkRead)
		notice.POST("/markReadAll", noticeH.MarkReadAll)
		notice.GET("/readUsers/list", noticeH.ReadUsers)
		notice.GET("/:noticeId", noticeH.GetInfo)
		notice.POST("", noticeH.Add)
		notice.PUT("", noticeH.Edit)
		notice.DELETE("/:noticeIds", noticeH.Remove)
	}
	operlog := r.Group("/monitor/operlog")
	{
		operlog.GET("/list", operLogH.List)
		operlog.DELETE("/clean", operLogH.Clean)
		operlog.DELETE("/:operIds", operLogH.Remove)
	}
	logininfor := r.Group("/monitor/logininfor")
	{
		logininfor.GET("/list", logininforH.List)
		logininfor.DELETE("/clean", logininforH.Clean)
		logininfor.GET("/unlock/:userName", logininforH.Unlock)
		logininfor.DELETE("/:infoIds", logininforH.Remove)
	}
	online := r.Group("/monitor/online")
	{
		online.GET("/list", onlineH.List)
		online.DELETE("/:tokenId", onlineH.ForceLogout)
	}
	monitorCache := r.Group("/monitor/cache")
	{
		monitorCache.GET("", cacheH.Info)
		monitorCache.GET("/getNames", cacheH.Names)
		monitorCache.GET("/getKeys/:cacheName", cacheH.Keys)
		monitorCache.GET("/getValue/:cacheName/:cacheKey", cacheH.Value)
		monitorCache.DELETE("/clearCacheKey/:cacheKey", cacheH.ClearCacheKey)
		monitorCache.DELETE("/clearCacheAll", cacheH.ClearCacheAll)
	}

	// ---------- 定时任务（10.0.0） ----------
	jobH := handler.NewJobHandler(dao.NewJobDAO(d.DB), dao.NewJobLogDAO(d.DB))
	jobLogH := handler.NewJobLogHandler(dao.NewJobLogDAO(d.DB))
	job := r.Group("/monitor/job")
	{
		job.GET("/list", jobH.List)
		job.PUT("/changeStatus", jobH.ChangeStatus)
		job.PUT("/run", jobH.Run)
		job.GET("/:jobId", jobH.GetInfo)
		job.POST("", jobH.Add)
		job.PUT("", jobH.Edit)
		job.DELETE("/:jobIds", jobH.Remove)
	}
	jobLog := r.Group("/monitor/jobLog")
	{
		jobLog.GET("/list", jobLogH.List)
		jobLog.GET("/:jobLogId", jobLogH.GetInfo)
		jobLog.DELETE("/clean", jobLogH.Clean)
		jobLog.DELETE("/:jobLogIds", jobLogH.Remove)
	}

	// 调度器加载启用态任务（暂停态不加载；changeStatus 恢复时从库补载，规避 Python 版踩坑 5）
	// DB 为 nil（纯路由测试）时跳过
	if d.DB != nil {
		jobH.LoadEnabledJobs(context.Background())
	}

	// ---------- 代码生成器（11.0.0，降级范围：数据层端点） ----------
	genH := handler.NewGenHandler(dao.NewGenTableDAO(d.DB))
	gen := r.Group("/tool/gen")
	{
		gen.GET("/list", genH.List)
		gen.GET("/db/list", genH.DbList)
		gen.POST("/importTable", genH.ImportTable)
		gen.GET("/:tableId", genH.GetInfo)
		gen.DELETE("/:tableIds", genH.Remove)
	}

	return r
}

// c0 非 nil context。
func c0() context.Context {
	return context.Background()
}
