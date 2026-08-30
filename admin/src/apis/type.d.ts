export namespace IApi {
	// ==================== 通用 ====================

	/** 分页响应（后端 dto.PageResp：pageNum/pageSize/total/list） */
	interface PageResp<T> {
		pageNum: number;
		pageSize: number;
		total: number;
		list: T[];
	}

	// ==================== 认证 ====================

	/** 登录请求（dto.LoginBody） */
	interface LoginReq {
		username: string;
		password: string;
		captchaId: string;
		captchaValue: string;
	}

	/** 登录响应（dto.LoginResp，含 Refresh Token） */
	interface LoginRes {
		accessToken: string;
		refreshToken: string;
		expiresIn: number;
		username: string;
		nickname: string;
	}

	/** 验证码响应（dto.CaptchaResp） */
	interface CaptchaRes {
		captchaId: string;
		img: string; // base64 data-url
	}

	/** 刷新令牌响应（dto.RefreshResp，仅新 Access Token） */
	interface RefreshRes {
		accessToken: string;
		expiresIn: number;
	}

	/** 修改密码请求（dto.ResetPasswordBody） */
	interface ResetPasswordReq {
		oldPassword: string;
		newPassword: string;
	}

	/** 修改个人信息请求（dto.ProfileBody，指针字段，nil=不改） */
	interface ProfileReq {
		nickname?: string;
		phone?: string;
		email?: string;
		sex?: number;
		avatar?: string;
	}

	// ==================== 用户 ====================

	/** 角色简述（dto.BriefRole） */
	interface BriefRole {
		role_name: string;
		role_key: string;
	}

	/** 部门简述（dto.BriefDept） */
	interface BriefDept {
		id: number;
		dept_name: string;
	}

	/** 用户列表项（dto.UserItem，脱敏：无 id，行 key 与编辑定位用 username） */
	interface UserItem {
		username: string;
		nickname: string;
		phone: string;
		email: string;
		sex: number; // 0未知 1男 2女
		avatar: string;
		status: number; // 0停用 1正常 
		loginIp: string;
		dept: BriefDept | null;
		roles: BriefRole[];
		createdAt: string;
	}

	/** 本人信息响应（dto.UserInfoResp = UserItem + 角色并集菜单，菜单为扁平 model.Menu） */
	interface UserInfoResp extends UserItem {
		menus: MenuRaw[];
	}

	/** 新增用户（dto.UserAddBody） */
	interface UserAddReq {
		username: string;
		password: string;
		nickname: string;
		phone: string;
		email: string;
		sex: number; // 0未知 1男 2女（后端已放开 0/2 值）
		deptId: number;
		roleKeys: string[];
		remark: string;
	}

	/** 修改用户（dto.UserEditBody，指针字段，nil=不改；roleKeys 空数组=清空） */
	interface UserEditReq {
		password?: string;
		deptId?: number;
		roleKeys?: string[];
		status?: number;
	}

	// ==================== 角色 ====================

	/**
	 * 角色列表项 / 角色详情 = 原生 model.Role（snake_case，预加载 menus/depts）。
	 * status：0禁用 1正常；data_scope：1全部 2自定义 3本部门 4本部门及以下 5仅本人。
	 */
	interface RoleItem {
		id: number;
		role_name: string;
		role_key: string;
		role_sort: number;
		data_scope: number;
		status: number;
		remark: string;
		menus: MenuRaw[];
		depts: DeptRaw[];
		createdAt: string;
		updatedAt: string;
	}

	/** 新增角色（dto.RoleAddBody，menuIds/deptIds 为空数组表示无授权） */
	interface RoleAddReq {
		roleName: string;
		roleKey: string;
		roleSort: number;
		dataScope: number;
		menuIds: number[];
		deptIds: number[];
		remark: string;
	}

	/** 修改角色（dto.RoleEditBody，menuIds 传 null=不改，空数组=清空） */
	interface RoleEditReq {
		roleName?: string;
		roleSort?: number;
		dataScope?: number;
		status?: number;
		menuIds?: number[];
		deptIds?: number[];
		remark?: string;
	}

	// ==================== 菜单 ====================

	/**
	 * 原始 model.Menu（snake_case）。
	 * /user/info、/role/info 返回扁平数组；/menu/tree 返回同构的嵌套树（children）。
	 * status 为 model 原值：1=正常 0=停用（前端 buildMenuTree 以 status===1 过滤）。
	 * 写接口 DTO（MenuAddReq/MenuEditReq）已与读接口统一为同一语义：1=正常 0=停用。
	 */
	interface MenuRaw {
		id: number;
		parent_id: number;
		menu_name: string;
		order_num: number;
		path: string;
		component: string;
		query: string;
		menu_type: "M" | "C" | "F"; // M目录 C菜单 F按钮
		status: number;
		perms: string;
		icon: string;
		visible: number; // 0隐藏 1显示 
		remark: string;
		children?: MenuRaw[];
		createdAt: string;
		updatedAt: string;
	}

	/** 新增菜单（dto.MenuAddBody，status：1正常 0停用，与 model 一致） */
	interface MenuAddReq {
		parentId: number;
		menuName: string;
		menuType: "M" | "C" | "F";
		orderNum: number;
		path: string;
		component: string;
		query: string;
		status: number;
		perms: string;
		icon: string;
		visible: number;
		remark: string;
	}

	/** 修改菜单（dto.MenuEditBody） */
	interface MenuEditReq {
		parentId?: number;
		menuName?: string;
		menuType?: "M" | "C" | "F";
		orderNum?: number;
		path?: string;
		component?: string;
		query?: string;
		status?: number;
		perms?: string;
		icon?: string;
		visible?: number;
		remark?: string;
	}

	// ==================== 部门 ====================

	/** 原始 model.Dept（/dept/tree 返回；snake_case；status：0停用 1正常 ） */
	interface DeptRaw {
		id: number;
		dept_name: string;
		parent_id: number;
		order_num: number;
		leader: string;
		phone: string;
		email: string;
		status: number;
		remark: string;
		children?: DeptRaw[];
		createdAt: string;
		updatedAt: string;
	}

	/** 新增部门（dto.DeptAddBody，parentId=0 为根） */
	interface DeptAddReq {
		parentId: number;
		deptName: string;
		orderNum: number;
		leader: string;
		phone: string;
		email: string;
		remark: string;
	}

	/** 修改部门（dto.DeptEditBody） */
	interface DeptEditReq {
		parentId?: number;
		deptName?: string;
		orderNum?: number;
		leader?: string;
		phone?: string;
		email?: string;
		status?: number;
		remark?: string;
	}

	// ==================== 字典 ====================

	/** 原生 model.DictType（status：0停用 1正常 ） */
	interface DictTypeItem {
		id: number;
		dict_name: string;
		dict_type: string;
		status: number;
		createdAt: string;
		updatedAt: string;
	}

	/** 新增字典类型（dto.DictTypeAddBody） */
	interface DictTypeAddReq {
		dictName: string;
		dictType: string;
		status: number;
	}

	/** 修改字典类型（dto.DictTypeEditBody） */
	interface DictTypeEditReq {
		dictName?: string;
		dictType?: string;
		status?: number;
	}

	/** 原生 model.DictData（status：0停用 1正常 ） */
	interface DictDataItem {
		id: number;
		dict_sort: number;
		dict_label: string;
		dict_value: string;
		dict_type: string;
		css_class: string;
		list_class: string;
		is_default: number;
		status: number;
		createdAt: string;
		updatedAt: string;
	}

	/** 新增字典数据（dto.DictDataAddBody） */
	interface DictDataAddReq {
		dictSort: number;
		dictLabel: string;
		dictValue: string;
		dictType: string;
		cssClass: string;
		listClass: string;
		isDefault: number;
		status: number;
	}

	/** 修改字典数据（dto.DictDataEditBody，dict_type 不可变） */
	interface DictDataEditReq {
		dictSort?: number;
		dictLabel?: string;
		dictValue?: string;
		cssClass?: string;
		listClass?: string;
		isDefault?: number;
		status?: number;
	}

	// ==================== 日志 ====================

	/** 登录日志项（dto.LoginLogItem） */
	interface LoginLogItem {
		id: number;
		username: string;
		ip: string;
		ipAddr: string;
		browser: string;
		os: string;
		status: number; // 0成功 1失败
		msg: string;
		loginTime: string;
	}

	/** 操作日志项（dto.OperationLogItem） */
	interface OperationLogItem {
		id: number;
		title: string;
		requestMethod: string;
		operName: string;
		operUrl: string;
		operIp: string;
		status: number; // 0正常 1异常
		errorMsg: string;
		operTime: string;
	}
}
