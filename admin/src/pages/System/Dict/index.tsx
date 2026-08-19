import { useCallback, useEffect, useRef, useState } from "react";
import { App, Button, Popconfirm, Space, Tabs, Tag, Typography } from "antd";
import { ProTable } from "@/components/ProTable";
import type { ProTableRef, ProTableColumnType } from "@/components/ProTable";
import { ProForm } from "@/components/ProForm";
import type { ProFormFields, ProFormInstance } from "@/components/ProForm";
import { Api, IApi } from "@/apis";
import { usePerms } from "@/hooks";
import dayjs from "dayjs";

// 字典类型/数据状态：0正常 1停用（model.DictType/DictData）
const STATUS_TEXT: Record<number, string> = { 0: "正常", 1: "停用" };
const STATUS_OPTIONS = [
	{ label: "正常", value: 0 },
	{ label: "停用", value: 1 },
];
// 是否默认：0否 1是
const DEFAULT_TEXT: Record<number, string> = { 0: "否", 1: "是" };
const DEFAULT_OPTIONS = [
	{ label: "否", value: 0 },
	{ label: "是", value: 1 },
];

// ==================== 字典类型 Tab ====================

const DictTypeTab: React.FC<{ onViewData: () => void }> = ({ onViewData }) => {
	const { message } = App.useApp();
	const { hasPerm } = usePerms();
	const tableRef = useRef<ProTableRef<IApi.DictTypeItem>>(null);
	const addFormRef = useRef<ProFormInstance>(null);
	const editFormRef = useRef<ProFormInstance>(null);

	const addFields: ProFormFields = [
		{
			type: "input",
			label: "字典名称",
			name: "dictName",
			fieldProps: { maxLength: 100, placeholder: "2-100 个字符" },
			rules: [
				{ required: true, message: "请输入字典名称" },
				{ min: 2, max: 100, message: "字典名称长度需为 2-100 位" },
			],
		},
		{
			type: "input",
			label: "字典类型",
			name: "dictType",
			fieldProps: {
				maxLength: 100,
				placeholder: "唯一编码，如 sys_normal_disable",
			},
			rules: [
				{ required: true, message: "请输入字典类型编码" },
				{ min: 2, max: 100, message: "字典类型长度需为 2-100 位" },
			],
		},
		{
			type: "radio",
			label: "状态",
			name: "status",
			fieldProps: { options: STATUS_OPTIONS },
			rules: [{ required: true, message: "请选择状态" }],
		},
	];

	const handleAdd = useCallback(
		async (_title: string, values: Record<string, any>) => {
			await Api.addDictType({
				dictName: values.dictName,
				dictType: values.dictType,
				status: values.status,
			});
			message.success("新增字典类型成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	const editFields: ProFormFields = [
		{
			type: "input",
			label: "字典名称",
			name: "dictName",
			fieldProps: { maxLength: 100 },
			rules: [
				{ required: true, message: "请输入字典名称" },
				{ min: 2, max: 100, message: "字典名称长度需为 2-100 位" },
			],
		},
		{
			type: "input",
			label: "字典类型",
			name: "dictType",
			fieldProps: {
				maxLength: 100,
				placeholder: "修改将迁移该类型下的字典数据",
			},
			rules: [
				{ required: true, message: "请输入字典类型编码" },
				{ min: 2, max: 100, message: "字典类型长度需为 2-100 位" },
			],
		},
		{
			type: "radio",
			label: "状态",
			name: "status",
			fieldProps: { options: STATUS_OPTIONS },
			rules: [{ required: true, message: "请选择状态" }],
		},
	];

	const openEdit = useCallback((record: IApi.DictTypeItem) => {
		editFormRef.current?.open({
			title: `编辑字典类型：${record.dict_name}`,
			initialValues: {
				dictName: record.dict_name,
				dictType: record.dict_type,
				status: record.status,
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
			await Api.editDictType(id, {
				dictName: values.dictName,
				dictType: values.dictType,
				status: values.status,
			});
			message.success("修改字典类型成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	const handleDelete = useCallback(
		async (record: IApi.DictTypeItem) => {
			await Api.deleteDictType(record.id);
			message.success("删除字典类型成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	const columns: ProTableColumnType<IApi.DictTypeItem>[] = [
		{ title: "字典名称", dataIndex: "dict_name", width: 180 },
		{
			title: "字典类型",
			dataIndex: "dict_type",
			width: 200,
			render: (value) => (
				<Typography.Text code>{value as string}</Typography.Text>
			),
		},
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
			width: 190,
			fixed: "right",
			render: (_, record) => (
				<Space size={0}>
					<Button type="link" size="small" onClick={onViewData}>
						查看数据
					</Button>
					{hasPerm("system:dict:type:edit") && (
						<Button type="link" size="small" onClick={() => openEdit(record)}>
							编辑
						</Button>
					)}
					{hasPerm("system:dict:type:delete") && (
						<Popconfirm
							title="确定删除该字典类型吗？"
							description="将级联删除该类型下的全部字典数据"
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
			<ProTable<IApi.DictTypeItem>
				ref={tableRef}
				rowKey="id"
				api={Api.getDictTypeList}
				columns={columns}
				title="字典类型"
				toolbarExtra={
					hasPerm("system:dict:type:add") ? (
						<Button
							type="primary"
							onClick={() =>
								addFormRef.current?.open({
									title: "新增字典类型",
									initialValues: { status: 0 },
								})
							}
						>
							新增类型
						</Button>
					) : undefined
				}
				search={{
					fields: [
						{ type: "input", label: "字典名称", name: "dictName" },
						{ type: "input", label: "字典类型", name: "dictType" },
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
				ref={addFormRef}
				fields={addFields}
				onConfirm={handleAdd}
				width={520}
			/>
			<ProForm
				ref={editFormRef}
				fields={editFields}
				onConfirm={handleEdit}
				width={520}
			/>
		</>
	);
};

// ==================== 字典数据 Tab ====================

const DictDataTab: React.FC = () => {
	const { message } = App.useApp();
	const { hasPerm } = usePerms();
	const tableRef = useRef<ProTableRef<IApi.DictDataItem>>(null);
	const addFormRef = useRef<ProFormInstance>(null);
	const editFormRef = useRef<ProFormInstance>(null);
	const [typeOptions, setTypeOptions] = useState<
		{ label: string; value: string }[]
	>([]);
	// 最近一次查询使用的字典类型：新增数据时作为默认归属
	const [lastType, setLastType] = useState<string>();

	useEffect(() => {
		Api.getDictTypeList({ pageNumber: 1, pageSize: 100 })
			.then((res) =>
				setTypeOptions(
					res.data.map((t) => ({ label: t.dict_name, value: t.dict_type })),
				),
			)
			.catch((err) => console.error("加载字典类型失败", err));
	}, []);

	// 列表：dictType/dictLabel/status 透传（dictType 为空 = 不筛选，展示全部）
	const dictDataApi = useCallback(async (params: Record<string, unknown>) => {
		const dictType = params.dictType ? String(params.dictType) : undefined;
		setLastType(dictType);
		return Api.getDictDataList({
			pageNumber: params.pageNumber as number,
			pageSize: params.pageSize as number,
			dictType,
			dictLabel: params.dictLabel ? String(params.dictLabel) : undefined,
			status: params.status as number | undefined,
		});
	}, []);

	const addFields: ProFormFields = [
		{
			type: "select",
			label: "字典类型",
			name: "dictType",
			fieldProps: {
				options: typeOptions,
				placeholder: "请选择所属字典类型",
				showSearch: true,
			},
			rules: [{ required: true, message: "请选择字典类型" }],
		},
		{
			type: "input",
			label: "数据标签",
			name: "dictLabel",
			fieldProps: { maxLength: 100, placeholder: "展示名，如 男" },
			rules: [
				{ required: true, message: "请输入数据标签" },
				{ min: 1, max: 100, message: "数据标签长度需为 1-100 位" },
			],
		},
		{
			type: "input",
			label: "数据键值",
			name: "dictValue",
			fieldProps: { maxLength: 100, placeholder: "实际值，如 1" },
			rules: [
				{ required: true, message: "请输入数据键值" },
				{ min: 1, max: 100, message: "数据键值长度需为 1-100 位" },
			],
		},
		{
			type: "input-number",
			label: "显示顺序",
			name: "dictSort",
			fieldProps: { min: 0, style: { width: "100%" } },
		},
		{
			type: "input",
			label: "样式属性",
			name: "cssClass",
			fieldProps: { maxLength: 100, placeholder: "选填" },
		},
		{
			type: "input",
			label: "回显样式",
			name: "listClass",
			fieldProps: {
				maxLength: 100,
				placeholder: "如 success / danger / warning（选填）",
			},
		},
		{
			type: "radio",
			label: "是否默认",
			name: "isDefault",
			fieldProps: { options: DEFAULT_OPTIONS },
			rules: [{ required: true, message: "请选择是否默认" }],
		},
		{
			type: "radio",
			label: "状态",
			name: "status",
			fieldProps: { options: STATUS_OPTIONS },
			rules: [{ required: true, message: "请选择状态" }],
		},
	];

	const handleAdd = useCallback(
		async (_title: string, values: Record<string, any>) => {
			await Api.addDictData({
				dictSort: values.dictSort ?? 0,
				dictLabel: values.dictLabel,
				dictValue: values.dictValue,
				dictType: values.dictType,
				cssClass: values.cssClass ?? "",
				listClass: values.listClass ?? "",
				isDefault: values.isDefault,
				status: values.status,
			});
			message.success("新增字典数据成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	const editFields: ProFormFields = [
		{
			type: "select",
			label: "字典类型",
			name: "dictType",
			disabled: true,
			fieldProps: {
				options: typeOptions,
				placeholder: "所属类型不可修改",
				disabled: true,
				showSearch: true,
			},
			rules: [{ required: true, message: "请选择字典类型" }],
		},
		{
			type: "input",
			label: "数据标签",
			name: "dictLabel",
			fieldProps: { maxLength: 100 },
			rules: [
				{ required: true, message: "请输入数据标签" },
				{ min: 1, max: 100, message: "数据标签长度需为 1-100 位" },
			],
		},
		{
			type: "input",
			label: "数据键值",
			name: "dictValue",
			fieldProps: { maxLength: 100 },
			rules: [
				{ required: true, message: "请输入数据键值" },
				{ min: 1, max: 100, message: "数据键值长度需为 1-100 位" },
			],
		},
		{
			type: "input-number",
			label: "显示顺序",
			name: "dictSort",
			fieldProps: { min: 0, style: { width: "100%" } },
		},
		{
			type: "input",
			label: "样式属性",
			name: "cssClass",
			fieldProps: { maxLength: 100 },
		},
		{
			type: "input",
			label: "回显样式",
			name: "listClass",
			fieldProps: { maxLength: 100 },
		},
		{
			type: "radio",
			label: "是否默认",
			name: "isDefault",
			fieldProps: { options: DEFAULT_OPTIONS },
			rules: [{ required: true, message: "请选择是否默认" }],
		},
		{
			type: "radio",
			label: "状态",
			name: "status",
			fieldProps: { options: STATUS_OPTIONS },
			rules: [{ required: true, message: "请选择状态" }],
		},
	];

	const openEdit = useCallback((record: IApi.DictDataItem) => {
		editFormRef.current?.open({
			title: `编辑字典数据：${record.dict_label}`,
			initialValues: {
				dictType: record.dict_type,
				dictLabel: record.dict_label,
				dictValue: record.dict_value,
				dictSort: record.dict_sort,
				cssClass: record.css_class,
				listClass: record.list_class,
				isDefault: record.is_default,
				status: record.status,
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
			await Api.editDictData(id, {
				dictSort: values.dictSort ?? 0,
				dictLabel: values.dictLabel,
				dictValue: values.dictValue,
				cssClass: values.cssClass ?? "",
				listClass: values.listClass ?? "",
				isDefault: values.isDefault,
				status: values.status,
			});
			message.success("修改字典数据成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	const handleDelete = useCallback(
		async (record: IApi.DictDataItem) => {
			await Api.deleteDictData(record.id);
			message.success("删除字典数据成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	const columns: ProTableColumnType<IApi.DictDataItem>[] = [
		{ title: "数据标签", dataIndex: "dict_label", width: 140 },
		{
			title: "数据键值",
			dataIndex: "dict_value",
			width: 140,
			render: (value) => (
				<Typography.Text code>{value as string}</Typography.Text>
			),
		},
		{ title: "字典类型", dataIndex: "dict_type", width: 170 },
		{ title: "显示顺序", dataIndex: "dict_sort", width: 90 },
		{
			title: "默认",
			dataIndex: "is_default",
			width: 70,
			render: (value) => (
				<Tag color={(value as number) === 1 ? "blue" : "default"}>
					{DEFAULT_TEXT[(value as number) ?? 0] ?? "-"}
				</Tag>
			),
		},
		{
			title: "回显样式",
			dataIndex: "list_class",
			width: 100,
			render: (value) => (value ? <Tag>{value as string}</Tag> : "-"),
		},
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
			width: 130,
			fixed: "right",
			render: (_, record) => (
				<Space size={0}>
					{hasPerm("system:dict:data:edit") && (
						<Button type="link" size="small" onClick={() => openEdit(record)}>
							编辑
						</Button>
					)}
					{hasPerm("system:dict:data:delete") && (
						<Popconfirm
							title="确定删除该字典数据吗？"
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
			<ProTable<IApi.DictDataItem>
				ref={tableRef}
				rowKey="id"
				api={dictDataApi}
				columns={columns}
				title="字典数据"
				toolbarExtra={
					hasPerm("system:dict:data:add") ? (
						<Button
							type="primary"
							onClick={() =>
								addFormRef.current?.open({
									title: "新增字典数据",
									// 默认归属最近筛选的字典类型（未筛选则手动选择）
									initialValues: {
										dictType: lastType,
										dictSort: 0,
										isDefault: 0,
										status: 0,
									},
								})
							}
						>
							新增数据
						</Button>
					) : undefined
				}
				search={{
					fields: [
						{
							type: "select",
							label: "字典类型",
							name: "dictType",
							fieldProps: {
								options: typeOptions,
								allowClear: true,
								showSearch: true,
								placeholder: "全部类型",
							},
						},
						{ type: "input", label: "数据标签", name: "dictLabel" },
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
				ref={addFormRef}
				fields={addFields}
				onConfirm={handleAdd}
				width={560}
			/>
			<ProForm
				ref={editFormRef}
				fields={editFields}
				onConfirm={handleEdit}
				width={560}
			/>
		</>
	);
};

// ==================== 页面 ====================

/**
 * 字典管理：类型 + 数据两个 Tab。
 * - 类型 Tab：字典类型 CRUD（编码变更后端联动迁移数据与缓存）
 * - 数据 Tab：字典数据 CRUD，按「字典类型/标签/状态」搜索过滤
 * - 本页不调用 /api/dict/data/type（按计划，仅管理端列表 CRUD）
 */
const DictPage: React.FC = () => {
	const [activeTab, setActiveTab] = useState("type");
	return (
		<Tabs
			activeKey={activeTab}
			onChange={setActiveTab}
			items={[
				{
					key: "type",
					label: "字典类型",
					children: <DictTypeTab onViewData={() => setActiveTab("data")} />,
				},
				{ key: "data", label: "字典数据", children: <DictDataTab /> },
			]}
		/>
	);
};

export default DictPage;
