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
import { useTranslation } from "react-i18next";

// 操作状态：0正常 1异常（dto.OperationLogItem.Status）
const STATUS_TEXT: Record<number, string> = {
	0: "logs.normal",
	1: "logs.abnormal",
};
const STATUS_OPTIONS = [
	{ label: "logs.normal", value: 0 },
	{ label: "logs.abnormal", value: 1 },
];

/**
 * 操作日志：只读分页查询 + 导出 / 清空。
 * - 数据源 GET /api/log/operation（RBAC monitor:log:operation）
 * - 搜索 title（模糊）/operName（模糊）/status/startTime/endTime
 * - 导出 GET /api/log/operation/export（monitor:log:operation:export，携带当前筛选）
 * - 清空 POST /api/log/operation/clear（monitor:log:operation:clear，物理删除全部）
 * - 登出等后台动作也会以操作日志「退出登录」形式写入，此处一并展示
 */
const OperationLogPage: React.FC = () => {
	const { message, modal } = App.useApp();
	const { t } = useTranslation();
	const { hasPerm } = usePerms();
	const tableRef = useRef<ProTableRef<IApi.OperationLogItem>>(null);
	// 最近一次已查询的筛选参数（导出复用：导出的即当前表格所查）
	const lastQueryRef = useRef<Record<string, unknown>>({});

	const operationLogApi = useCallback(async (params: Record<string, unknown>) => {
		const query: Record<string, unknown> = {
			title: params.title ? String(params.title) : undefined,
			operName: params.operName ? String(params.operName) : undefined,
			status: params.status as number | undefined,
			startTime: toDateTime(params.startTime),
			endTime: toDateTime(params.endTime),
		};
		lastQueryRef.current = query;
		return Api.getOperationLogList({
			pageNumber: params.pageNumber as number,
			pageSize: params.pageSize as number,
			...query,
		});
	}, []);

	// 导出当前筛选结果（错误提示由请求层统一弹出，无需在此重复）
	const handleExport = async () => {
		try {
			const blob = await Api.exportOperationLogs({ ...lastQueryRef.current });
			downloadBlob(blob, "operation_logs.csv");
			message.success(t("logs.exportSuccess"));
		} catch {
			/* noop */
		}
	};

	const confirmClear = () => {
		modal.confirm({
			title: t("logs.clearOperation"),
			content: t("logs.clearOperationConfirm"),
			okText: t("logs.clear"),
			okButtonProps: { danger: true },
			onOk: async () => {
				try {
					await Api.clearOperationLogs();
					message.success(t("logs.operationCleared"));
					tableRef.current?.refresh();
				} catch {
					/* noop */
				}
			},
		});
	};

	const columns: ProTableColumnType<IApi.OperationLogItem>[] = [
		{ title: t("logs.moduleTitle"), dataIndex: "title", width: 140, ellipsis: true, render: (v) => v || "-" },
		{
			title: t("logs.requestMethod"),
			dataIndex: "requestMethod",
			width: 110,
			render: (value) => <Typography.Text code>{value as string}</Typography.Text>,
		},
		{ title: t("logs.operator"), dataIndex: "operName", width: 120, render: (v) => v || "-" },
		{ title: t("logs.requestUrl"), dataIndex: "operUrl", width: 240, ellipsis: true, render: (v) => v || "-" },
		{
			title: t("logs.operationIp"),
			dataIndex: "operIp",
			width: 130,
			render: (value) => <Typography.Text code>{value as string}</Typography.Text>,
		},
		{
			title: t("logs.status"),
			dataIndex: "status",
			width: 80,
			render: (value) => (
				<Tag color={(value as number) === 0 ? "green" : "red"}>
					{t(STATUS_TEXT[(value as number) ?? 0] ?? "-")}
				</Tag>
			),
		},
		{
			title: t("logs.errorMessage"),
			dataIndex: "errorMsg",
			width: 220,
			ellipsis: true,
			render: (v) => (v ? <Typography.Text type="danger">{v as string}</Typography.Text> : "-"),
		},
		{
			title: t("logs.operationTime"),
			dataIndex: "operTime",
			width: 180,
			render: (time) => dayjs(time).format("YYYY-MM-DD HH:mm:ss"),
		},
	];

	return (
		<ProTable<IApi.OperationLogItem>
			ref={tableRef}
			rowKey="id"
			api={operationLogApi}
			columns={columns}
			title={t("logs.operationTitle")}
			toolbarExtra={
				<>
					{hasPerm("monitor:log:operation:export") && (
						<Button icon={<DownloadOutlined />} onClick={handleExport}>
							{t("table.export")}
						</Button>
					)}
					{hasPerm("monitor:log:operation:clear") && (
						<Button danger icon={<DeleteOutlined />} onClick={confirmClear}>
							{t("logs.clear")}
						</Button>
					)}
				</>
			}
			search={{
				fields: [
					{ type: "input", label: t("logs.moduleTitle"), name: "title" },
					{ type: "input", label: t("logs.operator"), name: "operName" },
					{
						type: "select",
						label: t("logs.status"),
						name: "status",
						fieldProps: {
							options: STATUS_OPTIONS.map((item) => ({
								...item,
								label: t(item.label),
							})),
							allowClear: true,
						},
					},
					{
						type: "date-picker",
						label: t("logs.startTime"),
						name: "startTime",
						fieldProps: { showTime: true, style: { width: "100%" }, placeholder: t("logs.startTime") },
					},
					{
						type: "date-picker",
						label: t("logs.endTime"),
						name: "endTime",
						fieldProps: { showTime: true, style: { width: "100%" }, placeholder: t("logs.endTime") },
					},
				],
			}}
		/>
	);
};

export default OperationLogPage;
