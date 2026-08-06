import { computed, watch } from "vue";
import { useMainStore } from "@/stores/main";

function createFaviconCanvas(color, hasIssues = false) {
	const canvas = document.createElement("canvas");
	canvas.width = 32;
	canvas.height = 32;
	const ctx = canvas.getContext("2d");

	// Background circle
	ctx.beginPath();
	ctx.arc(16, 16, 14, 0, 2 * Math.PI);
	ctx.fillStyle = color;
	ctx.fill();

	// White inner icon - checkmark for good, X for issues
	ctx.strokeStyle = "#fff";
	ctx.lineWidth = 3;
	ctx.lineCap = "round";
	ctx.lineJoin = "round";

	if (hasIssues) {
		// X mark
		ctx.beginPath();
		ctx.moveTo(11, 11);
		ctx.lineTo(21, 21);
		ctx.moveTo(21, 11);
		ctx.lineTo(11, 21);
		ctx.stroke();
	} else {
		// Checkmark
		ctx.beginPath();
		ctx.moveTo(10, 16);
		ctx.lineTo(14, 20);
		ctx.lineTo(22, 12);
		ctx.stroke();
	}

	return canvas.toDataURL("image/png");
}

function setFavicon(dataUrl) {
	let link = document.querySelector("link[rel*='icon']");
	if (!link) {
		link = document.createElement("link");
		link.rel = "icon";
		document.head.appendChild(link);
	}
	link.type = "image/png";
	link.href = dataUrl;
}

export function useFaviconStatus() {
	const store = useMainStore();

	const services = computed(() => store.services || []);
	const onlineCount = computed(
		() => services.value.filter((s) => s.online).length,
	);
	const totalCount = computed(() => services.value.length);

	const statusColor = computed(() => {
		if (totalCount.value === 0) return "#9ca3af"; // gray - no services
		if (onlineCount.value === totalCount.value) return "#22c55e"; // green - all online
		if (onlineCount.value === 0) return "#ef4444"; // red - all offline
		return "#f59e0b"; // amber - partial
	});

	const hasIssues = computed(() => {
		return totalCount.value > 0 && onlineCount.value < totalCount.value;
	});

	watch(
		[statusColor, hasIssues],
		([color, issues]) => {
			const favicon = createFaviconCanvas(color, issues);
			setFavicon(favicon);
		},
		{ immediate: true },
	);
}
