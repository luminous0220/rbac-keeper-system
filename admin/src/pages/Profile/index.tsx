import { useCallback, useEffect, useState } from "react";
import { App, Card, Form, Tabs } from "antd";
import { useNavigate, useSearchParams } from "react-router";
import { ProForm } from "@/components/ProForm";
import type { ProFormFields } from "@/components/ProForm";
import { Api } from "@/apis";
import { useAuthStore } from "@/stores";
import { clearAuthToken } from "@/libs";
import { CFG } from "@/constants";
import { useTranslation } from "react-i18next";

/**
 * 个人信息页：个人信息 / 修改密码 两个 Tab。
 * 支持 ?tab=password 直达修改密码（TopHeader 下拉进入）。
 * 个人信息保存成功 → 同步 auth store（TopHeader 昵称/头像即时更新）。
 * 修改密码成功后后端强制旧令牌全部失效 → 清缓存并回登录页。
 */
const ProfilePage: React.FC = () => {
	const { message } = App.useApp();
	const { t } = useTranslation();
	const navigate = useNavigate();
	const [searchParams, setSearchParams] = useSearchParams();
	// 初始 Tab 由 URL 决定（hash 路由下 useSearchParams 同样生效）
	const [activeTab, setActiveTab] = useState(
		searchParams.get("tab") === "password" ? "password" : "info",
	);
	const setUserInfo = useAuthStore((s) => s.setUserInfo);
	const resetAuth = useAuthStore((s) => s.resetAuth);
	const [infoForm] = Form.useForm();
	const [pwdForm] = Form.useForm();

	// 切换 Tab 时同步 URL，保持 /profile?tab=password 可直达
	const handleTabChange = useCallback(
		(key: string) => {
			setActiveTab(key);
			if (key === "password") {
				setSearchParams({ tab: "password" });
			} else {
				setSearchParams({});
			}
		},
		[setSearchParams],
	);

	// 挂载时刷新本人信息并回填表单（TopHeader 昵称等也可能已被后台更新过）
	const loadProfile = useCallback(async () => {
		try {
			const info = await Api.getUserInfo();
			setUserInfo(info);
			infoForm.setFieldsValue({
				username: info.username,
				nickname: info.nickname,
				phone: info.phone,
				email: info.email,
				sex: info.sex,
				avatar: info.avatar,
			});
		} catch (err) {
			console.error(t("profile.loadFailed"), err);
		}
	}, [infoForm, setUserInfo, t]);

	useEffect(() => {
		loadProfile();
	}, [loadProfile]);

	// 个人信息字段：username 只读，其余可编辑
	const sexOptions = [
		{ label: t("profile.unknown"), value: 0 },
		{ label: t("profile.male"), value: 1 },
		{ label: t("profile.female"), value: 2 },
	];
	const infoFields: ProFormFields = [
		{
			type: "input",
			label: t("profile.username"),
			name: "username",
			disabled: true,
			fieldProps: {
				disabled: true,
				placeholder: t("profile.usernameImmutable"),
			},
		},
		{
			type: "input",
			label: t("profile.nickname"),
			name: "nickname",
			rules: [{ required: true, message: t("profile.nicknameRequired") }],
		},
		{
			type: "input",
			label: t("profile.phone"),
			name: "phone",
			fieldProps: { maxLength: 11 },
			rules: [{ pattern: /^1[3-9]\d{9}$/, message: t("profile.phoneInvalid") }],
		},
		{
			type: "input",
			label: t("profile.email"),
			name: "email",
			rules: [{ type: "email", message: t("profile.emailInvalid") }],
		},
		{
			type: "radio",
			label: t("profile.sex"),
			name: "sex",
			fieldProps: { options: sexOptions },
		},
		{
			type: "input",
			label: t("profile.avatar"),
			name: "avatar",
			fieldProps: { placeholder: t("profile.avatarPlaceholder") },
		},
	];

	// 保存个人信息：成功后用后端返回的最新信息覆盖 store → TopHeader 即时同步
	const handleInfoSubmit = useCallback(
		async (_title: string, values: Record<string, any>) => {
			const res = await Api.updateProfile({
				nickname: values.nickname,
				phone: values.phone ?? "",
				email: values.email ?? "",
				sex: values.sex ?? 0,
				avatar: values.avatar ?? "",
			});
			setUserInfo(res);
			message.success(t("profile.saved"));
		},
		[message, setUserInfo, t],
	);

	// 修改密码字段：新密码与后端约束对齐（8-32 位字母数字），二次确认仅前端
	const pwdFields: ProFormFields = [
		{
			type: "input-password",
			label: t("profile.oldPassword"),
			name: "oldPassword",
			rules: [{ required: true, message: t("profile.oldPasswordRequired") }],
		},
		{
			type: "input-password",
			label: t("profile.newPassword"),
			name: "newPassword",
			fieldProps: { placeholder: t("profile.passwordPlaceholder") },
			rules: [
				{ required: true, message: t("profile.newPasswordRequired") },
				{ min: 8, max: 32, message: t("profile.passwordLength") },
				{
					pattern: /^[a-zA-Z0-9]+$/,
					message: t("profile.passwordChars"),
				},
			],
		},
		{
			type: "input-password",
			label: t("profile.confirmPassword"),
			name: "confirmPassword",
			fieldProps: { placeholder: t("profile.confirmPasswordPlaceholder") },
			// 依赖新密码字段，输入变化时自动重新校验一致性
			formItemProps: { dependencies: ["newPassword"] },
			rules: [
				{ required: true, message: t("profile.confirmPasswordRequired") },
				{
					validator: (_rule: any, value: string) =>
						!value || pwdForm.getFieldValue("newPassword") === value
							? Promise.resolve()
							: Promise.reject(new Error(t("profile.passwordMismatch"))),
				},
			],
		},
	];

	// 提交修改密码：成功即旧令牌全部失效，清缓存回登录页
	const handlePwdSubmit = useCallback(
		async (_title: string, values: Record<string, any>) => {
			await Api.resetPassword({
				oldPassword: values.oldPassword,
				newPassword: values.newPassword,
			});
			message.success(t("profile.passwordUpdated"));
			clearAuthToken();
			resetAuth();
			navigate(CFG.LOGIN_PATH, { replace: true });
		},
		[message, navigate, resetAuth, t],
	);

	return (
		<Card className="h-full">
			<Tabs
				activeKey={activeTab}
				onChange={handleTabChange}
				items={[
					{
						key: "info",
						label: t("profile.info"),
						children: (
							<div className="max-w-xl">
								<ProForm
									form={infoForm}
									type="pure"
									fields={infoFields}
									defaultButtons={{
										confirmText: t("common.save"),
										hideReset: true,
									}}
									onConfirm={handleInfoSubmit}
								/>
							</div>
						),
					},
					{
						key: "password",
						label: t("profile.changePassword"),
						children: (
							<div className="max-w-xl">
								<ProForm
									form={pwdForm}
									type="pure"
									fields={pwdFields}
									defaultButtons={{
										confirmText: t("profile.confirmUpdate"),
										hideReset: true,
									}}
									onConfirm={handlePwdSubmit}
								/>
							</div>
						),
					},
				]}
			/>
		</Card>
	);
};

export default ProfilePage;
