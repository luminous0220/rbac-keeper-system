/**
 * 把 antd 表单/搜索控件产生的 Dayjs 值转为后端时间字符串（YYYY-MM-DD HH:mm:ss）。
 * 后端日志查询 startTime/endTime 使用 time_format:"2006-01-02 15:04:05"。
 * 非 Dayjs（undefined/空字符串等）一律返回 undefined，不参与筛选。
 */
export function toDateTime(v: unknown): string | undefined {
	if (v && typeof v === "object" && typeof (v as { format?: unknown }).format === "function") {
		const fmt = (v as { format: (s: string) => string }).format("YYYY-MM-DD HH:mm:ss");
		return fmt;
	}
	return undefined;
}
