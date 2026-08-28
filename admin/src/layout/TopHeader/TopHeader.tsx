import { Header } from "antd/es/layout/layout";
import {
	Dropdown,
	Avatar,
	Space,
	Breadcrumb,
	AutoComplete,
	Input,
	Tag,
	theme,
	App,
} from "antd";
import { CollapseButton, CollapseProps } from "./CollapseButton";
import { ThemeToggle } from "@/components/ThemeToggle";
import { useAuthStore } from "@/stores";
import { IconUser, IconKey, IconLogout, IconSearch } from "@tabler/icons-react";
import { IconMap, clearAuthToken } from "@/libs";
import { useNavigate, useLocation } from "react-router";
import { CFG } from "@/constants";
import type { MenuProps } from "antd";
import type { IApi } from "@/apis";
import { useMemo, useState } from "react";
import { Api } from "@/apis";
import { Setting } from "./Setting";

interface TopHeaderProps extends CollapseProps {
	className?: string;
}

/** 扁平化的菜单节点（用于面包屑和搜索） */
interface FlatMenuNode {
	path: string;
	title: string;
	icon: string;
	/** 菜单类型 M目录 C菜单 F按钮 */
	menuType: string;
	/** 从根到当前节点的面包屑文本，如 "系统管理 > 用户管理" */
	breadcrumbText: string;
	/** 从根到当前节点的面包屑路径数组 */
	trail: { path: string; title: string; icon: string }[];
}

/** 递归扁平化菜单树，返回所有可导航的菜单项 */
function flattenMenus(
	menus: IApi.MenuRaw[],
	trail: { path: string; title: string; icon: string }[] = [],
): FlatMenuNode[] {
	const result: FlatMenuNode[] = [];
	for (const menu of menus) {
		const currentTrail = [
			...trail,
			{ path: menu.path, title: menu.menu_name, icon: menu.icon },
		];
		result.push({
			path: menu.path,
			title: menu.menu_name,
			icon: menu.icon,
			menuType: menu.menu_type,
			breadcrumbText: currentTrail.map((t) => t.title).join(" > "),
			trail: currentTrail,
		});
		if (menu.children) {
			result.push(...flattenMenus(menu.children, currentTrail));
		}
	}
	return result;
}

/** 在菜单树中查找当前路径的面包屑轨迹 */
function findBreadcrumbTrail(
	menus: IApi.MenuRaw[],
	targetPath: string,
	trail: { path: string; title: string; icon: string; menuType: string }[] = [],
): { path: string; title: string; icon: string; menuType: string }[] | null {
	for (const menu of menus) {
		const current = {
			path: menu.path,
			title: menu.menu_name,
			icon: menu.icon,
			menuType: menu.menu_type,
		};
		if (menu.path === targetPath) {
			return [...trail, current];
		}
		if (menu.children) {
			const found = findBreadcrumbTrail(menu.children, targetPath, [
				...trail,
				current,
			]);
			if (found) return found;
		}
	}
	return null;
}

export const TopHeader = ({ setCollapsed, collapsed }: TopHeaderProps) => {
	const navigate = useNavigate();
	const location = useLocation();
	const { message } = App.useApp();
	const userInfo = useAuthStore((s) => s.userInfo);
	const menus = useAuthStore((s) => s.menus);
	const resetAuth = useAuthStore((s) => s.resetAuth);
	const [searchText, setSearchText] = useState("");
	const [inputFocused, setInputFocused] = useState(false);
	const { token } = theme.useToken();
	// 扁平化菜单（用于搜索）
	const flatMenus = useMemo(() => flattenMenus(menus), [menus]);
	// 面包屑
	const breadcrumbItems = useMemo(() => {
		const trail = findBreadcrumbTrail(menus, location.pathname);
		if (!trail || trail.length === 0) return [];
		return trail.map((item, index) => ({
			title: (
				<Space size={4}>
					{IconMap[item.icon]}
					<span>{item.title}</span>
				</Space>
			),
			onClick:
				index < trail.length - 1 && item.menuType !== "M"
					? () => navigate(item.path)
					: undefined,
		}));
	}, [menus, location.pathname, navigate]);

	// 搜索选项
	const searchOptions = useMemo(() => {
		const kw = searchText.toLowerCase();
		const source = searchText
			? flatMenus.filter(
					(item) =>
						item.title.toLowerCase().includes(kw) ||
						item.breadcrumbText.toLowerCase().includes(kw),
				)
			: flatMenus;
		return source.slice(0, 16).map((item) => ({
			value: item.path,
			disabled: item.menuType === "M",
			label: (
				<div className="flex items-center gap-2 py-0.5">
					<span>{IconMap[item.icon]}</span>
					<span className="text-xs">{item.breadcrumbText}</span>
				</div>
			),
		}));
	}, [searchText, flatMenus]);

	// 选中搜索结果 → 导航并清空
	const handleSearchSelect = (value: string) => {
		setSearchText("");
		navigate(value);
	};

	// 退出登录：先调后端吊销令牌（失败也继续），再清本地缓存并回登录页
	const handleLogout = async () => {
		try {
			await Api.logout();
		} catch (err) {
			console.error("退出登录接口调用失败，继续本地清理", err);
		}
		clearAuthToken();
		resetAuth();
		message.success("已退出登录");
		navigate(CFG.LOGIN_PATH, { replace: true });
	};

	// 用户下拉菜单项
	const userMenuItems: MenuProps["items"] = [
		{
			key: "profile",
			label: "个人信息",
			icon: <IconUser stroke={1} />,
			onClick: () => {
				navigate("/profile");
			},
		},
		{
			key: "password",
			label: "修改密码",
			icon: <IconKey stroke={1} />,
			onClick: () => {
				navigate("/profile?tab=password");
			},
		},
		{
			type: "divider",
		},
		{
			key: "logout",
			label: "退出登录",
			icon: <IconLogout stroke={1} />,
			danger: true,
			onClick: handleLogout,
		},
	];

	// 角色标签：最多展示 3 个，超出折叠为 +N
	const roleTags = (userInfo?.roles ?? []).slice(0, 3).map((role) => (
		<Tag
			key={role.role_key}
			color="blue"
			className="!m-0 hidden md:inline-flex"
		>
			{role.role_name}
		</Tag>
	));
	const extraRoles = (userInfo?.roles ?? []).length - 3;
	const nickname = userInfo?.nickname || userInfo?.username || "";

	return (
		<div className="w-full pt-4 relative z-[99] px-4">
			<Header
				className="flex h-top items-center justify-between shadow-soft-lg backdrop-blur "
				style={{
					padding: "0 16px",
					backgroundColor: token.colorBgContainer,
					borderRadius: "14px",
				}}
			>
				{/* 左侧：折叠按钮 + 面包屑 + 搜索 */}
				<div className="flex items-center gap-6 flex-1 min-w-0">
					<CollapseButton collapsed={collapsed} setCollapsed={setCollapsed} />

					<AutoComplete
						className="max-w-64 max-md:hidden h-[32px]"
						options={searchOptions}
						showSearch={{ onSearch: setSearchText }}
						onSelect={handleSearchSelect}
						onClear={() => setSearchText("")}
						value={searchText}
						open={inputFocused && searchOptions.length > 0 ? true : undefined}
						popupMatchSelectWidth={false}
						getPopupContainer={(trigger) => trigger.parentElement!}
					>
						<Input
							className="w-[326px] h-[32px]"
							placeholder="搜索菜单..."
							prefix={
								<IconSearch size={14} stroke={1.5} className="text-gray-400" />
							}
							allowClear
							onFocus={() => setInputFocused(true)}
							onBlur={() => setInputFocused(false)}
						/>
					</AutoComplete>

					<Breadcrumb className="shrink-0 max-md:hidden" items={breadcrumbItems} />
				</div>

				{/* 右侧：主题切换 + 系统设置 + 用户头像/角色标签 */}
				<div className="flex items-center gap-2">
					<ThemeToggle />

					<Setting />

					<Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
						<Space className="cursor-pointer">
							<Avatar
								size={32}
								src={userInfo?.avatar || undefined}
								style={{
									background: "#EAF3FF",
									cursor: "pointer",
								}}
							>
								{nickname.charAt(0) || "🐼"}
							</Avatar>
							<span className="hidden md:inline text-sm max-w-24 truncate">
								{nickname}
							</span>
							<Space size={4}>{roleTags}</Space>
							{extraRoles > 0 && (
								<Tag className="!m-0 hidden md:inline-flex">+{extraRoles}</Tag>
							)}
						</Space>
					</Dropdown>
				</div>
			</Header>
		</div>
	);
};
