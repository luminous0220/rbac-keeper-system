import { Button, Dropdown } from "antd";
import type { MenuProps } from "antd";
import { IconLanguage } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";
import i18n from "@/i18n";

export const LanguageSwitcher = () => {
	const { t } = useTranslation();
	const items: MenuProps["items"] = [
		{ key: "zh-CN", label: t("layout.chinese") },
		{ key: "en-US", label: t("layout.english") },
	];

	return (
		<Dropdown
			menu={{
				items,
				selectedKeys: [i18n.resolvedLanguage ?? "zh-CN"],
				onClick: ({ key }) => {
					void i18n.changeLanguage(key);
				},
			}}
			placement="bottomRight"
		>
			<Button
				type="text"
				icon={<IconLanguage stroke={1} />}
				aria-label={t("layout.language")}
				title={t("layout.language")}
			/>
		</Dropdown>
	);
};
