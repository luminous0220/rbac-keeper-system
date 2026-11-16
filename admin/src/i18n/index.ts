import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import enUS from "./locales/en-US";
import zhCN from "./locales/zh-CN";

const resources = {
	"zh-CN": {
		translation: zhCN,
	},
	"en-US": {
		translation: enUS,
	},
} as const;

const savedLanguage = localStorage.getItem("admin-language");
const initialLanguage = savedLanguage === "en-US" ? "en-US" : "zh-CN";
document.documentElement.lang = initialLanguage;

void i18n.use(initReactI18next).init({
	resources,
	lng: initialLanguage,
	fallbackLng: "zh-CN",
	interpolation: {
		escapeValue: false,
	},
});

i18n.on("languageChanged", (language) => {
	localStorage.setItem("admin-language", language);
	document.documentElement.lang = language;
});

export default i18n;
