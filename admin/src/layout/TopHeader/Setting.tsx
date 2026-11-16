import { Button, Drawer } from "antd";
import { useState } from "react";
import { useThemeStore } from "@/stores/theme";
import { ThemePanel } from "@/components/ThemePanel";
import { IconSettings } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

// 可控制的主题抽屉组件
export const Setting = () => {
	const [open, setOpen] = useState(false);
	const { colorPrimary } = useThemeStore();
	const { t } = useTranslation();

	const showDrawer = () => {
		setOpen(true);
	};

	const onClose = () => {
		setOpen(false);
	};

	return (
		<div className="flex items-center gap-2">
			<Button
				type="text"
				icon={<IconSettings stroke={1} />}
				onClick={showDrawer}
				aria-label={t("layout.settings")}
				style={{
					color: open ? colorPrimary : undefined,
					transition: "all 0.2s",
				}}
			/>


			<Drawer
				title={t("layout.settings")}
				placement="right"
				onClose={onClose}
				open={open}
				size={320}
			>
				<ThemePanel />
			</Drawer>
		</div>
	);
};
