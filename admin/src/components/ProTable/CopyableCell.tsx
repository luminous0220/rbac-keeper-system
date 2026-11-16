import { IconCopy } from "@tabler/icons-react";
import { Button } from "antd";
import React, { useCallback } from "react";
import { useTranslation } from "react-i18next";

export const CopyableCell: React.FC<{
	text: string;
	children?: React.ReactNode;
	className?: string;
}> = React.memo(({ text, children, className = "" }) => {
	const { t } = useTranslation();
	const handleCopy = useCallback(
		async (e: React.MouseEvent) => {
			e.stopPropagation();
			try {
				await navigator.clipboard.writeText(String(text));
				window.$message?.success?.(t("table.copySuccess"));
			} catch {
				window.$message?.error?.(t("table.copyFailed"));
			}
		},
		[text, t],
	);

	return (
		<div
			className={`flex items-center justify-between gap-1 group/row ${className}`}
		>
			{children || <span className="truncate">{text}</span>}
			<Button
				type="text"
				size="small"
				className="opacity-0 group-hover/row:opacity-100 transition-opacity shrink-0"
				icon={<IconCopy size={14} stroke={1.5} />}
				onClick={handleCopy}
			/>
		</div>
	);
});
