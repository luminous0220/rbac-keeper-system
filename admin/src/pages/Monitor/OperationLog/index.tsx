import { useCallback } from "react";
import { Tag, Typography } from "antd";
import { ProTable } from "@/components/ProTable";
import type { ProTableColumnType } from "@/components/ProTable";
import { Api, IApi } from "@/apis";
import { toDateTime } from "@/libs/date";

// 操作状态：0正常 1异常（dto.OperationLogItem.Status）
const STATUS_TEXT: Record<number, string> = { 0: "正常", 1: "异常" };
const STATUS_OPTIONS = [
	{ label: "正常", value: 0 },
	{ label: "异常", value: 1 },
];

/**
 * 操作日志：只读分页查询。
 * - 数据源 GET /api/log/operation（RBAC monitor:log:operation）
 * - 搜索 title（模糊）/operName（模糊）/status/startTime/endTime
 * - 登出等后台动作也会以操作日志「退出登录」形式写入，此处一并展示
 */
const OperationLogPage: React.FC = () => {
	const operationLogApi = useCallback(async (params: Record<string, unknown>) => {
		return Api.getOperationLogList({
			pageNumber: params.pageNumber as number,
			pageSize: params.pageSize as number,
			title: params.title ? String(params.title) : undefined,
			operName: params.operName ? String(params.operName) : undefined,
			status: params.status as number | undefined,
			startTime: toDateTime(params.startTime),
			endTime: toDateTime(params.endTime),
		});
	}, []);

	const columns: ProTableColumnType<IApi.OperationLogItem>[] = [
		{ title: "模块标题", dataIndex: "title", width: 140, ellipsis: true, render: (v) => v || "-" },
		{
			title: "请求方式",
			dataIndex: "requestMethod",
			width: 110,
			render: (value) => <Typography.Text code>{value as string}</Typography.Text>,
		},
		{ title: "操作人员", dataIndex: "operName", width: 120, render: (v) => v || "-" },
		{ title: "请求 URL", dataIndex: "operUrl", width: 240, ellipsis: true, render: (v) => v || "-" },
		{
			title: "操作 IP",
			dataIndex: "operIp",
			width: 130,
			render: (value) => <Typography.Text code>{value as string}</Typography.Text>,
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
			title: "异常信息",
			dataIndex: "errorMsg",
			width: 220,
			ellipsis: true,
			render: (v) => (v ? <Typography.Text type="danger">{v as string}</Typography.Text> : "-"),
		},
		{ title: "操作时间", dataIndex: "operTime", width: 180 },
	];

	return (
		<ProTable<IApi.OperationLogItem>
			rowKey="id"
			api={operationLogApi}
			columns={columns}
			title="操作日志"
			search={{
				fields: [
					{ type: "input", label: "模块标题", name: "title" },
					{ type: "input", label: "操作人员", name: "operName" },
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
						fieldProps: { showTime: true, style: { width: "100%" }, placeholder: "开始时间" },
					},
					{
						type: "date-picker",
						label: "结束时间",
						name: "endTime",
						fieldProps: { showTime: true, style: { width: "100%" }, placeholder: "结束时间" },
					},
				],
			}}
		/>
	);
};

export default OperationLogPage;
