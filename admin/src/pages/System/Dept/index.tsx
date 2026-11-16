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
import { useTranslation } from "react-i18next";

// 部门状态与用户/角色相反：0=正常 1=停用（model.Dept）
const STATUS_TEXT: Record<number, string> = {
	1: "common.enabled",
	0: "common.disabled",
};
const STATUS_OPTIONS = [
	{ label: "common.enabled", value: 1 },
	{ label: "common.disabled", value: 0 },
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
	const { t } = useTranslation();
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
				label: t("department.parent"),
				name: "parentId",
				fieldProps: {
					treeData: deptTreeData,
					allowClear: true,
					showSearch: true,
					treeDefaultExpandAll: true,
					placeholder: t("department.rootHint"),
				},
			},
			{
				type: "input",
				label: t("department.name"),
				name: "deptName",
				fieldProps: { maxLength: 30, placeholder: t("department.namePlaceholder") },
				rules: [
					{ required: true, message: t("department.nameRequired") },
					{ min: 2, max: 30, message: t("department.nameLength") },
				],
			},
			{
				type: "input-number",
				label: t("department.order"),
				name: "orderNum",
				fieldProps: { min: 0, style: { width: "100%" } },
			},
			{
				type: "input",
				label: t("department.leader"),
				name: "leader",
				fieldProps: { maxLength: 20 },
				rules: [{ max: 20, message: t("department.leaderLength") }],
			},
			{
				type: "input",
				label: t("system.phone"),
				name: "phone",
				fieldProps: { maxLength: 11, placeholder: t("department.phonePlaceholder") },
				rules: [{ max: 11, message: t("department.phoneLength") }],
			},
			{
				type: "input",
				label: t("system.email"),
				name: "email",
				fieldProps: { maxLength: 50 },
				rules: [{ type: "email", message: t("validation.emailInvalid") }],
			},
			{
				type: "textarea",
				label: t("system.note"),
				name: "remark",
				fieldProps: {
					maxLength: 500,
					rows: 3,
					placeholder: t("department.remarkPlaceholder"),
				},
			},
		];
		// 编辑模式追加 status
		if (isEdit) {
			fields.push({
				type: "radio",
				label: t("system.status"),
				name: "status",
				fieldProps: {
					options: STATUS_OPTIONS.map((item) => ({
						...item,
						label: t(item.label),
					})),
				},
				rules: [{ required: true, message: t("department.statusRequired") }],
			});
		}
		return fields;
	}, [formMode.type, deptTreeData, t]);

	// 打开新增：可在 action 列传入父部门以预选 parentId
	const openAdd = useCallback((parent?: IApi.DeptRaw) => {
		setFormMode({ type: "add" });
		formRef.current?.open({
			title: parent
				? `${t("department.addChild")}: ${parent.dept_name}`
				: t("department.add"),
			initialValues: {
				parentId: parent?.id,
				orderNum: 0,
			},
		});
	}, [t]);

	// 打开编辑
	const openEdit = useCallback((record: IApi.DeptRaw) => {
		setFormMode({ type: "edit", record });
		formRef.current?.open({
			title: `${t("department.edit")}: ${record.dept_name}`,
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
	}, [t]);

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
				message.success(t("department.addSuccess"));
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
				message.success(t("department.editSuccess"));
			}
			tableRef.current?.refresh();
		},
		[formMode.type, message, t],
	);

	// ---- 删除部门 ----
	const handleDelete = useCallback(
		async (record: IApi.DeptRaw) => {
			await Api.deleteDept(record.id);
			message.success(t("department.deleteSuccess"));
			tableRef.current?.refresh();
		},
		[message, t],
	);

	// ---- 列配置 ----
	const columns: ProTableColumnType<IApi.DeptRaw>[] = [
		{ title: t("department.name"), dataIndex: "dept_name", width: 220 },
		{ title: t("system.displayOrder"), dataIndex: "order_num", width: 90 },
		{ title: t("department.leader"), dataIndex: "leader", width: 100 },
		{ title: t("system.phone"), dataIndex: "phone", width: 130 },
		{ title: t("system.email"), dataIndex: "email", width: 180, ellipsis: true },
		{
			title: t("system.status"),
			dataIndex: "status",
			width: 80,
			render: (value) => (
				<Tag color={(value as number) === 1 ? "green" : "red"}>
					{t(STATUS_TEXT[(value as number) ?? 0] ?? "-")}
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
							{t("department.addChild")}
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
							{t("system.edit")}
						</Button>
					)}
					{hasPerm("system:dept:remove") && (
						<Popconfirm
							title={t("department.deleteConfirm")}
							description={t("department.deleteBlocked")}
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
			<ProTable<IApi.DeptRaw>
				ref={tableRef}
				rowKey="id"
				api={deptTreeApi}
				columns={columns}
				title={t("department.title")}
				index={false}
				toolbarExtra={
					hasPerm("system:dept:add") ? (
						<Button
							type="primary"
							icon={<PlusOutlined />}
							onClick={() => openAdd()}
						>
							{t("department.add")}
						</Button>
					) : undefined
				}
				search={{
					fields: [
						{ type: "input", label: t("department.name"), name: "deptName" },
						{
							type: "select",
							label: t("system.status"),
							name: "status",
							fieldProps: {
								options: STATUS_OPTIONS.map((item) => ({
									...item,
									label: t(item.label),
								})),
								allowClear: true,
							},
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
