import { useCallback, useMemo, useRef, useState } from "react";
import { SelectorTrigger } from "./components/SelectorTrigger";
import {
	TableSelectorModal,
	type TableSelectorModalRef,
} from "./TableSelectorModal";
import type {
	TableSelectorNode,
	TableSelectorKey,
	TableSelectorProps,
} from "./types";

export type {
	TableSelectorNode,
	TableSelectorKey,
	TableSelectorProps,
} from "./types";

/**
 * @description TableSelector 表格选择器（触发器 + 弹窗）
 * 适合作为表单项使用，自带触发器，点击打开弹窗进行选择。
 * checkedKeys/onChange 使用简单 key 列表（与行 rowKey 类型一致），
 * title/desc 等展示信息由组件内部从 Modal 确认时返回的完整行数据中推导并缓存。
 * checkedKeys 不传时为非受控模式，内部维护选中状态。
 * 如需编程式控制弹窗，请使用 TableSelectorModal。
 */
export function TableSelector<
	T extends Record<string, any> = Record<string, any>,
>({
	title,
	checkedKeys: controlledCheckedKeys,
	onChange,
	placeholder = "请选择",
	disabled = false,
	width,
	selectionType,
	renderTrigger,
	...restProps
}: TableSelectorProps<T>) {
	const isControlled = controlledCheckedKeys !== undefined;
	const [internalCheckedKeys, setInternalCheckedKeys] = useState<TableSelectorKey[]>(
		controlledCheckedKeys ?? [],
	);
	// 缓存 key -> TableSelectorNode：Modal 确认后写入；删除已选项时移除。
	// 外部直接受控传 checkedKeys 时缓存可能为空，触发器会展示空 title 直至下次 Modal 确认。
	const [itemCache, setItemCache] = useState<Map<TableSelectorKey, TableSelectorNode>>(
		() => new Map(),
	);
	const currentCheckedKeys = isControlled ? controlledCheckedKeys : internalCheckedKeys;

	const modalRef = useRef<TableSelectorModalRef>(null);

	// 由 keys + 缓存推导 TableSelectorNode[]（带 title）给触发器展示
	const selectedItems = useMemo<TableSelectorNode[]>(() => {
		return currentCheckedKeys.map((key) => {
			const cached = itemCache.get(key);
			if (cached) {
				return {
					value: cached.value,
					title: cached.title,
					desc: cached.desc,
				};
			}
			// 缓存缺失：fallback 用 key 构造
			return { value: key, title: "" };
		});
	}, [currentCheckedKeys, itemCache]);

	const handleOpen = useCallback(() => {
		if (disabled) return;
		modalRef.current?.open({
			title,
			checkedKeys: currentCheckedKeys,
		});
	}, [disabled, currentCheckedKeys, title]);

	// 弹窗确认：拿到的是完整 TableSelectorNode[]，更新缓存并回写 keys
	const handleConfirm = useCallback(
		(selected: TableSelectorNode[]) => {
			const newKeys = selected.map((s) => s.value);
			setItemCache((prev) => {
				const next = new Map(prev);
				for (const item of selected) next.set(item.value, item);
				return next;
			});
			if (!isControlled) setInternalCheckedKeys(newKeys);
			onChange?.(newKeys);
		},
		[isControlled, onChange],
	);

	const removeTag = useCallback(
		(item: TableSelectorNode) => {
			const newKeys = currentCheckedKeys.filter((k) => k !== item.value);
			setItemCache((prev) => {
				const next = new Map(prev);
				next.delete(item.value);
				return next;
			});
			if (!isControlled) setInternalCheckedKeys(newKeys);
			onChange?.(newKeys);
		},
		[currentCheckedKeys, isControlled, onChange],
	);

	return (
		<>
			<SelectorTrigger
				selectedItems={selectedItems}
				placeholder={placeholder}
				disabled={disabled}
				onOpen={handleOpen}
				onRemoveTag={removeTag}
				renderTrigger={renderTrigger}
			/>
			<TableSelectorModal<T>
				ref={modalRef}
				title={title}
				width={width}
				selectionType={selectionType}
				onConfirm={handleConfirm}
				{...restProps}
			/>
		</>
	);
}
