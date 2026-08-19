package constants

// 数据权限常量
const (
	// 全部
	DsAll int8 = 1
	// 自定义
	DsCustom int8 = 2
	// 本部门
	DsDept int8 = 3
	// 本部门及以下
	DsSubDept int8 = 4
	// 仅本人
	DsSelf int8 = 5
)

// dataScope        int8

const (
	// uint 用户数据权限值（1-5）
	CtxDataScope = "dataScope"
	// []uint DsCustom/DsSubDept 的授权部门集合
	CtxDataScopeDepts = "dataScopeDepts"
	// uint   DsDept/DsSelf 的本部门
	CtxDataScopeDept = "dataScopeDept"
	// uint   DsSelf 的本人 ID
	CtxDataScopeUserID = "dataScopeUserID"
)
