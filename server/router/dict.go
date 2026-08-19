package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"rbackeeper/internal/controller"
	"rbackeeper/middleware"
)

// registerDictRoutes 字典模块
func registerDictRoutes(base *gin.RouterGroup) {
	dict := &controller.DictController{}
	dictg := base.Group("/dict")
	dictg.Use(middleware.JWTAuth(), middleware.RateLimit(50, 100), middleware.CaptureBody())
	{
		// 字典类型分页列表
		dictg.GET("/type/list", middleware.RBAC("system:dict:type:list", "*"), dict.TypeList)
		// 新增字典类型
		dictg.POST("/type/add", middleware.Idempotent(10*time.Second), middleware.RBAC("system:dict:type:add", "*"), dict.TypeAdd)
		// 编辑字典类型
		dictg.POST("/type/edit", middleware.Idempotent(10*time.Second), middleware.RBAC("system:dict:type:edit", "*"), dict.TypeEdit)
		// 删除字典类型（级联清理其字典数据）
		dictg.POST("/type/delete", middleware.Idempotent(10*time.Second), middleware.RBAC("system:dict:type:delete", "*"), dict.TypeDelete)
		// 字典数据分页列表
		dictg.GET("/data/list", middleware.RBAC("system:dict:data:list", "*"), dict.DataList)
		// 按字典类型 key 取数据（供下拉框等使用）
		dictg.GET("/data/type", middleware.RBAC("system:dict:data:type", "*"), dict.DataByType)
		// 新增字典数据
		dictg.POST("/data/add", middleware.Idempotent(10*time.Second), middleware.RBAC("system:dict:data:add", "*"), dict.DataAdd)
		// 编辑字典数据
		dictg.POST("/data/edit", middleware.Idempotent(10*time.Second), middleware.RBAC("system:dict:data:edit", "*"), dict.DataEdit)
		// 删除字典数据
		dictg.POST("/data/delete", middleware.Idempotent(10*time.Second), middleware.RBAC("system:dict:data:delete", "*"), dict.DataDelete)
	}
}
