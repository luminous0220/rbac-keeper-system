import { useCallback, useMemo, useRef, useState } from "react";
import {
	App,
	Button,
	Popconfirm,
	Space,
	Tag,
	TreeSelect,
	Typography,
} from "antd";
import type { FormInstance } from "antd";
import { ProTable } from "@/components/ProTable";
import type { ProTableRef, ProTableColumnType } from "@/components/ProTable";
import { ProForm } from "@/components/ProForm";
import type { ProFormFields, ProFormInstance } from "@/components/ProForm";
import { Api, IApi } from "@/apis";
import { usePerms } from "@/hooks";
import dayjs from "dayjs";

// 菜单类型：M目录 C菜单 F按钮（与 model.Menu.menu_type 一致）
const MENU_TYPE_OPTIONS = [
	{ label: "目录", value: "M" },
	{ label: "菜单", value: "C" },
	{ label: "按钮", value: "F" },
];
const MENU_TYPE_TEXT: Record<string, { text: string; color: string }> = {
	M: { text: "目录", color: "orange" },
	C: { text: "菜单", color: "blue" },
	F: { text: "按钮", color: "default" },
};
// 菜单状态与 model 一致：1=正常 0=停用（前端 buildMenuTree 以 status===1 过滤）
const MENU_STATUS_TEXT: Record<number, string> = { 1: "正常", 0: "停用" };
const MENU_STATUS_OPTIONS = [
	{ label: "正常", value: 1 },
	{ label: "停用", value: 0 },
];
// 显示状态：0显示 1隐藏
const VISIBLE_TEXT: Record<number, string> = { 0: "显示", 1: "隐藏" };
const VISIBLE_OPTIONS = [
	{ label: "显示", value: 0 },
	{ label: "隐藏", value: 1 },
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
 * - 顶部 menuType 用 Radio(buttonStyle) 模拟「目录/菜单/按钮」三 Tab
 * - 目录(M)/菜单(C) 显示：图标/路由地址/显示状态；菜单(C) 额外显示：组件路径/路由参数
 * - 按钮(F) 仅显示：菜单名称/顺序/权限标识(必填三段式)/状态/备注
 * - 上级菜单按当前类型过滤可选父级（F 可挂 M/C 下，M/C 仅挂 M 目录下）
 */
function buildMenuFields(
	menuTypeDisabled: boolean,
	dirTree: any[],
	menuDirTree: any[],
): ProFormFields {
	return [
		{
			type: "radio",
			label: "菜单类型",
			name: "menuType",
			fieldProps: {
				options: MENU_TYPE_OPTIONS,
				optionType: "button",
				buttonStyle: "solid",
				disabled: menuTypeDisabled,
			},
			rules: [{ required: true, message: "请选择菜单类型" }],
		},
		{
			label: "上级菜单",
			name: "parentId",
			// 自定义渲染：树需随 menuType 切换过滤范围，故不走静态 fieldProps
			render: (value, form) => {
				const type = (form.getFieldValue("menuType") as string) ?? "M";
				return (
					<TreeSelect
						value={(value as number) ?? undefined}
						treeData={type === "F" ? menuDirTree : dirTree}
						allowClear
						showSearch
						treeDefaultExpandAll
						placeholder={type === "F" ? "请选择所属菜单" : "不选择则为顶级目录"}
						onChange={(v) => form.setFieldValue("parentId", v)}
					/>
				);
			},
		},
		{
			type: "input",
			label: "菜单名称",
			name: "menuName",
			fieldProps: { maxLength: 50, placeholder: "2-50 个字符" },
			rules: [
				{ required: true, message: "请输入菜单名称" },
				{ min: 2, max: 50, message: "菜单名称长度需为 2-50 位" },
			],
		},
		{
			type: "input-number",
			label: "显示顺序",
			name: "orderNum",
			fieldProps: { min: 0, style: { width: "100%" } },
		},
		{
			type: "input",
			label: "图标",
			name: "icon",
			fieldProps: {
				maxLength: 100,
				placeholder: "图标名称，如 user / system（选填）",
			},
			visible: (values) => values.menuType !== "F",
		},
		{
			type: "input",
			label: "路由地址",
			name: "path",
			fieldProps: {
				maxLength: 200,
				placeholder: "如 system（目录）或 user（菜单）",
			},
			visible: (values) => values.menuType !== "F",
			// 按钮(F) 下该字段隐藏且无需校验（antd 隐藏字段仍会校验，用 RuleRender 按类型动态判断）
			rules: [
				menuTypeValidator((form, value) =>
					form.getFieldValue("menuType") === "F"
						? null
						: value
							? null
							: "请输入路由地址",
				),
			],
		},
		{
			type: "input",
			label: "组件路径",
			name: "component",
			fieldProps: {
				maxLength: 255,
				placeholder: "如 system/user/index（菜单必填）",
			},
			visible: (values) => values.menuType === "C",
			rules: [
				menuTypeValidator((form, value) =>
					form.getFieldValue("menuType") !== "C"
						? null
						: value
							? null
							: "请输入组件路径",
				),
			],
		},
		{
			type: "input",
			label: "路由参数",
			name: "query",
			fieldProps: { maxLength: 255, placeholder: "选填" },
			visible: (values) => values.menuType === "C",
		},
		{
			type: "input",
			label: "权限标识",
			name: "perms",
			fieldProps: {
				maxLength: 100,
				placeholder: "三段式：模块:实体:动作，如 system:user:list（按钮必填）",
			},
			rules: [
				menuTypeValidator((form, value) => {
					const isF = form.getFieldValue("menuType") === "F";
					const v = value ? String(value) : "";
					if (isF && !v) return "按钮必须填写权限标识";
					if (v && !PERMS_RE.test(v)) {
						return "权限标识需为 模块:实体:动作 三段式（小写字母/数字/下划线）";
					}
					return null;
				}),
			],
		},
		{
			type: "radio",
			label: "显示状态",
			name: "visible",
			fieldProps: { options: VISIBLE_OPTIONS },
			visible: (values) => values.menuType !== "F",
		},
		{
			type: "radio",
			label: "菜单状态",
			name: "status",
			fieldProps: { options: MENU_STATUS_OPTIONS },
			rules: [{ required: true, message: "请选择菜单状态" }],
		},
		{
			type: "textarea",
			label: "备注",
			name: "remark",
			fieldProps: {
				maxLength: 500,
				rows: 3,
				placeholder: "请输入备注（选填）",
			},
		},
	];
}

/**
 * 菜单管理：树形表格 + 新增 / 编辑 / 删除。
 * - 数据源 /api/menu/tree（原生 model.Menu，snake_case，无分页 → total 置 0 隐藏分页条）
 * - 新增表单三 Tab（目录/菜单/按钮）按 menu_type 显隐字段；编辑时 menuType 置灰不可切换
 * - 写接口 status 语义已与 model 统一为 1=正常 0=停用
 * - 删除走 /api/menu/delete?id=（存在子菜单或按钮时后端拦截）
 */
const MenuPage: React.FC = () => {
	const { message } = App.useApp();
	const { hasPerm } = usePerms();
	const tableRef = useRef<ProTableRef<IApi.MenuRaw>>(null);
	const addFormRef = useRef<ProFormInstance>(null);
	const editFormRef = useRef<ProFormInstance>(null);
	// 上级菜单下拉：随表格数据同步刷新
	const [menuTree, setMenuTree] = useState<IApi.MenuRaw[]>([]);

	// 父级候选树：F 按钮可挂 M/C 下；目录/菜单仅挂 M 目录下（与 seed 结构一致）
	const dirTreeData = useMemo(
		() => toTreeSelect(filterByType(menuTree, ["M"])),
		[menuTree],
	);
	const menuDirTreeData = useMemo(
		() => toTreeSelect(filterByType(menuTree, ["M", "C"])),
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

	const addFields = useMemo(
		() => buildMenuFields(false, dirTreeData, menuDirTreeData),
		[dirTreeData, menuDirTreeData],
	);
	const editFields = useMemo(
		() => buildMenuFields(true, dirTreeData, menuDirTreeData),
		[dirTreeData, menuDirTreeData],
	);

	// 打开新增：指定父级时按父类型预选 menuType（M 下默认菜单 C；C 下默认按钮 F）
	const openAdd = useCallback((parent?: IApi.MenuRaw) => {
		const base: Record<string, any> = { orderNum: 0, status: 1, visible: 0 };
		if (parent) {
			base.menuType = parent.menu_type === "M" ? "C" : "F";
			base.parentId = parent.id;
		} else {
			base.menuType = "M";
		}
		addFormRef.current?.open({
			title: parent
				? `新增${parent.menu_type === "M" ? "菜单" : "按钮"}`
				: "新增菜单",
			initialValues: base,
		});
	}, []);

	const handleAdd = useCallback(
		async (_title: string, values: Record<string, any>) => {
			await Api.addMenu({
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
			});
			message.success("新增菜单成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	// ---- 编辑菜单 ----
	const openEdit = useCallback((record: IApi.MenuRaw) => {
		editFormRef.current?.open({
			title: `编辑菜单：${record.menu_name}`,
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
	}, []);

	const handleEdit = useCallback(
		async (
			_title: string,
			values: Record<string, any>,
			extraParams?: Record<string, any>,
		) => {
			const id = extraParams?.id as number;
			await Api.editMenu(id, {
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
			});
			message.success("修改菜单成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	// ---- 删除菜单 ----
	const handleDelete = useCallback(
		async (record: IApi.MenuRaw) => {
			await Api.deleteMenu(record.id);
			message.success("删除菜单成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	// ---- 列配置 ----
	const columns: ProTableColumnType<IApi.MenuRaw>[] = [
		{ title: "菜单名称", dataIndex: "menu_name", width: 200 },
		{
			title: "类型",
			dataIndex: "menu_type",
			width: 80,
			render: (value) => {
				const cfg = MENU_TYPE_TEXT[(value as string) ?? ""];
				return cfg ? <Tag color={cfg.color}>{cfg.text}</Tag> : "-";
			},
		},
		{ title: "显示顺序", dataIndex: "order_num", width: 90 },
		{
			title: "权限标识",
			dataIndex: "perms",
			width: 190,
			ellipsis: true,
			render: (value) =>
				value ? <Typography.Text code>{value as string}</Typography.Text> : "-",
		},
		{
			title: "路由地址",
			dataIndex: "path",
			width: 130,
			render: (value) => (value ? String(value) : "-"),
		},
		{
			title: "组件路径",
			dataIndex: "component",
			width: 190,
			ellipsis: true,
			render: (value) => (value ? String(value) : "-"),
		},
		{
			title: "可见",
			dataIndex: "visible",
			width: 70,
			render: (value) => VISIBLE_TEXT[(value as number) ?? 0] ?? "-",
		},
		{
			title: "状态",
			dataIndex: "status",
			width: 80,
			render: (value) => (
				<Tag color={(value as number) === 1 ? "green" : "red"}>
					{MENU_STATUS_TEXT[(value as number) ?? 0] ?? "-"}
				</Tag>
			),
		},
		{
			title: "创建时间",
			dataIndex: "createdAt",
			width: 170,
			render: (time) => dayjs(time).format("YYYY-MM-DD HH:mm:ss"),
		},
		{
			title: "操作",
			key: "action",
			width: 190,
			fixed: "right",
			render: (_, record) => (
				<Space size={0}>
					{hasPerm("system:menu:add") && record.menu_type !== "F" && (
						<Button type="link" size="small" onClick={() => openAdd(record)}>
							新增{record.menu_type === "M" ? "菜单" : "按钮"}
						</Button>
					)}
					{hasPerm("system:menu:edit") && (
						<Button type="link" size="small" onClick={() => openEdit(record)}>
							编辑
						</Button>
					)}
					{hasPerm("system:menu:remove") && (
						<Popconfirm
							title="确定删除该菜单吗？"
							description="存在子菜单或按钮时将被拦截"
							okText="删除"
							cancelText="取消"
							okButtonProps={{ danger: true }}
							onConfirm={() => handleDelete(record)}
						>
							<Button type="link" size="small" danger>
								删除
							</Button>
						</Popconfirm>
					)}
				</Space>
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
				title="菜单管理"
				index={false}
				toolbarExtra={
					hasPerm("system:menu:add") ? (
						<Button type="primary" onClick={() => openAdd()}>
							新增菜单
						</Button>
					) : undefined
				}
				search={{
					fields: [
						{ type: "input", label: "菜单名称", name: "menuName" },
						{
							type: "select",
							label: "状态",
							name: "status",
							fieldProps: { options: MENU_STATUS_OPTIONS, allowClear: true },
						},
					],
				}}
			/>
			{/* 新增菜单弹窗（三 Tab：目录/菜单/按钮） */}
			<ProForm
				ref={addFormRef}
				fields={addFields}
				onConfirm={handleAdd}
				width={640}
			/>
			{/* 编辑菜单弹窗（menuType 置灰不可切换） */}
			<ProForm
				ref={editFormRef}
				fields={editFields}
				onConfirm={handleEdit}
				width={640}
			/>
		</>
	);
};

export default MenuPage;
