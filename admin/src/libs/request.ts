import axios, {
	type AxiosInstance,
	type AxiosRequestConfig,
	AxiosError,
	type AxiosResponse,
	type InternalAxiosRequestConfig,
} from "axios";

import { getToken, getRefreshToken, setToken, clearAuthToken } from "@/libs";
import { useAuthStore } from "@/stores";

import { CFG, HTTP_STATUS, BUSINESS_CODE } from "@/constants";

export interface HttpRequestConfig extends InternalAxiosRequestConfig {
	noToken?: boolean;
}

interface IConfig extends AxiosRequestConfig {
	noToken?: boolean;
	noErrorMessage?: boolean;
}

interface IReqParams {
	url: string;
	params?: any;
	config?: IConfig;
}

// * 基础配置
const DEFAULT_CONFIG = {
	// baseURL: import.meta.env.VITE_BASE_URL as string,
	timeout: CFG.AXIOS_TIME_OUT as number,
	withCredential: true,
	headers: {
		Accept: "application/json, text/plain, */*",
		"Content-Type": "application/json",
		"X-Requested-With": "XMLHttpRequest",
	},
};

// ---- 登录失效统一处理：清空凭证 + 重置用户态 + 回登录页 ----
function handleAuthFailed(msg?: string) {
	clearAuthToken();
	useAuthStore.getState().resetAuth();
	if (msg) window.$message.error(msg);
	window.location.hash = CFG.LOGIN_PATH;
}

// ---- 刷新令牌并发控制：多个 401 并发时只发起一次刷新，其余请求挂起等待 ----
let isRefreshing = false;
/** 刷新期间等待重放的请求队列 */
let refreshQueue: Array<{ resolve: (token: string | null) => void }> = [];

/**
 * 尝试用 Refresh Token 静默续期 Access Token（单例）。
 * 成功返回新 token；无 refresh token 或刷新失败返回 null。
 * 刷新请求走 noToken：避免自身触发续期导致死循环。
 */
function tryRefreshToken(): Promise<string | null> {
	const refreshToken = getRefreshToken();
	if (!refreshToken) return Promise.resolve(null);

	// 已有刷新进行中：挂起等待，由发起者刷新完成后统一唤醒
	if (isRefreshing) {
		return new Promise((resolve) => {
			refreshQueue.push({ resolve });
		});
	}

	isRefreshing = true;
	return http
		.post<{ accessToken: string }>({
			url: "/api/auth/refresh",
			params: { refreshToken },
			config: { noToken: true, noErrorMessage: true },
		})
		.then((res) => {
			setToken(res.accessToken);
			return res.accessToken;
		})
		.catch(() => null)
		.finally(() => {
			isRefreshing = false;
			// 唤醒所有等待请求：刷新成功拿到新 token，失败为 null 走登出
			const token = getToken() ?? null;
			refreshQueue.forEach((item) => item.resolve(token));
			refreshQueue = [];
		});
}

// * 创建axios实例的类
class RequestHttp {
	// * Axios实例
	private requestServer: AxiosInstance;
	constructor(config: AxiosRequestConfig) {
		this.requestServer = axios.create(config);

		// 请求拦截器
		this.requestServer.interceptors.request.use(
			async (confi: HttpRequestConfig) => {
				const { headers, noToken } = confi;
				// 如果是白名单里的请求url，则不需要 token
				if (noToken) {
					return confi;
				}

				const token = getToken();
				headers.setAuthorization("Bearer " + token);

				return confi;
			},
			(error: AxiosError) => {
				return Promise.reject(error);
			},
		);

		// 响应拦截器
		this.requestServer.interceptors.response.use(
			async (interRes: AxiosResponse) => {
				const config = interRes.config as IConfig;

				// blob 下载（如日志 CSV 导出）：
				// 成功时后端返回真文件字节（text/csv…），直接给调用方 Blob；
				// 出错时后端仍以 HTTP 200 + JSON {code,msg,data} 信封返回，axios 在 responseType=blob 下会把它
				// 变成“装着 JSON 的 Blob”，此处解析还原成信封，交给下方统一错误/401 续期逻辑处理。
				if (interRes.config.responseType === "blob" && interRes.data instanceof Blob) {
					const blob = interRes.data as Blob;
					const looksJson =
						blob.type.includes("application/json") || blob.type.includes("text/plain");
					if (looksJson) {
						try {
							const parsed = JSON.parse(await blob.text());
							if (parsed && typeof parsed.code === "number") {
								interRes.data = parsed; // 还原信封，走统一逻辑
							}
						} catch {
							// 非 JSON 文本（带 BOM 的 CSV 等）→ 视为文件，保持 Blob
						}
					}
					if (interRes.data instanceof Blob) {
						return interRes.data; // 真文件直接给调用方
					}
				}

				// 后端统一响应：{ code, msg, data }，code === 0 为成功（server/response/response.go）
				const { code, msg, data } = interRes.data || {};

				// 成功：解包 data
				if (code === BUSINESS_CODE.SUCCESS) return data;

				// 登录失效（40101/40102/40103）：已认证请求先尝试静默续期，
				// 刷新成功则用新 token 重放原请求；刷新失败才清空凭证回登录页。
				// 公开接口（如登录失败）同样返回 40101，只提示不跳转。
				if (
					!config.noToken &&
					(code === BUSINESS_CODE.UNAUTHORIZED ||
						code === BUSINESS_CODE.TOKEN_EXPIRED ||
						code === BUSINESS_CODE.TOKEN_INVALID)
				) {
					// 已用新 token 重试过仍失败：不再续期，直接登出
					if ((interRes.config as IConfig & { _retried?: boolean })._retried) {
						handleAuthFailed(msg || "登录失效，请重新登录");
						return Promise.reject(data);
					}

					const newToken = await tryRefreshToken();
					if (newToken) {
						// 用新 token 重放原请求（标记 _retried，避免循环续期）
						const retryConfig = {
							...interRes.config,
							_retried: true,
						} as InternalAxiosRequestConfig;
						retryConfig.headers.set("Authorization", `Bearer ${newToken}`);
						return this.requestServer(retryConfig);
					}
					// 续期失败（refresh token 缺失/过期）：清凭证并回登录页
					handleAuthFailed(msg || "登录已过期，请重新登录");
					return Promise.reject(data);
				}

				// 其他业务错误：默认弹出后端 msg，noErrorMessage 可静默
				if (!config.noErrorMessage) {
					window.$message.error(msg || "请求失败");
				}
				console.error(interRes);
				return Promise.reject(data);
			},
			async (error: AxiosError) => {
				const { response, status } = error;
				console.error(response);

				// 网络异常
				if (!window.navigator.onLine) {
					window.$message.error("网络请求超时，请您重试");
					return Promise.reject(response);
				}

				// HTTP 401：未认证/登录过期（少见，业务错误走 body code）
				if (status === HTTP_STATUS.UNAUTHORIZED) {
					handleAuthFailed("登录失效，请重新登录");
					return Promise.reject(response);
				}

				window.$message.error("请求错误，请联系系统管理员");
				return Promise.reject(error);
			},
		);
	}

	get<T>(p: IReqParams): Promise<T> {
		const { url, params, config } = p;
		// 响应拦截器已将 {code,msg,data} 解包为 data，故断言返回值即 T（仅类型层面，无运行时改动）
		return this.requestServer.get(url, { params, ...config }) as unknown as Promise<T>;
	}
	post<T>(p: IReqParams): Promise<T> {
		const { url, params, config } = p;
		return this.requestServer.post(url, params, config) as unknown as Promise<T>;
	}
	put<T>(p: IReqParams): Promise<T> {
		const { url, params, config } = p;
		return this.requestServer.put(url, params, config) as unknown as Promise<T>;
	}
	delete<T>(p: IReqParams): Promise<T> {
		const { url, params, config } = p;
		return this.requestServer.delete(url, { params, ...config }) as unknown as Promise<T>;
	}
}

export const http = new RequestHttp(DEFAULT_CONFIG);
