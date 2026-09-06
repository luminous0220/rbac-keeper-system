import { memo } from "react";
import type { ReactNode } from "react";
import { Alert, Button, Space } from "antd";

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
	if (selectedCount === 0) return null;

	return (
		<div className="shrink-0 mb-1">
			<Alert
				type="info"
				showIcon
				title={`已选择 ${selectedCount} 条`}
				action={
					<Space size={8} wrap>
						{actions}
						<Button  onClick={onClear}>
							取消选择
						</Button>
					</Space>
				}
			/>
		</div>
	);
});
