<template>
    <div class="col-12 mt-4">
        <div class="row mb-5">
            <div class="col-12 text-center">
                <h2 class="font-weight-bold mb-3">Platform Documentation</h2>
                <p class="text-muted lead">Comprehensive guide to the Enterprise Monitoring Platform</p>
                <hr style="width: 60px; border-top: 3px solid #007bff; margin: 20px auto;">
            </div>
        </div>

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
        </div>
    </div>
</template>

<script>
export default {
	name: "Help",
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
