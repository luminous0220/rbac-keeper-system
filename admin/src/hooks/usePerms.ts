import { useCallback } from "react";
import { useAuthStore } from "@/stores";

/**
 * 按钮权限 hook：从 auth store 的 perms（F 型菜单权限并集）判断当前用户是否拥有指定权限。
 * 页面内配合按钮显隐使用：const { hasPerm } = usePerms(); hasPerm("system:user:add")
 */
export function usePerms() {
	const perms = useAuthStore((s) => s.perms);

	const hasPerm = useCallback(
		(perm: string) => {
			if (!perm) return true;
			return perms.includes(perm);
		},
		[perms],
	);

	return { hasPerm, perms };
}
