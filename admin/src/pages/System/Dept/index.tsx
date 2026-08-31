import { useCallback, useMemo, useRef, useState } from "react";
import { App, Button, Popconfirm, Tag } from "antd";
import { PlusOutlined, EditOutlined, DeleteOutlined } from "@ant-design/icons";
import { ProTable } from "@/components/ProTable";
import type { ProTableRef, ProTableColumnType } from "@/components/ProTable";
import { ProForm } from "@/components/ProForm";
import type { ProFormFields, ProFormInstance } from "@/components/ProForm";
import { Api, IApi } from "@/apis";
import { usePerms } from "@/hooks";
import dayjs from "dayjs";

// 部门状态与用户/角色相反：0=正常 1=停用（model.Dept）
const STATUS_TEXT: Record<number, string> = { 1: "正常", 0: "停用" };
const STATUS_OPTIONS = [
	{ label: "正常", value: 1 },
	{ label: "停用", value: 0 },
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
	const formRef = useRef<ProFormInstance>(null);
	// 区分当前弹窗是新增还是编辑，并保存编辑时的记录
	const [formMode, setFormMode] = useState<{
		type: "add" | "edit";
		record?: IApi.DeptRaw;
	}>({ type: "add" });
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

	// 动态生成表单字段：新增模式无 status 字段；编辑模式需 status
	const getFormFields = useCallback((): ProFormFields => {
		const isEdit = formMode.type === "edit";
		const fields: ProFormFields = [
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
		// 编辑模式追加 status
		if (isEdit) {
			fields.push({
				type: "radio",
				label: "状态",
				name: "status",
				fieldProps: { options: STATUS_OPTIONS },
				rules: [{ required: true, message: "请选择状态" }],
			});
		}
		return fields;
	}, [formMode.type, deptTreeData]);

	// 打开新增：可在 action 列传入父部门以预选 parentId
	const openAdd = useCallback((parent?: IApi.DeptRaw) => {
		setFormMode({ type: "add" });
		formRef.current?.open({
			title: parent ? `新增子部门：${parent.dept_name}` : "新增部门",
			initialValues: {
				parentId: parent?.id,
				orderNum: 0,
			},
		});
	}, []);

	// 打开编辑
	const openEdit = useCallback((record: IApi.DeptRaw) => {
		setFormMode({ type: "edit", record });
		formRef.current?.open({
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

	// 统一提交：根据 formMode 走新增或编辑
	const handleFormSubmit = useCallback(
		async (
			_title: string,
			values: Record<string, any>,
			extraParams?: Record<string, any>,
		) => {
			if (formMode.type === "add") {
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
			} else {
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
			}
			tableRef.current?.refresh();
		},
		[formMode.type, message],
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
				<Tag color={(value as number) === 1 ? "green" : "red"}>
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
			width: 240,
			fixed: "right",
			render: (_, record) => (
				<div className="flex items-center gap-2">
					{hasPerm("system:dept:add") && (
						<Button
							size="small"
							color="primary"
							variant="filled"
							icon={<PlusOutlined />}
							onClick={() => openAdd(record)}
						>
							新增子部门
						</Button>
					)}
					{hasPerm("system:dept:edit") && (
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
					{hasPerm("system:dept:remove") && (
						<Popconfirm
							title="确定删除该部门吗？"
							description="存在子部门或部门下存在用户时将被拦截"
							okText="删除"
							cancelText="取消"
							okButtonProps={{ danger: true }}
							onConfirm={() => handleDelete(record)}
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
							icon={<PlusOutlined />}
							onClick={() => openAdd()}
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
			<ProForm
				ref={formRef}
				type="drawer"
				fields={getFormFields()}
				onConfirm={handleFormSubmit}
				width={560}
			/>
		</>
	);
};

export default DeptPage;
