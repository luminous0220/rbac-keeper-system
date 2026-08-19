import { useCallback } from "react";
import { Tag, Typography } from "antd";
import { ProTable } from "@/components/ProTable";
import type { ProTableColumnType } from "@/components/ProTable";
import { Api, IApi } from "@/apis";
import { toDateTime } from "@/libs/date";
import dayjs from "dayjs";

// 登录状态：0成功 1失败（dto.LoginLogItem.Status）
const STATUS_TEXT: Record<number, string> = { 0: "成功", 1: "失败" };
const STATUS_OPTIONS = [
	{ label: "成功", value: 0 },
	{ label: "失败", value: 1 },
];

/**
 * 登录日志：只读分页查询。
 * - 数据源 GET /api/log/login（RBAC monitor:log:login，DataScope 已在 DAO 应用）
 * - 搜索 username（模糊）/ip（精确）/status/startTime/endTime
 * - Dayjs 时间值在 api 包装层统一转 YYYY-MM-DD HH:mm:ss
 */
const LoginLogPage: React.FC = () => {
	const loginLogApi = useCallback(async (params: Record<string, unknown>) => {
		return Api.getLoginLogList({
			pageNumber: params.pageNumber as number,
			pageSize: params.pageSize as number,
			username: params.username ? String(params.username) : undefined,
			ip: params.ip ? String(params.ip) : undefined,
			status: params.status as number | undefined,
			startTime: toDateTime(params.startTime),
			endTime: toDateTime(params.endTime),
		});
	}, []);

	const columns: ProTableColumnType<IApi.LoginLogItem>[] = [
		{ title: "用户名", dataIndex: "username", width: 130 },
		{
			title: "登录 IP",
			dataIndex: "ip",
			width: 140,
			render: (value) => (
				<Typography.Text code>{value as string}</Typography.Text>
			),
		},
		{
			title: "归属地",
			dataIndex: "ipAddr",
			width: 150,
			ellipsis: true,
			render: (v) => v || "-",
		},
		{
			title: "浏览器",
			dataIndex: "browser",
			width: 160,
			ellipsis: true,
			render: (v) => v || "-",
		},
		{
			title: "操作系统",
			dataIndex: "os",
			width: 120,
			ellipsis: true,
			render: (v) => v || "-",
		},
		{
			title: "状态",
			dataIndex: "status",
			width: 80,
			render: (value) => (
				<Tag color={(value as number) === 0 ? "green" : "red"}>
					{STATUS_TEXT[(value as number) ?? 0] ?? "-"}
				</Tag>
			),
		},
		{
			title: "提示消息",
			dataIndex: "msg",
			width: 220,
			ellipsis: true,
			render: (v) => v || "-",
		},
		{
			title: "登录时间",
			dataIndex: "loginTime",
			width: 180,
			render: (time) => dayjs(time).format("YYYY-MM-DD HH:mm:ss"),
		},
	];

	return (
		<ProTable<IApi.LoginLogItem>
			rowKey="id"
			api={loginLogApi}
			columns={columns}
			title="登录日志"
			search={{
				fields: [
					{ type: "input", label: "用户名", name: "username" },
					{ type: "input", label: "登录 IP", name: "ip" },
					{
						type: "select",
						label: "状态",
						name: "status",
						fieldProps: { options: STATUS_OPTIONS, allowClear: true },
					},
					{
						type: "date-picker",
						label: "开始时间",
						name: "startTime",
						fieldProps: {
							showTime: true,
							style: { width: "100%" },
							placeholder: "开始时间",
						},
					},
					{
						type: "date-picker",
						label: "结束时间",
						name: "endTime",
						fieldProps: {
							showTime: true,
							style: { width: "100%" },
							placeholder: "结束时间",
						},
					},
				],
			}}
		/>
	);
};

export default LoginLogPage;
