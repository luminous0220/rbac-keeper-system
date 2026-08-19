import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { App, Button, Form, Popconfirm, Space, Tag } from "antd";
import { ProTable } from "@/components/ProTable";
import type { ProTableRef, ProTableColumnType } from "@/components/ProTable";
import { ProForm } from "@/components/ProForm";
import type { ProFormFields, ProFormInstance } from "@/components/ProForm";
import { TreeSelectorModal } from "@/components/ModalSelector";
import type {
	TreeSelectorModalRef,
	TreeNodeData,
	TreeSelectorNode,
} from "@/components/ModalSelector";
import { Api, IApi } from "@/apis";
import { usePerms } from "@/hooks";
import dayjs from "dayjs";

// 数据权限文案（与后端常量一致）
const DATA_SCOPE_OPTIONS = [
	{ label: "全部数据", value: 1 },
	{ label: "自定义数据", value: 2 },
	{ label: "本部门数据", value: 3 },
	{ label: "本部门及以下数据", value: 4 },
	{ label: "仅本人数据", value: 5 },
];
const DATA_SCOPE_TEXT: Record<number, string> = Object.fromEntries(
	DATA_SCOPE_OPTIONS.map((o) => [o.value, o.label]),
);
// 角色状态：0禁用 1正常
const STATUS_TEXT: Record<number, string> = { 0: "禁用", 1: "正常" };
const STATUS_OPTIONS = [
	{ label: "正常", value: 1 },
	{ label: "禁用", value: 0 },
];

/** 菜单树 → TreeSelectorModal 节点数据 */
function toMenuTreeNodes(nodes: IApi.MenuRaw[]): TreeNodeData[] {
	return nodes.map((m) => ({
		value: m.id,
		title: m.menu_name,
		children: m.children?.length ? toMenuTreeNodes(m.children) : undefined,
	}));
}
/** 部门树 → TreeSelectorModal 节点数据 */
function toDeptTreeNodes(nodes: IApi.DeptRaw[]): TreeNodeData[] {
	return nodes.map((d) => ({
		value: d.id,
		title: d.dept_name,
		children: d.children?.length ? toDeptTreeNodes(d.children) : undefined,
	}));
}
/** 扁平化 id → 名称 映射（用于触发器回显选中项名称） */
function collectNameMap<T extends { id: number; children?: T[] }>(
	nodes: T[],
	getName: (n: T) => string,
): Map<number, string> {
	const map = new Map<number, string>();
	const walk = (list: T[]) => {
		for (const n of list) {
			map.set(n.id, getName(n));
			if (n.children?.length) walk(n.children);
		}
	};
	walk(nodes);
	return map;
}

/** 树选择触发器：展示已选名称，点击打开 TreeSelectorModal */
function TreeTrigger({
	value,
	nameMap,
	placeholder,
	onOpen,
}: {
	value?: number[];
	nameMap: Map<number, string>;
	placeholder: string;
	onOpen: () => void;
}) {
	const ids = value ?? [];
	const shown = ids.slice(0, 3).map((id) => nameMap.get(id) ?? `#${id}`);
	return (
		<div
			onClick={onOpen}
			className="w-full min-h-[32px] px-2 py-1 border border-[var(--ant-color-border)] rounded-md cursor-pointer hover:border-[var(--ant-color-primary)] flex items-center flex-wrap gap-1"
		>
			{ids.length === 0 ? (
				<span className="text-[var(--ant-color-text-placeholder)]">
					{placeholder}
				</span>
			) : (
				<Space size={4} wrap>
					{shown.map((n) => (
						<Tag key={n}>{n}</Tag>
					))}
					{ids.length > shown.length && <Tag>+{ids.length - shown.length}</Tag>}
				</Space>
			)}
		</div>
	);
}

/**
 * 角色管理：列表 / 新增 / 编辑 / 删除。
 * - 列表返回原生 model.Role（snake_case，含 menus/depts 预加载）
 * - menuIds / deptIds 用 TreeSelectorModal 选择（deptIds 仅在 dataScope=2 时显示且必填）
 * - 删除走 /api/role/delete?roleKey=（后端校验绑定用户则拦截；super 角色禁止删除）
 * - role_key 创建后不可变，编辑时置灰
 */
const RolePage: React.FC = () => {
	const { message } = App.useApp();
	const { hasPerm } = usePerms();
	const tableRef = useRef<ProTableRef<IApi.RoleItem>>(null);
	const addRoleFormRef = useRef<ProFormInstance>(null);
	const editRoleFormRef = useRef<ProFormInstance>(null);
	const menuTreeModalRef = useRef<TreeSelectorModalRef>(null);
	const deptTreeModalRef = useRef<TreeSelectorModalRef>(null);
	// 当前树选择写入的目标表单与字段（add/edit 共用两个弹窗选择器）
	const targetFormRef = useRef<ProFormInstance | null>(null);
	const targetFieldRef = useRef<"menuIds" | "deptIds">("menuIds");

	const [menuTree, setMenuTree] = useState<IApi.MenuRaw[]>([]);
	const [deptTree, setDeptTree] = useState<IApi.DeptRaw[]>([]);

	useEffect(() => {
		Api.getMenuTree()
			.then(setMenuTree)
			.catch((err) => console.error("加载菜单树失败", err));
		Api.getDeptTree()
			.then(setDeptTree)
			.catch((err) => console.error("加载部门树失败", err));
	}, []);

	const menuTreeData = useMemo(() => toMenuTreeNodes(menuTree), [menuTree]);
	const deptTreeData = useMemo(() => toDeptTreeNodes(deptTree), [deptTree]);
	const menuNameMap = useMemo(
		() => collectNameMap(menuTree, (m) => m.menu_name),
		[menuTree],
	);
	const deptNameMap = useMemo(
		() => collectNameMap(deptTree, (d) => d.dept_name),
		[deptTree],
	);

	// ---- 树选择联动 ----
	const openMenuSelector = useCallback(
		(formRef: React.MutableRefObject<ProFormInstance | null>) => {
			const current = (formRef.current?.getFieldValue("menuIds") ??
				[]) as number[];
			targetFormRef.current = formRef.current;
			targetFieldRef.current = "menuIds";
			menuTreeModalRef.current?.open({
				title: "选择菜单权限",
				checkedKeys: current.map((id) => ({
					value: id,
					title: menuNameMap.get(id) ?? `#${id}`,
					type: "child",
				})),
			});
		},
		[menuNameMap],
	);

	const openDeptSelector = useCallback(
		(formRef: React.MutableRefObject<ProFormInstance | null>) => {
			const current = (formRef.current?.getFieldValue("deptIds") ??
				[]) as number[];
			targetFormRef.current = formRef.current;
			targetFieldRef.current = "deptIds";
			deptTreeModalRef.current?.open({
				title: "选择自定义数据权限部门",
				checkedKeys: current.map((id) => ({
					value: id,
					title: deptNameMap.get(id) ?? `#${id}`,
					type: "child",
				})),
			});
		},
		[deptNameMap],
	);

	const handleTreeConfirm = useCallback((selected: TreeSelectorNode[]) => {
		targetFormRef.current?.setFieldValue(
			targetFieldRef.current,
			selected.map((s) => s.value as number),
		);
	}, []);

	// ---- 新增角色 ----
	const addRoleFields: ProFormFields = [
		{
			type: "input",
			label: "角色名称",
			name: "roleName",
			fieldProps: { maxLength: 30, placeholder: "2-30 个字符" },
			rules: [
				{ required: true, message: "请输入角色名称" },
				{ min: 2, max: 30, message: "角色名称长度需为 2-30 位" },
			],
		},
		{
			type: "input",
			label: "角色标识",
			name: "roleKey",
			fieldProps: {
				maxLength: 32,
				placeholder: "字母数字组合，创建后不可修改",
			},
			rules: [
				{ required: true, message: "请输入角色标识" },
				{ min: 2, max: 32, message: "角色标识长度需为 2-32 位" },
				{ pattern: /^[a-zA-Z0-9]+$/, message: "角色标识只能包含字母和数字" },
			],
		},
		{
			type: "input-number",
			label: "显示顺序",
			name: "roleSort",
			fieldProps: { min: 0, style: { width: "100%" } },
		},
		{
			type: "radio",
			label: "数据权限",
			name: "dataScope",
			fieldProps: { options: DATA_SCOPE_OPTIONS },
			rules: [{ required: true, message: "请选择数据权限" }],
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

	const handleRoleAdd = useCallback(
		async (_title: string, values: Record<string, any>) => {
			await Api.addRole({
				roleName: values.roleName,
				roleKey: values.roleKey,
				roleSort: values.roleSort ?? 0,
				dataScope: values.dataScope,
				menuIds: values.menuIds ?? [],
				deptIds: values.deptIds ?? [],
				remark: values.remark ?? "",
			});
			message.success("新增角色成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	// ---- 编辑角色 ----
	const editRoleFields: ProFormFields = [
		{
			type: "input",
			label: "角色名称",
			name: "roleName",
			fieldProps: { maxLength: 30 },
			rules: [
				{ required: true, message: "请输入角色名称" },
				{ min: 2, max: 30, message: "角色名称长度需为 2-30 位" },
			],
		},
		{
			type: "input",
			label: "角色标识",
			name: "roleKey",
			disabled: true,
			fieldProps: { disabled: true, placeholder: "创建后不可修改" },
		},
		{
			type: "input-number",
			label: "显示顺序",
			name: "roleSort",
			fieldProps: { min: 0, style: { width: "100%" } },
		},
		{
			type: "radio",
			label: "数据权限",
			name: "dataScope",
			fieldProps: { options: DATA_SCOPE_OPTIONS },
			rules: [{ required: true, message: "请选择数据权限" }],
		},
		{
			type: "radio",
			label: "状态",
			name: "status",
			fieldProps: { options: STATUS_OPTIONS },
			rules: [{ required: true, message: "请选择状态" }],
		},
		{
			type: "textarea",
			label: "备注",
			name: "remark",
			fieldProps: { maxLength: 500, rows: 3 },
		},
	];

	const openEdit = useCallback((record: IApi.RoleItem) => {
		editRoleFormRef.current?.open({
			title: `编辑角色：${record.role_name}`,
			initialValues: {
				roleName: record.role_name,
				roleKey: record.role_key,
				roleSort: record.role_sort,
				dataScope: record.data_scope,
				status: record.status,
				remark: record.remark,
				menuIds: record.menus?.map((m) => m.id) ?? [],
				deptIds: record.depts?.map((d) => d.id) ?? [],
			},
			extraParams: { roleKey: record.role_key },
		});
	}, []);

	const handleRoleEdit = useCallback(
		async (
			_title: string,
			values: Record<string, any>,
			extraParams?: Record<string, any>,
		) => {
			const roleKey = extraParams?.roleKey as string;
			await Api.editRole(roleKey, {
				roleName: values.roleName,
				roleSort: values.roleSort ?? 0,
				dataScope: values.dataScope,
				status: values.status,
				remark: values.remark ?? "",
				menuIds: values.menuIds ?? [],
				deptIds: values.deptIds ?? [],
			});
			message.success("修改角色成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	// ---- 删除角色 ----
	const handleRoleDelete = useCallback(
		async (record: IApi.RoleItem) => {
			await Api.deleteRole(record.role_key);
			message.success("删除角色成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	// ---- 列配置 ----
	const columns: ProTableColumnType<IApi.RoleItem>[] = [
		{ title: "角色名称", dataIndex: "role_name", width: 140 },
		{ title: "角色标识", dataIndex: "role_key", width: 120 },
		{ title: "显示顺序", dataIndex: "role_sort", width: 90 },
		{
			title: "数据权限",
			dataIndex: "data_scope",
			width: 140,
			render: (value) => DATA_SCOPE_TEXT[(value as number) ?? 1] ?? "-",
		},
		{
			title: "菜单授权",
			dataIndex: "menus",
			width: 100,
			render: (_, record) => record.menus?.length ?? 0,
		},
		{
			title: "状态",
			dataIndex: "status",
			width: 80,
			render: (value) => (
				<Tag color={(value as number) === 1 ? "green" : "red"}>
					{STATUS_TEXT[(value as number) ?? 0] ?? "-"}
				</Tag>
			),
		},
		{ title: "备注", dataIndex: "remark", ellipsis: true },
		{
			title: "创建时间",
			dataIndex: "createdAt",
			width: 170,
			render: (time) => dayjs(time).format("YYYY-MM-DD HH:mm:ss"),
		},
		{
			title: "操作",
			key: "action",
			width: 140,
			fixed: "right",
			render: (_, record) => (
				<Space size={0}>
					{hasPerm("system:role") && (
						<Button type="link" size="small" onClick={() => openEdit(record)}>
							编辑
						</Button>
					)}
					{hasPerm("system:role:remove") && record.role_key !== "super" && (
						<Popconfirm
							title="确定删除该角色吗？"
							description="若角色已分配给用户将被拦截"
							okText="删除"
							cancelText="取消"
							okButtonProps={{ danger: true }}
							onConfirm={() => handleRoleDelete(record)}
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

	// 树选择触发器（表单 children 内嵌，add/edit 各一套）
	const renderTreeTriggers = (
		formRef: React.MutableRefObject<ProFormInstance | null>,
	) => (
		<>
			<Form.Item label="菜单权限" name="menuIds">
				<TreeTrigger
					nameMap={menuNameMap}
					placeholder="请选择菜单权限"
					onOpen={() => openMenuSelector(formRef)}
				/>
			</Form.Item>
			{/* 自定义数据权限：仅 dataScope=2 时显示且必填（与后端 required_if 对齐） */}
			<Form.Item
				noStyle
				shouldUpdate={(prev, curr) => prev.dataScope !== curr.dataScope}
			>
				{({ getFieldValue }) =>
					getFieldValue("dataScope") === 2 ? (
						<Form.Item
							label="自定义数据权限部门"
							name="deptIds"
							rules={[{ required: true, message: "请选择部门" }]}
						>
							<TreeTrigger
								nameMap={deptNameMap}
								placeholder="请选择部门"
								onOpen={() => openDeptSelector(formRef)}
							/>
						</Form.Item>
					) : null
				}
			</Form.Item>
		</>
	);

	return (
		<>
			<ProTable<IApi.RoleItem>
				ref={tableRef}
				rowKey="id"
				api={async (params) => {
					const list = await Api.getRoleList(params);
					return { data: list };
				}}
				columns={columns}
				title="角色管理"
				toolbarExtra={
					hasPerm("system:role:add") ? (
						<Button
							type="primary"
							onClick={() =>
								addRoleFormRef.current?.open({
									title: "新增角色",
									initialValues: { roleSort: 0, dataScope: 1 },
								})
							}
						>
							新增角色
						</Button>
					) : undefined
				}
				search={{
					fields: [
						{ type: "input", label: "角色名称", name: "roleName" },
						{ type: "input", label: "角色标识", name: "roleKey" },
						{
							type: "select",
							label: "状态",
							name: "status",
							fieldProps: { options: STATUS_OPTIONS, allowClear: true },
						},
					],
				}}
			/>

			{/* 新增角色弹窗 */}
			<ProForm
				ref={addRoleFormRef}
				fields={addRoleFields}
				onConfirm={handleRoleAdd}
				width={640}
			>
				{renderTreeTriggers(addRoleFormRef)}
			</ProForm>
			{/* 编辑角色弹窗 */}
			<ProForm
				ref={editRoleFormRef}
				fields={editRoleFields}
				onConfirm={handleRoleEdit}
				width={640}
			>
				{renderTreeTriggers(editRoleFormRef)}
			</ProForm>

			{/* 菜单权限 / 部门选择弹窗（add/edit 共用） */}
			<TreeSelectorModal
				ref={menuTreeModalRef}
				treeData={menuTreeData}
				width={820}
				onConfirm={handleTreeConfirm}
			/>
			<TreeSelectorModal
				ref={deptTreeModalRef}
				treeData={deptTreeData}
				width={820}
				onConfirm={handleTreeConfirm}
			/>
		</>
	);
};

export default RolePage;
