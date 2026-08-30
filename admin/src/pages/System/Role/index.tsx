import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { App, Button, Popconfirm, Tag } from "antd";
import { PlusOutlined, EditOutlined, DeleteOutlined } from "@ant-design/icons";
import { ProTable } from "@/components/ProTable";
import type { ProTableRef, ProTableColumnType } from "@/components/ProTable";
import { ProForm } from "@/components/ProForm";
import type { ProFormFields, ProFormInstance } from "@/components/ProForm";
import { TreeSelector } from "@/components/ModalSelector";
import type { TreeNodeData } from "@/components/ModalSelector";
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

/** 菜单树 → TreeSelector 节点数据 */
function toMenuTreeNodes(nodes: IApi.MenuRaw[]): TreeNodeData[] {
	return nodes.map((m) => ({
		value: m.id,
		title: m.menu_name,
		children: m.children?.length ? toMenuTreeNodes(m.children) : undefined,
	}));
}
/** 部门树 → TreeSelector 节点数据 */
function toDeptTreeNodes(nodes: IApi.DeptRaw[]): TreeNodeData[] {
	return nodes.map((d) => ({
		value: d.id,
		title: d.dept_name,
		children: d.children?.length ? toDeptTreeNodes(d.children) : undefined,
	}));
}

/**
 * 角色管理：列表 / 新增 / 编辑 / 删除。
 * - 列表返回原生 model.Role（snake_case，含 menus/depts 预加载）
 * - menuIds / deptIds 用 TreeSelector 选择（deptIds 仅在 dataScope=2 时显示且必填）
 * - 删除走 /api/role/delete?roleKey=（后端校验绑定用户则拦截；super 角色禁止删除）
 * - role_key 创建后不可变，编辑时置灰
 */
const RolePage: React.FC = () => {
	const { message } = App.useApp();
	const { hasPerm } = usePerms();
	const tableRef = useRef<ProTableRef<IApi.RoleItem>>(null);
	const formRef = useRef<ProFormInstance>(null);
	// 区分当前弹窗是新增还是编辑，并保存编辑时的记录
	const [formMode, setFormMode] = useState<{
		type: "add" | "edit";
		record?: IApi.RoleItem;
	}>({ type: "add" });

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

	// 动态生成表单字段：编辑 mode 下 roleKey 置灰且追加 status；新增无需 status
	const getFormFields = useCallback((): ProFormFields => {
		const isEdit = formMode.type === "edit";

		const fields: ProFormFields = [
			{
				type: "input",
				label: "角色名称",
				name: "roleName",
				fieldProps: {
					maxLength: 30,
					placeholder: "2-30 个字符",
				},
				rules: [
					{ required: true, message: "请输入角色名称" },
					{ min: 2, max: 30, message: "角色名称长度需为 2-30 位" },
				],
			},
			{
				type: "input",
				label: "角色标识",
				name: "roleKey",
				disabled: isEdit,
				fieldProps: {
					maxLength: 32,
					placeholder: isEdit
						? "创建后不可修改"
						: "字母数字组合，创建后不可修改",
				},
				rules: isEdit
					? undefined
					: [
							{ required: true, message: "请输入角色标识" },
							{ min: 2, max: 32, message: "角色标识长度需为 2-32 位" },
							{
								pattern: /^[a-zA-Z0-9]+$/,
								message: "角色标识只能包含字母和数字",
							},
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
				formItemProps: { className: "col-span-3" },
				fieldProps: { options: DATA_SCOPE_OPTIONS },
				rules: [{ required: true, message: "请选择数据权限" }],
			},
			// 菜单权限：自定义渲染 TreeSelector（受控：checkedKeys/onChange 直接使用 id 数组）
			{
				label: "菜单权限",
				name: "menuIds",
				formItemProps: { className: "col-span-3" },
				render: (value, form) => (
					<TreeSelector
						title="选择菜单权限"
						treeData={menuTreeData}
						checkedKeys={(value as number[] | undefined) ?? []}
						onChange={(keys) => {
							form.setFieldValue(
								"menuIds",
								keys.map((k) => Number(k)),
							);
						}}
						placeholder="请选择菜单权限"
						width={820}
					/>
				),
			},
			// 自定义数据权限：仅 dataScope=2 时显示（StandardField 已支持自定义字段的 visible 函数）
			{
				label: "自定义数据权限部门",
				name: "deptIds",
				formItemProps: { className: "col-span-3" },
				visible: (values) => values.dataScope === 2,
				rules: [
					{ required: true, message: "请选择数据权限部门" },
					{
						validator: (_rule: unknown, value: unknown) => {
							const v = value as number[] | undefined;
							return v && v.length > 0 ? Promise.resolve() : Promise.reject();
						},
					},
				],

				// rules: (values) =>
				// 	values.dataScope === 2
				// 		? [
				// 				{ required: true, message: "请选择数据权限部门" },
				// 				{
				// 					validator: (_rule: unknown, value: unknown) => {
				// 						const v = value as number[] | undefined;
				// 						return v && v.length > 0
				// 							? Promise.resolve()
				// 							: Promise.reject();
				// 					},
				// 				},
				// 			]
				// 		: [],
				render: (value, form) => (
					<TreeSelector
						title="选择自定义数据权限部门"
						treeData={deptTreeData}
						checkedKeys={(value as number[] | undefined) ?? []}
						onChange={(keys) => {
							form.setFieldValue(
								"deptIds",
								keys.map((k) => Number(k)),
							);
						}}
						placeholder="请选择部门"
						width={820}
					/>
				),
			},
			{
				type: "textarea",
				label: "备注",
				name: "remark",
				formItemProps: { className: "col-span-3" },
				fieldProps: {
					maxLength: 500,
					rows: 3,
					placeholder: "请输入备注（选填）",
				},
			},
		];

		// 编辑模式追加状态字段
		if (isEdit) {
			fields.splice(5, 0, {
				type: "radio",
				label: "状态",
				name: "status",
				fieldProps: { options: STATUS_OPTIONS },
				rules: [{ required: true, message: "请选择状态" }],
			});
		}

		return fields;
	}, [formMode.type, menuTreeData, deptTreeData]);

	// 打开新增
	const openAdd = useCallback(() => {
		setFormMode({ type: "add" });
		formRef.current?.open({
			title: "新增角色",
			initialValues: { roleSort: 0, dataScope: 1 },
		});
	}, []);

	// 打开编辑
	const openEdit = useCallback((record: IApi.RoleItem) => {
		setFormMode({ type: "edit", record });
		formRef.current?.open({
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

	// 统一提交
	const handleFormSubmit = useCallback(
		async (
			_title: string,
			values: Record<string, any>,
			extraParams?: Record<string, any>,
		) => {
			const payload = {
				roleName: values.roleName,
				roleSort: values.roleSort ?? 0,
				dataScope: values.dataScope,
				remark: values.remark ?? "",
				menuIds: values.menuIds ?? [],
				deptIds: values.deptIds ?? [],
			};
			if (formMode.type === "add") {
				await Api.addRole({
					...payload,
					roleKey: values.roleKey,
				});
				message.success("新增角色成功");
			} else {
				const roleKey = extraParams?.roleKey as string;
				await Api.editRole(roleKey, {
					...payload,
					status: values.status,
				});
				message.success("修改角色成功");
			}
			tableRef.current?.refresh();
		},
		[formMode.type, message],
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
			width: 180,
			fixed: "right",
			render: (_, record) => (
				<div className="flex items-center gap-2">
					{hasPerm("system:role:edit") && (
						<Button
							size="small"
							color="primary"
							variant="filled"
							icon={<EditOutlined />}
							onClick={() => openEdit(record)}
						>
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
							<Button
								size="small"
								color="danger"
								variant="filled"
								icon={<DeleteOutlined />}
							>
								删除
							</Button>
						</Popconfirm>
					)}
				</div>
			),
		},
	];

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
						<Button type="primary" icon={<PlusOutlined />} onClick={openAdd}>
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

			{/* 角色表单：新增/编辑共用，menuIds/deptIds 用 TreeSelector 渲染 */}
			<ProForm
				ref={formRef}
				fields={getFormFields()}
				onConfirm={handleFormSubmit}
				className="grid grid-cols-3 gap-4"
				width={840}
			/>
		</>
	);
};

export default RolePage;
