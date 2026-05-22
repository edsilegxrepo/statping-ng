<template>
  <div v-if="selectedDate" class="card text-black-50 bg-white mt-3 mb-5">
    <div class="card-header text-capitalize d-flex justify-content-between align-items-center">
      <span>Hourly Failure Breakdown: {{ format(selectedDate, 'MMMM do, yyyy') }} (UTC) - {{ totalFailures }} total failures</span>
      <button class="btn btn-sm btn-outline-secondary" @click="$emit('close')">
        <font-awesome-icon icon="times" />
      </button>
    </div>
    <div class="card-body">
      <div v-if="loading" class="text-center py-5">
        <font-awesome-icon icon="circle-notch" size="3x" spin />
      </div>
      <apexchart v-else width="100%" height="250" type="bar" :options="chartOptions" :series="series"></apexchart>
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
  computed: {
    totalFailures() {
      if (!this.series || this.series.length === 0 || !this.series[0].data) {
        return 0;
      }
      return this.series[0].data.reduce((sum, val) => sum + val, 0);
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
        plotOptions: {
          bar: {
            colors: {
              ranges: [
                {
                  from: 0,
                  to: 0,
                  color: '#28a745' // Green for Healthy
                },
                {
                  from: 1,
                  to: 10,
                  color: '#98EE99' // Light Green for Minor Failures
                },
                {
                  from: 11,
                  to: 30,
                  color: '#FFEB3B' // Yellow for Moderate Failures
                },
                {
                  from: 31,
                  to: 60,
                  color: '#FF9800' // Orange for Major Failures
                },
                {
                  from: 61,
                  to: 1000000,
                  color: '#F44336' // Red for Critical Failures
                }
              ]
            }
          }
        },
        xaxis: {
          type: 'category',
          categories: Array.from({length: 24}, (_, i) => `${i}:00`),
          title: { text: 'Hour of Day (UTC)' }
        },
        yaxis: {
          title: { text: 'Failures' },
          min: 0,
          forceNiceScale: true
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
      
      // Since the heatmap aggregates and displays data by UTC day, 
      // we must query the exact 24-hour range in UTC for the selected day.
      const year = this.selectedDate.getFullYear();
      const month = this.selectedDate.getMonth();
      const day = this.selectedDate.getDate();
      
      const startUnix = Math.floor(Date.UTC(year, month, day, 0, 0, 0) / 1000);
      const endUnix = Math.floor(Date.UTC(year, month, day, 23, 59, 59) / 1000);
      
      const data = await Api.service_failures_data(
        this.service.id, 
        startUnix, 
        endUnix, 
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
