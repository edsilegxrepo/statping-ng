import DOMPurify from "dompurify";

export function sanitizeHtml(html) {
	if (!html) return "";
	return DOMPurify.sanitize(html);
}

export function isNumeric(value) {
	return !isNaN(value) && !isNaN(parseFloat(value));
}

export function toUnix(date) {
	return Math.floor(new Date(date).getTime() / 1000);
}

export function fromUnix(ts) {
	return new Date(parseInt(ts, 10) * 1000).toISOString();
}

export function now() {
	return new Date();
}

export function nowSubtract(seconds) {
	return new Date(Date.now() - seconds * 1000);
}

export function beginningOf(period, date = new Date()) {
	const d = new Date(date);
	d.setHours(0, 0, 0, 0);
	return d;
}

export function endOf(period, date = new Date()) {
	const d = new Date(date);
	d.setHours(23, 59, 59, 999);
	return d;
}

export function format(date, _formatStr) {
	// TODO: Implement custom format string support
	return new Date(date).toLocaleString();
}

export function niceDate(date) {
	return new Date(date).toLocaleDateString();
}

export function serviceLink(service) {
	if (!service) return "/";
	return `/service/${service.permalink || service.id}`;
}

export function convertToChartData(data) {
	if (!data || !Array.isArray(data)) return { data: [] };
	return {
		data: data.map((d) => ({
			x: new Date(d.timeframe).getTime(),
			y: d.amount,
		})),
	};
}

export function humanTime(ms) {
	if (!ms) return "0ms";
	if (ms >= 10000) return `${Math.round(ms / 1000)} ms`;
	return `${ms} μs`;
}
