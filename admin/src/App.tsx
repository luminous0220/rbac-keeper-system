// src/App.tsx
import { ConfigProvider, App as Message, theme } from "antd";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import zhCN from "antd/locale/zh_CN";
import enUS from "antd/locale/en_US";
import dayjs from "dayjs";
import "dayjs/locale/zh-cn";
import "dayjs/locale/en";
import { useTranslation } from "react-i18next";
import { MessageHolder, FullLoadingHolder } from "./hooks";
import { useThemeStore } from "./stores";
import { AppRouter } from "./routers";

const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			staleTime: 1000 * 60 * 5,
		},
	},
});

const App = () => {
	const { isDark, colorPrimary } = useThemeStore();
	const { i18n } = useTranslation();
	const isEnglish = i18n.resolvedLanguage === "en-US";
	dayjs.locale(isEnglish ? "en" : "zh-cn");

	return (
		<div className="size-full">
			<ConfigProvider
				locale={isEnglish ? enUS : zhCN}
				theme={{
					algorithm: isDark ? theme.darkAlgorithm : theme.defaultAlgorithm,
					token: {
						colorPrimary,
						colorBgBase: isDark ? "#0A0A0A" : "#fff",
						colorTextBase: isDark ? "#fff" : "#030229",
						colorBgContainer: isDark ? "#161618" : "#FFFFFF",
						colorBgLayout: isDark ? "#000000" : "#F8FAFE",
						borderRadius: 10,
					},
					components: {
						Modal: {
							colorBgElevated: isDark ? "#0A0A0A" : "#fff",
						},
					},
				}}
			>
				<QueryClientProvider client={queryClient}>
					<Message
						className="size-full"
						message={{ duration: 3 }}
						notification={{ duration: 3 }}
					>
						<MessageHolder />
						<AppRouter />
					</Message>
				</QueryClientProvider>
			</ConfigProvider>
			<FullLoadingHolder />
		</div>
	);
};

export default App;
