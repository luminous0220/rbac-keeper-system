import dayjs, { type Dayjs } from "dayjs";
import type { CalendarReminder } from "./types.d";

export const COLOR_OPTIONS = [
	"#ff4d4f",
	"#fa8c16",
	"#faad14",
	"#52c41a",
	"#13c2c2",
	"#1677ff",
	"#722ed1",
	"#eb2f96",
];

export const DEFAULT_REMINDER_COLOR = "#1677ff";

export const WEEKDAYS = [0, 1, 2, 3, 4, 5, 6];

export const buildCalendarDays = (month: Dayjs): Dayjs[] => {
	const startOfMonth = month.startOf("month");
	const startDay = startOfMonth.day();
	const daysInMonth = month.daysInMonth();
	const days: Dayjs[] = [];

	for (let i = startDay - 1; i >= 0; i--) {
		days.push(startOfMonth.subtract(i + 1, "day"));
	}
	for (let i = 0; i < daysInMonth; i++) {
		days.push(startOfMonth.add(i, "day"));
	}
	const remaining = 42 - days.length;
	for (let i = 0; i < remaining; i++) {
		days.push(startOfMonth.add(daysInMonth + i, "day"));
	}

	return days;
};

export const isToday = (date: Dayjs): boolean =>
	date.format("YYYY-MM-DD") === dayjs().format("YYYY-MM-DD");

export const isOtherMonth = (date: Dayjs, currentMonth: Dayjs): boolean =>
	date.month() !== currentMonth.month() || date.year() !== currentMonth.year();

export const isOverdue = (reminder: CalendarReminder): boolean =>
	!reminder.completed && dayjs(reminder.dateTime).isBefore(dayjs());
