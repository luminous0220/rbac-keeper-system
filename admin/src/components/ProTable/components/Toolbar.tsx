import { memo } from "react";
import { Button, Dropdown, Popover, Typography } from "antd";
import {
	IconRefresh,
	IconBaselineDensityLarge,
	IconSettings,
	IconDownload,
} from "@tabler/icons-react";
import type { MenuProps } from "antd";
import type { ProTableDensity, ProTableToolbarProps } from "../types.d";
import { TableDragSort } from "./TableDragSort";
import { useTranslation } from "react-i18next";

export const Toolbar = memo(function Toolbar({
	title,
	toolbarExtra,
	density,
	onDensityChange,
	onRefresh,
	columns,
	columnConfigs,
	onColumnConfigsChange,
	isMobile,
	exportable,
	onExport,
	selectedCount,
}: ProTableToolbarProps) {
	const { t } = useTranslation();
	const densityItems: MenuProps["items"] = [
		{ key: "small", label: t("table.densityCompact") },
		{ key: "middle", label: t("table.densityDefault") },
		{ key: "large", label: t("table.densityComfortable") },
	];

	return (
		<div className="flex items-center justify-between mb-3 shrink-0 gap-2 flex-wrap ">
			<div className="flex items-center gap-2 min-w-0">
				{title && (
					<span
						className={`text-base font-medium whitespace-nowrap ${
							isMobile && "flex-wrap"
						}`}
					>
						{title}
					</span>
				)}
			</div>
			<div
				className={`flex items-center gap-1.5  ${isMobile ? "flex-wrap	" : ""}`}
			>
				{toolbarExtra}
				{exportable && (
					<Button
						icon={<IconDownload size={18} stroke={1.5} />}
						onClick={onExport}
						size="middle"
					>
						{selectedCount
							? t("table.exportSelected", { count: selectedCount })
							: t("table.export")}
					</Button>
				)}
				<Button
					icon={<IconRefresh size={18} stroke={1.5} />}
					onClick={onRefresh}
					size="middle"
				/>
				<Dropdown
					menu={{
						items: densityItems,
						selectedKeys: [density],
						onClick: ({ key }) => onDensityChange(key as ProTableDensity),
					}}
				>
					<Button
						icon={<IconBaselineDensityLarge size={18} stroke={1.5} />}
						size="middle"
					/>
				</Dropdown>

				<Popover
					content={
						<div className="max-w-64 mt-[-8px]">
							<Typography.Text type="secondary" className="text-xs">
								{t("table.columnHints")}
							</Typography.Text>
							<TableDragSort
								columns={columns}
								columnConfigs={columnConfigs}
								onColumnConfigsChange={onColumnConfigsChange}
							/>
						</div>
					}
					title={t("table.columnSettings")}
					trigger="click"
					placement="bottomRight"
				>
					<Button icon={<IconSettings />} type="text" />
				</Popover>
			</div>
		</div>
	);
});
