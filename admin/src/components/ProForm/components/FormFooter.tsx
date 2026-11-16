import { Button, Form } from "antd";
import { memo } from "react";
import type { ProFormProps } from "../types.d";
import { useTranslation } from "react-i18next";

/** Footer 按钮组件（重置 + 确认） */
export const FormFooter = memo(function FormFooter({
	footer,
	loading,
	onConfirm,
	onReset,
	className,
}: {
	footer: NonNullable<ProFormProps["defaultButtons"]>;
	loading: boolean;
	onConfirm: () => void;
	onReset: () => void;
	className?: string;
}) {
	const { t } = useTranslation();
	const {
		confirmText = t("common.confirm"),
		resetText = t("common.reset"),
		hideConfirm,
		hideReset,
		confirmStyle,
	} = footer;

	return (
		<Form.Item className="mb-0">
			<div
				className={`flex items-center justify-end w-full gap-4 ${className}`}
			>
				{!hideReset && <Button onClick={onReset}>{resetText}</Button>}
				{!hideConfirm && (
					<Button
						type="primary"
						{...confirmStyle}
						loading={loading}
						onClick={onConfirm}
					>
						{confirmText}
					</Button>
				)}
			</div>
		</Form.Item>
	);
});
