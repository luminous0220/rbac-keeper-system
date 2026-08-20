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

// 性别展示：UserItem.sex 按 model 语义 1男 2女（0 为新增接口遗留的未知态）
const SEX_TEXT: Record<number, string> = { 0: "未知", 1: "男", 2: "女" };
// 账号状态：model 语义 0禁用 1正常
const STATUS_TEXT: Record<number, string> = { 0: "停用", 1: "正常" };
const STATUS_OPTIONS = [
	{ label: "正常", value: 1 },
	{ label: "停用", value: 0 },
];
// 性别选项：新增/编辑/列表统一按 model 语义 0未知 1男 2女（后端已对齐为 omitempty,oneof=0 1 2）
const SEX_OPTIONS = [
	{ label: "未知", value: 0 },
	{ label: "男", value: 1 },
	{ label: "女", value: 2 },
];

const presets = [
	"orange",
	"volcano",
	"cyan",
	"lime",
	"magenta",
	"red",
	"gold",
	"green",
	"blue",
	"geekblue",
	"purple",
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
	const { hasPerm } = usePerms();
	const tableRef = useRef<ProTableRef<IApi.UserItem>>(null);
	const addFormRef = useRef<ProFormInstance>(null);
	const editFormRef = useRef<ProFormInstance>(null);

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
			.catch((err) => console.error("加载部门树失败", err))
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
			.catch((err) => console.error("加载角色列表失败", err));
	}, [fetchDeptTree]);

	const deptTreeData = useMemo(() => toTreeSelect(deptTree), [deptTree]);

	// 左侧部门树数据：顶部固定一个"全部部门"根节点，选中即清空部门筛选
	const [searchValue, setSearchValue] = useState("");
	const deptFilterTree = useMemo<TreeNodeItem[]>(() => {
		const originalTree = [
			{ title: "全部部门", key: 0, children: toTreeData(deptTree) },
		];
		return filterTree(originalTree, searchValue);
	}, [deptTree, searchValue]);

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

	// ---- 新增用户 ----
	const addFields: ProFormFields = [
		{
			type: "input",
			label: "用户名",
			name: "username",
			fieldProps: { maxLength: 30, placeholder: "6-30 位字母或数字" },
			rules: [
				{ required: true, message: "请输入用户名" },
				{ min: 6, max: 30, message: "用户名长度需为 6-30 位" },
				{ pattern: /^[a-zA-Z0-9]+$/, message: "用户名只能包含字母和数字" },
			],
		},
		{
			type: "input-password",
			label: "密码",
			name: "password",
			fieldProps: { placeholder: "8-32 位，需同时包含字母和数字" },
			rules: [
				{ required: true, message: "请输入密码" },
				{ min: 8, max: 32, message: "密码长度需为 8-32 位" },
				{
					pattern: /^(?=.*[A-Za-z])(?=.*\d).{8,32}$/,
					message: "密码需同时包含字母和数字",
				},
			],
		},
		{
			type: "input",
			label: "昵称",
			name: "nickname",
			fieldProps: {
				maxLength: 30,
				placeholder: "请输入昵称",
			},
			rules: [{ required: true, message: "请输入昵称" }],
		},
		{
			type: "input",
			label: "手机号",
			name: "phone",
			fieldProps: { maxLength: 11, placeholder: "请输入正确的手机号" },
			rules: [{ pattern: /^1[3-9]\d{9}$/, message: "请输入正确的手机号" }],
		},
		{
			type: "input",
			label: "邮箱",
			name: "email",
			fieldProps: { maxLength: 50, placeholder: "请输入正确的邮箱" },
			rules: [{ type: "email", message: "请输入正确的邮箱" }],
		},
		{
			type: "radio",
			label: "性别",
			name: "sex",
			fieldProps: { options: SEX_OPTIONS }, // 默认值经 open({ initialValues }) 注入为 1（男）
		},
		{
			type: "tree-select",
			label: "所属部门",
			name: "deptId",
			fieldProps: {
				treeDefaultExpandAll: true,
				treeData: deptTreeData,
				placeholder: "请选择部门",
				showSearch: true,
			},
			rules: [{ required: true, message: "请选择部门" }],
		},
		{
			type: "select",
			label: "角色",
			name: "roleKeys",
			fieldProps: {
				options: roleOptions,
				mode: "multiple",
				placeholder: "请选择角色（可多选）",
			},
			rules: [{ required: true, message: "请至少选择一个角色" }],
		},
		{
			type: "textarea",
			label: "备注",
			name: "remark",
			formItemProps: {
				className: "col-span-2",
			},
			fieldProps: {
				maxLength: 500,
				rows: 3,
				placeholder: "请输入备注（选填）",
			},
		},
	];

	const handleAddSubmit = useCallback(
		async (_title: string, values: Record<string, any>) => {
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
			message.success("新增用户成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	// ---- 编辑用户 ----
	const editFields: ProFormFields = [
		{
			type: "input-password",
			label: "密码",
			name: "password",
			fieldProps: { placeholder: "留空则不修改密码" },
			// 可选字段：为空不校验，非空需满足后端复杂度
			rules: [
				{
					validator: (_rule: any, value: string) => {
						if (!value) return Promise.resolve();
						if (value.length < 8 || value.length > 32) {
							return Promise.reject(new Error("密码长度需为 8-32 位"));
						}
						if (!/(?=.*[A-Za-z])(?=.*\d)/.test(value)) {
							return Promise.reject(new Error("密码需同时包含字母和数字"));
						}
						return Promise.resolve();
					},
				},
			],
		},
		{
			type: "tree-select",
			label: "所属部门",
			name: "deptId",
			fieldProps: {
				treeDefaultExpandAll: true,
				treeData: deptTreeData,
				placeholder: "请选择部门",
				showSearch: true,
			},
		},
		{
			type: "select",
			label: "角色",
			name: "roleKeys",
			fieldProps: {
				options: roleOptions,
				mode: "multiple",
				placeholder: "请选择角色（可多选）",
			},
		},
		{
			type: "radio",
			label: "状态",
			name: "status",
			fieldProps: { options: STATUS_OPTIONS },
			rules: [{ required: true, message: "请选择状态" }],
		},
	];

	const openEdit = useCallback((record: IApi.UserItem) => {
		editFormRef.current?.open({
			title: `编辑用户：${record.username}`,
			initialValues: {
				password: "",
				deptId: record.dept?.id,
				roleKeys: record.roles?.map((r) => r.role_key) ?? [],
				status: record.status,
			},
			extraParams: { username: record.username },
		});
	}, []);

	const handleEditSubmit = useCallback(
		async (
			_title: string,
			values: Record<string, any>,
			extraParams?: Record<string, any>,
		) => {
			const username = extraParams?.username as string;
			const body: IApi.UserEditReq = {
				deptId: values.deptId,
				roleKeys: values.roleKeys ?? [],
				status: values.status,
			};
			// 密码为空则不传（后端 nil=不改）
			if (values.password) body.password = values.password;
			await Api.editUser(username, body);
			message.success("修改成功");
			tableRef.current?.refresh();
		},
		[message],
	);

	// ---- 列配置 ----
	const columns: ProTableColumnType<IApi.UserItem>[] = useMemo(() => {
		if (roleOptions.length <= 0 || deptFlatenOptions.length <= 0) return [];
		return [
			{
				title: "用户名",
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
			{ title: "昵称", dataIndex: "nickname", width: 120 },
			{
				title: "部门",
				dataIndex: "dept",
				width: 120,
				filters: deptFlatenOptions.map((item) => ({
					text: item.text,
					value: item.value,
				})),
				onFilter: (value, record) => record.dept?.id == value,
				render: (_, record) => record.dept?.dept_name ?? "-",
			},
			{
				title: "角色",
				dataIndex: "roles",
				filters: roleOptions.map((item) => ({
					text: item.label,
					value: item.value,
				})),
				onFilter: (value, record) =>
					record.roles?.some((role) => role.role_key === value),
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
				title: "手机号",
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
			{ title: "邮箱", dataIndex: "email", width: 160, ellipsis: true },
			{
				title: "性别",
				dataIndex: "sex",
				width: 70,
				filters: SEX_OPTIONS.map((item) => ({
					text: item.label,
					value: item.value,
				})),
				render: (value) => SEX_TEXT[(value as number) ?? 0] ?? "未知",
			},
			{
				title: "状态",
				dataIndex: "status",
				width: 80,
				onFilter: (value, record) => record.sex == value,
				render: (value) => (
					<Tag color={(value as number) === 1 ? "green" : "red"}>
						{STATUS_TEXT[(value as number) ?? 0] ?? "-"}
					</Tag>
				),
			},
			{
				title: "创建时间",
				dataIndex: "createdAt",
				width: 170,
				render: (time) => dayjs(time).format("YYYY-MM-DD HH:mm:ss"),
			},
			{
				title: "操作",
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
						编辑
					</Button>
				),
			},
		];
	}, [roleOptions, openEdit, deptFlatenOptions]);

	// ---- 表格配置 ----
	const tableProps: ProTableProps<IApi.UserItem> = useMemo(() => {
		return {
			rowKey: "username",
			api: Api.getUserList,
			columns: columns, // 建议你的 columns 变量也使用 useMemo 包裹
			title: "用户管理",
			externalParams: externalParams, // 如果 externalParams 是动态的，确保它也是稳定引用
			search: {
				fields: [
					{
						type: "input",
						label: "用户名",
						name: "username",
						fieldProps: { placeholder: "请输入" },
					},
					{
						type: "input",
						label: "手机号",
						name: "phone",
						fieldProps: { placeholder: "请输入" },
					},
					{
						type: "select",
						label: "状态",
						name: "status",
						fieldProps: {
							options: STATUS_OPTIONS,
							allowClear: true,
							placeholder: "点击选择",
						},
					},
				],
			},
			toolbarExtra: !hasPerm("system:user:add") ? undefined : (
				<Button
					type="primary"
					onClick={() =>
						addFormRef.current?.open({
							title: "新增用户",
							initialValues: {
								sex: 1, // 默认男
								// 左侧选中了部门时预填所属部门
								...(selectedDeptId ? { deptId: selectedDeptId } : {}),
							},
						})
					}
				>
					新增用户
				</Button>
			),
		};
	}, [columns, externalParams, hasPerm, selectedDeptId]);

	return (
		<>
			<div className="flex gap-2 items-start">
				{/* 左侧部门树：选中部门后右侧表格按「该部门 + 全部子部门」过滤（data-scope 仍叠加生效） */}
				<Card
					className="w-[223px] self-stretch"
					title="部门"
					size="small"
					loading={deptLoading}
					extra={
						<Tooltip title="刷新">
							<Button
								icon={<ReloadOutlined />}
								onClick={fetchDeptTree}
							/>
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
						placeholder="搜索部门"
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
			{/* 新增用户弹窗 */}
			<ProForm
				className={"grid grid-cols-2 gap-4"}
				ref={addFormRef}
				fields={addFields}
				onConfirm={handleAddSubmit}
				width={600}
			/>
			{/* 编辑用户弹窗 */}
			<ProForm
				ref={editFormRef}
				fields={editFields}
				onConfirm={handleEditSubmit}
				width={560}
			/>
		</>
	);
};

export default UserPage;
