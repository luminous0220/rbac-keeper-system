import { Card, Empty } from "antd";
import { useTranslation } from "react-i18next";

/** 页面开发占位：后续阶段替换为真实 CRUD 页面 */
export const PagePlaceholder = ({ title }: { title: string }) => {
	const { t } = useTranslation();
	return (
		<Card className="h-full">
			<Empty description={t("pagePlaceholder.development", { title })} />
		</Card>
	);
};
