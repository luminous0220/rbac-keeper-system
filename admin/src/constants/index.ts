export const CFG = {
	TOKEN_KEY: "ADMIN_TOKEN",
	// Refresh Token 存储键（独立于 Access Token，失效后无法续期需重新登录）
	REFRESH_TOKEN_KEY: "ADMIN_REFRESH_TOKEN",
	HOME_PATH: "/dash-board",
	LOGIN_PATH: "/login",
	REMEMBER_KEY: "REMEMBER_LOGIN_INFO",
	SYSTEM_NAME: "rbackeeper",
	AXIOS_TIME_OUT: 30 * 1000,
};

export * from "./http.enum";
