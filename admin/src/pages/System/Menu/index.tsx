import { useCallback, useMemo, useRef, useState } from "react";
import { App, Button, Popconfirm, Tag, TreeSelect, Typography } from "antd";
import { PlusOutlined, EditOutlined, DeleteOutlined } from "@ant-design/icons";
import type { FormInstance } from "antd";
import { ProTable } from "@/components/ProTable";
import type { ProTableRef, ProTableColumnType } from "@/components/ProTable";
import { ProForm } from "@/components/ProForm";
import type { ProFormFields, ProFormInstance } from "@/components/ProForm";
import { Api, IApi } from "@/apis";
import { usePerms } from "@/hooks";
import dayjs from "dayjs";
import IconSelect from "@/components/IconSelect";
import { IconMap } from "@/libs";
import { useTranslation } from "react-i18next";

// 菜单类型：M菜单 C目录 F功能（与 model.Menu.menu_type 一致）
const MENU_TYPE_OPTIONS = [
	{ label: "menu.typeMenu", value: "M" },
	{ label: "menu.typeDirectory", value: "C" },
	{ label: "menu.typeFunction", value: "F" },
];
const MENU_TYPE_TEXT: Record<string, { text: string; color: string }> = {
	M: { text: "menu.typeMenu", color: "blue" },
	C: { text: "menu.typeDirectory", color: "orange" },
	F: { text: "menu.typeFunction", color: "default" },
};
// 菜单状态与 model 一致：1=正常 0=停用（前端 buildMenuTree 以 status===1 过滤）
const MENU_STATUS_TEXT: Record<number, string> = {
	1: "common.enabled",
	0: "common.disabled",
};
const MENU_STATUS_OPTIONS = [
	{ label: "common.enabled", value: 1 },
	{ label: "common.disabled", value: 0 },
];
// 显示状态：0隐藏 1显示
const VISIBLE_TEXT: Record<number, string> = {
	1: "menu.shown",
	0: "menu.hidden",
};
const VISIBLE_OPTIONS = [
	{ label: "menu.shown", value: 1 },
	{ label: "menu.hidden", value: 0 },
];
// 权限标识三段式：模块:实体:动作（与后端 validateMenuPerms 一致）
const PERMS_RE = /^[a-z][a-z0-9_]*:[a-z][a-z0-9_]*:[a-z][a-z0-9_]*$/;

/**
 * 按表单值动态校验的规则（antd RuleRender 形式，返回单个 RuleObject）。
 * 用于「随 menuType 显隐」的字段：隐藏字段仍会参与校验，故需按类型动态返回校验结果。
 * @param check 返回错误文案，null 表示通过
 */
function menuTypeValidator(
	check: (form: FormInstance, value: unknown) => string | null,
) {
	return (form: FormInstance) => ({
		validator: (_rule: unknown, value: unknown) => {
			const msg = check(form, value);
			return msg ? Promise.reject(new Error(msg)) : Promise.resolve();
		},
	});
}

/** 菜单树 → antd TreeSelect 数据结构 */
function toTreeSelect(nodes: IApi.MenuRaw[]): any[] {
	return nodes.map((m) => ({
		value: m.id,
		title: m.menu_name,
		children: m.children?.length ? toTreeSelect(m.children) : undefined,
	}));
}

/** 过滤菜单树：仅保留 menu_type ∈ types 的节点（祖先链同步裁剪，保证树完整） */
function filterByType(nodes: IApi.MenuRaw[], types: string[]): IApi.MenuRaw[] {
	const set = new Set(types);
	const walk = (list: IApi.MenuRaw[]): IApi.MenuRaw[] =>
		list
			.filter((n) => set.has(n.menu_type))
			.map((n) => ({
				...n,
				children: n.children?.length ? walk(n.children) : undefined,
			}));
	return walk(nodes);
}

/**
 * 构建菜单表单字段（新增/编辑共用，编辑时 menuType 置灰不可切换）。
 * - 顶部 menuType 用 Radio(buttonStyle) 模拟「菜单/目录/功能」三 Tab
 * - 菜单(M)/目录(C) 显示：图标/路由地址/显示状态；菜单(M) 额外显示：组件路径/路由参数
 * - 功能(F) 仅显示：菜单名称/顺序/权限标识(必填三段式)/状态/备注
 * - 上级菜单按当前类型过滤可选父级（F 可挂 M 下，M/C 仅挂 C 目录下）
 */
function buildMenuFields(
	menuTypeDisabled: boolean,
	dirTree: any[],
	menuDirTree: any[],
	t: ReturnType<typeof useTranslation>["t"],
): ProFormFields {
	return [
		{
			type: "radio",
			label: t("menu.typeLabel"),
			name: "menuType",
			disabled: menuTypeDisabled,
			fieldProps: {
				block: true,
				options: MENU_TYPE_OPTIONS.map((item) => ({
					...item,
					label: t(item.label),
				})),
				optionType: "button",
				buttonStyle: "solid",
				disabled: true,
			},
			rules: [{ required: true, message: t("menu.typeLabel") }],
		},
		{
			label: t("menu.parent"),
			name: "parentId",
			// 自定义渲染：树需随 menuType 切换过滤范围，故不走静态 fieldProps
			render: (value, form) => {
				const type = (form.getFieldValue("menuType") as string) ?? "C";
				return (
					<TreeSelect
						value={(value as number) ?? undefined}
						treeData={type === "F" ? menuDirTree : dirTree}
						allowClear
						showSearch
						treeDefaultExpandAll
						placeholder={
							type === "F" ? t("menu.selectOwnerMenu") : t("menu.rootHint")
						}
						onChange={(v) => form.setFieldValue("parentId", v)}
					/>
				);
			},
		},
		{
			type: "input",
			label: t("menu.name"),
			name: "menuName",
			fieldProps: { maxLength: 50, placeholder: t("menu.namePlaceholder") },
			rules: [
				{ required: true, message: t("menu.nameRequired") },
				{ min: 2, max: 50, message: t("menu.nameLength") },
			],
		},
		{
			type: "input-number",
			label: t("menu.order"),
			name: "orderNum",
			fieldProps: { min: 0, style: { width: "100%" } },
		},
		{
			label: t("menu.icon"),
			name: "icon",
			render: (value: string, form) => {
				return (
					<IconSelect
						value={value}
						onChange={(v) => {
							form.setFieldValue("icon", v);
						}}
					/>
				);
			},
			visible: (values) => values.menuType !== "F",
		},
		{
			type: "input",
			label: t("menu.route"),
			name: "path",
			fieldProps: {
				maxLength: 200,
				placeholder: t("menu.routePlaceholder"),
			},
			visible: (values) => values.menuType !== "F",
			// 功能(F) 下该字段隐藏且无需校验（antd 隐藏字段仍会校验，用 RuleRender 按类型动态判断）
			rules: [
				menuTypeValidator((form, value) =>
					form.getFieldValue("menuType") === "F"
						? null
						: value
							? null
							: t("menu.enterRoute"),
				),
			],
		},
		{
			type: "input",
			label: t("menu.component"),
			name: "component",
			fieldProps: {
				maxLength: 255,
				placeholder: t("menu.componentPlaceholder"),
			},
			visible: (values) => values.menuType === "M",
			rules: [
				menuTypeValidator((form, value) =>
					form.getFieldValue("menuType") !== "M"
						? null
						: value
							? null
							: t("menu.enterComponent"),
				),
			],
		},
		{
			type: "input",
			label: t("menu.routeParams"),
			name: "query",
			fieldProps: { maxLength: 255, placeholder: t("menu.optional") },
			visible: (values) => values.menuType === "M",
		},
		{
			type: "input",
			label: t("menu.permission"),
			name: "perms",
			fieldProps: {
				maxLength: 100,
				placeholder: t("menu.permissionPlaceholder"),
			},
			// 菜单(M)/目录(C) 不再填写权限标识，仅功能(F) 显示
			visible: (values) => values.menuType === "F",
			rules: [
				menuTypeValidator((form, value) => {
					// 仅功能(F) 校验；菜单/目录隐藏字段不参与校验
					if (form.getFieldValue("menuType") !== "F") return null;
					const v = value ? String(value) : "";
					if (!v) return t("menu.permissionRequired");
					if (!PERMS_RE.test(v)) {
						return t("menu.permissionFormat");
					}
					return null;
				}),
			],
		},
		{
			type: "radio",
			label: t("menu.visible"),
			name: "visible",
			fieldProps: {
				options: VISIBLE_OPTIONS.map((item) => ({
					...item,
					label: t(item.label),
				})),
			},
			visible: (values) => values.menuType !== "F",
		},
		{
			type: "radio",
			label: t("menu.status"),
			name: "status",
			fieldProps: {
				options: MENU_STATUS_OPTIONS.map((item) => ({
					...item,
					label: t(item.label),
				})),
			},
			rules: [{ required: true, message: t("menu.statusRequired") }],
		},
		{
			type: "textarea",
			label: t("menu.note"),
			name: "remark",
			fieldProps: {
				maxLength: 500,
				rows: 3,
				placeholder: t("department.remarkPlaceholder"),
			},
		},
	];
}

/**
 * 菜单管理：树形表格 + 新增 / 编辑 / 删除。
 * - 数据源 /api/menu/tree（原生 model.Menu，snake_case，无分页 → total 置 0 隐藏分页条）
 * - 新增表单三 Tab（菜单/目录/功能）按 menu_type 显隐字段；编辑时 menuType 置灰不可切换
 * - 写接口 status 语义已与 model 统一为 1=正常 0=停用
 * - 删除走 /api/menu/delete?id=（存在子菜单或功能时后端拦截）
 */
const MenuPage: React.FC = () => {
	const { message } = App.useApp();
	const { t } = useTranslation();
	const { hasPerm } = usePerms();
	const tableRef = useRef<ProTableRef<IApi.MenuRaw>>(null);
	const formRef = useRef<ProFormInstance>(null);
	// 区分当前弹窗是新增还是编辑，并保存编辑时的记录
	const [formMode, setFormMode] = useState<{
		type: "add" | "edit";
		record?: IApi.MenuRaw;
	}>({ type: "add" });
	// 上级菜单下拉：随表格数据同步刷新
	const [menuTree, setMenuTree] = useState<IApi.MenuRaw[]>([]);

	// 父级候选树：F 功能可挂 M 菜单下；目录/菜单仅挂 C 目录下（与 seed 结构一致）
	const dirTreeData = useMemo(
		() => toTreeSelect(filterByType(menuTree, ["C"])),
		[menuTree],
	);
	const menuDirTreeData = useMemo(
		() => toTreeSelect(filterByType(menuTree, ["M"])),
		[menuTree],
	);

	// 树形表格 API：无分页
	const menuTreeApi = useCallback(async (params: Record<string, unknown>) => {
		const data = await Api.getMenuTree({
			menuName: params.menuName ? String(params.menuName) : undefined,
			status: params.status as number | undefined,
		});
		setMenuTree(data); // 同步上级菜单下拉
		return { data, total: 0 };
	}, []);

	// 动态生成字段：编辑模式 menuType 置灰
	const getFormFields = useCallback((): ProFormFields => {
		const menuTypeDisabled = formMode.type === "edit";
		return buildMenuFields(menuTypeDisabled, dirTreeData, menuDirTreeData, t);
	}, [formMode.type, dirTreeData, menuDirTreeData, t]);

	// 打开新增：指定父级时按父类型预选 menuType（C 目录下默认菜单 M；M 菜单下默认功能 F）
	const openAdd = useCallback((parent?: IApi.MenuRaw) => {
		setFormMode({ type: "add" });
		const base: Record<string, any> = { orderNum: 0, status: 1, visible: 0 };
		if (parent) {
			base.menuType = parent.menu_type === "C" ? "M" : "F";
			base.parentId = parent.id;
		} else {
			base.menuType = "C";
		}
		formRef.current?.open({
			title: parent
				? `${t("menu.addItem")}${parent.menu_type === "C" ? t("menu.typeMenu") : t("menu.typeFunction")}`
				: t("menu.add"),
			initialValues: base,
		});
	}, [t]);

	// 打开编辑
	const openEdit = useCallback((record: IApi.MenuRaw) => {
		setFormMode({ type: "edit", record });
		formRef.current?.open({
			title: `${t("menu.edit")}: ${record.menu_name}`,
			// 根菜单 parent_id=0 置 undefined，展示占位「不选择则为顶级目录」
			initialValues: {
				menuType: record.menu_type,
				parentId: record.parent_id || undefined,
				menuName: record.menu_name,
				orderNum: record.order_num,
				icon: record.icon,
				path: record.path,
				component: record.component,
				query: record.query,
				perms: record.perms,
				visible: record.visible,
				status: record.status,
				remark: record.remark,
			},
			extraParams: { id: record.id },
		});
	}, [t]);

	// 统一提交
	const handleFormSubmit = useCallback(
		async (
			_title: string,
			values: Record<string, any>,
			extraParams?: Record<string, any>,
		) => {
			const payload = {
				parentId: values.parentId ?? 0,
				menuName: values.menuName,
				menuType: values.menuType,
				orderNum: values.orderNum ?? 0,
				path: values.path ?? "",
				component: values.component ?? "",
				query: values.query ?? "",
				status: values.status,
				perms: values.perms ?? "",
				icon: values.icon ?? "",
				visible: values.visible ?? 0,
				remark: values.remark ?? "",
			};
			if (formMode.type === "add") {
				await Api.addMenu(payload);
				message.success(t("menu.addSuccess"));
			} else {
				const id = extraParams?.id as number;
				await Api.editMenu(id, payload);
				message.success(t("menu.editSuccess"));
			}
			tableRef.current?.refresh();
		},
		[formMode.type, message, t],
	);

	// ---- 删除菜单 ----
	const handleDelete = useCallback(
		async (record: IApi.MenuRaw) => {
			await Api.deleteMenu(record.id);
			message.success(t("menu.deleteSuccess"));
			tableRef.current?.refresh();
		},
		[message, t],
	);

	// ---- 列配置 ----
	const columns: ProTableColumnType<IApi.MenuRaw>[] = [
		{
			title: t("menu.name"),
			minWidth:240,
			dataIndex: "menu_name",
			render: (v, record) => (
				<div className="flex-center gap-2">
					<span className="size-8 min-w-8 flex-center text-base bg-[#EBF7FF] rounded-lg text-[var(--p)] dark:bg-[#070D10]">
						{IconMap[record.icon] ?? record.menu_name[0]} 
					</span>
					<span>{v}</span>
				</div>
			),
			width: 200,
		},
		{
			title: t("menu.type"),
			dataIndex: "menu_type",
			width: 80,
			render: (value) => {
				const cfg = MENU_TYPE_TEXT[(value as string) ?? ""];
				return cfg ? <Tag color={cfg.color}>{t(cfg.text)}</Tag> : "-";
			},
		},
		{ title: t("system.displayOrder"), dataIndex: "order_num", width: 90 },
		{
			title: t("menu.permission"),
			dataIndex: "perms",
			width: 190,
			ellipsis: true,
			render: (value) =>
				value ? <Typography.Text code>{value as string}</Typography.Text> : "-",
		},
		{
			title: t("menu.route"),
			dataIndex: "path",
			width: 130,
			render: (value) =>
				value ? <Typography.Text code>{value as string}</Typography.Text> : "-",
		},
		{
			title: t("menu.component"),
			dataIndex: "component",
			width: 190,
			ellipsis: true,
			render: (value) =>
				value ? <Typography.Text code>{value as string}</Typography.Text> : "-",
		},
		{
			title: t("menu.visible"),
			dataIndex: "visible",
			width: 70,
			render: (value) => (
				<Tag color={(value as number) === 1 ? "green" : "red"}>
					{t(VISIBLE_TEXT[(value as number) ?? 0] ?? "-")}
				</Tag>
			),
		},
		{
			title: t("system.status"),
			dataIndex: "status",
			width: 80,
			render: (value) => (
				<Tag color={(value as number) === 1 ? "green" : "red"}>
					{t(MENU_STATUS_TEXT[(value as number) ?? 0] ?? "-")}
				</Tag>
			),
		},
		{
			title: t("system.createdAt"),
			dataIndex: "createdAt",
			width: 170,
			render: (time) => dayjs(time).format("YYYY-MM-DD HH:mm:ss"),
		},
		{
			title: t("system.action"),
			key: "action",
			width: 260,
			fixed: "right",
			render: (_, record) => (
				<div className="flex items-center gap-2">
					{hasPerm("system:menu:add") && record.menu_type !== "F" && (
						<Button
							size="small"
							color="primary"
							variant="filled"
							icon={<PlusOutlined />}
							onClick={() => openAdd(record)}
						>
							{t("menu.addItem")}
							{record.menu_type === "C"
								? t("menu.typeMenu")
								: t("menu.typeFunction")}
						</Button>
					)}
					{hasPerm("system:menu:edit") && (
						<Button
							size="small"
							color="primary"
							variant="filled"
							icon={<EditOutlined />}
							onClick={() => openEdit(record)}
						>
							{t("system.edit")}
						</Button>
					)}
					{hasPerm("system:menu:remove") && (
						<Popconfirm
							title={t("menu.deleteConfirm")}
							description={t("menu.deleteBlocked")}
							okText={t("system.delete")}
							cancelText={t("common.cancel")}
							okButtonProps={{ danger: true }}
							onConfirm={() => handleDelete(record)}
						>
							<Button
								size="small"
								color="danger"
								variant="filled"
								icon={<DeleteOutlined />}
							>
								{t("system.delete")}
							</Button>
						</Popconfirm>
					)}
				</div>
			),
		},
	];

	return (
		<>
			<ProTable<IApi.MenuRaw>
				ref={tableRef}
				rowKey="id"
				api={menuTreeApi}
				columns={columns}
				title={t("menu.title")}
				index={false}
				toolbarExtra={
					hasPerm("system:menu:add") ? (
						<Button
							type="primary"
							icon={<PlusOutlined />}
							onClick={() => openAdd()}
						>
							{t("menu.add")}
						</Button>
					) : undefined
				}
				search={{
					fields: [
						{ type: "input", label: t("menu.name"), name: "menuName" },
						{
							type: "select",
							label: t("system.status"),
							name: "status",
							fieldProps: {
								options: MENU_STATUS_OPTIONS.map((item) => ({
									...item,
									label: t(item.label),
								})),
								allowClear: true,
							},
						},
					],
				}}
			/>
			{/* 菜单表单：新增/编辑共用同一抽屉，编辑时 menuType 置灰不可切换 */}
			<ProForm
				type="drawer"
				ref={formRef}
				fields={getFormFields()}
				onConfirm={handleFormSubmit}
			/>
		</>
	);
};

export default MenuPage;
