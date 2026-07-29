<template>
  <div class="col-12">
  <div class="text-center" style="width:210px" v-if="!loaded">
    <font-awesome-icon icon="circle-notch" class="h-25 text-dim" spin/>
  </div>
  <apexchart v-else width="100%" height="80" type="bar" :options="chartOpts" :series="data"></apexchart>
  </div>
</template>

<script>
import Api from "@/API";

const timeoptions = {
	weekday: "long",
	year: "numeric",
	month: "long",
	day: "numeric",
	hour: "numeric",
	minute: "numeric",
};

export default {
	name: "FailuresBarChart",
	props: {
		service: {
			required: true,
			type: Object,
		},
		group: {
			required: true,
			type: String,
		},
		start: {
			required: true,
			type: String,
		},
		end: {
			required: true,
			type: String,
		},
	},
	data() {
		return {
			data: null,
			loaded: false,
			chartOpts: {
				chart: {
					type: "bar",
					height: 150,
					sparkline: {
						enabled: true,
					},
					animations: {
						enabled: false,
					},
				},
				xaxis: {
					type: "datetime",
				},
				showPoint: false,
				fullWidth: true,
				chartPadding: { top: 0, right: 0, bottom: 0, left: 80 },
				stroke: {
					curve: "straight",
				},
				fill: {
					opacity: 0.8,
				},
				yaxis: {
					min: 0,
					forceNiceScale: true,
				},
				plotOptions: {
					bar: {
						colors: {
							ranges: [
								{
									from: 1,
									to: 2,
									color: "#f58e49", // Orange for minor
								},
								{
									from: 3,
									to: 10,
									color: "#e01a1a", // Red for major
								},
								{
									from: 11,
									to: Infinity,
									color: "#9b0909", // Dark Red for critical spikes
								},
							],
						},
					},
				},
				tooltip: {
					theme: false,
					enabled: true,
					custom: ({ series, seriesIndex, dataPointIndex, w }) => {
						let val = series[seriesIndex][dataPointIndex];
						let ts = w.globals.seriesX[seriesIndex][dataPointIndex];
						const dt = new Date(ts).toLocaleDateString("en-us", timeoptions);
						let _ago = `${(dataPointIndex - 12) * -1} hours ago`;
						if ((dataPointIndex - 12) * -1 === 0) {
							_ago = `Current hour`;
						}
						return `<div class="chart_list_tooltip font-2 mb-4">${val} Failures<br>${dt}</div>`;
					},
					fixed: {
						enabled: true,
						position: "topLeft",
						offsetX: 0,
						offsetY: 0,
					},
					x: {
						formatter: (value) => {
							return value;
						},
					},
					y: {
						show: false,
					},
				},
				title: {
					enabled: false,
				},
				subtitle: {
					enabled: false,
				},
			},
		};
	},
	async mounted() {
		await this.loadFailures();
	},
	watch: {
		group(_o, _n) {
			this.loaded = false;
			this.loadFailures();
			this.loaded = true;
		},
		start(_o, _n) {
			this.loaded = false;
			this.loadFailures();
			this.loaded = true;
		},
		end(_o, _n) {
			this.loaded = false;
			this.loadFailures();
			this.loaded = true;
		},
	},
	methods: {
		convertChartData(data) {
			if (!data) {
				return [];
			}
			let arr = [];
			data.forEach((d, _k) => {
				arr.push({
					x: d.timeframe,
					y: d.amount,
				});
			});
			return arr;
		},
		async loadFailures() {
			this.loaded = false;
			const startEnd = this.startEndParams(
				this.parseISO(this.start),
				this.parseISO(this.end),
				this.group,
			);
			const data = await Api.service_failures_data(
				this.service.id,
				startEnd.start,
				startEnd.end,
				this.group,
				true,
			);
			this.loaded = true;
			this.data = [{ data: this.convertChartData(data) }];
		},
	},
};
</script>
