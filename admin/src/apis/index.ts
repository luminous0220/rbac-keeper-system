import { http } from "@/libs";
import { IApi } from "./type.d";

/**
 * 分页参数适配：ProTable 传 {pageNumber,pageSize,...}，后端收 {pageNum,pageSize,...}。
 * 列表页统一走 ProTable，故此处统一转换，页面无需感知两种命名。
 */
function pageParams(params: Record<string, unknown>): Record<string, unknown> {
	const { pageNumber, pageSize, ...rest } = params;
	return {
		pageNum: pageNumber ?? 1,
		pageSize: pageSize ?? 10,
		...rest,
	};
}

/**
 * 分页响应适配：后端返回 {pageNum,pageSize,total,list}，ProTable 期望 {data,total}。
 */
function pageResult<T>(res: IApi.PageResp<T>): { data: T[]; total: number } {
	return { data: res.list, total: res.total };
}

/** 列表接口统一包装：GET 分页，参数/响应适配 */
function pageApi<T>(url: string) {
	return (params: Record<string, unknown>) =>
		http
			.get<IApi.PageResp<T>>({ url, params: pageParams(params) })
			.then(pageResult);
}

/**
 * 编辑/删除类写接口：路径通过 query 定位记录（如 /api/user/edit?username=），
 * body 为编辑内容。query 放在 config.params，body 放在 params。
 */
function postByIdentifier<T = void>(
	url: string,
	identifier: Record<string, string | number>,
	body?: T,
) {
	return http.post<void>({
		url,
		params: body,
		config: { params: identifier },
	});
}

export const Api = {
	// ==================== 认证 ====================

	/** 获取图形验证码 */
	getCaptcha: () =>
		http.get<IApi.CaptchaRes>({
			url: "/api/auth/captcha",
			config: { noToken: true },
		}),

	/** 登录（返回 Access + Refresh 双令牌） */
	login: (params: IApi.LoginReq) =>
		http.post<IApi.LoginRes>({
			url: "/api/auth/login",
			params,
			config: { noToken: true },
		}),

	/** 用 Refresh Token 静默续期 Access Token（仅请求拦截器内部使用） */
	refresh: (refreshToken: string) =>
		http.post<IApi.RefreshRes>({
			url: "/api/auth/refresh",
			params: { refreshToken },
			config: { noToken: true, noErrorMessage: true },
		}),

	/** 登出：吊销当前 Access Token */
	logout: () =>
		http.post<void>({
			url: "/api/auth/logout",
		}),

	/** 修改密码：成功后旧令牌全部失效，需重新登录 */
	resetPassword: (params: IApi.ResetPasswordReq) =>
		http.post<void>({
			url: "/api/auth/reset",
			params,
		}),

	/** 修改个人信息（昵称/手机/邮箱/性别/头像），返回最新本人信息 */
	updateProfile: (params: IApi.ProfileReq) =>
		http.post<IApi.UserItem>({
			url: "/api/auth/profile",
			params,
		}),

	/** 当前用户权限：菜单权限列表 + 数据权限部门列表（引导期权限唯一来源） */
	getPermissions: () =>
		http.get<IApi.PermissionsResp>({
			url: "/api/auth/permissions",
		}),

	// ==================== 用户 ====================

	/** 本人基本信息（不含菜单） */
	getUserInfo: () =>
		http.get<IApi.UserItem>({
			url: "/api/user/info",
		}),

	/** 用户分页列表 */
	getUserList: pageApi<IApi.UserItem>("/api/user/list"),

	/** 新增用户 */
	addUser: (params: IApi.UserAddReq) =>
		http.post<void>({
			url: "/api/user/add",
			params,
		}),

	/** 编辑用户（按 username 定位，body 为编辑内容） */
	editUser: (username: string, params: IApi.UserEditReq) =>
		postByIdentifier("/api/user/edit", { username }, params),

	// ==================== 角色 ====================

	/** 角色分页列表（原生 model.Role，含 menus/depts） */
	getRoleList: (params?: Record<string, unknown>) =>
		http.get<IApi.RoleItem[]>({
			url: "/api/role/list",
			params,
		}),

	/** 新增角色 */
	addRole: (params: IApi.RoleAddReq) =>
		http.post<void>({
			url: "/api/role/add",
			params,
		}),

	/** 编辑角色（按 roleKey 定位） */
	editRole: (roleKey: string, params: IApi.RoleEditReq) =>
		postByIdentifier("/api/role/edit", { roleKey }, params),

	/** 删除角色（按 roleKey 定位；绑定用户时后端拦截） */
	deleteRole: (roleKey: string) =>
		postByIdentifier("/api/role/delete", { roleKey }),

	// ==================== 菜单 ====================

	/** 菜单树（原生 model.Menu 嵌套树；status 1正常 0停用） */
	getMenuTree: (params?: { menuName?: string; status?: number }) =>
		http.get<IApi.MenuRaw[]>({
			url: "/api/menu/tree",
			params,
		}),

	/** 新增菜单 */
	addMenu: (params: IApi.MenuAddReq) =>
		http.post<void>({
			url: "/api/menu/add",
			params,
		}),

	/** 编辑菜单（按 id 定位） */
	editMenu: (id: number, params: IApi.MenuEditReq) =>
		postByIdentifier("/api/menu/edit", { id }, params),

	/** 删除菜单 */
	deleteMenu: (id: number) => postByIdentifier("/api/menu/delete", { id }),

	// ==================== 部门 ====================

	/** 部门树（原生 model.Dept，snake_case，status 0停用 1正常 ） */
	getDeptTree: (params?: { deptName?: string; status?: number }) =>
		http.get<IApi.DeptRaw[]>({
			url: "/api/dept/tree",
			params,
		}),

	/** 新增部门 */
	addDept: (params: IApi.DeptAddReq) =>
		http.post<void>({
			url: "/api/dept/add",
			params,
		}),

	/** 编辑部门（按 id 定位） */
	editDept: (id: number, params: IApi.DeptEditReq) =>
		postByIdentifier("/api/dept/edit", { id }, params),

	/** 删除部门 */
	deleteDept: (id: number) => postByIdentifier("/api/dept/delete", { id }),

	// ==================== 字典 ====================

	/** 字典类型分页列表（原生 model.DictType） */
	getDictTypeList: pageApi<IApi.DictTypeItem>("/api/dict/type/list"),

	/** 新增字典类型 */
	addDictType: (params: IApi.DictTypeAddReq) =>
		http.post<void>({
			url: "/api/dict/type/add",
			params,
		}),

	/** 编辑字典类型（按 id 定位） */
	editDictType: (id: number, params: IApi.DictTypeEditReq) =>
		postByIdentifier("/api/dict/type/edit", { id }, params),

	/** 删除字典类型（级联清理其字典数据） */
	deleteDictType: (id: number) =>
		postByIdentifier("/api/dict/type/delete", { id }),

	/** 批量删除字典类型 */
	deleteDictTypes: (ids: number[]) =>
		http.post<void>({
			url: "/api/dict/type/batch-delete",
			params: { ids },
		}),

	/** 字典数据分页列表（按 dictType 过滤） */
	getDictDataList: pageApi<IApi.DictDataItem>("/api/dict/data/list"),

	/** 新增字典数据 */
	addDictData: (params: IApi.DictDataAddReq) =>
		http.post<void>({
			url: "/api/dict/data/add",
			params,
		}),

	/** 编辑字典数据（按 id 定位，dictType 不可变） */
	editDictData: (id: number, params: IApi.DictDataEditReq) =>
		postByIdentifier("/api/dict/data/edit", { id }, params),

	/** 删除字典数据 */
	deleteDictData: (id: number) =>
		postByIdentifier("/api/dict/data/delete", { id }),

	/** 批量删除字典数据 */
	deleteDictDatas: (ids: number[]) =>
		http.post<void>({
			url: "/api/dict/data/batch-delete",
			params: { ids },
		}),

	// ==================== 日志 ====================

	/** 登录日志分页列表 */
	getLoginLogList: pageApi<IApi.LoginLogItem>("/api/log/login"),

	/** 操作日志分页列表 */
	getOperationLogList: pageApi<IApi.OperationLogItem>("/api/log/operation"),
};

export * from "./type.d";
