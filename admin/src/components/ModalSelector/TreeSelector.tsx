import { useCallback, useMemo, useRef, useState } from "react";
import { SelectorTrigger } from "./components/SelectorTrigger";
import { TreeSelectorModal, type TreeSelectorModalRef } from "./TreeSelectorModal";
import type { TreeSelectorKey, TreeSelectorProps } from "./types";
import { findNode, getDescendantKeys } from "./utils";

export type {
	TreeNodeData,
	TreeSelectorNode,
	TreeSelectorKey,
	TreeSelectorProps,
} from "./types";

/**
 * @description TreeSelector 树形选择器（触发器 + 弹窗）
 * 适合作为表单项使用，自带触发器，点击打开弹窗进行选择。
 * 支持 treeData 静态数据或 api 动态获取数据。
 * checkedKeys/onChange 使用简单 key 列表（与 TreeNodeData.value 类型一致），
 * title 等展示信息由组件内部从 treeData 推导，无需调用方手动拼装。
 * checkedKeys 不传时为非受控模式，内部维护选中状态。
 * 如需编程式控制弹窗，请使用 TreeSelectorModal。
 */
export const TreeSelector: React.FC<TreeSelectorProps> = ({
	title,
	checkedKeys: controlledCheckedKeys,
	onChange,
	treeData = [],
	api,
	placeholder = "请选择",
	disabled = false,
	width,
	renderTrigger,
}) => {
	const isControlled = controlledCheckedKeys !== undefined;
	const [internalCheckedKeys, setInternalCheckedKeys] = useState<TreeSelectorKey[]>(
		controlledCheckedKeys ?? [],
	);
	const currentCheckedKeys = isControlled
		? controlledCheckedKeys
		: internalCheckedKeys;

	const modalRef = useRef<TreeSelectorModalRef>(null);

	// 由 key 列表从 treeData 推导 TreeSelectorNode[]（带 title），
	// 触发器 SelectorTrigger 需要 title 才能展示 Tag
	const selectedNodes = useMemo(() => {
		if (!currentCheckedKeys?.length) return [];
		return currentCheckedKeys
			.map((key) => {
				const node = findNode(treeData, key);
				if (!node) return null;
				const isParent = !!node.children?.length;
				return {
					value: node.value as number | string,
					title: node.title,
					type: (isParent ? "parent" : "child") as "parent" | "child",
				};
			})
			.filter((n): n is NonNullable<typeof n> => n !== null);
	}, [currentCheckedKeys, treeData]);

	const handleOpen = useCallback(() => {
		if (disabled) return;
		modalRef.current?.open({
			title,
			treeData,
			checkedKeys: currentCheckedKeys,
		});
	}, [disabled, currentCheckedKeys, title, treeData]);

	// 弹窗确认：拿到的是 key 列表，直接回写到受控/内部状态
	const handleConfirm = useCallback(
		(selected: TreeSelectorKey[]) => {
			if (!isControlled) setInternalCheckedKeys(selected);
			onChange?.(selected);
		},
		[isControlled, onChange],
	);

	const removeTag = useCallback(
		(item: { value: number | string }) => {
			let newSelected = currentCheckedKeys.filter(
				(s) => s !== item.value,
			);
			const nodeData = findNode(treeData, item.value);
			if (nodeData?.children?.length) {
				const descendantKeys = getDescendantKeys(treeData, item.value);
				newSelected = newSelected.filter(
					(s) => !descendantKeys.includes(s),
				);
			}
			if (!isControlled) setInternalCheckedKeys(newSelected);
			onChange?.(newSelected);
		},
		[currentCheckedKeys, isControlled, treeData, onChange],
	);

	return (
		<>
			<SelectorTrigger
				selectedItems={selectedNodes}
				placeholder={placeholder}
				disabled={disabled}
				onOpen={handleOpen}
				onRemoveTag={removeTag}
				renderTrigger={renderTrigger}
			/>
			<TreeSelectorModal
				ref={modalRef}
				title={title}
				treeData={treeData}
				api={api}
				width={width}
				onConfirm={handleConfirm}
			/>
		</>
	);
};
