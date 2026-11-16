import { useCallback, useMemo, useRef, useState } from "react";
import { SelectorTrigger } from "./components/SelectorTrigger";
import { ListSelectorModal, type ListSelectorModalRef } from "./ListSelectorModal";
import type {
	ListItemData,
	ListSelectorKey,
	ListSelectorNode,
	ListSelectorProps,
} from "./types";
import { useTranslation } from "react-i18next";

export type {
	ListItemData,
	ListSelectorNode,
	ListSelectorKey,
	ListSelectorProps,
} from "./types";

/**
 * @description ListSelector 列表选择器（触发器 + 弹窗）
 * 适合作为表单项使用，自带触发器，点击打开弹窗进行选择。
 * 数据通过 api 分页获取，不支持 dataSource。
 * checkedKeys/onChange 使用简单 key 列表（与 ListItemData.value 类型一致），
 * title/desc 等展示信息由组件内部从 Modal 确认时返回的完整数据中推导并缓存。
 * checkedKeys 不传时为非受控模式，内部维护选中状态。
 * 如需编程式控制弹窗，请使用 ListSelectorModal。
 */
export const ListSelector: React.FC<ListSelectorProps> = ({
	title,
	checkedKeys: controlledCheckedKeys,
	onChange,
	api,
	placeholder: providedPlaceholder,
	disabled = false,
	width,
	renderTrigger,
}) => {
	const { t } = useTranslation();
	const placeholder = providedPlaceholder ?? t("common.select");
	const isControlled = controlledCheckedKeys !== undefined;
	const [internalCheckedKeys, setInternalCheckedKeys] = useState<ListSelectorKey[]>(
		controlledCheckedKeys ?? [],
	);
	// 缓存 key -> ListItemData：Modal 确认后写入；删除已选项时移除。
	// 外部直接受控传 checkedKeys 时缓存可能为空，触发器会展示空 title 直至下次 Modal 确认。
	const [itemCache, setItemCache] = useState<Map<ListSelectorKey, ListItemData>>(
		() => new Map(),
	);
	const currentCheckedKeys = isControlled ? controlledCheckedKeys : internalCheckedKeys;

	const modalRef = useRef<ListSelectorModalRef>(null);

	// 由 keys + 缓存推导 ListSelectorNode[]（带 title）给触发器展示
	const selectedItems = useMemo<ListSelectorNode[]>(() => {
		return currentCheckedKeys.map((key) => {
			const cached = itemCache.get(key);
			if (cached) {
				return {
					value: cached.value,
					title: cached.title,
					desc: cached.desc,
				};
			}
			// 缓存缺失：fallback 用 key 构造（title 留空，避免误显示）
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

	// 弹窗确认：拿到的是完整 ListItemData[]，更新缓存并回写 keys
	const handleConfirm = useCallback(
		(selected: ListItemData[]) => {
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
		(item: ListSelectorNode) => {
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
			<ListSelectorModal
				ref={modalRef}
				title={title}
				width={width}
				api={api}
				onConfirm={handleConfirm}
			/>
		</>
	);
};
