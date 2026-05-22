<template>
    <apexchart v-if="ready" width="100%" height="350" type="heatmap" :options="plotOptions" :series="series"></apexchart>
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
              borderColor: '#f1f1f1',
              padding: {
                top: 10,
                right: 20,
                bottom: 10,
                left: 20
              }
            },
            colors: [ "#cb3d36" ],
            xaxis: {
              tickAmount: 30,
              min: 1,
              max: 31,
              type: "numeric",
              labels: {
                show: true,
                enabled: true,
                formatter: (value) => `${value}`,
              },
              tooltip: {
                enabled: true,
                formatter: function(value, { series, seriesIndex, dataPointIndex, w }) {
                  // Assuming each 'x' value is already set as the day of the month.
                  // The series name for each series is set as the full month name.
                  const month = w.globals.seriesNames[seriesIndex]; // Gets the month from the series name.
                  const year = new Date().getFullYear(); // Assumes current year; adjust as needed.
                  return `${dataPointIndex} ${month} ${year}`; // Formats the tooltip's title to show a full date.
                }
              }
            },
            yaxis: {
              labels: {
                show: true,
                style: {
                  fontSize: '11px'
                }
              }
            },
            plotOptions: {
              heatmap: {
                enableShades: true,
                useFillColorAsStroke: false,
                colorScale: {
                  ranges: [{
                      from: -1000000,
                      to: 0,
                      color: '#f3f3f3',
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
            const monthData = []
            let current = this.firstDayOfMonth(this.now())

            for (let i=0; i<6; i++) {
                let monthStart = this.addMonths(current, -i)
                let monthEnd = this.lastDayOfMonth(monthStart)
                monthData.push(await this.heatmapData(monthStart, monthEnd))
            }

            this.series = monthData
            this.ready = true
          },
          async heatmapData(start, end) {
              const failuresData = await Api.service_failures_data(this.service.id, this.toUnix(start), this.toUnix(end), "24h", true)
              const hitsData = await Api.service_hits(this.service.id, this.toUnix(start), this.toUnix(end), "24h", true);

              // Merge the data
              const mergedData = this.mergeData(failuresData, hitsData);
              console.log(mergedData)
              return {name: start.toLocaleString('en-us', { month: 'long'}), data: mergedData}
          },
          mergeData(failuresData, hitsData) {
            let data = {}
            
            // Process hits data
            hitsData.forEach(d => {
              let date = this.parseISO(d.timeframe);
              data[date] = -d.amount
            });
            
            // Process failures data
            failuresData.forEach(d => {
              let date = this.parseISO(d.timeframe);
              data[date] = d.amount
            });

            let dataArr = []
            for (let i = 0; i < 31; i++) {
              if (failuresData[i] && failuresData[i].amount > 0) {
                // If any failures return failure amount
                dataArr.push(failuresData[i].amount)
              } else if (hitsData[i] && hitsData[i].amount > 0) {
                // Make neg if only success
                // dataArr.push(-hitsData[i].amount)
                dataArr.push(0)
              } else {
                dataArr.push(0)
              }
            }
            
            // Convert map to array
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
