import { useCallback, useRef } from "react";
import { App, Button, Tag, Typography } from "antd";
import { DownloadOutlined, DeleteOutlined } from "@ant-design/icons";
import { ProTable } from "@/components/ProTable";
import type { ProTableColumnType, ProTableRef } from "@/components/ProTable";
import { Api, IApi } from "@/apis";
import { usePerms } from "@/hooks";
import { downloadBlob } from "@/libs";
import { toDateTime } from "@/libs/date";
import dayjs from "dayjs";

// 登录状态：0成功 1失败（dto.LoginLogItem.Status）
const STATUS_TEXT: Record<number, string> = { 0: "成功", 1: "失败" };
const STATUS_OPTIONS = [
	{ label: "成功", value: 0 },
	{ label: "失败", value: 1 },
];

/**
 * 登录日志：只读分页查询 + 导出 / 清空。
 * - 数据源 GET /api/log/login（RBAC monitor:log:login）
 * - 搜索 username（模糊）/ip（精确）/status/startTime/endTime
 * - 导出 GET /api/log/login/export（monitor:log:login:export，携带当前筛选）
 * - 清空 POST /api/log/login/clear（monitor:log:login:clear，物理删除全部）
 * - Dayjs 时间值在 api 包装层统一转 YYYY-MM-DD HH:mm:ss
 */
const LoginLogPage: React.FC = () => {
	const { message, modal } = App.useApp();
	const { hasPerm } = usePerms();
	const tableRef = useRef<ProTableRef<IApi.LoginLogItem>>(null);
	// 最近一次已查询的筛选参数（导出复用：导出的即当前表格所查）
	const lastQueryRef = useRef<Record<string, unknown>>({});

	const loginLogApi = useCallback(async (params: Record<string, unknown>) => {
		const query: Record<string, unknown> = {
			username: params.username ? String(params.username) : undefined,
			ip: params.ip ? String(params.ip) : undefined,
			status: params.status as number | undefined,
			startTime: toDateTime(params.startTime),
			endTime: toDateTime(params.endTime),
		};
		lastQueryRef.current = query;
		return Api.getLoginLogList({
			pageNumber: params.pageNumber as number,
			pageSize: params.pageSize as number,
			...query,
		});
	}, []);

	// 导出当前筛选结果（错误提示由请求层统一弹出，无需在此重复）
	const handleExport = async () => {
		try {
			const blob = await Api.exportLoginLogs({ ...lastQueryRef.current });
			downloadBlob(blob, "login_logs.csv");
			message.success("导出成功");
		} catch {
			/* noop */
		}
	};

	const confirmClear = () => {
		modal.confirm({
			title: "清空登录日志",
			content: "确定清空全部登录日志吗？该操作不可恢复。",
			okText: "清空",
			okButtonProps: { danger: true },
			onOk: async () => {
				try {
					await Api.clearLoginLogs();
					message.success("登录日志已清空");
					tableRef.current?.refresh();
				} catch {
					/* noop */
				}
			},
		});
	};

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
			ref={tableRef}
			rowKey="id"
			api={loginLogApi}
			columns={columns}
			title="登录日志"
			toolbarExtra={
				<>
					{hasPerm("monitor:log:login:export") && (
						<Button icon={<DownloadOutlined />} onClick={handleExport}>
							导出
						</Button>
					)}
					{hasPerm("monitor:log:login:clear") && (
						<Button danger icon={<DeleteOutlined />} onClick={confirmClear}>
							清空
						</Button>
					)}
				</>
			}
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
