<template>
    <apexchart v-if="ready" width="100%" height="400" type="heatmap" :options="plotOptions" :series="series"></apexchart>
</template>

<script>
  import Api from "../../API"

  export default {
      name: 'ServiceHeatmap',
      props: {
          service: {
              type: Object,
              required: true
          }
      },
      async created() {
          await this.chartHeatmap()
      },
      data() {
        return {
          ready: false,
          mergedData: [],
          data: [],
          outageSeverity: {
            minor: { start: 1, end: 30 },
            moderate: { start: 30, end: 120 },
            major: { start: 120, end: 240 },
            critical: { start: 240 }
          },
          plotOptions: {
            tooltip: { 
              enabled: true,
              custom: function({series, seriesIndex, dataPointIndex, w}) {
                const failures = series[seriesIndex][dataPointIndex];
                if (failures > 0) { return  `<div class="p-2"><strong>${failures} Failures</strong><br><small>Click to view hourly breakdown</small></div>` }
                return ''
              }
            },
            chart: {
              selection: {
                enabled: false
              },
              zoom: {
                enabled: false
              },
              toolbar: {
                show: false
              },
              events: {
                dataPointSelection: (event, chartContext, config) => {
                    const monthName = config.w.globals.seriesNames[config.seriesIndex];
                    const day = config.dataPointIndex + 1;
                    const year = new Date().getFullYear();
                    const selectedDate = new Date(`${monthName} ${day}, ${year}`);
                    this.$emit('selected-day', selectedDate);
                }
              }
            },
            grid: {
              show: true,
              borderColor: '#dee2e6', // Subtle, clean grey border
              position: 'front', // Draw the grid lines ON TOP of the solid boxes so they are visible
              xaxis: {
                lines: {
                  show: false // No vertical lines
                }
              },
              yaxis: {
                lines: {
                  show: true // Horizontal lines separating each month
                }
              },
              padding: {
                top: 10,
                right: 20,
                bottom: 10,
                left: 20
              }
            },
            stroke: {
              show: true,
              width: 1.5, // Thin border separating each box
              colors: ['#ffffff'] // White color to cleanly isolate each day box
            },
            dataLabels: {
              enabled: false
            },
            colors: [ "#cb3d36" ],
            xaxis: {
              type: "category",
              labels: {
                show: true,
              },
              tooltip: {
                enabled: true,
                formatter: function(value, { series, seriesIndex, dataPointIndex, w }) {
                  const month = w.globals.seriesNames[seriesIndex];
                  const year = new Date().getFullYear();
                  return `${dataPointIndex + 1} ${month} ${year}`;
                }
              }
            },
            yaxis: {
              labels: {
                show: true,
                style: {
                  fontSize: '12px',
                  fontWeight: 'bold'
                }
              }
            },
            plotOptions: {
              heatmap: {
                enableShades: false,
                useFillColorAsStroke: false,
                colorScale: {
                  ranges: [{
                      from: -1000000,
                      to: 0,
                      color: '#f8f9fa',
                      name: 'Healthy',
                    },
                    {
                      from: 1,
                      to: 30,
                      color: '#98EE99',
                      name: 'Minor',
                    },
                    {
                      from: 31,
                      to: 120,
                      color: '#FFEB3B',
                      name: 'Moderate',
                    },
                    {
                      from: 121,
                      to: 240,
                      color: '#FF9800',
                      name: 'Major',
                    },
                    {
                      from: 241,
                      to: 1000000,
                      color: '#F44336',
                      name: 'Critical',
                    }
                  ]
                }
              }
            }
          },
          series: [{
            data: []
          }],
        }
      },
      methods: {
          async chartHeatmap() {
            this.ready = false;
            const months = [];
            const current = this.firstDayOfMonth(this.now());

            // Generate 6 months in chronological order (Oldest to Newest)
            for (let i = 5; i >= 0; i--) {
              let start = this.addMonths(current, -i);
              let end = this.lastDayOfMonth(start);
              months.push({ start, end });
            }

            const results = await Promise.all(months.map(m => this.heatmapData(m.start, m.end)));
            this.series = results;
            this.ready = true;
          },
          async heatmapData(start, end) {
              const failuresData = await Api.service_failures_data(this.service.id, this.toUnix(start), this.toUnix(end), "24h", true)
              const dataArr = this.mergeData(failuresData);
              return {
                name: start.toLocaleString('en-us', { month: 'long'}), 
                data: dataArr
              }
          },
          mergeData(failuresData) {
            const dataArr = [];
            // Initialize with 31 days of zeros (Healthy)
            for (let i = 0; i < 31; i++) {
              dataArr.push({ x: (i + 1).toString(), y: 0 });
            }
            
            // Map actual failure data to the correct day index
            failuresData.forEach(d => {
              const day = new Date(d.timeframe).getUTCDate();
              if (day >= 1 && day <= 31) {
                dataArr[day - 1].y = d.amount;
              }
            });
            
            return dataArr;
          },
          getDayColor({ value, seriesIndex, w }) {
              // No data points for day
              if (value === 0) {
                return '#e9e9e9';
              } 
              // No failures for day
              else if (value < 0) {
                return '#4CAF50';
              } 
              // Some failures for the day
              else {
                // Determine the severity and return the corresponding color class
                const outageSeverity = value;
                if (outageSeverity >= this.outageSeverity.minor.start && outageSeverity < this.outageSeverity.minor.end) {
                  return '#98EE99'; // Light green
                } else if (outageSeverity >= this.outageSeverity.moderate.start && outageSeverity < this.outageSeverity.moderate.end) {
                  return '#FFEB3B'; // Yellow
                } else if (outageSeverity >= this.outageSeverity.major.start && outageSeverity < this.outageSeverity.major.end) {
                  return '#FF9800'; // Orange
                } else if (outageSeverity >= this.outageSeverity.critical.start) {
                  return '#F44336'; // Red
                }
              }
              // Default, shouldn't get here
              return '#e9e9e9';
            }
      }
  }
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
