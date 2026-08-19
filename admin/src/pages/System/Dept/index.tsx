import { useCallback, useMemo, useRef, useState } from "react";
import { App, Button, Popconfirm, Space, Tag } from "antd";
import { ProTable } from "@/components/ProTable";
import type { ProTableRef, ProTableColumnType } from "@/components/ProTable";
import { ProForm } from "@/components/ProForm";
import type { ProFormFields, ProFormInstance } from "@/components/ProForm";
import { Api, IApi } from "@/apis";
import { usePerms } from "@/hooks";
import dayjs from "dayjs";

// 部门状态与用户/角色相反：0=正常 1=停用（model.Dept）
const STATUS_TEXT: Record<number, string> = { 0: "正常", 1: "停用" };
const STATUS_OPTIONS = [
	{ label: "正常", value: 0 },
	{ label: "停用", value: 1 },
];

/** 部门树 → antd TreeSelect 数据结构（上级部门选择） */
function toTreeSelect(nodes: IApi.DeptRaw[]): any[] {
	return nodes.map((d) => ({
		value: d.id,
		title: d.dept_name,
		children: d.children?.length ? toTreeSelect(d.children) : undefined,
	}));
}

/**
 * 部门管理：树形表格 + 新增 / 编辑 / 删除。
 * - 数据源 /api/dept/tree（原生 model.Dept，snake_case，无分页 → total 置 0 隐藏分页条）
 * - 搜索 deptName/status 透传后端（后端 FilterTreeWithAncestors 返回匹配节点 + 祖先链）
 * - 删除走 /api/dept/delete?id=（后端校验存在子部门或部门下有用户则拦截）
 * - 新增部门默认 status=1（正常）由后端 service 写入，表单不展示
 */
const DeptPage: React.FC = () => {
	const { message } = App.useApp();
	const { hasPerm } = usePerms();
	const tableRef = useRef<ProTableRef<IApi.DeptRaw>>(null);
	const addFormRef = useRef<ProFormInstance>(null);
	const editFormRef = useRef<ProFormInstance>(null);
	// 上级部门下拉选项：随表格数据同步刷新（CRUD 后 refresh 时自动更新）
	const [deptTree, setDeptTree] = useState<IApi.DeptRaw[]>([]);

	const deptTreeData = useMemo(() => toTreeSelect(deptTree), [deptTree]);

	// 树形表格 API：无分页，返回 {data, total:0} 使 ProTable 不渲染分页条
	const deptTreeApi = useCallback(async (params: Record<string, unknown>) => {
		const data = await Api.getDeptTree({
			deptName: params.deptName ? String(params.deptName) : undefined,
			status: params.status as number | undefined,
		});
		setDeptTree(data); // 同步上级部门下拉
		return { data, total: 0 };
	}, []);

	// ---- 新增部门 ----
	const addFields: ProFormFields = [
		{
			type: "tree-select",
			label: "上级部门",
			name: "parentId",
			fieldProps: {
				treeData: deptTreeData,
				allowClear: true,
				showSearch: true,
				treeDefaultExpandAll: true,
				placeholder: "不选择则为顶级部门",
			},
		},
		{
			type: "input",
			label: "部门名称",
			name: "deptName",
			fieldProps: { maxLength: 30, placeholder: "2-30 个字符" },
			rules: [
				{ required: true, message: "请输入部门名称" },
				{ min: 2, max: 30, message: "部门名称长度需为 2-30 位" },
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
			label: "负责人",
			name: "leader",
			fieldProps: { maxLength: 20 },
			rules: [{ max: 20, message: "负责人长度不能超过 20 位" }],
		},
		{
			type: "input",
			label: "联系电话",
			name: "phone",
			fieldProps: { maxLength: 11, placeholder: "手机号或分机号" },
			rules: [{ max: 11, message: "联系电话长度不能超过 11 位" }],
		},
		{
			type: "input",
			label: "邮箱",
			name: "email",
			fieldProps: { maxLength: 50 },
			rules: [{ type: "email", message: "请输入正确的邮箱" }],
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

	const handleAdd = useCallback(
		async (_title: string, values: Record<string, any>) => {
			await Api.addDept({
				parentId: values.parentId ?? 0,
				deptName: values.deptName,
				orderNum: values.orderNum ?? 0,
				leader: values.leader ?? "",
				phone: values.phone ?? "",
				email: values.email ?? "",
				remark: values.remark ?? "",
			});
			message.success("新增部门成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	// ---- 编辑部门 ----
	const editFields: ProFormFields = [
		{
			type: "tree-select",
			label: "上级部门",
			name: "parentId",
			fieldProps: {
				treeData: deptTreeData,
				allowClear: true,
				showSearch: true,
				treeDefaultExpandAll: true,
				placeholder: "不选择则为顶级部门",
			},
		},
		{
			type: "input",
			label: "部门名称",
			name: "deptName",
			fieldProps: { maxLength: 30 },
			rules: [
				{ required: true, message: "请输入部门名称" },
				{ min: 2, max: 30, message: "部门名称长度需为 2-30 位" },
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
			label: "负责人",
			name: "leader",
			fieldProps: { maxLength: 20 },
			rules: [{ max: 20, message: "负责人长度不能超过 20 位" }],
		},
		{
			type: "input",
			label: "联系电话",
			name: "phone",
			fieldProps: { maxLength: 11 },
			rules: [{ max: 11, message: "联系电话长度不能超过 11 位" }],
		},
		{
			type: "input",
			label: "邮箱",
			name: "email",
			fieldProps: { maxLength: 50 },
			rules: [{ type: "email", message: "请输入正确的邮箱" }],
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

	const openEdit = useCallback((record: IApi.DeptRaw) => {
		editFormRef.current?.open({
			title: `编辑部门：${record.dept_name}`,
			// 根部门 parent_id=0 置 undefined，展示占位「不选择则为顶级部门」
			initialValues: {
				parentId: record.parent_id || undefined,
				deptName: record.dept_name,
				orderNum: record.order_num,
				leader: record.leader,
				phone: record.phone,
				email: record.email,
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
			await Api.editDept(id, {
				parentId: values.parentId ?? 0,
				deptName: values.deptName,
				orderNum: values.orderNum ?? 0,
				leader: values.leader ?? "",
				phone: values.phone ?? "",
				email: values.email ?? "",
				status: values.status,
				remark: values.remark ?? "",
			});
			message.success("修改部门成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	// ---- 删除部门 ----
	const handleDelete = useCallback(
		async (record: IApi.DeptRaw) => {
			await Api.deleteDept(record.id);
			message.success("删除部门成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	// ---- 列配置 ----
	const columns: ProTableColumnType<IApi.DeptRaw>[] = [
		{ title: "部门名称", dataIndex: "dept_name", width: 220 },
		{ title: "显示顺序", dataIndex: "order_num", width: 90 },
		{ title: "负责人", dataIndex: "leader", width: 100 },
		{ title: "联系电话", dataIndex: "phone", width: 130 },
		{ title: "邮箱", dataIndex: "email", width: 180, ellipsis: true },
		{
			title: "状态",
			dataIndex: "status",
			width: 80,
			render: (value) => (
				<Tag color={(value as number) === 0 ? "green" : "red"}>
					{STATUS_TEXT[(value as number) ?? 0] ?? "-"}
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
			width: 200,
			fixed: "right",
			render: (_, record) => (
				<Space size={0}>
					{hasPerm("system:dept:add") && (
						<Button
							type="link"
							size="small"
							onClick={() =>
								addFormRef.current?.open({
									title: `新增子部门：${record.dept_name}`,
									initialValues: { parentId: record.id, orderNum: 0 },
								})
							}
						>
							新增子部门
						</Button>
					)}
					{hasPerm("system:dept:edit") && (
						<Button type="link" size="small" onClick={() => openEdit(record)}>
							编辑
						</Button>
					)}
					{hasPerm("system:dept:remove") && (
						<Popconfirm
							title="确定删除该部门吗？"
							description="存在子部门或部门下存在用户时将被拦截"
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
			<ProTable<IApi.DeptRaw>
				ref={tableRef}
				rowKey="id"
				api={deptTreeApi}
				columns={columns}
				title="部门管理"
				index={false}
				toolbarExtra={
					hasPerm("system:dept:add") ? (
						<Button
							type="primary"
							onClick={() =>
								addFormRef.current?.open({
									title: "新增部门",
									initialValues: { orderNum: 0 },
								})
							}
						>
							新增部门
						</Button>
					) : undefined
				}
				search={{
					fields: [
						{ type: "input", label: "部门名称", name: "deptName" },
						{
							type: "select",
							label: "状态",
							name: "status",
							fieldProps: { options: STATUS_OPTIONS, allowClear: true },
						},
					],
				}}
			/>
			{/* 新增部门弹窗 */}
			<ProForm
				ref={addFormRef}
				fields={addFields}
				onConfirm={handleAdd}
				width={560}
			/>
			{/* 编辑部门弹窗 */}
			<ProForm
				ref={editFormRef}
				fields={editFields}
				onConfirm={handleEdit}
				width={560}
			/>
		</>
	);
};

export default DeptPage;
