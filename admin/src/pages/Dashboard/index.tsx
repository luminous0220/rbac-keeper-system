import { Row, Col } from "antd";
import { useMemo } from "react";
import * as echarts from "echarts";
import { useChartColors } from "@/hooks/useChartColors";
import { StatCard, type StatCardData } from "@/components/Card";
import { Chart } from "@/components/Chart";
import { useAuthStore } from "@/stores";

// Dashboard 静态占位数据（已确认无需后端接口，仅做展示）
const STAT_CARDS: StatCardData[] = [
	{
		id: "users",
		title: "用户总数",
		value: 12847,
		prefix: "",
		suffix: "",
		tooltip: "系统累计注册用户",
		sparklineType: "line",
		sparklineData: [20, 35, 28, 45, 38, 60, 72, 65, 80, 74, 92, 88],
	},
	{
		id: "depts",
		title: "部门数量",
		value: 12,
		tooltip: "当前组织架构部门数",
		sparklineType: "bar",
		sparklineData: [5, 8, 6, 9, 12, 10, 14],
	},
	{
		id: "roles",
		title: "角色数量",
		value: 8,
		tooltip: "已配置的角色数",
		sparklineType: "line",
		sparklineData: [2, 4, 3, 5, 6, 8],
	},
	{
		id: "ops",
		title: "今日操作",
		value: 1234,
		tooltip: "今日累计操作日志数",
		sparklineType: "bar",
		sparklineData: [300, 420, 380, 510, 640, 580, 720, 810, 760, 920, 1030, 1234],
	},
];

// 分类销售额（横向柱状图）
const SALE_DATA = [
	{ name: "研发", value: 4200 },
	{ name: "市场", value: 3600 },
	{ name: "销售", value: 5200 },
	{ name: "财务", value: 2100 },
	{ name: "人力", value: 1500 },
];

// 流量来源占比（环形图）
const SOURCE_DATA = [
	{ name: "直接访问", value: 335 },
	{ name: "搜索引擎", value: 468 },
	{ name: "外链", value: 198 },
	{ name: "社交媒体", value: 210 },
	{ name: "邮件", value: 122 },
];

// 实时趋势（面积图）
const TREND_DATA = [
	{ time: "10:00", value: 860 },
	{ time: "11:00", value: 1180 },
	{ time: "12:00", value: 1520 },
	{ time: "13:00", value: 1390 },
	{ time: "14:00", value: 1960 },
	{ time: "15:00", value: 2400 },
	{ time: "16:00", value: 2780 },
	{ time: "17:00", value: 3250 },
];

const Dashboard: React.FC = () => {
	const colors = useChartColors();
	const userInfo = useAuthStore((s) => s.userInfo);
	const nickname = userInfo?.nickname || userInfo?.username || "管理员";

	const horizontalBarOption = useMemo(
		() => ({
			tooltip: { trigger: "axis" as const },
			grid: { top: 10, bottom: 10, left: 10, right: 30 },
			xAxis: {
				type: "value" as const,
				axisLabel: { color: colors.textSecondary, fontSize: 11 },
				splitLine: { lineStyle: { color: colors.borderLight } },
			},
			yAxis: {
				type: "category" as const,
				data: SALE_DATA.map((d) => d.name),
				axisLine: { show: false },
				axisTick: { show: false },
				axisLabel: { color: colors.textPrimary, fontSize: 12 },
			},
			series: [
				{
					type: "bar" as const,
					data: SALE_DATA.map((d) => ({
						value: d.value,
						itemStyle: {
							borderRadius: [0, 4, 4, 0],
							color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
								{ offset: 0, color: colors.primary },
								{ offset: 1, color: colors.primaryLight },
							]),
						},
					})),
					barWidth: "50%",
				},
			],
		}),
		[colors],
	);

	const donutOption = useMemo(() => {
		const total = SOURCE_DATA.reduce((sum, item) => sum + item.value, 0);
		const colorsList = [
			"#FFF6DE",
			"#8BDFDD",
			"#F48F68",
			"#FFE394",
			"#59B292",
			"#FFC94D",
			"#FAE7CB",
			"#FA6781",
		];
		const dataWithColor = SOURCE_DATA.map((item, index) => ({
			...item,
			itemStyle: { color: colorsList[index] ?? colorsList[0] },
		}));
		return {
			tooltip: {
				trigger: "item" as const,
				formatter: (params: any) =>
					`${params.name}<br/>${params.value} (${((params.value / total) * 100).toFixed(1)}%)`,
			},
			legend: {
				bottom: 0,
				textStyle: { color: colors.textSecondary, fontSize: 11 },
				itemWidth: 10,
				itemHeight: 10,
				itemGap: 12,
			},
			series: [
				{
					type: "pie" as const,
					radius: ["40%", "80%"],
					center: ["50%", "45%"],
					data: dataWithColor,
					label: {
						show: true,
						position: "outside" as const,
						fontSize: 10,
						color: colors.textPrimary,
					},
					labelLine: {
						show: true,
						length: 10,
						length2: 8,
						smooth: true,
						lineStyle: { width: 1, color: colors.textSecondary },
					},
					emphasis: { scale: true },
					avoidLabelOverlap: true,
				},
			],
		};
	}, [colors]);

	const areaLineOption = useMemo(
		() => ({
			tooltip: {
				trigger: "axis" as const,
				formatter: (params: any) => {
					const p = params[0];
					return `<div style="color:${colors.textPrimary}">${p.axisValue}<br/>¥${p.data.toLocaleString()}</div>`;
				},
			},
			grid: { top: 20, bottom: 20, left: 20, right: 20 },
			xAxis: {
				type: "category" as const,
				data: TREND_DATA.map((d) => d.time),
				axisLine: { show: false },
				axisTick: { show: false },
				axisLabel: { color: colors.textSecondary, fontSize: 10 },
			},
			yAxis: {
				type: "value" as const,
				axisLine: { show: false },
				axisTick: { show: false },
				axisLabel: { color: colors.textSecondary, fontSize: 10 },
				splitLine: { lineStyle: { color: colors.borderLight } },
			},
			series: [
				{
					type: "line" as const,
					data: TREND_DATA.map((d) => d.value),
					smooth: true,
					symbol: "none",
					lineStyle: { color: colors.primary, width: 2 },
					areaStyle: {
						color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
							{ offset: 0, color: colors.primary + "60" },
							{ offset: 1, color: colors.primary + "10" },
						]),
					},
				},
			],
		}),
		[colors],
	);

	return (
		<div className="space-y-4">
			<div className="overflow-hidden rounded-2xl bg-gradient-to-r from-[#E6F0FF] to-[#f8fafe] shadow-soft dark:from-[rgb(37,119,245,0.3)] dark:to-[rgb(37,119,245,0.1)]">
				<div className="flex flex-col gap-4 p-6 md:flex-row md:items-center md:justify-between">
					<div>
						<h2 className="text-lg font-semibold text-[#1f2a44] dark:text-white">
							欢迎回来，{nickname} 👋
						</h2>
						<p className="mt-1 text-sm text-[#7c8aa5]">
							今天又是充满效率的一天，祝你工作顺利。
						</p>
					</div>
					<div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-white text-3xl shadow-soft">
						🐼
					</div>
				</div>
			</div>

			<Row gutter={[16, 16]}>
				{STAT_CARDS.map((item) => (
					<Col xs={24} sm={12} md={8} lg={6} key={item.id}>
						<StatCard data={item} colors={colors} />
					</Col>
				))}
			</Row>

			<Row gutter={[16, 16]}>
				<Col xs={24} lg={12}>
					<Chart type="card" title="分类销售额" option={horizontalBarOption} />
				</Col>
				<Col xs={24} lg={12}>
					<Chart type="card" title="流量来源占比" option={donutOption} />
				</Col>
			</Row>

			<Row gutter={[16, 16]}>
				<Col xs={24} lg={24}>
					<Chart
						type="card"
						title="实时销售额趋势"
						option={areaLineOption}
						height={280}
					/>
				</Col>
			</Row>
		</div>
	);
};

export default Dashboard;
