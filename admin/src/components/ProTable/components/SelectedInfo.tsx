import { memo } from "react";
import type { ReactNode } from "react";
import { Alert, Button, Space } from "antd";
import { useTranslation } from "react-i18next";

export const SelectedInfo = memo(function SelectedInfo({
	selectedCount,
	onClear,
	actions,
}: {
	selectedCount: number;
	onClear: () => void;
	/** 批量操作按钮（如批量删除），由 ProTable 经 selectedActions 渲染后注入 */
	actions?: ReactNode;
}) {
	const { t } = useTranslation();
	if (selectedCount === 0) return null;

	return (
		<div className="shrink-0 mb-1">
			<Alert
				type="info"
				showIcon
				title={t("table.selectedCount", { count: selectedCount })}
				action={
					<Space size={8} wrap>
						{actions}
						<Button  onClick={onClear}>
							{t("table.cancelSelection")}
						</Button>
					</Space>
				}
			/>
		</div>
	);
});
