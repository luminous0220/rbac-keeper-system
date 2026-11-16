import {
	useState,
	useEffect,
	useLayoutEffect,
	useCallback,
	useMemo,
	useRef,
	useImperativeHandle,
} from "react";
import { Table, Card, Form } from "antd";
import type { TableColumnType } from "antd";
import ExcelJS from "exceljs";
import { saveAs } from "file-saver";
import type {
	ProTableProps,
	ProTableColumnType,
	ProTableColumnConfig,
	ProTableDensity,
	ProTablePagination,
	ProTableRef,
} from "../types.d";
import {
	DEFAULT_PAGE_SIZE,
	PAGE_SIZE_OPTIONS,
	SELECTION_COL_WIDTH,
	EXPAND_COL_WIDTH,
	getColumnKey,
	initColumnConfigs,
	getOrderedColumns,
	createIndexColumn,
} from "../utils";
import { useGlobalStore } from "@/stores";
import { ProForm } from "@/components/ProForm";
import { DndContext, DragEndEvent } from "@dnd-kit/core";
import { restrictToVerticalAxis } from "@dnd-kit/modifiers";
import {
	SortableContext,
	arrayMove,
	verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { Toolbar } from "./Toolbar";
import { SelectedInfo } from "./SelectedInfo";
import { DragHandle, SortableRow } from "./DragSortRow";
import { SearchOutlined } from "@ant-design/icons";
import { throttle } from "lodash";
import { useTranslation } from "react-i18next";

export function ProTableInner<T>(
	props: ProTableProps<T>,
	ref: React.ForwardedRef<ProTableRef<T>>,
) {
	const { t } = useTranslation();
	const {
		api,
		auto = true,
		title,
		toolbarExtra,
		index = true,
		indexWidth = 80,
		onSelectRows,
		selectedActions,
		onPageChange,
		columns = [],
		dataSource: externalDataSource,
		total: externalTotal,
		scroll,
		defaultPagination = { pageNumber: 1, pageSize: DEFAULT_PAGE_SIZE },
		search,
		externalParams,
		exportable = false,
		exportFileName,
		onExport,
		dragSort = false,
		onDragSortEnd,
		rowKey,
		...restProps
	} = props;

	// 内部状态
	const [dataSource, setDataSource] = useState<T[]>(externalDataSource ?? []);
	const [total, setTotal] = useState<number>(externalTotal ?? 0);
	const [loading, setLoading] = useState<boolean>(false);
	const [pagination, setPagination] =
		useState<ProTablePagination>(defaultPagination);
	const [density, setDensity] = useState<ProTableDensity>("middle");
	const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
	const [selectedRows, setSelectedRows] = useState<T[]>([]);
	const [columnConfigs, setColumnConfigs] = useState<ProTableColumnConfig[]>(
		() => initColumnConfigs(columns),
	);
	// 筛选表单状态
	const [searchValues, setSearchValues] = useState<Record<string, unknown>>(
		search?.initialValues ?? {},
	);
	const [searchForm] = Form.useForm();
	// 响应式状态
	const isMobile = useGlobalStore((s) => s.isMobile);

	// 是否已初始化（防止重复请求）
	const initializedRef = useRef(false);
	// 记录初始分页（避免 mount 时两个 effect 各触发一次 refresh）
	const prevPaginationRef = useRef(pagination);
	// 强制刷新标识
	const forceRefreshRef = useRef(false);
	// 序号列专用的"显示"分页：仅在 dataSource 更新时同步，避免翻页瞬间序号先于数据变化
	const [displayedPagination, setDisplayedPagination] = useState({
		pageNumber: defaultPagination.pageNumber,
		pageSize: defaultPagination.pageSize,
	});
	useLayoutEffect(() => {
		setDisplayedPagination({
			pageNumber: pagination.pageNumber,
			pageSize: pagination.pageSize,
		});
		// 仅依赖 dataSource，让序号与新数据同步刷新，避免分页变化早于数据导致的序号闪烁
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [dataSource]);

	const debounceRef = useRef<ReturnType<typeof setTimeout>>(null);

	const handleDelayCloseLoading = () => {
		if (debounceRef.current) clearTimeout(debounceRef.current);
		debounceRef.current = setTimeout(() => {
			setLoading(false);
		}, 300);
	};

	const ignore = useRef(false);

	// 刷新数据（合并筛选参数）
	const refresh = useCallback(async () => {
		if (!api || ignore.current) return;
		ignore.current = true;
		setLoading(true);
		try {
			const result = await api({
				pageNumber: pagination.pageNumber,
				pageSize: pagination.pageSize,
				...searchValues,
				...externalParams,
			});
			if (!result || !result.data) return;
			setDataSource(result.data);
			setTotal(result.total || 0);
		} finally {
			handleDelayCloseLoading();
			ignore.current = false;
		}
	}, [api, pagination.pageNumber, pagination.pageSize, searchValues, externalParams]);

	// 自动请求
	useEffect(() => {
		if (!api || !auto) return;
		if (initializedRef.current) return;

		initializedRef.current = true;
		refresh();
	}, [api, auto, refresh]);

	// 分页变化时重新请求（跳过初始值，避免与自动请求重复）
	useEffect(() => {
		if (!api || !auto || !initializedRef.current) return;
		const prev = prevPaginationRef.current;
		if (
			prev.pageNumber === pagination.pageNumber &&
			prev.pageSize === pagination.pageSize &&
			!forceRefreshRef.current
		)
			return;
		prevPaginationRef.current = pagination;
		refresh();
		forceRefreshRef.current = false;
	}, [pagination, api, auto, refresh]);

	// 筛选值变化时重新请求（重置分页到第一页）
	useEffect(() => {
		if (!api || !auto || !initializedRef.current) return;
		setPagination((prev) =>
			prev.pageNumber === 1 ? prev : { ...prev, pageNumber: 1 },
		);
		forceRefreshRef.current = true;
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [searchValues]);

	// 外部参数变化时刷新（重置分页到第一页，与筛选值变化行为一致）。
	// 非第 1 页时改分页走「分页变更」刷新；已在第 1 页时直接刷新，让新参数立即生效。
	const prevExternalParamsRef = useRef<Record<string, unknown> | undefined>(
		externalParams,
	);
	useEffect(() => {
		if (!api || !auto || !initializedRef.current) return;
		if (prevExternalParamsRef.current === externalParams) return;
		prevExternalParamsRef.current = externalParams;
		if (pagination.pageNumber === 1) {
			refresh();
		} else {
			setPagination((prev) => ({ ...prev, pageNumber: 1 }));
		}
	}, [externalParams, api, auto, refresh, pagination.pageNumber]);

	// 同步外部数据
	useEffect(() => {
		if (externalDataSource !== undefined) {
			setDataSource(externalDataSource);
		}
	}, [externalDataSource]);

	useEffect(() => {
		if (externalTotal !== undefined) {
			setTotal(externalTotal);
		}
	}, [externalTotal]);

	// columns 的稳定 key
	const columnsKey = useMemo(() => {
		return columns.map((col, idx) => getColumnKey(col, idx)).join("|");
	}, [columns]);

	// 同步 columns 配置
	useEffect(() => {
		setColumnConfigs(initColumnConfigs(columns));
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [columnsKey]);

	// 处理分页变化
	const handlePaginationChange = useCallback(
		(page: number, pageSize: number) => {
			const newPagination = { pageNumber: page, pageSize };
			setPagination(newPagination);
			onPageChange?.(newPagination);
		},
		[onPageChange],
	);

	// 处理行选择
	const handleRowSelectionChange = useCallback(
		(newSelectedRowKeys: React.Key[]) => {
			setSelectedRowKeys(newSelectedRowKeys);
			const selected = dataSource.filter((row, idx) =>
				newSelectedRowKeys.includes((row as any)?.id ?? idx),
			);
			setSelectedRows(selected);
			onSelectRows?.(selected);
		},
		[onSelectRows, dataSource],
	);

	// 清空选择
	const handleClearSelection = useCallback(() => {
		setSelectedRowKeys([]);
		setSelectedRows([]);
		onSelectRows?.([]);
	}, [onSelectRows]);

	// 拖拽排序结束回调
	const handleDragEnd = useCallback(
		({ active, over }: DragEndEvent) => {
			if (!over || active.id === over.id) return;
			setDataSource((prev) => {
				const activeIndex = prev.findIndex((record) => {
					const key =
						typeof rowKey === "function"
							? (rowKey as (r: T) => string)(record)
							: rowKey
								? String((record as any)[rowKey])
								: String(prev.indexOf(record));
					return key === String(active.id);
				});
				const overIndex = prev.findIndex((record) => {
					const key =
						typeof rowKey === "function"
							? (rowKey as (r: T) => string)(record)
							: rowKey
								? String((record as any)[rowKey])
								: String(prev.indexOf(record));
					return key === String(over.id);
				});
				const newData = arrayMove(prev, activeIndex, overIndex);
				onDragSortEnd?.(newData);
				return newData;
			});
		},
		[rowKey, onDragSortEnd],
	);

	// 计算拖拽排序所需的 ids
	const sortableIds = useMemo(() => {
		if (typeof rowKey === "function") {
			return dataSource.map((record) => (rowKey as (r: T) => string)(record));
		}
		if (typeof rowKey === "string") {
			return dataSource.map((record) => String((record as any)[rowKey]));
		}
		return dataSource.map((_, i) => String(i));
	}, [dataSource, rowKey]);

	// 计算实际渲染的列
	const renderedColumns = useMemo(() => {
		let result = getOrderedColumns(columns, columnConfigs);

		// 拖拽手柄列（最前面）
		if (dragSort) {
			const dragColumn: ProTableColumnType<T> = {
				key: "__drag_handle__",
				width: 80,
				align: "center",
				hideable: false,
				fixable: false,
				render: () => <DragHandle />,
			};
			result = [dragColumn, ...result];
		}

		if (index) {
			result = [
				createIndexColumn(
					displayedPagination.pageNumber,
					displayedPagination.pageSize,
					indexWidth,
				),
				...result,
			];
		}

		// 移动端去除所有列的固定定位，避免横向滚动时布局错乱
		if (isMobile) {
			result = result.map((col) => ({ ...col, fixed: undefined }));
		}

		return result as TableColumnType<T>[];
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [
		columnsKey,
		columnConfigs,
		dragSort,
		index,
		isMobile,
		displayedPagination.pageNumber,
		displayedPagination.pageSize,
	]);

	// 导出 Excel / 自定义导出
	const handleExport = useCallback(async () => {
		// 自定义导出回调：直接调用，由调用方自行处理导出逻辑
		if (onExport) {
			try {
				setLoading(true);
				await onExport(selectedRows);
			} finally {
				handleDelayCloseLoading();
			}
			return;
		}

		// 内置 Excel 导出
		if (!selectedRows.length) {
			window.$message?.warning?.(t("table.exportSelectRequired"));
			return;
		}
		window.$message?.info?.(t("table.exporting"));
		try {
			// 获取当前可见的列（排除无 dataIndex 的列和序号列）
			const exportColumns = renderedColumns.filter(
				(col: any) => col.dataIndex && col.key !== "__index__",
			);
			const workbook = new ExcelJS.Workbook();
			const worksheet = workbook.addWorksheet("Sheet1");

			const header = worksheet.addRow(
				exportColumns.map((col: any) => col.title ?? ""),
			);
			header.eachCell((cell, colNumber) => {
				cell.font = { bold: true };
				cell.fill = {
					type: "pattern",
					pattern: "solid",
					fgColor: { argb: "FFE6F3FF" },
				};
				cell.border = {
					top: { style: "thin", color: { argb: "FF000000" } },
					left: { style: "thin", color: { argb: "FF000000" } },
					bottom: { style: "thin", color: { argb: "FF000000" } },
					right: { style: "thin", color: { argb: "FF000000" } },
				};
				const col = exportColumns[colNumber - 1] as any;
				worksheet.getColumn(colNumber).width =
					typeof col?.width === "number"
						? Math.max(8, Math.floor(col.width / 8))
						: 20;
			});

			selectedRows.forEach((row) => {
				const dataRow = worksheet.addRow(
					exportColumns.map(
						(col: any) => (row as any)[col.dataIndex as string] ?? "",
					),
				);
				dataRow.eachCell((cell) => {
					cell.border = {
						top: { style: "thin", color: { argb: "FF000000" } },
						left: { style: "thin", color: { argb: "FF000000" } },
						bottom: { style: "thin", color: { argb: "FF000000" } },
						right: { style: "thin", color: { argb: "FF000000" } },
					};
				});
			});

			const buffer = await workbook.xlsx.writeBuffer();
			const fileName = exportFileName
				? `${exportFileName}.xlsx`
				: title
					? `${title}-${t("table.fileSuffix")}.xlsx`
					: `${new Date().toISOString().slice(0, 10)}.xlsx`;
			saveAs(
				new Blob([buffer], {
					type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
				}),
				fileName,
			);
		} catch {
			window.$message?.error?.(t("table.exportFailed"));
		}
	}, [onExport, selectedRows, renderedColumns, title, exportFileName, t]);

	// 密度变化回调
	const handleDensityChange = useCallback(
		(newDensity: ProTableDensity) => setDensity(newDensity),
		[],
	);

	// 列配置变化回调
	const handleColumnConfigsChange = useCallback(
		(newConfigs: ProTableColumnConfig[]) => setColumnConfigs(newConfigs),
		[],
	);

	// 筛选表单查询回调
	// 1. 使用 useMemo 缓存节流函数，确保组件生命周期内只有一个节流实例
	const throttleSetSearchValues = useMemo(
		() =>
			throttle((values: Record<string, unknown>) => {
				setSearchValues(values);
			}, 800), // 500ms 节流延迟
		[setSearchValues], // 空依赖数组，确保只在组件挂载时创建一次
	);

	// 2. 使用 useCallback 包裹，保持引用稳定，方便传给子组件
	const handleSearch = useCallback(
		(values: Record<string, unknown>) => {
			forceRefreshRef.current = true;
			throttleSetSearchValues(values);
		},
		[throttleSetSearchValues],
	);

	useEffect(() => {
		return () => {
			throttleSetSearchValues.cancel(); // 组件卸载时取消未执行的防抖任务
		};
	}, [throttleSetSearchValues]);

	// 分页配置
	const paginationConfig = useMemo(() => {
		if (!total || total === 0) return false as const;

		return {
			current: pagination.pageNumber,
			pageSize: pagination.pageSize,
			total,
			pageSizeOptions: PAGE_SIZE_OPTIONS,
			showSizeChanger: true,
			showQuickJumper: true,
			showTotal: (total: number, range: [number, number]) =>
				`${range[0]}-${range[1]} 共 ${total} 条`,
			onChange: handlePaginationChange,
		};
	}, [
		total,
		pagination.pageNumber,
		pagination.pageSize,
		handlePaginationChange,
	]);

	// 行选择配置（selectedActions 内含批量操作时同样需要开启选择列）
	const rowSelectionConfig = useMemo(() => {
		if (!onSelectRows && !exportable && !selectedActions) return undefined;

		return {
			selectedRowKeys,
			onChange: handleRowSelectionChange,
			type: "checkbox" as const,
		};
	}, [
		onSelectRows,
		exportable,
		selectedActions,
		selectedRowKeys,
		handleRowSelectionChange,
	]);

	// 虚拟列表内容宽度：渲染列宽之和 + 选择列/展开列宽度（虚拟表格要求数值型 scroll.x，
	// 字符串会被 VirtualTable 强转成 1 导致表格被截断）
	const virtualScrollX = useMemo(() => {
		const columnsWidth = renderedColumns.reduce(
			(sum, col) => sum + (typeof col.width === "number" ? col.width : 0),
			0,
		);
		return (
			columnsWidth +
			(rowSelectionConfig ? SELECTION_COL_WIDTH : 0) +
			(restProps.expandable ? EXPAND_COL_WIDTH : 0)
		);
	}, [renderedColumns, rowSelectionConfig, restProps.expandable]);

	// 合并滚动配置：虚拟列表强制数值型 scroll.x（有调用方传入时校验其类型），
	// 其余场景沿用调用方配置或默认值
	const mergedScroll = useMemo(() => {
		if (restProps.virtual) {
			const x = typeof scroll?.x === "number" ? scroll.x : virtualScrollX;
			return { ...scroll, x };
		}
		if (scroll) return scroll;
		if (isMobile) return { x: "max-content", y: "420px" };
		return { x: "max-content", y: "calc(100vh - 420px)" };
	}, [scroll, isMobile, restProps.virtual, virtualScrollX]);

	// 稳定的 Table props（无拖拽时的基础配置，拖拽时在此基础上覆盖 components）
	const baseTableProps = useMemo(
		() => ({
			indentSize: 40,
			...restProps,
			rowKey,
			columns: renderedColumns,
			dataSource,
			loading,
			size: density,
			pagination: paginationConfig,
			...(rowSelectionConfig ? { rowSelection: rowSelectionConfig } : {}),
			scroll: mergedScroll,
		}),
		[
			restProps,
			rowKey,
			renderedColumns,
			dataSource,
			density,
			loading,
			paginationConfig,
			rowSelectionConfig,
			mergedScroll,
		],
	);

	// 拖拽模式下的 Table props（合并 SortableRow）
	const tableProps = useMemo(() => {
		if (!dragSort) return baseTableProps;

		return {
			...baseTableProps,
			components: {
				body: {
					row: SortableRow,
				},
			},
		};
	}, [baseTableProps, dragSort]);

	// 暴露实例方法给父组件
	useImperativeHandle(
		ref,
		() => ({
			refresh,
			getDataSource: () => dataSource,
			clearSelected: () => {
				setSelectedRowKeys([]);
				setSelectedRows([]);
				onSelectRows?.([]);
			},
			setSearchValues: (values: Record<string, unknown>) => {
				handleSearch(values);
			},
			reset: () => {
				setSearchValues(search?.initialValues ?? {});
				setPagination(defaultPagination);
			},
			setLoading,
		}),
		[
			refresh,
			dataSource,
			onSelectRows,
			search?.initialValues,
			defaultPagination,
			handleSearch,
		],
	);

	return (
		<>
			{/* 筛选表单 */}
			{search && (
				<Card className="mb-2" classNames={{ body: "!p-4 !px-6" }}>
					<ProForm
						type="pure"
						form={searchForm}
						fields={search.fields}
						initialValues={search.initialValues}
						layout="inline"
						labelCol={{}}
						wrapperCol={{}}
						collapsible={{
							defaultShowCount: search.defaultShowCount ?? 6,
						}}
						disabled={loading}
						defaultButtons={{
							confirmText: t("common.search"),
							resetText: t("common.reset"),
							confirmIcon: <SearchOutlined />,
						}}
						onReset={async (values: Record<string, any>) => {
							handleSearch(values);
						}}
						onConfirm={async (_, values: Record<string, unknown>) => {
							handleSearch(values);
						}}
					/>
				</Card>
			)}

			<Card className="flex flex-col" classNames={{ body: "!p-4 !px-6" }}>
				{/* 工具栏 */}
				<Toolbar
					title={title}
					toolbarExtra={toolbarExtra}
					density={density}
					onDensityChange={handleDensityChange}
					onRefresh={refresh}
					columns={columns as ProTableColumnType<unknown>[]}
					columnConfigs={columnConfigs}
					onColumnConfigsChange={handleColumnConfigsChange}
					isMobile={isMobile}
					exportable={exportable}
					onExport={handleExport}
					selectedCount={selectedRows.length}
				/>

				{/* 选中信息：批量操作按钮随实时选中渲染，未选中（count 为 0）时整条隐藏 */}
				{(onSelectRows || selectedActions) && (
					<SelectedInfo
						selectedCount={selectedRowKeys.length}
						onClear={handleClearSelection}
						actions={selectedActions?.({
							selectedRows,
							selectedCount: selectedRows.length,
						})}
					/>
				)}

				{/* 表格容器 */}
				<div className=" mt-1 flex-1 min-h-0 overflow-hidden">
					{dragSort ? (
						<DndContext
							modifiers={[restrictToVerticalAxis]}
							onDragEnd={handleDragEnd}
						>
							<SortableContext
								items={sortableIds}
								strategy={verticalListSortingStrategy}
							>
								<Table<T> {...tableProps} style={{ height: "100%" }} />
							</SortableContext>
						</DndContext>
					) : (
						<Table<T> {...tableProps} style={{ height: "100%" }} />
					)}
				</div>
			</Card>
		</>
	);
}
