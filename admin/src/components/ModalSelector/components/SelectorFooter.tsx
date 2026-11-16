import { Button, Space } from "antd";
import type { SelectorFooterProps } from "../types";
import { useTranslation } from "react-i18next";

/**
 * @description 选择器底部操作栏 — 显示已选数量 + 取消/确定按钮
 * TreeSelector 和 TableSelector 共用
 */
export const SelectorFooter: React.FC<SelectorFooterProps> = ({
	selectedCount,
	onCancel,
	onConfirm,
}) => {
	const { t } = useTranslation();
	return (
		<div className="flex justify-between items-center mt-4 pt-4 border-t border-[var(--ant-color-border)]">
			<span className="text-gray-500 text-sm">
				{t("selector.selectedCount", { count: selectedCount })}
			</span>
			<Space>
				<Button size="large" onClick={onCancel}>
					{t("selector.cancel")}
				</Button>
				<Button size="large" type="primary" onClick={onConfirm}>
					{t("selector.confirm")}
				</Button>
			</Space>
		</div>
	);
};
