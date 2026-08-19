import { Card, Empty } from "antd";

/** 页面开发占位：后续阶段替换为真实 CRUD 页面 */
export const PagePlaceholder = ({ title }: { title: string }) => (
	<Card className="h-full">
		<Empty description={`${title} 开发中，敬请期待`} />
	</Card>
);
