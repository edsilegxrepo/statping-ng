<template>
  <div>
    <TopNav />
    <PageHeader title="Platform Documentation" description="Comprehensive guide to the Enterprise Monitoring Platform" />
    <div class="col-12 mt-4">

        <div class="row">
            <!-- Platform Overview -->
            <div class="col-md-6 mb-4">
                <div class="card h-100 shadow-sm border-0">
                    <div class="card-body p-4">
                        <div class="d-flex align-items-center mb-3">
                            <div class="bg-primary-soft p-3 rounded mr-3">
                                <font-awesome-icon icon="info-circle" class="text-primary" size="2x"/>
                            </div>
                            <h4 class="mb-0 font-weight-bold">Platform Overview</h4>
                        </div>
                        <p class="text-muted">An advanced, open-source monitoring solution designed for high-availability enterprise environments. It provides real-time visibility into the health and performance of your entire digital infrastructure.</p>
                        <ul class="list-unstyled mt-3 small text-muted">
                            <li class="mb-2"><font-awesome-icon icon="check" class="text-success mr-2"/> Real-time health tracking</li>
                            <li class="mb-2"><font-awesome-icon icon="check" class="text-success mr-2"/> Public status page for transparency</li>
                            <li class="mb-2"><font-awesome-icon icon="check" class="text-success mr-2"/> Historical uptime reporting</li>
                        </ul>
                    </div>
                </div>
            </div>

            <!-- Architecture -->
            <div class="col-md-6 mb-4">
                <div class="card h-100 shadow-sm border-0">
                    <div class="card-body p-4">
                        <div class="d-flex align-items-center mb-3">
                            <div class="bg-primary-soft p-3 rounded mr-3">
                                <font-awesome-icon icon="layer-group" class="text-primary" size="2x"/>
                            </div>
                            <h4 class="mb-0 font-weight-bold">Architecture</h4>
                        </div>
                        <p class="text-muted">Built on a highly efficient, multi-threaded Go engine, the platform utilizes a stateless design for maximum reliability and non-blocking performance.</p>
                        <ul class="list-unstyled mt-3 small text-muted">
                            <li class="mb-2"><font-awesome-icon icon="microchip" class="text-primary mr-2"/> <strong>Polling Engine:</strong> Multi-threaded Go routines</li>
                            <li class="mb-2"><font-awesome-icon icon="database" class="text-primary mr-2"/> <strong>Storage:</strong> Support for SQLite, MySQL, and Postgres</li>
                            <li class="mb-2"><font-awesome-icon icon="shield-alt" class="text-primary mr-2"/> <strong>DNS:</strong> Pure-Go resolver with cache bypass</li>
                        </ul>
                    </div>
                </div>
            </div>

            <!-- Monitoring Types -->
            <div class="col-md-6 mb-4">
                <div class="card h-100 shadow-sm border-0">
                    <div class="card-body p-4">
                        <div class="d-flex align-items-center mb-3">
                            <div class="bg-primary-soft p-3 rounded mr-3">
                                <font-awesome-icon icon="broadcast-tower" class="text-primary" size="2x"/>
                            </div>
                            <h4 class="mb-0 font-weight-bold">Monitor Types</h4>
                        </div>
                        <p class="text-muted">A versatile array of monitor types to ensure every layer of your stack is covered from HTTP to raw network protocols.</p>
                        <div class="row small text-muted mt-3">
                            <div class="col-6">
                                <p><strong>HTTP/HTTPS:</strong> Status codes, Regex matching</p>
                                <p><strong>TCP/UDP:</strong> Port connectivity</p>
                            </div>
                            <div class="col-6">
                                <p><strong>gRPC:</strong> Health check protocol support</p>
                                <p><strong>ICMP:</strong> Latency and packet loss</p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Alerts & Notifications -->
            <div class="col-md-6 mb-4">
                <div class="card h-100 shadow-sm border-0">
                    <div class="card-body p-4">
                        <div class="d-flex align-items-center mb-3">
                            <div class="bg-primary-soft p-3 rounded mr-3">
                                <font-awesome-icon icon="bell" class="text-primary" size="2x"/>
                            </div>
                            <h4 class="mb-0 font-weight-bold">Alerting & Alerts</h4>
                        </div>
                        <p class="text-muted">Intelligent notification system that triggers instantly upon service failure, ensuring your team is the first to know.</p>
                        <div class="d-flex flex-wrap mt-3">
                            <span class="badge badge-light p-2 mr-2 mb-2">Email</span>
                            <span class="badge badge-light p-2 mr-2 mb-2">Slack</span>
                            <span class="badge badge-light p-2 mr-2 mb-2">Discord</span>
                            <span class="badge badge-light p-2 mr-2 mb-2">Telegram</span>
                            <span class="badge badge-light p-2 mr-2 mb-2">Webhooks</span>
                            <span class="badge badge-light p-2 mr-2 mb-2">Twilio</span>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Additional Functionality -->
            <div class="col-12 mb-4">
                <div class="card shadow-sm border-0 bg-light">
                    <div class="card-body p-4">
                        <h4 class="font-weight-bold mb-4">Core Functionality</h4>
                        <div class="row">
                            <div class="col-md-4 mb-3">
                                <h6><font-awesome-icon icon="users" class="text-primary mr-2"/> User Management</h6>
                                <p class="small text-muted">Role-based access control (RBAC) to manage team access levels.</p>
                            </div>
                            <div class="col-md-4 mb-3">
                                <h6><font-awesome-icon icon="calendar-alt" class="text-primary mr-2"/> Incidents & Maintenance</h6>
                                <p class="small text-muted">Communicate scheduled downtime and track incident resolution.</p>
                            </div>
                            <div class="col-md-4 mb-3">
                                <h6><font-awesome-icon icon="terminal" class="text-primary mr-2"/> API & Integrations</h6>
                                <p class="small text-muted">Full REST API for custom integrations and automated management.</p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Understanding Response Times -->
            <div class="col-md-6 mb-4">
                <div class="card h-100 shadow-sm border-0">
                    <div class="card-body p-4">
                        <div class="d-flex align-items-center mb-3">
                            <div class="bg-primary-soft p-3 rounded mr-3">
                                <font-awesome-icon icon="stopwatch" class="text-primary" size="2x"/>
                            </div>
                            <h4 class="mb-0 font-weight-bold">Response Times</h4>
                        </div>
                        <p class="text-muted small">How to interpret latency values:</p>
                        <table class="table table-sm small mb-0">
                            <tbody>
                                <tr>
                                    <td><span class="badge bg-success text-white">&lt; 100ms</span></td>
                                    <td>Excellent - Fast response</td>
                                </tr>
                                <tr>
                                    <td><span class="badge bg-primary text-white">100-300ms</span></td>
                                    <td>Good - Normal for most services</td>
                                </tr>
                                <tr>
                                    <td><span class="badge bg-warning text-dark">300-500ms</span></td>
                                    <td>Acceptable - May indicate load</td>
                                </tr>
                                <tr>
                                    <td><span class="badge bg-danger text-white">&gt; 500ms</span></td>
                                    <td>Slow - Investigate if persistent</td>
                                </tr>
                            </tbody>
                        </table>
                        <p class="text-muted small mt-3 mb-0"><strong>Note:</strong> Expected latency varies by service type. Database queries may be faster than external API calls.</p>
                    </div>
                </div>
            </div>

            <!-- Uptime SLA Reference -->
            <div class="col-md-6 mb-4">
                <div class="card h-100 shadow-sm border-0">
                    <div class="card-body p-4">
                        <div class="d-flex align-items-center mb-3">
                            <div class="bg-primary-soft p-3 rounded mr-3">
                                <font-awesome-icon icon="percentage" class="text-primary" size="2x"/>
                            </div>
                            <h4 class="mb-0 font-weight-bold">Uptime SLA Reference</h4>
                        </div>
                        <p class="text-muted small">What uptime percentages mean in practice:</p>
                        <table class="table table-sm small mb-0">
                            <thead>
                                <tr>
                                    <th>Uptime</th>
                                    <th>Downtime/Month</th>
                                    <th>Downtime/Year</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr>
                                    <td><strong>99%</strong></td>
                                    <td>~7 hours</td>
                                    <td>~3.6 days</td>
                                </tr>
                                <tr>
                                    <td><strong>99.9%</strong></td>
                                    <td>~44 minutes</td>
                                    <td>~8.7 hours</td>
                                </tr>
                                <tr>
                                    <td><strong>99.95%</strong></td>
                                    <td>~22 minutes</td>
                                    <td>~4.4 hours</td>
                                </tr>
                                <tr>
                                    <td><strong>99.99%</strong></td>
                                    <td>~4 minutes</td>
                                    <td>~52 minutes</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>

            <!-- Reading the Status Page | FAQ side by side -->
            <div class="col-md-6 mb-4">
                <div class="card h-100 shadow-sm border-0">
                    <div class="card-body p-4">
                        <h4 class="font-weight-bold mb-4"><font-awesome-icon icon="eye" class="text-primary mr-2"/>Reading the Status Page</h4>
                        <h6>Status Indicators</h6>
                        <ul class="list-unstyled small mb-3">
                            <li class="mb-2"><span class="badge bg-success text-white">Online</span> Service responding normally</li>
                            <li class="mb-2"><span class="badge bg-danger text-white">Offline</span> Service not responding</li>
                            <li class="mb-2"><span class="badge bg-warning text-dark">Degraded</span> Slow or partial errors</li>
                        </ul>
                        <h6>Top Banner Metrics</h6>
                        <ul class="small text-muted mb-3">
                            <li><strong>X/Y Online:</strong> Services currently up</li>
                            <li><strong>XX% (24h):</strong> Average uptime</li>
                            <li><strong>Last Check:</strong> Most recent check time</li>
                        </ul>
                        <h6>Service Details</h6>
                        <p class="small text-muted mb-0">Click any service to view response times, uptime charts, failure history, and statistics.</p>
                    </div>
                </div>
            </div>

            <div class="col-md-6 mb-4">
                <div class="card h-100 shadow-sm border-0 bg-light">
                    <div class="card-body p-4">
                        <h4 class="font-weight-bold mb-4"><font-awesome-icon icon="question-circle" class="text-primary mr-2"/>Frequently Asked Questions</h4>
                        <p class="mb-1"><strong>How often are services checked?</strong></p>
                        <p class="small text-muted mb-3">Each service has a configurable interval, typically 30 seconds to 5 minutes.</p>

                        <p class="mb-1"><strong>How is uptime calculated?</strong></p>
                        <p class="small text-muted mb-3">Uptime = (successful checks / total checks) × 100% over the selected period.</p>

                        <p class="mb-1"><strong>Why might a service show as offline?</strong></p>
                        <p class="small text-muted mb-3">Network issues, server errors (5xx), timeouts, or the service is genuinely down.</p>

                        <p class="mb-1"><strong>Can I subscribe to updates?</strong></p>
                        <p class="small text-muted mb-0">Contact your administrator to be added to notification channels.</p>
                    </div>
                </div>
            </div>
        </div>
    </div>
  </div>
</template>

<script>
import PageHeader from "@/components/Elements/PageHeader.vue";
import TopNav from "@/components/Index/TopNav.vue";

export default {
	name: "Help",
	components: { TopNav, PageHeader },
};
</script>

<style scoped>
.bg-primary-soft {
    background-color: rgba(0, 123, 255, 0.1);
}
.card {
    transition: transform 0.2s ease-in-out, box-shadow 0.2s ease-in-out;
    border-radius: 12px;
}
.card:hover {
    transform: translateY(-5px);
    box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.08) !important;
}
.badge-light {
    background-color: #f8f9fa;
    border: 1px solid #dee2e6;
    color: #495057;
}
</style>
