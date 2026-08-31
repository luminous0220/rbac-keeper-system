import { lazy, Suspense } from "react";
import { RouteObject } from "react-router";
import { Spin } from "antd";
import { IApi } from "@/apis";

// 空页面组件（未配置组件路径的占位）
const EmptyPage = () => (
	<div className="min-h-full flex items-center justify-center">
		<div className="text-center">
			<h1 className="text-2xl font-bold text-gray-800 dark:text-white mb-4">
				页面开发中
			</h1>
			<p className="text-gray-500 dark:text-gray-400">敬请期待...</p>
		</div>
	</div>
);

// 懒加载组件包装器（带 fallback，保证过渡动画可见）
const LazyWrapper = (
	Component: React.LazyExoticComponent<React.ComponentType<any>>,
) => (
	<Suspense
		fallback={
			<div className="size-full flex items-center justify-center">
				<Spin />
			</div>
		}
	>
		<Component />
	</Suspense>
);

// 拼接父路径与子路径，统一输出绝对路径（/system/user）
// 后端菜单 path 为相对路径（如 user/role），导航/路由均需绝对路径。
function joinPath(parent: string, child: string): string {
	const p = parent.replace(/\/+$/, "");
	const c = (child || "").replace(/^\/+/, "");
	if (!c) return p || "/";
	return `${p}/${c}`;
}

/**
 * 由扁平菜单列表构建菜单树（/user/info 返回扁平 model.Menu，需前端建树）。
 * - 仅保留启用（status===1）且非功能（menu_type!=='F'）的菜单
 * - 按 parent_id 挂载，同级按 order_num 升序
 * - 逐层拼接 path 为绝对路径，供侧边栏导航与路由匹配
 */
export const buildMenuTree = (flat: IApi.MenuRaw[]): IApi.MenuRaw[] => {
	const nodes = flat
		.filter((m) => m.menu_type !== "F")
		.slice()
		.sort((a, b) => a.order_num - b.order_num);

	const map = new Map<number, IApi.MenuRaw>();
	const roots: IApi.MenuRaw[] = [];
	nodes.forEach((m) => {
		m.children = [];
		map.set(m.id, m);
	});
	nodes.forEach((m) => {
		if (m.parent_id === 0 || !map.has(m.parent_id)) {
			roots.push(m);
		} else {
			map.get(m.parent_id)!.children!.push(m);
		}
	});

	// 递归：拼接绝对路径 + 排序子级 + 清理空 children
	const normalize = (list: IApi.MenuRaw[], parentPath: string) => {
		for (const node of list) {
			node.path = joinPath(parentPath, node.path);
			if (node.children?.length) {
				node.children.sort((a, b) => a.order_num - b.order_num);
				normalize(node.children, node.path);
			} else {
				delete node.children;
			}
		}
	};
	normalize(roots, "");
	return roots;
};

// 生成动态路由（目录/菜单类型且含路径的节点生成路由，F 功能不生成）
export const generateRoutes = (menus: IApi.MenuRaw[]): RouteObject[] => {
	const routes: RouteObject[] = [];
	menus.forEach((menu) => {
		if (menu.menu_type !== "F" && menu.path) {
			routes.push({
				path: menu.path,
				element: loadComponent(menu.component),
			});
		}
		if (menu.children?.length) {
			routes.push(...generateRoutes(menu.children));
		}
	});
	return routes;
};

// 动态加载组件页面
const dynamicImport = (componentPath: string) => {
	if (!componentPath) return <EmptyPage />;
	// 使用 Vite 支持的动态导入格式
	const modules = import.meta.glob("@/pages/**/*.tsx");
	const path = `/src/pages/${componentPath}`;
	if (modules[path]) {
		return modules[path];
	}
	const path2 = `/src/pages${componentPath}`;
	if (modules[path2]) {
		return modules[path2];
	}

	// 尝试其他路径格式
	const path3 = `/src/pages/${componentPath}`;
	if (modules[path3]) {
		return modules[path3];
	}

	return null;
};

export const loadComponent = (
	componentPath: string | undefined,
): React.ReactNode => {
	if (!componentPath) return <EmptyPage />;

	const importer = dynamicImport(componentPath);
	if (importer) {
		const LazyComponent = lazy(
			importer as () => Promise<{ default: React.ComponentType<any> }>,
		);
		return LazyWrapper(LazyComponent);
	}

	return <EmptyPage />;
};
