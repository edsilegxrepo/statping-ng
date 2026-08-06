<template>
  <div class="text-center" style="width: 210px" v-if="!loaded">
    <font-awesome-icon icon="circle-notch" class="h-25 text-dim" spin />
  </div>
  <apexchart v-else width="240" height="30" type="bar" :options="chartOpts" :series="data"></apexchart>
</template>

<script setup>
import { onMounted, reactive, ref, watch } from "vue";
import Api from "@/API";

const props = defineProps({
	service: {
		required: true,
		type: Object,
	},
	timeframe: {
		required: true,
		type: String,
	},
});

const data = ref(null);
const loaded = ref(false);

const timeoptions = {
	weekday: "long",
	year: "numeric",
	month: "long",
	day: "numeric",
	hour: "numeric",
	minute: "numeric",
};

const chartOpts = reactive({
	chart: {
		type: "bar",
		height: 50,
		sparkline: { enabled: true },
		animations: { enabled: false },
	},
	xaxis: { type: "datetime" },
	showPoint: false,
	fullWidth: true,
	chartPadding: { top: 0, right: 0, bottom: 0, left: 0 },
	stroke: { curve: "straight" },
	fill: { opacity: 0.4 },
	yaxis: { min: 0, max: 5 },
	plotOptions: {
		bar: {
			colors: {
				ranges: [
					{ from: 0, to: 1, color: "#cfcfcf" },
					{ from: 2, to: 3, color: "#f58e49" },
					{ from: 3, to: 20, color: "#e01a1a" },
					{ from: 21, to: Infinity, color: "#9b0909" },
				],
			},
		},
	},
	tooltip: {
		theme: false,
		enabled: true,
		custom: ({ series, seriesIndex, dataPointIndex, w }) => {
			const val = series[seriesIndex][dataPointIndex];
			const ts = w.globals.seriesX[seriesIndex][dataPointIndex];
			const dt = new Date(ts).toLocaleDateString("en-us", timeoptions);
			return `<div class="chart_list_tooltip">${val - 1} Failures<br>${dt}</div>`;
		},
		fixed: { enabled: true, position: "topLeft", offsetX: 0, offsetY: 0 },
		x: { formatter: (value) => value },
		y: { show: false },
	},
	title: { enabled: false },
	subtitle: { enabled: false },
});

onMounted(() => {
	loadFailures();
});

watch(
	() => props.timeframe,
	() => {
		loaded.value = false;
		loadFailures();
	},
);

function toUnix(date) {
	return Math.floor(date.getTime() / 1000);
}

function nowSubtract(seconds) {
	return new Date(Date.now() - seconds * 1000);
}

function beginningOf(period, date = new Date()) {
	const d = new Date(date);
	d.setHours(0, 0, 0, 0);
	return d;
}

function endOf(period, date = new Date()) {
	const d = new Date(date);
	d.setHours(23, 59, 59, 999);
	return d;
}

function convertChartData(chartData) {
	if (!chartData) return [];
	return chartData.map((d) => ({
		x: d.timeframe,
		y: d.amount + 1,
	}));
}

async function loadFailures() {
	loaded.value = false;
	let start = 43200;
	let group = "12h";
	if (props.timeframe === "3h") {
		start = 10800;
		group = "5m";
	} else if (props.timeframe === "12h") {
		start = 43200;
		group = "1h";
	} else if (props.timeframe === "24h") {
		start = 86400;
		group = "2h";
	} else if (props.timeframe === "7d") {
		start = 86400 * 7;
		group = "24h";
	}

	const startTime = beginningOf("day", nowSubtract(start));
	const endTime = endOf("day", new Date());

	const failuresData = await Api.service_failures_data(
		props.service.id,
		toUnix(startTime),
		toUnix(endTime),
		group,
		true,
	);
	data.value = [{ data: convertChartData(failuresData) }];
	loaded.value = true;
}
</script>
