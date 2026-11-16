import { useEffect, useState, useRef, useCallback } from "react";
import { Navigate, useLocation } from "react-router";
import { useAuthStore } from "@/stores/auth";
import { getToken, clearAuthToken } from "@/libs/storage";
import { CFG } from "@/constants";
import { Api } from "@/apis";
import { buildMenuTree } from "./utils";
import { useTranslation } from "react-i18next";
interface AuthGuardProps {
	children: React.ReactNode;
}

export const AuthGuard: React.FC<AuthGuardProps> = ({ children }) => {
	const [loading, setLoading] = useState(true);
	const { t } = useTranslation();
	const location = useLocation();
	const isRequestingRef = useRef(false);

	const { setAuthData, setIsOnline, resetAuth } = useAuthStore();
	const isOnline = useAuthStore((state) => state.isOnline);

	const reset = useCallback(() => {
		resetAuth();
		clearAuthToken();
		setLoading(false);
	}, [resetAuth]);

	useEffect(() => {
		const bootstrap = async () => {
			// 防止重复请求（React StrictMode / 并发重挂载）
			if (isRequestingRef.current) return;
			isRequestingRef.current = true;

			const token = getToken();
			if (!token) {
				reset();
				return;
			}

			// 已在登录态（本会话内已引导过），无需重复拉取
			if (isOnline) {
				setLoading(false);
				return;
			}

			try {
				// 1. 并行拉取本人基本信息 + 权限（菜单权限列表 + 数据权限部门列表）
				const [info, perm] = await Promise.all([
					Api.getUserInfo(),
					Api.getPermissions(),
				]);

				// 2. 角色标识（多角色）
				const roleKeys = info.roles?.map((r) => r.role_key) ?? [];

				// 3. 权限并集：以 F 型功能 perms 为主；M/C 菜单/目录的 perms 也一并纳入，
				//    供模块级功能判断使用（seed 中角色管理缺 system:role:edit 功能，需用 system:role 兜底）
				const perms = (perm.menus ?? [])
					.filter((m) => m.perms)
					.map((m) => m.perms);

				// 4. 写 store：菜单树 / 权限 / 数据范围（部门列表由后端按 dataScope 返回）
				setAuthData({
					userInfo: info,
					menus: buildMenuTree(perm.menus ?? []),
					perms,
					roleKeys,
					depts: perm.depts ?? [],
				});
				setIsOnline(true);
				setLoading(false);
			} catch {
				reset();
			} finally {
				isRequestingRef.current = false;
			}
		};

		bootstrap();
	}, [isOnline, reset, setAuthData, setIsOnline]);

	// 全屏 loading
	if (loading) {
		return (
			<img
				src={`${import.meta.env.BASE_URL}loading.svg`}
				className="size-[25%] abs-center"
				alt={t("common.loading")}
			/>
		);
	}

	// 未登录，跳转登录页
	if (!isOnline || !getToken()) {
		return <Navigate to={CFG.LOGIN_PATH} replace state={{ from: location }} />;
	}

	return <>{children}</>;
};

// 已登录用户重定向组件
export const RedirectIfAuthenticated = ({
	children,
}: {
	children: React.ReactNode;
}) => {
	const token = getToken();
	if (token) {
		return <Navigate to={CFG.HOME_PATH} replace />;
	}
	return <>{children}</>;
};
