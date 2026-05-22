<template>
  <div v-if="selectedDate" class="card text-black-50 bg-white mt-3 mb-5">
    <div class="card-header text-capitalize d-flex justify-content-between align-items-center">
      <span>Hourly Failure Breakdown: {{ format(selectedDate, 'MMMM do, yyyy') }}</span>
      <button class="btn btn-sm btn-outline-secondary" @click="$emit('close')">
        <font-awesome-icon icon="times" />
      </button>
    </div>
    <div class="card-body">
      <div v-if="loading" class="text-center py-5">
        <font-awesome-icon icon="circle-notch" size="3x" spin />
      </div>
      <apexchart v-else width="100%" height="250" type="line" :options="chartOptions" :series="series"></apexchart>
    </div>
  </div>
</template>

<script>
import Api from "../../API"

export default {
  name: "DailyFailuresChart",
  props: {
    service: {
      type: Object,
      required: true
    },
    selectedDate: {
      type: Date,
      default: null
    }
  },
  data() {
    return {
      loading: false,
      series: [],
      chartOptions: {
        chart: {
          id: 'daily-failures',
          toolbar: { show: false },
          zoom: { enabled: false }
        },
        stroke: {
          curve: 'smooth',
          width: 3
        },
        colors: ['#e01a1a'],
        xaxis: {
          type: 'category',
          categories: Array.from({length: 24}, (_, i) => `${i}:00`),
          title: { text: 'Hour of Day' }
        },
        yaxis: {
          title: { text: 'Failures' },
          min: 0,
          forceNiceScale: true
        },
        markers: {
          size: 4
        },
        tooltip: {
          y: {
            formatter: (val) => `${val} Failures`
          }
        }
      }
    }
  },
  watch: {
    selectedDate: {
      immediate: true,
      handler(newVal) {
        if (newVal) {
          this.fetchDailyData();
        }
      }
    }
  },
  methods: {
    async fetchDailyData() {
      this.loading = true;
      const start = this.beginningOf('day', this.selectedDate);
      const end = this.endOf('day', this.selectedDate);
      
      const data = await Api.service_failures_data(
        this.service.id, 
        this.toUnix(start), 
        this.toUnix(end), 
        "1h", 
        true
      );

      const hourlyData = Array(24).fill(0);
      data.forEach(d => {
        const hour = new Date(d.timeframe).getUTCHours();
        if (hour >= 0 && hour < 24) {
          hourlyData[hour] = d.amount;
        }
      });

      this.series = [{
        name: 'Failures',
        data: hourlyData
      }];
      this.loading = false;
    }
  }
}
</script>
