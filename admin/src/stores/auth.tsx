import { create } from "zustand";
import { devtools } from "zustand/middleware";
import { IApi } from "@/apis";

// Auth State：登录用户 + 引导期汇总的权限数据
export interface AuthState {
	isOnline: boolean;
	/** 本人信息（含角色并集菜单，来自 /api/user/info） */
	userInfo: IApi.UserInfoResp | null;
	/** 已构建的菜单树（仅启用、含 F 按钮，用于侧边栏/动态路由/按钮权限） */
	menus: IApi.MenuRaw[];
	/** 权限并集（F 按钮 perms 为主，另含 M/C 目录/菜单 perms 供模块级判断），供 hasPerm 使用 */
	perms: string[];
	/** 角色标识列表（多角色） */
	roleKeys: string[];
	/** 数据权限最小值（1全部 2自定义 3本部门 4本部门及以下 5仅本人） */
	dataScope: number;
	/** 自定义数据权限部门并集（role_depts 汇总） */
	depts: IApi.DeptRaw[];
}

// Auth Actions
interface AuthActions {
	setUserInfo: (info: IApi.UserInfoResp | null) => void;
	setAuthData: (partial: Partial<AuthState>) => void;
	setIsOnline: (status: boolean) => void;
	resetAuth: () => void;
}

const initialAuthState: AuthState = {
	isOnline: false,
	userInfo: null,
	menus: [],
	perms: [],
	roleKeys: [],
	dataScope: 5,
	depts: [],
};

export const useAuthStore = create<AuthState & AuthActions>()(
	devtools(
		(set) => ({
			...initialAuthState,

			setUserInfo: (info) => set({ userInfo: info }),

			setAuthData: (partial) => set(partial),

			setIsOnline: (status) => set({ isOnline: status }),

			resetAuth: () => set({ ...initialAuthState, isOnline: false }),
		}),
		{
			name: "auth",
		},
	),
);
