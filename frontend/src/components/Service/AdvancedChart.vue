<template>
    <div class="service-chart-container">
        <apexchart v-if="ready" width="100%" height="420" type="line" :options="main_chart_options" :series="main_chart"></apexchart>
    </div>
</template>

<script>
    import Api from "../../API";
    const timeoptions = { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric', hour: 'numeric', minute: 'numeric' };

    export default {
        name: 'AdvancedChart',
        props: {
          service: {
            type: Object,
            required: true
          },
          start: {
            type: String,
            required: true
          },
          end: {
            type: String,
            required: true
          },
          group: {
            type: String,
            required: true
          },
          updated: {
            type: Function,
            required: true
          },
        },
      data() {
        return {
          ready: false,
          loading: true,
          main_data: null,
          ping_data: null,
          expanded_data: null,
          failure_data: null,
          main_chart_options: {
            noData: {
              text: "Loading...",
              align: 'center',
              verticalAlign: 'middle',
              offsetX: 0,
              offsetY: -20,
              style: {
                color: "#bababa",
                fontSize: '27px'
              }
            },
            chart: {
              id: 'mainchart',
              stacked: true,
              events: {
                dataPointSelection: (event, chartContext, config) => {
                  window.console.log('slect')
                  window.console.log(event)
                },
                updated: (chartContext, config) => {
                  window.console.log('updated')
                },
                beforeZoom: (chartContext, { xaxis }) => {
                  const start = (xaxis.min / 1000).toFixed(0)
                  const end = (xaxis.max / 1000).toFixed(0)
                  window.console.log(start, end)
                  this.updated(this.fromUnix(start), this.fromUnix(end))
                  return {
                    xaxis: {
                      min: this.fromUnix(start),
                      max: this.fromUnix(end)
                    }
                  }
                },
                scrolled: (chartContext, { xaxis }) => {
                  window.console.log(xaxis)
                },
              },
              height: 500,
              width: "100%",
              type: "area",
              animations: {
                enabled: false,
                initialAnimation: {
                  enabled: true
                }
              },
              selection: {
                enabled: true
              },
              zoom: {
                enabled: true
              },
              toolbar: {
                show: true
              },
              stroke: {
                show: false,
                curve: 'stepline',
                lineCap: 'butt',
              },
            },
            grid: {
              show: true,
              borderColor: '#f8f9fa',
              padding: {
                top: 25, // Adds space at the top to completely prevent toolbar overlap
                bottom: 25 // Adds professional breathing room above the legend
              }
            },
            xaxis: {
              type: "datetime",
              labels: {
                show: true
              },
              tooltip: {
                enabled: false
              }
            },
            yaxis: [
              {
                title: { text: 'Latency & Ping' },
                labels: {
                  formatter: (value) => {
                    return this.humanTime(value)
                  }
                }
              },
              {
                show: false, // Hidden to share the left axis scale
                labels: {
                  formatter: (value) => {
                    return this.humanTime(value)
                  }
                }
              },
              {
                opposite: true, // Right axis
                title: { text: 'Failure Spikes' },
                labels: {
                  formatter: (value) => {
                    return value ? value.toFixed(0) : '0'
                  }
                },
                min: 0,
                forceNiceScale: true
              }
            ],
            markers: {
              size: 0,
              strokeWidth: 0,
              hover: {
                size: undefined,
                sizeOffset: 0
              }
            },
            tooltip: {
              theme: false,
              enabled: true,
              custom: ({ series, seriesIndex, dataPointIndex, w }) => {
                let ts = w.globals.seriesX[seriesIndex][dataPointIndex];
                const dt = new Date(ts).toLocaleDateString("en-us", timeoptions)
                
                let latencyVal = series[0][dataPointIndex];
                let pingVal = series[1][dataPointIndex];
                let failuresVal = series[2] ? series[2][dataPointIndex] : 0;
                
                let latText = this.humanTime(latencyVal);
                let pingText = this.humanTime(pingVal);
                let failText = failuresVal ? `${failuresVal} Failures` : '0 Failures';
                
                return `<div class="p-3" style="background: rgba(255, 255, 255, 0.98); border: 1px solid #dee2e6; border-radius: 6px; box-shadow: 0 4px 15px rgba(0, 0, 0, 0.12); min-width: 190px; pointer-events: none;">
                  <div style="margin-bottom: 6px; font-size: 12px; color: #495057; display: flex; justify-content: space-between;">
                    <span><strong>Latency:</strong></span>
                    <span style="color: #f1771f; font-weight: bold; margin-left: 10px;">${latText}</span>
                  </div>
                  <div style="margin-bottom: 6px; font-size: 12px; color: #495057; display: flex; justify-content: space-between;">
                    <span><strong>Ping:</strong></span>
                    <span style="color: #48d338; font-weight: bold; margin-left: 10px;">${pingText}</span>
                  </div>
                  <div style="margin-bottom: 6px; font-size: 12px; color: #495057; display: flex; justify-content: space-between;">
                    <span><strong>Failures:</strong></span>
                    <span style="color: #e01a1a; font-weight: bold; margin-left: 10px;">${failText}</span>
                  </div>
                  <hr style="margin: 8px 0; border: 0; border-top: 1px solid #e9ecef;">
                  <div style="font-size: 10px; color: #6c757d; text-align: right;">${dt}</div>
                </div>`
              },
              fixed: {
                enabled: false, // Changed to false so the metric box floats and follows the mouse dynamically
              },
              x: {
                show: true,
              },
              y: {
                formatter: undefined,
                title: {
                  formatter: (seriesName) => seriesName,
                },
              },
            },
            legend: {
              show: true,
              position: 'bottom', // Moved to bottom to avoid overlapping the top-right toolbar controls
              horizontalAlign: 'center',
              offsetY: 10
            },
            dataLabels: {
              enabled: false
            },
            floating: true,
            axisTicks: {
              show: true
            },
            axisBorder: {
              show: false
            },
            colors: ["#f1771f", "#48d338", "#e01a1a"],
            fill: {
              colors: ["#f1771f", "#48d338", "#e01a1a"],
              opacity: [0.5, 0.4, 0.8], // Transparent areas so they don't block each other, solid bars for failures
              type: 'solid'
            },
            stroke: {
              show: true,
              curve: 'smooth', // Smoother areas look much more professional than steplines
              lineCap: 'butt',
              colors: ["#f1771f", "#48d338", "transparent"], // No line stroke for the column bars
              width: [2, 2, 0],
            }
          },
          expanded_chart_options: {
            chart: {
              id: "chart1",
              height: 130,
              type: "bar",
              foreColor: "#ccc",
              brush: {
                target: "chart2",
                enabled: true
              },
              selection: {
                enabled: true,
                fill: {
                  color: "#fff",
                  opacity: 0.4
                },
                xaxis: {
                  min: new Date("27 Jul 2017 10:00:00").getTime(),
                  max: new Date("14 Aug 2999 10:00:00").getTime()
                }
              }
            },
            colors: ["#FF0080"],
            stroke: {
              width: 2
            },
            grid: {
              borderColor: "#444"
            },
            markers: {
              size: 0
            },
            xaxis: {
              type: "datetime",
              tooltip: {
                enabled: false
              }
            },
            yaxis: {
              tickAmount: 2
            }
          }
        }
      },
      async mounted() {
        await this.update_data();
      },
      computed: {
        main_chart () {
          const list = [
            {
              name: "Latency",
              type: "area",
              ...this.convertToChartData(this.main_data)
            },
            {
              name: "Ping",
              type: "area",
              ...this.convertToChartData(this.ping_data)
            }
          ]
          if (this.failure_data) {
            list.push({
              name: "Failures",
              type: "column",
              ...this.convertToChartData(this.failure_data)
            })
          }
          return list
        },
        expanded_chart () {
          return this.toBarData(this.expanded_data)
        },
        params () {
          return {start: this.toUnix(new Date(this.start)), end: this.toUnix(new Date(this.end))}
        },
      },
      watch: {
        start: function(n, o) {
          this.update_data()
        },
        end: function(n, o) {
          this.update_data()
        },
        group: function(n, o) {
          this.update_data()
        },
      },
      methods: {
          async update_data() {
            this.ready = false
            this.loading = true
            await this.chartHits()
            // await this.expanded_hits()
            this.loading = false
            this.ready = true
          },
        async expanded_hits() {
          this.expanded_data = await this.load_hits(0, 99999999999, "24h")
        },
        async chartHits() {
          this.main_data = await this.load_hits()
          this.ping_data = await this.load_ping()
          this.failure_data = await this.load_failures()
        },
        async load_hits(start=this.params.start, end=this.params.end, group=this.group) {
          return await Api.service_hits(this.service.id, start, end, group, false)
        },
        async load_ping(start=this.params.start, end=this.params.end, group=this.group) {
          return await Api.service_ping(this.service.id, start, end, group, false)
        },
        async load_failures(start=this.params.start, end=this.params.end, group=this.group) {
          return await Api.service_failures_data(this.service.id, start, end, group, true)
        }
      }
    }
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
