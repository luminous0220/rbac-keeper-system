import { useCallback, useMemo, useRef, useState } from "react";
import { App, Button, Popconfirm, Tag, Typography } from "antd";
import {
	PlusOutlined,
	EditOutlined,
	DeleteOutlined,
	EyeOutlined,
} from "@ant-design/icons";
import { ProTable } from "@/components/ProTable";
import type { ProTableRef, ProTableColumnType } from "@/components/ProTable";
import { ProForm } from "@/components/ProForm";
import type { ProFormFields, ProFormInstance } from "@/components/ProForm";
import { TableModalView } from "@/components/TableModalView";
import type { TableModalViewRef } from "@/components/TableModalView";
import { Api, IApi } from "@/apis";
import { usePerms } from "@/hooks";
import dayjs from "dayjs";

// 字典类型/数据状态：0停用 1正常（model.DictType/DictData，与 constants.StatusDisabled/Enabled 一致）
const STATUS_TEXT: Record<number, string> = { 0: "停用", 1: "正常" };
const STATUS_OPTIONS = [
	{ label: "正常", value: 1 },
	{ label: "停用", value: 0 },
];
// 是否默认：0否 1是
const DEFAULT_TEXT: Record<number, string> = { 0: "否", 1: "是" };
const DEFAULT_OPTIONS = [
	{ label: "否", value: 0 },
	{ label: "是", value: 1 },
];

/**
 * 字典管理：主从结构（无 Tab）。
 * - 主表：字典类型 CRUD（编码变更后端联动迁移数据与缓存）
 * - 详情：点击「查看数据」用 TableModalView 弹出该类型下的字典数据表（含增删改查）
 * - 本页不调用 /api/dict/data/type（按计划，仅管理端列表 CRUD）
 */
const DictPage: React.FC = () => {
	const { message, modal } = App.useApp();
	const { hasPerm } = usePerms();

	// ==================== 字典类型（主表） ====================
	const typeTableRef = useRef<ProTableRef<IApi.DictTypeItem>>(null);
	const typeFormRef = useRef<ProFormInstance>(null);
	const [typeFormMode, setTypeFormMode] = useState<{
		type: "add" | "edit";
		record?: IApi.DictTypeItem;
	}>({ type: "add" });

	// ==================== 字典数据（弹窗） ====================
	const dataModalRef = useRef<TableModalViewRef<IApi.DictDataItem>>(null);
	const dataFormRef = useRef<ProFormInstance>(null);
	const [dataFormMode, setDataFormMode] = useState<{
		type: "add" | "edit";
		record?: IApi.DictDataItem;
	}>({ type: "add" });

	// ---------- 字典类型：表单 ----------
	const getTypeFormFields = useCallback((): ProFormFields => {
		return [
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
	}, []);

	const openTypeAdd = useCallback(() => {
		setTypeFormMode({ type: "add" });
		typeFormRef.current?.open({
			title: "新增字典类型",
			initialValues: { status: 1 },
		});
	}, []);

	const openTypeEdit = useCallback((record: IApi.DictTypeItem) => {
		setTypeFormMode({ type: "edit", record });
		typeFormRef.current?.open({
			title: `编辑字典类型：${record.dict_name}`,
			initialValues: {
				dictName: record.dict_name,
				dictType: record.dict_type,
				status: record.status,
			},
			extraParams: { id: record.id },
		});
	}, []);

	const handleTypeSubmit = useCallback(
		async (
			_title: string,
			values: Record<string, any>,
			extraParams?: Record<string, any>,
		) => {
			const payload = {
				dictName: values.dictName,
				dictType: values.dictType,
				status: values.status,
			};
			if (typeFormMode.type === "add") {
				await Api.addDictType(payload);
				message.success("新增字典类型成功");
			} else {
				await Api.editDictType(extraParams?.id as number, payload);
				message.success("修改字典类型成功");
			}
			typeTableRef.current?.refresh();
		},
		[typeFormMode.type, message],
	);

	const handleTypeDelete = useCallback(
		async (record: IApi.DictTypeItem) => {
			await Api.deleteDictType(record.id);
			message.success("删除字典类型成功");
			typeTableRef.current?.refresh();
		},
		[message],
	);

	// ---------- 字典数据：表单 ----------
	const getDataFormFields = useCallback((): ProFormFields => {
		return [
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
				formItemProps: {
					className: "col-span-2",
				},
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
	}, []);

	// 新增：dict_type 通过 extraParams 带入，提交时取用（无需依赖外部状态）
	const openDataAdd = useCallback((type: IApi.DictTypeItem) => {
		setDataFormMode({ type: "add" });
		dataFormRef.current?.open({
			title: `新增字典数据：${type.dict_name}`,
			initialValues: { dictSort: 0, isDefault: 0, status: 1 },
			extraParams: { dictType: type.dict_type },
		});
	}, []);

	const openDataEdit = useCallback((record: IApi.DictDataItem) => {
		setDataFormMode({ type: "edit", record });
		dataFormRef.current?.open({
			title: `编辑字典数据：${record.dict_label}`,
			initialValues: {
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

	const handleDataSubmit = useCallback(
		async (
			_title: string,
			values: Record<string, any>,
			extraParams?: Record<string, any>,
		) => {
			const payload = {
				dictSort: values.dictSort ?? 0,
				dictLabel: values.dictLabel,
				dictValue: values.dictValue,
				cssClass: values.cssClass ?? "",
				listClass: values.listClass ?? "",
				isDefault: values.isDefault,
				status: values.status,
			};
			if (dataFormMode.type === "add") {
				await Api.addDictData({
					...payload,
					dictType: extraParams?.dictType as string,
				});
				message.success("新增字典数据成功");
			} else {
				await Api.editDictData(extraParams?.id as number, payload);
				message.success("修改字典数据成功");
			}
			// 直接刷新弹窗内表格（TableModalView 暴露了内部 ProTable 的 refresh）
			dataModalRef.current?.refresh();
		},
		[dataFormMode.type, message],
	);

	const handleDataDelete = useCallback(
		async (record: IApi.DictDataItem) => {
			await Api.deleteDictData(record.id);
			message.success("删除字典数据成功");
			dataModalRef.current?.refresh();
		},
		[message],
	);

	// 批量删除字典数据：rows 由 selectedActions 实时传入（无需父组件登记选中）
	const handleBatchDeleteData = useCallback(
		(rows: IApi.DictDataItem[]) => {
			if (!rows.length) {
				message.warning("请先选择要删除的字典数据");
				return;
			}
			modal.confirm({
				title: "批量删除字典数据",
				content: `确定删除选中的 ${rows.length} 个字典数据吗？`,
				okText: "删除",
				cancelText: "取消",
				okButtonProps: { danger: true },
				onOk: async () => {
					await Api.deleteDictDatas(rows.map((r) => r.id));
					message.success("批量删除字典数据成功");
					dataModalRef.current?.clearSelected();
					dataModalRef.current?.refresh();
				},
			});
		},
		[modal, message],
	);

	// 批量删除字典类型：rows 由 selectedActions 实时传入（无需父组件登记选中）
	const handleBatchDeleteTypes = useCallback(
		(rows: IApi.DictTypeItem[]) => {
			if (!rows.length) {
				message.warning("请先选择要删除的字典类型");
				return;
			}
			modal.confirm({
				title: "批量删除字典类型",
				content: `确定删除选中的 ${rows.length} 个字典类型吗？将级联删除这些类型下的全部字典数据。`,
				okText: "删除",
				cancelText: "取消",
				okButtonProps: { danger: true },
				onOk: async () => {
					await Api.deleteDictTypes(rows.map((t) => t.id));
					message.success("批量删除字典类型成功");
					typeTableRef.current?.clearSelected();
					typeTableRef.current?.refresh();
				},
			});
		},
		[modal, message],
	);

	// ---------- 字典数据：弹窗表格列 ----------
	const dataColumns = useMemo<ProTableColumnType<IApi.DictDataItem>[]>(
		() => [
			{ title: "数据标签", dataIndex: "dict_label", width: 140 },
			{
				title: "数据键值",
				dataIndex: "dict_value",
				width: 140,
				render: (value) => (
					<Typography.Text code>{value as string}</Typography.Text>
				),
			},
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
				width: 180,
				fixed: "right",
				render: (_, record) => (
					<div className="flex items-center gap-2">
						{hasPerm("system:dict:data:edit") && (
							<Button
								size="small"
								color="primary"
								variant="filled"
								icon={<EditOutlined />}
								onClick={() => openDataEdit(record)}
							>
								编辑
							</Button>
						)}
						{hasPerm("system:dict:data:delete") && (
							<Popconfirm
								title="确定删除该字典数据吗？"
								okText="删除"
								cancelText="取消"
								okButtonProps={{ danger: true }}
								onConfirm={() => handleDataDelete(record)}
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
		],
		[hasPerm, openDataEdit, handleDataDelete],
	);

	// 打开某类型的数据弹窗：api/toolbarExtra/selectedActions 按当前记录闭包生成，数据变更后走 ref.refresh
	const openData = useCallback(
		(record: IApi.DictTypeItem) => {
			dataModalRef.current?.open({
				title: `字典数据：${record.dict_name}`,
				columns: dataColumns,
				api: (params: Record<string, unknown>) =>
					Api.getDictDataList({
						pageNumber: params.pageNumber as number,
						pageSize: params.pageSize as number,
						dictType: record.dict_type,
						dictLabel: params.dictLabel
							? String(params.dictLabel)
							: undefined,
						status: params.status as number | undefined,
					}),
				toolbarExtra: hasPerm("system:dict:data:add") ? (
					<Button
						type="primary"
						icon={<PlusOutlined />}
						onClick={() => openDataAdd(record)}
					>
						新增数据
					</Button>
				) : null,
				// 批量删除放入选中信息栏：渲染函数实时拿到选中行/个数，无需父组件登记选中
				selectedActions: ({ selectedRows, selectedCount }) =>
					hasPerm("system:dict:data:delete") && (
						<Button
							danger
							icon={<DeleteOutlined />}
							onClick={() => handleBatchDeleteData(selectedRows)}
						>
							批量删除({selectedCount})
						</Button>
					),
				search: {
					fields: [
						{ type: "input", label: "数据标签", name: "dictLabel" },
						{
							type: "select",
							label: "状态",
							name: "status",
							fieldProps: { options: STATUS_OPTIONS, allowClear: true },
						},
					],
				},
			});
		},
		[dataColumns, hasPerm, openDataAdd, handleBatchDeleteData],
	);

	// ---------- 字典类型：主表列 ----------
	const typeColumns: ProTableColumnType<IApi.DictTypeItem>[] = [
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
			width: 260,
			fixed: "right",
			render: (_, record) => (
				<div className="flex items-center gap-2">
					{hasPerm("system:dict:data:list") && (
						<Button
							size="small"
							color="primary"
							variant="filled"
							icon={<EyeOutlined />}
							onClick={() => openData(record)}
						>
							查看数据
						</Button>
					)}
					{hasPerm("system:dict:type:edit") && (
						<Button
							size="small"
							color="primary"
							variant="filled"
							icon={<EditOutlined />}
							onClick={() => openTypeEdit(record)}
						>
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
							onConfirm={() => handleTypeDelete(record)}
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
			<ProTable<IApi.DictTypeItem>
				ref={typeTableRef}
				rowKey="id"
				api={Api.getDictTypeList}
				columns={typeColumns}
				title="字典类型"
				toolbarExtra={
					hasPerm("system:dict:type:add") ? (
						<Button
							type="primary"
							icon={<PlusOutlined />}
							onClick={openTypeAdd}
						>
							新增类型
						</Button>
					) : null
				}
				// 批量删除放入选中信息栏：未选中行时不展示，选中后按钮随实时个数联动
				selectedActions={({ selectedRows, selectedCount }) =>
					hasPerm("system:dict:type:delete") && (
						<Button
							danger
							icon={<DeleteOutlined />}
							onClick={() => handleBatchDeleteTypes(selectedRows)}
						>
							批量删除({selectedCount})
						</Button>
					)
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
			{/* 数据类型 */}
			<ProForm
				ref={typeFormRef}
				fields={getTypeFormFields()}
				onConfirm={handleTypeSubmit}
			/>
			<TableModalView ref={dataModalRef} title="字典数据" width={1280} />
			{/* 数据字典（zIndex 高于 TableModalView 弹窗，保证编辑/新增表单悬浮在其上） */}
			<ProForm
				ref={dataFormRef}
				className="grid grid-cols-2 gap-4"
				zIndex={1100}
				fields={getDataFormFields()}
				onConfirm={handleDataSubmit}
			/>
		</>
	);
};

export default DictPage;
