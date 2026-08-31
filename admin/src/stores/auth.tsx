import { create } from "zustand";
import { devtools } from "zustand/middleware";
import { IApi } from "@/apis";

// Auth State：登录用户 + 引导期汇总的权限数据
export interface AuthState {
	isOnline: boolean;
	/** 本人基本信息（来自 /api/user/info） */
	userInfo: IApi.UserItem | null;
	/** 已构建的菜单树（仅启用、含 F 功能，用于侧边栏/动态路由/功能权限） */
	menus: IApi.MenuRaw[];
	/** 权限并集（F 功能 perms 为主，另含 M/C 菜单/目录 perms 供模块级判断），供 hasPerm 使用 */
	perms: string[];
	/** 角色标识列表（多角色） */
	roleKeys: string[];
	/** 自定义数据权限部门并集（role_depts 汇总） */
	depts: IApi.DeptRaw[];
}

// Auth Actions
interface AuthActions {
	setUserInfo: (info: IApi.UserItem | null) => void;
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
