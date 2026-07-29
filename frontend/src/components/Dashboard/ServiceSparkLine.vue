<template>
  <div v-if="series && series.length">
    <apexchart width="100%" height="100" type="bar" :options="chartOpts" :series="series"></apexchart>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  series: {
    type: Array,
    default: () => [],
  },
  title: {
    type: String,
    default: '',
  },
  subtitle: {
    type: String,
    default: '',
  },
})

const timeoptions = {
  weekday: 'long',
  year: 'numeric',
  month: 'long',
  day: 'numeric',
  hour: 'numeric',
  minute: 'numeric',
}

function humanTime(ms) {
  if (!ms) return '0ms'
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

const chartOpts = computed(() => ({
  chart: {
    type: 'bar',
    height: 180,
    sparkline: {
      enabled: true,
    },
  },
  showPoint: false,
  fullWidth: true,
  chartPadding: { top: 0, right: 0, bottom: 0, left: 0 },
  stroke: {
    curve: 'straight',
  },
  fill: {
    opacity: 0.3,
  },
  yaxis: {
    min: 0,
  },
  colors: ['#b3bdc3'],
  tooltip: {
    theme: false,
    enabled: true,
    custom: ({ series, seriesIndex, dataPointIndex, w }) => {
      let ts = w.globals.seriesX[seriesIndex][dataPointIndex]
      const dt = new Date(ts).toLocaleDateString('en-us', timeoptions)
      let val = series[seriesIndex][dataPointIndex]
      return `<div class="chartmarker"><span class="">Average Response Time: ${humanTime(val)}</span><span>${dt}</span></div>`
    },
    fixed: {
      enabled: true,
      position: 'bottomLeft',
      offsetX: 0,
      offsetY: -15,
    },
    x: {
      show: true,
    },
    y: {
      formatter: (value) => `${value} %`,
    },
  },
  title: {
    text: props.title,
    offsetX: 0,
    style: {
      fontSize: '18px',
      cssClass: 'apexcharts-yaxis-title',
    },
  },
  subtitle: {
    text: props.subtitle,
    offsetX: 0,
    offsetY: 20,
    style: {
      fontSize: '9px',
      cssClass: 'apexcharts-yaxis-title',
    },
  },
}))
</script>
