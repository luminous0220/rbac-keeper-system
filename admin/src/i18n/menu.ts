const menuKeysByPath: Record<string, string> = {
	"/system": "navigation.system",
	"/monitor": "navigation.monitor",
};

const menuKeysByComponent: Record<string, string> = {
	"Dashboard/index.tsx": "navigation.dashboard",
	"System/User/index.tsx": "navigation.user",
	"System/Role/index.tsx": "navigation.role",
	"System/Menu/index.tsx": "navigation.menu",
	"System/Dept/index.tsx": "navigation.department",
	"System/Dict/index.tsx": "navigation.dictionary",
	"Monitor/LoginLog/index.tsx": "navigation.loginLog",
	"Monitor/OperationLog/index.tsx": "navigation.operationLog",
};

export function translateMenuTitle(
	path: string,
	component: string,
	title: string,
	translate: (key: string) => string,
): string {
	const key = menuKeysByComponent[component] ?? menuKeysByPath[path];
	return key ? translate(key) : title;
}
