import { useState, useCallback, useRef, forwardRef, useImperativeHandle } from "react";
import { ProModal } from "@/components/ProModal";
import { ProTable } from "@/components/ProTable";
import type { ProTableColumnType, ProTableRef } from "@/components/ProTable";
import type {
	TableModalViewProps,
	TableModalViewRef,
	TableModalOpenProps,
} from "./type";
import { useTranslation } from "react-i18next";

export type { TableModalViewProps, TableModalViewRef, TableModalOpenProps } from "./type";

/**
 * @description TableModalView 表格弹窗查看组件
 * 通过 ref.open(props) 打开 ProModal 并渲染 ProTable，
 * ref.close() 关闭弹窗。无触发器，完全由父组件编程式控制。
 */
function TableModalViewInner<T extends Record<string, any> = Record<string, any>>(
	{ title: defaultTitle, width = 1280 }: TableModalViewProps,
	ref: React.ForwardedRef<TableModalViewRef<T>>,
) {
	const { t } = useTranslation();
	const [isOpen, setIsOpen] = useState(false);
	const [tableProps, setTableProps] = useState<TableModalOpenProps<T>>({});
	const tableRef = useRef<ProTableRef<T>>(null);

	const open = useCallback((props: TableModalOpenProps<T>) => {
		setTableProps(props);
		setIsOpen(true);
	}, []);

	const close = useCallback(() => {
		setIsOpen(false);
	}, []);

	// 暴露 open/close 及内部 ProTable 实例方法（代理到 tableRef，供父组件直接 refresh 等）
	useImperativeHandle(
		ref,
		() => ({
			open,
			close,
			refresh: () => tableRef.current?.refresh() ?? Promise.resolve(),
			getDataSource: () => tableRef.current?.getDataSource() ?? [],
			clearSelected: () => tableRef.current?.clearSelected(),
			setSearchValues: (values: Record<string, unknown>) =>
				tableRef.current?.setSearchValues(values),
			reset: () => tableRef.current?.reset(),
			setLoading: (loading: boolean) => tableRef.current?.setLoading(loading),
		}),
		[open, close],
	);

	const {
		title: modalTitle,
		columns = [],
		rowKey = "id",
		...restTableProps
	} = tableProps;

	return (
		<ProModal
			open={isOpen}
			onCancel={close}
			title={modalTitle ?? defaultTitle ?? t("table.viewData")}
			width={width}
			footer={null}
			destroyOnHidden
		>
			<ProTable<T>
				ref={tableRef}
				columns={columns as ProTableColumnType<T>[]}
				rowKey={rowKey}
				index={true}
				scroll={restTableProps.scroll ?? { y: "420px" }}
				{...restTableProps}
			/>
		</ProModal>
	);
}

/**
 * TableModalView（forwardRef + memo）
 * 通过 ref.open() / ref.close() 控制弹窗开闭
 */
export const TableModalView = forwardRef(TableModalViewInner) as <
	T extends Record<string, any> = Record<string, any>,
>(
	props: TableModalViewProps & {
		ref?: React.Ref<TableModalViewRef<T>>;
	},
) => React.ReactElement | null;
