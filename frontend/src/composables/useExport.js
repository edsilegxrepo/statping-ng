export function getTimestamp() {
	const now = new Date();
	return (
		now.getFullYear().toString() +
		(now.getMonth() + 1).toString().padStart(2, "0") +
		now.getDate().toString().padStart(2, "0") +
		now.getHours().toString().padStart(2, "0") +
		now.getMinutes().toString().padStart(2, "0") +
		now.getSeconds().toString().padStart(2, "0")
	);
}

export function sanitizeFilename(name) {
	return name
		.replace(/[^a-zA-Z0-9-_]/g, "_")
		.replace(/_+/g, "_")
		.replace(/^_|_$/g, "");
}

export function downloadFile(content, filename, mimeType) {
	const blob = new Blob([content], { type: mimeType });
	const url = URL.createObjectURL(blob);
	const a = document.createElement("a");
	a.href = url;
	a.download = filename;
	document.body.appendChild(a);
	a.click();
	document.body.removeChild(a);
	URL.revokeObjectURL(url);
}

export function exportTSV(headers, rows, filename) {
	const tsv = [headers.join("\t"), ...rows.map((r) => r.join("\t"))].join("\n");
	const safeName = sanitizeFilename(filename);
	downloadFile(
		tsv,
		`${safeName}_${getTimestamp()}.tsv`,
		"text/tab-separated-values",
	);
}

export function exportJSON(data, filename) {
	const json = JSON.stringify(data, null, 2);
	const safeName = sanitizeFilename(filename);
	downloadFile(json, `${safeName}_${getTimestamp()}.json`, "application/json");
}
