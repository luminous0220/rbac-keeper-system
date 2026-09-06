/**
 * 触发浏览器下载后端返回的 Blob 文件（如日志 CSV 导出）。
 * 通过临时 <a download> + createObjectURL 实现，结束后及时释放 URL。
 */
export function downloadBlob(blob: Blob, filename: string) {
	const url = URL.createObjectURL(blob);
	const a = document.createElement("a");
	a.href = url;
	a.download = filename;
	document.body.appendChild(a);
	a.click();
	document.body.removeChild(a);
	URL.revokeObjectURL(url);
}
