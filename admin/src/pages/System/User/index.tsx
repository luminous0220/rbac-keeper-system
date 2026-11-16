import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
	App,
	Button,
	Card,
	Empty,
	Input,
	Space,
	Tag,
	Tooltip,
	Tree,
} from "antd";
import { CopyableCell, ProTable } from "@/components/ProTable";
import type {
	ProTableRef,
	ProTableColumnType,
	ProTableProps,
} from "@/components/ProTable";
import { ProForm } from "@/components/ProForm";
import type { ProFormFields, ProFormInstance } from "@/components/ProForm";
import { Api, IApi } from "@/apis";
import { usePerms } from "@/hooks";
import dayjs from "dayjs";
import { EditOutlined, ReloadOutlined } from "@ant-design/icons";
import { filterTree } from "@/libs";
import { IconSearch } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";
import i18n from "@/i18n";

// 性别展示：UserItem.sex 按 model 语义 1男 2女（0 为新增接口遗留的未知态）
const SEX_TEXT: Record<number, string> = {
	0: "profile.unknown",
	1: "profile.male",
	2: "profile.female",
};
// 账号状态：model 语义 0禁用 1正常
const STATUS_TEXT: Record<number, string> = {
	0: "common.disabled",
	1: "common.enabled",
};
const STATUS_OPTIONS = [
	{ label: "common.enabled", value: 1 },
	{ label: "common.disabled", value: 0 },
];
// 性别选项：新增/编辑/列表统一按 model 语义 0未知 1男 2女（后端已对齐为 omitempty,oneof=0 1 2）
const SEX_OPTIONS = [
	{ label: "profile.unknown", value: 0 },
	{ label: "profile.male", value: 1 },
	{ label: "profile.female", value: 2 },
];

const presets = [
	"blue",
	"green",
	"volcano",
	"gold",
	"purple",
	"orange",
	"cyan",
	"lime",
	"magenta",
	"red",
	"geekblue",
];

/** 部门树 → antd TreeSelect 数据结构 */
function toTreeSelect(nodes: IApi.DeptRaw[]): OptionItem[] {
	return nodes.map((d) => ({
		value: d.id,
		title: d.dept_name,
		children: d.children?.length ? toTreeSelect(d.children) : undefined,
	}));
}

/** 部门树 → antd Tree 数据结构（左侧部门筛选） */

function toTreeData(nodes: IApi.DeptRaw[]): TreeNodeItem[] {
	return nodes.map((d) => ({
		title: d.dept_name,
		key: d.id,
		children: d.children?.length ? toTreeData(d.children) : undefined,
	}));
}

/**
 * 用户管理：列表 / 新增 / 编辑。
 * - 列表行 key 用 username（UserItem 脱敏无 id）
 * - 新增走 /api/user/add；编辑走 /api/user/edit?username=
 * - 编辑仅支持：密码（留空不修改）/ 部门 / 角色 / 状态（后端 UserEditBody 约束）
 * - super 受后端保护不可编辑
 */
const UserPage: React.FC = () => {
	const { message } = App.useApp();
	const { t } = useTranslation();
	const { hasPerm } = usePerms();
	const tableRef = useRef<ProTableRef<IApi.UserItem>>(null);

	const [deptTree, setDeptTree] = useState<IApi.DeptRaw[]>([]);
	const [deptLoading, setDeptLoading] = useState(true);
	const [roleOptions, setRoleOptions] = useState<
		{ label: string; value: string }[]
	>([]);
	// 左侧部门树选中状态：selectedKeys 控展开高亮，selectedDeptId 决定列表的部门过滤
	const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([0]); // 默认选中"全部部门"
	const [selectedDeptId, setSelectedDeptId] = useState<number | undefined>();

	const fetchDeptTree = useCallback(() => {
		setDeptLoading(true); // 点击刷新时也应显示 loading
		Api.getDeptTree()
			.then(setDeptTree)
			.catch((err) =>
				console.error(i18n.t("system.departmentTreeLoadFailed"), err),
			)
			.finally(() => setDeptLoading(false));
	}, []);

	// 挂载时预取部门树与角色列表（表单下拉选项 + 左侧筛选树）
	useEffect(() => {
		fetchDeptTree();
		Api.getRoleList()
			.then((list) => {
				setRoleOptions(
					list.map((r) => ({ label: r.role_name, value: r.role_key })),
				);
			})
			.catch((err) => console.error(i18n.t("system.roleListLoadFailed"), err));
	}, [fetchDeptTree]);

	const deptTreeData = useMemo(() => toTreeSelect(deptTree), [deptTree]);

	// 左侧部门树数据：顶部固定一个"全部部门"根节点，选中即清空部门筛选
	const [searchValue, setSearchValue] = useState("");
	const deptFilterTree = useMemo<TreeNodeItem[]>(() => {
		const originalTree = [
			{ title: t("user.allDepartments"), key: 0, children: toTreeData(deptTree) },
		];
		return filterTree(originalTree, searchValue);
	}, [deptTree, searchValue, t]);

	// 部门树加载完成后默认全部展开（antd defaultExpandAll 只在首次挂载生效，异步数据需受控展开）
	const allDeptKeys = useMemo(() => {
		const keys: number[] = [];
		const walk = (nodes: IApi.DeptRaw[]) => {
			nodes.forEach((n) => {
				keys.push(n.id);
				if (n.children?.length) walk(n.children);
			});
		};
		walk(deptTree);
		return keys;
	}, [deptTree]);
	const [expandedKeys, setExpandedKeys] = useState<React.Key[]>([]);
	useEffect(() => {
		if (allDeptKeys.length) setExpandedKeys([0, ...allDeptKeys]);
	}, [allDeptKeys]);

	// 选中部门 → 注入 deptId 到表格查询（ProTable.externalParams 变化时自动回第 1 页刷新）
	const handleDeptSelect = useCallback((keys: React.Key[]) => {
		setSelectedKeys(keys);
		// key 为 0（全部部门）或点击已选中节点取消选择（keys 为空）→ 均清空部门筛选
		setSelectedDeptId((keys[0] as number) || undefined);
	}, []);
	const externalParams = useMemo(
		() => (selectedDeptId ? { deptId: selectedDeptId } : {}),
		[selectedDeptId],
	);

	const deptFlatenOptions = useMemo(() => {
		const options: OptionItem[] = [];
		const flatenDeptTreeData = (list: OptionItem[]) => {
			list.forEach((item) => {
				if (item.children && item.children.length > 0) {
					flatenDeptTreeData(item.children);
				}
				options.push({
					text: item.title,
					value: item.value,
				});
			});
		};

		flatenDeptTreeData(deptTreeData);
		return options;
	}, [deptTreeData]);

	// 1. 新增一个统一的 formRef
	const formRef = useRef<ProFormInstance>(null);

	// 2. 新增状态，用于区分是新增还是编辑，并存储当前编辑的用户数据
	const [formMode, setFormMode] = useState<{
		type: "add" | "edit";
		record?: IApi.UserItem;
	}>({ type: "add" });

	// 3. 动态生成表单字段的函数
	const getFormFields = useCallback((): ProFormFields => {
		const isEdit = formMode.type === "edit";

		const fields: ProFormFields = [
			// --- 公共字段 ---
			{
				type: "input",
				label: t("user.nickname"),
				name: "nickname",
				fieldProps: { maxLength: 30, placeholder: t("user.nicknameRequired") },
				rules: [{ required: true, message: t("user.nicknameRequired") }],
			},
			{
				type: "input",
				label: t("user.phone"),
				name: "phone",
				fieldProps: { maxLength: 11, placeholder: t("user.phonePlaceholder") },
				rules: [{ pattern: /^1[3-9]\d{9}$/, message: t("user.phoneInvalid") }],
			},
			{
				type: "input",
				label: t("user.email"),
				name: "email",
				fieldProps: { maxLength: 50, placeholder: t("user.emailInvalid") },
				rules: [{ type: "email", message: t("user.emailInvalid") }],
			},
			{
				type: "radio",
				label: t("user.sex"),
				name: "sex",
				fieldProps: {
					options: SEX_OPTIONS.map((item) => ({
						...item,
						label: t(item.label),
					})),
				},
			},
			{
				type: "tree-select",
				label: t("user.department"),
				name: "deptId",
				fieldProps: {
					treeDefaultExpandAll: true,
					treeData: deptTreeData,
					placeholder: t("user.selectDepartment"),
					showSearch: true,
				},
				rules: [{ required: true, message: t("user.selectDepartment") }],
			},
			{
				type: "select",
				label: t("user.role"),
				name: "roleKeys",
				fieldProps: {
					options: roleOptions,
					mode: "multiple",
					placeholder: t("user.selectRoles"),
				},
				rules: [{ required: true, message: t("user.roleRequired") }],
			},
			{
				type: "textarea",
				label: t("user.remark"),
				name: "remark",
				formItemProps: { className: "col-span-2" },
				fieldProps: {
					maxLength: 500,
					rows: 3,
					placeholder: t("user.remarkPlaceholder"),
				},
			},
		];

		// --- 根据模式插入特定字段 ---
		if (!isEdit) {
			// 新增模式：插入用户名和密码字段
			fields.unshift(
				{
					type: "input",
					label: t("user.name"),
					name: "username",
					fieldProps: { maxLength: 30, placeholder: t("user.usernamePlaceholder") },
					rules: [
						{ required: true, message: t("user.usernameRequired") },
						{ min: 6, max: 30, message: t("user.usernameLength") },
						{ pattern: /^[a-zA-Z0-9]+$/, message: t("user.usernameChars") },
					],
				},
				{
					type: "input-password",
					label: t("user.password"),
					name: "password",
					fieldProps: { placeholder: t("user.passwordPlaceholder") },
					rules: [
						{ required: true, message: t("user.passwordRequired") },
						{ min: 8, max: 32, message: t("user.passwordLength") },
						{
							pattern: /^(?=.*[A-Za-z])(?=.*\d).{8,32}$/,
							message: t("user.passwordRequirements"),
						},
					],
				},
			);
		} else {
			// 编辑模式：插入密码（非必填）和状态字段
			fields.unshift(
				{
					type: "input-password",
					label: t("user.password"),
					name: "password",
					fieldProps: { placeholder: t("user.passwordOptional") },
					rules: [
						{
							validator: (_rule: any, value: string) => {
								if (!value) return Promise.resolve();
								if (value.length < 8 || value.length > 32) {
									return Promise.reject(new Error(t("user.passwordLength")));
								}
								if (!/(?=.*[A-Za-z])(?=.*\d)/.test(value)) {
									return Promise.reject(new Error(t("user.passwordRequirements")));
								}
								return Promise.resolve();
							},
						},
					],
				},
				{
					type: "radio",
					label: t("user.status"),
					name: "status",
					fieldProps: {
						options: STATUS_OPTIONS.map((item) => ({
							...item,
							label: t(item.label),
						})),
					},
					rules: [{ required: true, message: t("department.statusRequired") }],
				},
			);
		}

		return fields;
	}, [formMode.type, deptTreeData, roleOptions, t]);

	// 新增表单
	const handleOpenAdd = useCallback(() => {
		setFormMode({ type: "add" });
		formRef.current?.open({
			title: t("user.add"),
			initialValues: {
				sex: 1, // 默认男
				...(selectedDeptId ? { deptId: selectedDeptId } : {}),
			},
		});
	}, [selectedDeptId, t]);

	// 编辑表单
	const openEdit = useCallback((record: IApi.UserItem) => {
		setFormMode({ type: "edit", record });
		formRef.current?.open({
			title: `${t("user.edit")}: ${record.username}`,
			initialValues: {
				password: "",
				deptId: record.dept?.id,
				roleKeys: record.roles?.map((r) => r.role_key) ?? [],
				status: record.status,
				nickname: record.nickname,
				phone: record.phone,
				email: record.email,
				sex: record.sex,
			},
			extraParams: { username: record.username }, // 将 username 存入 extraParams 以便提交时使用
		});
	}, [t]);

	// 提交处理函数
	const handleFormSubmit = useCallback(
		async (
			_title: string,
			values: Record<string, any>,
			extraParams?: Record<string, any>,
		) => {
			if (formMode.type === "add") {
				// --- 新增逻辑 ---
				await Api.addUser({
					username: values.username,
					password: values.password,
					nickname: values.nickname ?? "",
					phone: values.phone ?? "",
					email: values.email ?? "",
					sex: values.sex,
					deptId: values.deptId,
					roleKeys: values.roleKeys,
					remark: values.remark ?? "",
				});
				message.success(t("user.addSuccess"));
			} else {
				// --- 编辑逻辑 ---
				const username = extraParams?.username as string;
				const body: IApi.UserEditReq = {
					deptId: values.deptId,
					roleKeys: values.roleKeys ?? [],
					status: values.status,
				};
				if (values.password) body.password = values.password;

				await Api.editUser(username, body);
				message.success(t("user.editSuccess"));
			}

			tableRef.current?.refresh();
		},
		[formMode.type, message, t], // 依赖项中加入 formMode.type
	);

	// ---- 列配置 ----
	const columns: ProTableColumnType<IApi.UserItem>[] = useMemo(() => {
		if (roleOptions.length <= 0 || deptFlatenOptions.length <= 0) return [];
		return [
			{
				title: t("user.name"),
				dataIndex: "username",
				width: 120,
				render: (v) => (
					<>
						<CopyableCell text={String(v)}>
							<span>{v}</span>
						</CopyableCell>
					</>
				),
			},
			{ title: t("user.nickname"), dataIndex: "nickname", width: 120 },
			{
				title: t("system.department"),
				dataIndex: "dept",
				width: 120,
				render: (_, record) => record.dept?.dept_name ?? "-",
			},
			{
				title: t("system.role"),
				dataIndex: "roles",
				filters: roleOptions.map((item) => ({
					text: item.label,
					value: item.value,
				})),
				onFilter: (value, record) => {
					return record.roles?.some((role) => role.role_key === value);
				},
				render: (_, record) =>
					record.roles?.length ? (
						<Space size={4} wrap>
							{record.roles.map((r, i) => (
								<Tag key={r.role_key} color={presets[i]}>
									{r.role_name}
								</Tag>
							))}
						</Space>
					) : (
						"-"
					),
			},
			{
				title: t("user.phone"),
				dataIndex: "phone",
				width: 120,
				render: (v) => (
					<>
						<CopyableCell text={String(v)}>
							<span>{v}</span>
						</CopyableCell>
					</>
				),
			},
			{ title: t("user.email"), dataIndex: "email", width: 160, ellipsis: true },
			{
				title: t("user.sex"),
				dataIndex: "sex",
				width: 70,
				filters: SEX_OPTIONS.map((item) => ({
					text: t(item.label),
					value: item.value,
				})),
				onFilter: (value, record) => record.sex == value,

				render: (value) => t(SEX_TEXT[(value as number) ?? 0] ?? "profile.unknown"),
			},
			{
				title: t("system.status"),
				dataIndex: "status",
				width: 80,
				filters: STATUS_OPTIONS.map((item) => ({
					text: t(item.label),
					value: item.value,
				})),
				onFilter: (value, record) => record.status == value,
				render: (value) => (
					<Tag color={(value as number) === 1 ? "green" : "red"}>
						{t(STATUS_TEXT[(value as number) ?? 0] ?? "-")}
					</Tag>
				),
			},
			{
				title: t("system.createdAt"),
				dataIndex: "createdAt",
				width: 170,
				render: (time) => dayjs(time).format("YYYY-MM-DD HH:mm:ss"),
			},
			{
				title: t("system.action"),
				key: "action",
				width: 90,
				fixed: "right",
				render: (_, record) => (
					<Button
						size="small"
						color="primary"
						variant="filled"
						icon={<EditOutlined />}
						onClick={() => openEdit(record)}
					>
						{t("system.edit")}
					</Button>
				),
			},
		];
	}, [roleOptions, openEdit, deptFlatenOptions, t]);

	// ---- 表格配置 ----
	const tableProps: ProTableProps<IApi.UserItem> = useMemo(() => {
		return {
			rowKey: "username",
			api: Api.getUserList,
			columns: columns, // 建议你的 columns 变量也使用 useMemo 包裹
			title: t("user.title"),
			externalParams: externalParams, // 如果 externalParams 是动态的，确保它也是稳定引用
			search: {
				fields: [
					{
						type: "input",
						label: t("user.name"),
						name: "username",
						fieldProps: { placeholder: t("common.input") },
					},
					{
						type: "input",
						label: t("user.phone"),
						name: "phone",
						fieldProps: { placeholder: t("common.input") },
					},
					{
						type: "select",
						label: t("user.status"),
						name: "status",
						fieldProps: {
							allowClear: true,
							placeholder: t("common.select"),
							options: STATUS_OPTIONS.map((item) => ({
								...item,
								label: t(item.label),
							})),
						},
					},
				],
			},
			toolbarExtra: !hasPerm("system:user:add") ? undefined : (
				<Button
					type="primary"
					onClick={handleOpenAdd} // 改为调用新的 handleOpenAdd
				>
					{t("user.add")}
				</Button>
			),
		};
	}, [columns, externalParams, hasPerm, handleOpenAdd, t]);

	return (
		<>
			<div className="flex gap-2 items-start">
				{/* 左侧部门树：选中部门后右侧表格按「该部门 + 全部子部门」过滤（data-scope 仍叠加生效） */}
				<Card
					className="w-[223px] self-stretch"
					title={t("user.department")}
					size="small"
					loading={deptLoading}
					extra={
						<Tooltip title={t("system.refresh")}>
							<Button icon={<ReloadOutlined />} onClick={fetchDeptTree} />
						</Tooltip>
					}
					styles={{
						body: {
							padding: 8,
							overflow: "auto",
						},
					}}
				>
					<Input
						className="w-full mb-2"
						placeholder={t("system.searchDepartment")}
						prefix={
							<IconSearch size={14} stroke={1.5} className="text-gray-400" />
						}
						allowClear
						value={searchValue}
						onChange={(e) => setSearchValue(e.target.value)}
						style={{ width: "100%" }}
					/>
					{deptFilterTree.length == 0 ? (
						<Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
					) : (
						<Tree
							blockNode
							showLine
							treeData={deptFilterTree}
							selectedKeys={selectedKeys}
							expandedKeys={expandedKeys}
							onSelect={handleDeptSelect}
							onExpand={(keys) => setExpandedKeys(keys)}
						/>
					)}
				</Card>
				<div className="flex-1 min-w-0">
					<ProTable<IApi.UserItem> ref={tableRef} {...tableProps} />
				</div>
			</div>
			<ProForm
				className={"grid grid-cols-2 gap-4"}
				ref={formRef}
				fields={getFormFields()} // 使用动态生成的字段
				onConfirm={handleFormSubmit} // 使用合并后的提交函数
				width={720}
			/>
		</>
	);
};

export default UserPage;
