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
import { useTranslation } from "react-i18next";

// 字典类型/数据状态：0停用 1正常（model.DictType/DictData，与 constants.StatusDisabled/Enabled 一致）
const STATUS_TEXT: Record<number, string> = {
	0: "common.disabled",
	1: "common.enabled",
};
const STATUS_OPTIONS = [
	{ label: "common.enabled", value: 1 },
	{ label: "common.disabled", value: 0 },
];
// 是否默认：0否 1是
const DEFAULT_TEXT: Record<number, string> = { 0: "common.no", 1: "common.yes" };
const DEFAULT_OPTIONS = [
	{ label: "common.no", value: 0 },
	{ label: "common.yes", value: 1 },
];

/**
 * 字典管理：主从结构（无 Tab）。
 * - 主表：字典类型 CRUD（编码变更后端联动迁移数据与缓存）
 * - 详情：点击「查看数据」用 TableModalView 弹出该类型下的字典数据表（含增删改查）
 * - 本页不调用 /api/dict/data/type（按计划，仅管理端列表 CRUD）
 */
const DictPage: React.FC = () => {
	const { message, modal } = App.useApp();
	const { t } = useTranslation();
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
				label: t("dictionary.name"),
				name: "dictName",
				fieldProps: {
					maxLength: 100,
					placeholder: t("dictionary.namePlaceholder"),
				},
				rules: [
					{ required: true, message: t("dictionary.nameRequired") },
					{ min: 2, max: 100, message: t("dictionary.nameLength") },
				],
			},
			{
				type: "input",
				label: t("dictionary.type"),
				name: "dictType",
				fieldProps: {
					maxLength: 100,
					placeholder: t("dictionary.typePlaceholder"),
				},
				rules: [
					{ required: true, message: t("dictionary.typeRequired") },
					{ min: 2, max: 100, message: t("dictionary.typeLength") },
				],
			},
			{
				type: "radio",
				label: t("system.status"),
				name: "status",
				fieldProps: {
					options: STATUS_OPTIONS.map((item) => ({
						...item,
						label: t(item.label),
					})),
				},
				rules: [{ required: true, message: t("dictionary.statusRequired") }],
			},
		];
	}, [t]);

	const openTypeAdd = useCallback(() => {
		setTypeFormMode({ type: "add" });
		typeFormRef.current?.open({
			title: t("dictionary.addType"),
			initialValues: { status: 1 },
		});
	}, [t]);

	const openTypeEdit = useCallback((record: IApi.DictTypeItem) => {
		setTypeFormMode({ type: "edit", record });
		typeFormRef.current?.open({
			title: `${t("system.edit")}${t("dictionary.type")}: ${record.dict_name}`,
			initialValues: {
				dictName: record.dict_name,
				dictType: record.dict_type,
				status: record.status,
			},
			extraParams: { id: record.id },
		});
	}, [t]);

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
				message.success(t("dictionary.addTypeSuccess"));
			} else {
				await Api.editDictType(extraParams?.id as number, payload);
				message.success(t("dictionary.editTypeSuccess"));
			}
			typeTableRef.current?.refresh();
		},
		[typeFormMode.type, message, t],
	);

	const handleTypeDelete = useCallback(
		async (record: IApi.DictTypeItem) => {
			await Api.deleteDictType(record.id);
			message.success(t("dictionary.deleteTypeSuccess"));
			typeTableRef.current?.refresh();
		},
		[message, t],
	);

	// ---------- 字典数据：表单 ----------
	const getDataFormFields = useCallback((): ProFormFields => {
		return [
			{
				type: "input",
				label: t("dictionary.label"),
				name: "dictLabel",
				fieldProps: { maxLength: 100, placeholder: t("dictionary.labelPlaceholder") },
				rules: [
					{ required: true, message: t("dictionary.labelRequired") },
					{ min: 1, max: 100, message: t("dictionary.labelLength") },
				],
			},
			{
				type: "input",
				label: t("dictionary.value"),
				name: "dictValue",
				fieldProps: { maxLength: 100, placeholder: t("dictionary.valuePlaceholder") },
				rules: [
					{ required: true, message: t("dictionary.valueRequired") },
					{ min: 1, max: 100, message: t("dictionary.valueLength") },
				],
			},
			{
				type: "input-number",
				label: t("dictionary.order"),
				name: "dictSort",
				fieldProps: { min: 0, style: { width: "100%" } },
			},
			{
				type: "input",
				label: t("dictionary.style"),
				name: "cssClass",
				fieldProps: { maxLength: 100, placeholder: t("dictionary.optional") },
			},
			{
				type: "input",
				label: t("dictionary.echoStyle"),
				name: "listClass",
				formItemProps: {
					className: "col-span-2",
				},
				fieldProps: {
					maxLength: 100,
					placeholder: t("dictionary.echoStylePlaceholder"),
				},
			},
			{
				type: "radio",
				label: t("dictionary.isDefault"),
				name: "isDefault",
				fieldProps: {
					options: DEFAULT_OPTIONS.map((item) => ({
						...item,
						label: t(item.label),
					})),
				},
				rules: [{ required: true, message: t("dictionary.defaultRequired") }],
			},
			{
				type: "radio",
				label: t("system.status"),
				name: "status",
				fieldProps: {
					options: STATUS_OPTIONS.map((item) => ({
						...item,
						label: t(item.label),
					})),
				},
				rules: [{ required: true, message: t("dictionary.statusRequired") }],
			},
		];
	}, [t]);

	// 新增：dict_type 通过 extraParams 带入，提交时取用（无需依赖外部状态）
	const openDataAdd = useCallback((type: IApi.DictTypeItem) => {
		setDataFormMode({ type: "add" });
		dataFormRef.current?.open({
			title: `${t("dictionary.addData")}: ${type.dict_name}`,
			initialValues: { dictSort: 0, isDefault: 0, status: 1 },
			extraParams: { dictType: type.dict_type },
		});
	}, [t]);

	const openDataEdit = useCallback((record: IApi.DictDataItem) => {
		setDataFormMode({ type: "edit", record });
		dataFormRef.current?.open({
			title: `${t("system.edit")}${t("dictionary.dataTitle")}: ${record.dict_label}`,
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
	}, [t]);

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
				message.success(t("dictionary.addDataSuccess"));
			} else {
				await Api.editDictData(extraParams?.id as number, payload);
				message.success(t("dictionary.editDataSuccess"));
			}
			// 直接刷新弹窗内表格（TableModalView 暴露了内部 ProTable 的 refresh）
			dataModalRef.current?.refresh();
		},
		[dataFormMode.type, message, t],
	);

	const handleDataDelete = useCallback(
		async (record: IApi.DictDataItem) => {
			await Api.deleteDictData(record.id);
			message.success(t("dictionary.deleteDataSuccess"));
			dataModalRef.current?.refresh();
		},
		[message, t],
	);

	// 批量删除字典数据：rows 由 selectedActions 实时传入（无需父组件登记选中）
	const handleBatchDeleteData = useCallback(
		(rows: IApi.DictDataItem[]) => {
			if (!rows.length) {
				message.warning(t("dictionary.selectDataBeforeDelete"));
				return;
			}
			modal.confirm({
				title: t("dictionary.deleteData"),
				content: t("dictionary.deleteSelectedData", { count: rows.length }),
				okText: t("system.delete"),
				cancelText: t("common.cancel"),
				okButtonProps: { danger: true },
				onOk: async () => {
					await Api.deleteDictDatas(rows.map((r) => r.id));
					message.success(t("dictionary.batchDeleteDataSuccess"));
					dataModalRef.current?.clearSelected();
					dataModalRef.current?.refresh();
				},
			});
		},
		[modal, message, t],
	);

	// 批量删除字典类型：rows 由 selectedActions 实时传入（无需父组件登记选中）
	const handleBatchDeleteTypes = useCallback(
		(rows: IApi.DictTypeItem[]) => {
			if (!rows.length) {
				message.warning(t("dictionary.selectTypeBeforeDelete"));
				return;
			}
			modal.confirm({
				title: t("dictionary.deleteType"),
				content: t("dictionary.deleteSelectedTypes", { count: rows.length }),
				okText: t("system.delete"),
				cancelText: t("common.cancel"),
				okButtonProps: { danger: true },
				onOk: async () => {
					await Api.deleteDictTypes(rows.map((t) => t.id));
					message.success(t("dictionary.batchDeleteTypeSuccess"));
					typeTableRef.current?.clearSelected();
					typeTableRef.current?.refresh();
				},
			});
		},
		[modal, message, t],
	);

	// ---------- 字典数据：弹窗表格列 ----------
	const dataColumns = useMemo<ProTableColumnType<IApi.DictDataItem>[]>(
		() => [
			{ title: t("dictionary.label"), dataIndex: "dict_label", width: 140 },
			{
				title: t("dictionary.value"),
				dataIndex: "dict_value",
				width: 140,
				render: (value) => (
					<Typography.Text code>{value as string}</Typography.Text>
				),
			},
			{ title: t("dictionary.order"), dataIndex: "dict_sort", width: 90 },
			{
				title: t("dictionary.default"),
				dataIndex: "is_default",
				width: 70,
				render: (value) => (
					<Tag color={(value as number) === 1 ? "blue" : "default"}>
						{t(DEFAULT_TEXT[(value as number) ?? 0] ?? "-")}
					</Tag>
				),
			},
			{
				title: t("dictionary.echoStyle"),
				dataIndex: "list_class",
				width: 100,
				render: (value) => (value ? <Tag>{value as string}</Tag> : "-"),
			},
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
								{t("system.edit")}
							</Button>
						)}
						{hasPerm("system:dict:data:delete") && (
							<Popconfirm
								title={t("dictionary.deleteDataConfirm")}
								okText={t("system.delete")}
								cancelText={t("common.cancel")}
								okButtonProps={{ danger: true }}
								onConfirm={() => handleDataDelete(record)}
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
		],
		[hasPerm, openDataEdit, handleDataDelete, t],
	);

	// 打开某类型的数据弹窗：api/toolbarExtra/selectedActions 按当前记录闭包生成，数据变更后走 ref.refresh
	const openData = useCallback(
		(record: IApi.DictTypeItem) => {
			dataModalRef.current?.open({
				title: `${t("dictionary.dataTitle")}: ${record.dict_name}`,
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
						{t("dictionary.addData")}
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
							{t("dictionary.batchDeleteButton", { count: selectedCount })}
						</Button>
					),
				search: {
					fields: [
						{ type: "input", label: t("dictionary.label"), name: "dictLabel" },
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
				},
			});
		},
		[dataColumns, hasPerm, openDataAdd, handleBatchDeleteData, t],
	);

	// ---------- 字典类型：主表列 ----------
	const typeColumns: ProTableColumnType<IApi.DictTypeItem>[] = [
		{ title: t("dictionary.name"), dataIndex: "dict_name", width: 180 },
		{
			title: t("dictionary.type"),
			dataIndex: "dict_type",
			width: 200,
			render: (value) => (
				<Typography.Text code>{value as string}</Typography.Text>
			),
		},
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
							{t("dictionary.viewData")}
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
							{t("system.edit")}
						</Button>
					)}
					{hasPerm("system:dict:type:delete") && (
						<Popconfirm
							title={t("dictionary.deleteTypeConfirm")}
							description={t("dictionary.cascadeDelete")}
							okText={t("system.delete")}
							cancelText={t("common.cancel")}
							okButtonProps={{ danger: true }}
							onConfirm={() => handleTypeDelete(record)}
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
			<ProTable<IApi.DictTypeItem>
				ref={typeTableRef}
				rowKey="id"
				api={Api.getDictTypeList}
				columns={typeColumns}
				title={t("dictionary.typeTitle")}
				toolbarExtra={
					hasPerm("system:dict:type:add") ? (
						<Button
							type="primary"
							icon={<PlusOutlined />}
							onClick={openTypeAdd}
						>
							{t("dictionary.addTypeButton")}
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
							{t("dictionary.batchDeleteButton", { count: selectedCount })}
						</Button>
					)
				}
				search={{
					fields: [
						{ type: "input", label: t("dictionary.name"), name: "dictName" },
						{ type: "input", label: t("dictionary.type"), name: "dictType" },
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
			{/* 数据类型 */}
			<ProForm
				ref={typeFormRef}
				fields={getTypeFormFields()}
				onConfirm={handleTypeSubmit}
			/>
			<TableModalView ref={dataModalRef} title={t("dictionary.dataTitle")} width={1280} />
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
