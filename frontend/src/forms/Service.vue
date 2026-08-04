<template>
  <form v-if="service.type" @submit.prevent="saveService">
    <div class="card contain-card mb-4">
      <div class="card-header">{{ $t('service_info') }}</div>
      <div class="card-body">
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">{{ $t('service_name') }}</label>
          <div class="col-sm-8">
            <input
              v-model="service.name"
              @input="updatePermalink"
              id="name"
              type="text"
              name="name"
              class="form-control"
              placeholder="Server Name"
              required
              spellcheck="false"
              autocorrect="off"
            />
            <small class="form-text text-muted">Give your service a name you can recognize</small>
          </div>
        </div>
        <div class="form-group row">
          <label for="service_type" class="col-sm-4 col-form-label">{{ $t('service_type') }}</label>
          <div class="col-sm-8">
            <select v-model="service.type" @change="updateDefaultValues()" class="form-control" id="service_type">
              <option value="cmd">Command</option>
              <option value="http">HTTP {{ $t('service') }}</option>
              <option value="tcp">TCP {{ $t('service') }}</option>
              <option value="udp">UDP {{ $t('service') }}</option>
              <option value="icmp">ICMP Ping</option>
              <option value="grpc">gRPC {{ $t('service') }}</option>
              <option value="smtp">SMTP {{ $t('service') }}</option>
              <option value="imap">IMAP {{ $t('service') }}</option>
              <option value="storage">Cloud Storage</option>
              <option value="database">Database</option>
              <option value="tls">TLS Certificate</option>
              <option value="static">Static {{ $t('service') }}</option>
            </select>
            <small class="form-text text-muted"
              >Use HTTP if you are checking a website or use TCP if you are checking a server</small
            >
          </div>
        </div>
        <div class="form-group row">
          <label for="service_type" class="col-sm-4 col-form-label">{{ $t('group') }}</label>
          <div class="col-sm-8">
            <select v-model.number="service.group_id" class="form-control">
              <option value="0">No Group</option>
              <option v-for="group in cleanGroups" :key="group.id" :value="group.id">{{ group.name }}</option>
            </select>
            <small class="form-text text-muted">Attach this service to a group</small>
          </div>
        </div>
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">{{ $t('permalink') }}</label>
          <div class="col-sm-8">
            <input
              v-model="service.permalink"
              type="text"
              name="permalink"
              class="form-control"
              id="permalink"
              autocapitalize="none"
              spellcheck="true"
              placeholder="awesome_service"
            />
            <small class="form-text text-muted">Use text for the service URL rather than the service number.</small>
          </div>
        </div>

        <div class="form-group row">
          <label for="service_priority" class="col-sm-4 col-form-label">Priority</label>
          <div class="col-sm-8">
            <select v-model.number="service.priority" class="form-control" id="service_priority">
              <option :value="1">Critical</option>
              <option :value="2">High</option>
              <option :value="3">Normal</option>
              <option :value="4">Low</option>
            </select>
            <small class="form-text text-muted">Higher priority services are checked first when the queue is busy</small>
          </div>
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">{{ $t('service_public') }}</label>
          <div class="col-12 col-md-8 mt-1 mb-2">
            <span @click="service.public = !!service.public" class="switch float-left">
              <input
                v-model="service.public"
                type="checkbox"
                name="public-option"
                class="switch"
                id="switch-public"
                :checked="service.public"
              />
              <label v-if="service.public" for="switch-public">This service will be visible for everyone</label>
              <label v-if="!service.public" for="switch-public"
                >This service will only be visible for users and administrators.</label
              >
            </span>
          </div>
        </div>

        <div v-if="service.type !== 'static'" class="form-group row">
          <label for="service_interval" class="col-sm-4 col-form-label">{{ $t('check_interval') }}</label>
          <div class="col-sm-6">
            <span class="slider-info">{{ secondsHumanize(service.check_interval) }}</span>
            <input
              v-model.number="service.check_interval"
              type="range"
              class="slider"
              id="service_interval"
              min="1"
              max="1800"
              :step="1"
            />
            <small id="interval" class="form-text text-muted">Interval to check your service state</small>
          </div>
          <div class="col-sm-2">
            <input v-model.number="service.check_interval" type="number" name="check_interval" class="form-control" />
          </div>
        </div>
      </div>
    </div>

    <div v-if="service.type !== 'static' && service.type !== 'storage' && service.type !== 'database' && service.type !== 'tls'" class="card contain-card mb-4">
      <div class="card-header">Request Details</div>
      <div class="card-body">
        <div v-if="service.type !== 'cmd'" class="form-group row">
          <label for="service_url" class="col-sm-4 col-form-label">
            {{ $t('service_endpoint') }} {{ service.type === 'http' ? '(URL)' : '(Domain)' }}
          </label>
          <div class="col-sm-8">
            <input
              v-model="service.domain"
              type="url"
              class="form-control"
              id="service_url"
              :placeholder="service.type === 'http' ? 'https://google.com' : '192.168.1.1'"
              required
              autocapitalize="none"
              spellcheck="false"
            />
            <small class="form-text text-muted">Statping will attempt to connect to this address</small>
          </div>
        </div>

        <div v-if="service.type.match(/^(tcp|udp|grpc)$/)" class="form-group row">
          <label class="col-sm-4 col-form-label">Port</label>
          <div class="col-sm-8">
            <input
              v-model.number="service.port"
              type="number"
              name="port"
              class="form-control"
              id="service_port"
              placeholder="8080"
            />
          </div>
        </div>
        <div v-if="service.type.match(/^(smtp|imap)$/)" class="form-group row">
          <label class="col-sm-4 col-form-label">Port</label>
          <div class="col-sm-8">
            <input
              v-model.number="service.port"
              type="number"
              name="port"
              class="form-control"
              id="service_port"
              placeholder="587"
            />
          </div>
        </div>

        <div v-if="service.type.match(/^(http)$/)" class="form-group row">
          <label class="col-sm-4 col-form-label">{{ $t('service_check') }}</label>
          <div class="col-sm-8">
            <select v-model="service.method" name="method" class="form-control">
              <option value="GET">GET</option>
              <option value="HEAD">HEAD</option>
              <option value="POST">POST</option>
              <option value="DELETE">DELETE</option>
              <option value="PATCH">PATCH</option>
              <option value="PUT">PUT</option>
            </select>
            <small class="form-text text-muted"
              >A GET/HEAD request will simply request the endpoint, you can also send data with POST.</small
            >
          </div>
        </div>

        <div class="form-group row">
          <label class="col-sm-4 col-form-label">{{ $t('service_timeout') }}</label>
          <div class="col-sm-6">
            <span v-if="service.timeout >= 0" class="slider-info">{{ secondsHumanize(service.timeout) }}</span>
            <input v-model.number="service.timeout" type="range" id="timeout" name="timeout" class="slider" min="1" max="180" />
            <small class="form-text text-muted"
              >If the {{ service.type === 'cmd' ? 'command' : 'endpoint' }} does not
              {{ service.type === 'cmd' ? 'exit' : 'respond' }} within this time it will be considered to be offline</small
            >
          </div>
          <div class="col-sm-2">
            <input v-model.number="service.timeout" type="number" name="service_timeout" class="form-control" />
          </div>
        </div>

        <div
          v-if="service.type === 'cmd' || (service.type.match(/^(http)$/) && service.method.match(/^(POST|PATCH|DELETE|PUT)$/))"
          class="form-group row"
        >
          <label class="col-sm-4 col-form-label">{{
            service.type === 'cmd' ? 'Command Config (JSON)' : 'Optional Post Data (JSON)'
          }}</label>
          <div class="col-sm-8">
            <textarea
              v-model="service.post_data"
              class="form-control"
              :rows="service.type === 'cmd' && !service.post_data ? '21' : service.type === 'cmd' ? '5' : '3'"
              autocapitalize="none"
              spellcheck="false"
            ></textarea>
            <small class="form-text text-muted">{{
              service.type === 'cmd' ? 'Configure the command to run.' : 'Insert a JSON string to send data to the endpoint.'
            }}</small>
          </div>
        </div>
        <div v-if="service.type.match(/^(http)$/)" class="form-group row">
          <label class="col-sm-4 col-form-label">HTTP Headers</label>
          <div class="col-sm-8">
            <input
              v-model="service.headers"
              class="form-control"
              autocapitalize="none"
              spellcheck="false"
              placeholder="Authorization=1010101,Content-Type=application/json"
            />
            <small class="form-text text-muted">Comma delimited list of HTTP Headers (KEY=VALUE,KEY=VALUE)</small>
          </div>
        </div>
        <div v-if="service.type.match(/^(smtp|imap)$/)" class="form-group row">
          <label class="col-sm-4 col-form-label">Credentials</label>
          <div class="col-sm-8">
            <input
              v-model="service.headers"
              class="form-control"
              autocapitalize="none"
              spellcheck="false"
              placeholder="Username=user@domain.com,Password=secretpassword"
            />
            <small class="form-text text-muted"
              >Comma delimited list of IMAP/SMTP credentials (Username=user@domain.com,Password=secretpassword)</small
            >
          </div>
        </div>
        <div v-if="service.type.match(/^(http)$/)" class="form-group row">
          <label class="col-sm-4 col-form-label">{{ $t('expected_resp') }} (Regex)</label>
          <div class="col-sm-8">
            <textarea
              v-model="service.expected"
              class="form-control"
              rows="3"
              autocapitalize="none"
              spellcheck="false"
            ></textarea>
            <small class="form-text text-muted"
              >You can use plain text or insert
              <a target="_blank" href="https://regex101.com/r/I5bbj9/1">Regex</a> to validate the response</small
            >
          </div>
        </div>
        <div v-if="service.type.match(/^(cmd|http)$/)" class="form-group row">
          <label for="service_response_code" class="col-sm-4 col-form-label">{{ $t('expected_code') }}</label>
          <div class="col-sm-8">
            <input
              v-model="service.expected_status"
              type="number"
              name="expected_status"
              class="form-control"
              :placeholder="service.type === 'cmd' ? '0' : '200'"
              id="service_response_code"
            />
            <small v-if="service.type === 'cmd'" class="form-text text-muted">An exit code of 0 is success</small>
            <small v-if="service.type === 'http'" class="form-text text-muted"
              >A status code of 200 is success, or view all the
              <a target="_blank" href="https://www.restapitutorial.com/httpstatuscodes.html">HTTP Status Codes</a></small
            >
          </div>
        </div>

        <div v-if="service.type.match(/^(http)$/)" class="form-group row">
          <label class="col-12 col-md-4 col-form-label">{{ $t('follow_redir') }}</label>
          <div class="col-12 col-md-8 mt-1 mb-2 mb-md-0">
            <span @click="service.redirect = !!service.redirect" class="switch float-left">
              <input
                v-model="service.redirect"
                type="checkbox"
                name="redirect-option"
                class="switch"
                id="switch-redirect"
                :checked="service.redirect"
              />
              <label for="switch-redirect">Follow HTTP Redirects if server attempts</label>
            </span>
          </div>
        </div>
        <div v-if="service.type.match(/^(http|grpc|smtp|imap)$/)" class="form-group row">
          <label class="col-12 col-md-4 col-form-label">{{ $t('verify_ssl') }}</label>
          <div class="col-12 col-md-8 mt-1 mb-2 mb-md-0">
            <span @click="service.verify_ssl = !!service.verify_ssl" class="switch float-left">
              <input
                v-model="service.verify_ssl"
                type="checkbox"
                name="verify_ssl-option"
                class="switch"
                id="switch-verify-ssl"
                :checked="service.verify_ssl"
              />
              <label for="switch-verify-ssl" v-if="service.verify_ssl">Verify SSL Certificate for this service</label>
              <label for="switch-verify-ssl" v-if="!service.verify_ssl"
                >Skip SSL Certificate verification for this service</label
              >
            </span>
          </div>
        </div>

        <div v-if="service.type.match(/^(grpc)$/)" class="form-group row">
          <label class="col-12 col-md-4 col-form-label"
            ><a href="https://github.com/grpc/grpc/blob/master/doc/health-checking.md#grpc-health-checking-protocol"
              >GRPC Health Check</a
            ></label
          >
          <div class="col-12 col-md-8 mt-1 mb-2 mb-md-0">
            <span @click="service.grpc_health_check = !!service.grpc_health_check" class="switch float-left">
              <input
                v-model="service.grpc_health_check"
                type="checkbox"
                name="grpc_health_check-option"
                class="switch"
                id="switch-grpc-health-check"
                :checked="service.grpc_health_check"
              />
              <label for="switch-grpc-health-check" v-if="service.grpc_health_check"
                >Check against GRPC health check endpoint.</label
              >
              <label for="switch-grpc-health-check" v-if="!service.grpc_health_check"
                >Only checks if GRPC connection can be established.</label
              >
            </span>
          </div>
        </div>

        <div v-if="service.grpc_health_check" class="form-group row">
          <label class="col-sm-4 col-form-label">Expected Response</label>
          <div class="col-sm-8">
            <textarea
              v-model="service.expected"
              class="form-control"
              rows="3"
              autocapitalize="none"
              spellcheck="false"
              placeholder="status:SERVING"
            ></textarea>
            <small class="form-text text-muted"
              >Check
              <a target="_blank" href="https://pkg.go.dev/google.golang.org/grpc/health/grpc_health_v1?tab=doc#pkg-variables"
                >GPRC health check response codes</a
              >
              for more information.</small
            >
          </div>
        </div>

        <div v-if="service.grpc_health_check" class="form-group row">
          <label for="service_response_code" class="col-sm-4 col-form-label">Expected Status Code</label>
          <div class="col-sm-8">
            <input
              v-model="service.expected_status"
              type="number"
              name="expected_status"
              class="form-control"
              placeholder="1"
              id="service_response_code"
            />
            <small class="form-text text-muted"
              >A status code of 1 is success, or view all the
              <a
                target="_blank"
                href="https://pkg.go.dev/google.golang.org/grpc/health/grpc_health_v1?tab=doc#HealthCheckResponse_ServingStatus"
                >GRPC Status Codes</a
              ></small
            >
          </div>
        </div>

        <div v-if="service.type.match(/^(tcp|smtp|imap|http)$/)" class="form-group row">
          <label class="col-12 col-md-4 col-form-label">{{ $t('tls_cert') }}</label>
          <div class="col-12 col-md-8 mt-1 mb-2 mb-md-0">
            <span @click="use_tls = !!use_tls" class="switch float-left">
              <input v-model="use_tls" type="checkbox" name="verify_ssl-option" class="switch" id="switch-use-tls" :checked="use_tls" />
              <label for="switch-use-tls" v-if="use_tls">Custom TLS Certificates for mTLS services</label>
              <label for="switch-use-tls" v-if="!use_tls">Ignore TLS Certificates</label>
            </span>
          </div>
        </div>

        <div v-if="use_tls" class="form-group row">
          <label for="service_tls_cert" class="col-sm-4 col-form-label">TLS Client Certificate</label>
          <div class="col-sm-8">
            <textarea v-model="service.tls_cert" name="tls_cert" class="form-control" id="service_tls_cert"></textarea>
            <small class="form-text text-muted">Absolute path to TLS Client Certificate file or in PEM format</small>
          </div>
        </div>

        <div v-if="use_tls" class="form-group row">
          <label for="service_tls_cert_key" class="col-sm-4 col-form-label">TLS Client Key</label>
          <div class="col-sm-8">
            <textarea
              v-model="service.tls_cert_key"
              name="tls_cert_key"
              class="form-control"
              id="service_tls_cert_key"
            ></textarea>
            <small class="form-text text-muted">Absolute path to TLS Client Key file or in PEM format</small>
          </div>
        </div>

        <div v-if="use_tls" class="form-group row">
          <label for="service_tls_cert_chain" class="col-sm-4 col-form-label">Root CA</label>
          <div class="col-sm-8">
            <textarea
              v-model="service.tls_cert_root"
              name="tls_cert_key"
              class="form-control"
              id="service_tls_cert_chain"
            ></textarea>
            <small class="form-text text-muted">Absolute path to Root CA file or in PEM format (optional)</small>
          </div>
        </div>
      </div>
    </div>

    <!-- Cloud Storage Service Configuration -->
    <div v-if="service.type === 'storage'" class="card contain-card mb-4">
      <div class="card-header">Cloud Storage Configuration</div>
      <div class="card-body">
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Storage Backend</label>
          <div class="col-sm-8">
            <select v-model="service.storage_backend" class="form-control">
              <option value="gcs">Google Cloud Storage (GCS)</option>
              <option value="s3">Amazon S3</option>
              <option value="azureblob">Azure Blob Storage</option>
              <option value="sftp">SFTP</option>
              <option value="b2">Backblaze B2</option>
              <option value="dropbox">Dropbox</option>
              <option value="minio">MinIO (S3-compatible)</option>
            </select>
            <small class="form-text text-muted">Select the cloud storage provider (rclone backend)</small>
          </div>
        </div>
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Bucket / Container</label>
          <div class="col-sm-8">
            <input
              v-model="service.storage_bucket"
              type="text"
              class="form-control"
              placeholder="my-bucket-name"
              required
            />
            <small class="form-text text-muted">Name of the bucket or container to check</small>
          </div>
        </div>
        <div v-if="service.storage_backend && service.storage_backend.match(/^(s3|minio|azureblob)$/)" class="form-group row">
          <label class="col-sm-4 col-form-label">Region</label>
          <div class="col-sm-8">
            <input
              v-model="service.storage_region"
              type="text"
              class="form-control"
              placeholder="us-east-1"
            />
            <small class="form-text text-muted">Cloud region (e.g., us-east-1, westus2)</small>
          </div>
        </div>
        <div v-if="service.storage_backend && service.storage_backend.match(/^(s3|minio)$/)" class="form-group row">
          <label class="col-sm-4 col-form-label">Custom Endpoint</label>
          <div class="col-sm-8">
            <input
              v-model="service.storage_endpoint"
              type="text"
              class="form-control"
              placeholder="https://minio.example.com:9000"
            />
            <small class="form-text text-muted">Custom endpoint for S3-compatible storage (MinIO, etc.)</small>
          </div>
        </div>
        <div v-if="service.storage_backend === 'gcs'" class="form-group row">
          <label class="col-sm-4 col-form-label">GCP Project ID</label>
          <div class="col-sm-8">
            <input
              v-model="service.storage_project_id"
              type="text"
              class="form-control"
              placeholder="my-gcp-project"
            />
            <small class="form-text text-muted">Google Cloud Project ID</small>
          </div>
        </div>
        <div v-if="service.storage_backend === 'gcs'" class="form-group row">
          <label class="col-sm-4 col-form-label">Credentials File</label>
          <div class="col-sm-8">
            <input
              v-model="service.storage_cred_file"
              type="text"
              class="form-control"
              placeholder="/path/to/service-account.json"
            />
            <small class="form-text text-muted">Path to GCS service account JSON key file</small>
          </div>
        </div>
        <div v-if="service.storage_backend === 'gcs'" class="form-group row">
          <label class="col-12 col-md-4 col-form-label">Application Default Credentials</label>
          <div class="col-12 col-md-8 mt-1 mb-2 mb-md-0">
            <span class="switch float-left">
              <input
                v-model="service.storage_allow_adc"
                type="checkbox"
                class="switch"
                id="switch-adc"
                :checked="service.storage_allow_adc"
              />
              <label for="switch-adc">Allow Application Default Credentials (ADC)</label>
            </span>
          </div>
        </div>
        <div v-if="service.storage_backend && service.storage_backend.match(/^(s3|minio|azureblob|b2)$/)" class="form-group row">
          <label class="col-sm-4 col-form-label">Access Key / Client ID</label>
          <div class="col-sm-8">
            <input
              v-model="service.storage_access_key"
              type="text"
              class="form-control"
              placeholder="AKIAIOSFODNN7EXAMPLE"
            />
            <small class="form-text text-muted">Access key or client ID (encrypted at rest)</small>
          </div>
        </div>
        <div v-if="service.storage_backend && service.storage_backend.match(/^(s3|minio|azureblob|b2)$/)" class="form-group row">
          <label class="col-sm-4 col-form-label">Secret Key</label>
          <div class="col-sm-8">
            <input
              v-model="service.storage_secret_key"
              type="password"
              class="form-control"
              placeholder="••••••••••••••••"
            />
            <small class="form-text text-muted">Secret key (encrypted at rest)</small>
          </div>
        </div>
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">{{ $t('service_timeout') }}</label>
          <div class="col-sm-6">
            <span class="slider-info">{{ secondsHumanize(service.timeout) }}</span>
            <input v-model.number="service.timeout" type="range" class="slider" min="1" max="120" />
            <small class="form-text text-muted">Timeout for storage connectivity check</small>
          </div>
          <div class="col-sm-2">
            <input v-model.number="service.timeout" type="number" class="form-control" />
          </div>
        </div>
      </div>
    </div>

    <!-- Database Service Configuration -->
    <div v-if="service.type === 'database'" class="card contain-card mb-4">
      <div class="card-header">Database Configuration</div>
      <div class="card-body">
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Database Type</label>
          <div class="col-sm-8">
            <select v-model="service.database_type" class="form-control">
              <option value="postgres">PostgreSQL</option>
              <option value="mysql">MySQL / MariaDB</option>
              <option value="sqlite">SQLite</option>
              <option value="sqlserver">SQL Server</option>
              <option value="mongodb">MongoDB</option>
              <option value="oracle">Oracle</option>
            </select>
            <small class="form-text text-muted">Select the database engine</small>
          </div>
        </div>
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Connection String (DSN)</label>
          <div class="col-sm-8">
            <input
              v-model="service.database_dsn"
              type="password"
              class="form-control"
              :placeholder="getDsnPlaceholder(service.database_type)"
              required
            />
            <small class="form-text text-muted">Database connection string (encrypted at rest)</small>
          </div>
        </div>
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Health Check Query</label>
          <div class="col-sm-8">
            <input
              v-model="service.database_query"
              type="text"
              class="form-control"
              placeholder="SELECT 1"
            />
            <small class="form-text text-muted">Query to verify database connectivity (default: SELECT 1)</small>
          </div>
        </div>
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">{{ $t('service_timeout') }}</label>
          <div class="col-sm-6">
            <span class="slider-info">{{ secondsHumanize(service.timeout) }}</span>
            <input v-model.number="service.timeout" type="range" class="slider" min="1" max="120" />
            <small class="form-text text-muted">Timeout for database connectivity check</small>
          </div>
          <div class="col-sm-2">
            <input v-model.number="service.timeout" type="number" class="form-control" />
          </div>
        </div>
      </div>
    </div>

    <!-- TLS Certificate Service Configuration -->
    <div v-if="service.type === 'tls'" class="card contain-card mb-4">
      <div class="card-header">TLS Certificate Configuration</div>
      <div class="card-body">
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Host / Domain</label>
          <div class="col-sm-8">
            <input
              v-model="service.domain"
              type="text"
              class="form-control"
              placeholder="example.com"
              required
            />
            <small class="form-text text-muted">Domain or hostname to check TLS certificate</small>
          </div>
        </div>
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Port</label>
          <div class="col-sm-8">
            <input
              v-model.number="service.port"
              type="number"
              class="form-control"
              placeholder="443"
            />
            <small class="form-text text-muted">TLS port (default: 443)</small>
          </div>
        </div>
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Minimum Days Before Expiry</label>
          <div class="col-sm-6">
            <span class="slider-info">{{ service.tls_min_days }} days</span>
            <input v-model.number="service.tls_min_days" type="range" class="slider" min="0" max="90" />
            <small class="form-text text-muted">Alert if certificate expires within this many days (0 to disable)</small>
          </div>
          <div class="col-sm-2">
            <input v-model.number="service.tls_min_days" type="number" class="form-control" />
          </div>
        </div>
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">Expected SAN</label>
          <div class="col-sm-8">
            <input
              v-model="service.tls_expected_san"
              type="text"
              class="form-control"
              placeholder="*.example.com"
            />
            <small class="form-text text-muted">Expected Subject Alternative Name (optional)</small>
          </div>
        </div>
        <div class="form-group row">
          <label class="col-12 col-md-4 col-form-label">{{ $t('verify_ssl') }}</label>
          <div class="col-12 col-md-8 mt-1 mb-2 mb-md-0">
            <span class="switch float-left">
              <input
                v-model="service.verify_ssl"
                type="checkbox"
                class="switch"
                id="switch-tls-verify"
                :checked="service.verify_ssl"
              />
              <label for="switch-tls-verify" v-if="service.verify_ssl">Verify certificate chain</label>
              <label for="switch-tls-verify" v-if="!service.verify_ssl">Skip certificate chain verification</label>
            </span>
          </div>
        </div>
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">{{ $t('service_timeout') }}</label>
          <div class="col-sm-6">
            <span class="slider-info">{{ secondsHumanize(service.timeout) }}</span>
            <input v-model.number="service.timeout" type="range" class="slider" min="1" max="60" />
            <small class="form-text text-muted">Timeout for TLS handshake</small>
          </div>
          <div class="col-sm-2">
            <input v-model.number="service.timeout" type="number" class="form-control" />
          </div>
        </div>
      </div>
    </div>

    <div class="card contain-card mb-4">
      <div class="card-header">{{ $t('notification_opts') }}</div>
      <div class="card-body">
        <div class="form-group row">
          <label class="col-sm-4 col-form-label">{{ $t('notifications_enable') }}</label>
          <div class="col-12 col-md-8 mt-1 mb-2 mb-md-0">
            <span @click="service.allow_notifications = !!service.allow_notifications" class="switch float-left">
              <input
                v-model="service.allow_notifications"
                type="checkbox"
                name="allow_notifications-option"
                class="switch"
                id="switch-notifications"
                :checked="service.allow_notifications"
              />
              <label for="switch-notifications">Allow notifications to be sent for this service</label>
            </span>
          </div>
        </div>
        <div v-if="service.allow_notifications" class="form-group row">
          <label class="col-sm-4 col-form-label">{{ $t('notify_after') }}</label>
          <div class="col-sm-8">
            <span class="slider-info">{{
              service.notify_after === 0 ? 'First Failure' : service.notify_after + ' Failures'
            }}</span>
            <input v-model="service.notify_after" type="range" name="notify_after" class="slider" id="notify_after" min="0" max="20" />
            <small class="form-text text-muted"
              >Send Notification after
              {{ service.notify_after === 0 ? 'the first Failure' : service.notify_after + ' Failures' }}
            </small>
          </div>
        </div>
        <div v-if="service.allow_notifications" class="form-group row">
          <label class="col-sm-4 col-form-label">{{ $t('notify_all') }}</label>
          <div class="col-12 col-md-8 mt-1">
            <span @click="service.notify_all_changes = !!service.notify_all_changes" class="switch float-left">
              <input
                v-model="service.notify_all_changes"
                type="checkbox"
                name="notify_all-option"
                class="switch"
                id="notify_all"
                :checked="service.notify_all_changes"
              />
              <label v-if="service.notify_all_changes" for="notify_all"
                >Continuously send notifications when service is failing.</label
              >
              <label v-if="!service.notify_all_changes" for="notify_all"
                >Only notify one time when service hits an error</label
              >
            </span>
          </div>
        </div>
      </div>
    </div>

    <div class="card contain-card mb-4" v-if="service.type !== 'static'">
      <div class="card-header">Connection Test</div>
      <div class="card-body">
        <p class="text-muted small mb-3">Test connectivity and credentials before saving</p>
        <button type="button" @click="testConnection" :disabled="testing" class="btn btn-outline-primary">
          <span v-if="testing" class="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
          {{ testing ? 'Testing...' : 'Test Connection' }}
        </button>
        <div v-if="testResult" class="mt-3">
          <div :class="['alert', testResult.success ? 'alert-success' : 'alert-danger']" role="alert">
            <strong>{{ testResult.message }}</strong>
            <div v-if="testResult.latency" class="mt-1">
              <small>Latency: {{ Math.round(testResult.latency / 1000) }}ms</small>
            </div>
            <div v-if="testResult.info && Object.keys(testResult.info).length > 0" class="mt-2">
              <small v-for="(value, key) in testResult.info" :key="key" class="d-block">
                <strong>{{ formatInfoKey(key) }}:</strong> {{ value }}
              </small>
            </div>
            <div v-if="testResult.details && !testResult.success" class="mt-2">
              <small class="text-break">{{ truncateDetails(testResult.details) }}</small>
            </div>
            <div v-if="testResult.details && testResult.success" class="mt-2">
              <small class="text-muted">Response Preview:</small>
              <pre class="response-preview mt-1 mb-1">{{ showFullResponse ? testResult.details : truncateLines(testResult.details, 8) }}</pre>
              <button
                v-if="countLines(testResult.details) > 8"
                type="button"
                @click="showFullResponse = !showFullResponse"
                class="btn btn-sm btn-link p-0"
              >
                {{ showFullResponse ? 'Show less' : `Show all ${countLines(testResult.details)} lines` }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="form-group row">
      <div class="col-6">
        <button type="button" @click="cancelEdit" class="btn btn-outline-secondary btn-block">
          Cancel
        </button>
      </div>
      <div class="col-6">
        <button :disabled="loading" @click.prevent="saveService" type="submit" class="btn btn-success btn-block">
          {{ service.id ? $t('service_update') : $t('service_create') }}
        </button>
      </div>
    </div>

    <div class="alert alert-danger d-none" id="alerter" role="alert"></div>
  </form>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useMainStore } from '@/stores/main'
import Api from '@/API'

const props = defineProps({
  in_service: {
    type: Object,
    default: null,
  },
})

const router = useRouter()
const store = useMainStore()

const loading = ref(false)
const use_tls = ref(false)
const testing = ref(false)
const testResult = ref(null)
const showFullResponse = ref(false)

const service = reactive({
  name: '',
  type: 'cmd',
  domain: '',
  group_id: 0,
  method: 'GET',
  post_data: '',
  headers: '',
  expected: '',
  expected_status: 0,
  port: 80,
  check_interval: 60,
  timeout: 15,
  permalink: '',
  order: 1,
  priority: 3,
  verify_ssl: true,
  grpc_health_check: false,
  redirect: true,
  allow_notifications: true,
  notify_all_changes: true,
  notify_after: 2,
  public: true,
  tls_cert: '',
  tls_cert_key: '',
  tls_cert_root: '',
  // Cloud Storage fields
  storage_backend: '',
  storage_bucket: '',
  storage_region: '',
  storage_endpoint: '',
  storage_access_key: '',
  storage_secret_key: '',
  storage_cred_file: '',
  storage_project_id: '',
  storage_allow_adc: false,
  // Database fields
  database_type: '',
  database_dsn: '',
  database_query: '',
  // TLS Certificate fields
  tls_min_days: 30,
  tls_expected_san: '',
})

const cleanGroups = computed(() => store.groupsClean || [])

watch(
  () => props.in_service,
  (svr) => {
    if (svr) {
      Object.assign(service, svr)
      use_tls.value = !!svr.tls_cert
    }
  },
  { immediate: true }
)

onMounted(async () => {
  if (!store.groups) {
    const groups = await Api.groups()
    store.setGroups(groups)
  }
  update()
})

function update() {
  if (props.in_service) {
    Object.assign(service, props.in_service)
  }
  use_tls.value = service.tls_cert !== ''
}

function secondsHumanize(seconds) {
  if (seconds < 60) return `${seconds} seconds`
  if (seconds < 3600) return `${Math.floor(seconds / 60)} minutes`
  return `${Math.floor(seconds / 3600)} hours`
}

function updateDefaultValues() {
  if (service.type === 'cmd') {
    service.expected_status = 0
    service.expected = ''
    service.port = 0
    service.verify_ssl = false
    service.method = ''
  } else if (service.type === 'grpc') {
    service.expected_status = 1
    service.expected = 'status:SERVING'
    service.port = 50051
    service.verify_ssl = false
    service.method = ''
  } else if (service.type === 'storage') {
    service.timeout = 30
    service.storage_backend = 'gcs'
  } else if (service.type === 'database') {
    service.timeout = 30
    service.database_type = 'postgres'
    service.database_query = 'SELECT 1'
  } else if (service.type === 'tls') {
    service.port = 443
    service.timeout = 10
    service.verify_ssl = true
    service.tls_min_days = 30
  } else {
    service.expected_status = 200
    service.expected = ''
    service.port = 80
    service.verify_ssl = true
    service.method = 'GET'
  }
}

function getDsnPlaceholder(dbType) {
  const placeholders = {
    postgres: 'postgres://user:pass@localhost:5432/dbname?sslmode=disable',
    mysql: 'user:pass@tcp(localhost:3306)/dbname',
    sqlite: 'file:/path/to/database.db',
    sqlserver: 'sqlserver://user:pass@localhost:1433?database=dbname',
    mongodb: 'mongodb://user:pass@localhost:27017/dbname',
    oracle: 'oracle://user:pass@localhost:1521/service',
  }
  return placeholders[dbType] || 'connection-string'
}

function updatePermalink() {
  const a = 'àáâäæãåāăąçćčđďèéêëēėęěğǵḧîïíīįìłḿñńǹňôöòóœøōõőṕŕřßśšşșťțûüùúūǘůűųẃẍÿýžźż·/_,:;'
  const b = 'aaaaaaaaaacccddeeeeeeeegghiiiiiilmnnnnoooooooooprrsssssttuuuuuuuuuwxyyzzz------'
  const p = new RegExp(a.split('').join('|'), 'g')

  service.permalink = service.name
    .toLowerCase()
    .replace(/\s+/g, '-')
    .replace(p, (c) => b.charAt(a.indexOf(c)))
    .replace(/&/g, '-and-')
    .replace(/[^\w-]+/g, '')
    .replace(/--+/g, '-')
    .replace(/^-+/, '')
    .replace(/-+$/, '')
}

function cancelEdit() {
  router.push('/dashboard/services')
}

async function saveService() {
  const s = { ...service }
  loading.value = true
  delete s.failures
  delete s.created_at
  delete s.updated_at
  delete s.last_success
  delete s.latency
  delete s.online_24_hours
  s.check_interval = parseInt(s.check_interval, 10)
  s.timeout = parseInt(s.timeout, 10)
  s.port = parseInt(s.port, 10)
  s.notify_after = parseInt(s.notify_after, 10)
  s.expected_status = parseInt(s.expected_status, 10)
  s.order = parseInt(s.order, 10)

  if (s.id) {
    await Api.service_update(s)
  } else {
    await Api.service_create(s)
  }
  const services = await Api.services()
  store.setServices(services)
  loading.value = false
  router.push('/dashboard/services')
}

function truncateDetails(details, maxLen = 500) {
  if (!details || details.length <= maxLen) return details
  return details.substring(0, maxLen) + '...'
}

function truncateLines(text, maxLines = 8) {
  if (!text) return ''
  const lines = text.split('\n')
  if (lines.length <= maxLines) return text
  return lines.slice(0, maxLines).join('\n') + '\n...'
}

function countLines(text) {
  if (!text) return 0
  return text.split('\n').length
}

function formatInfoKey(key) {
  return key.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase())
}

async function testConnection() {
  testing.value = true
  testResult.value = null
  showFullResponse.value = false
  try {
    const s = { ...service }
    s.check_interval = parseInt(s.check_interval, 10) || 60
    s.timeout = parseInt(s.timeout, 10) || 15
    s.port = parseInt(s.port, 10) || 0
    s.expected_status = parseInt(s.expected_status, 10) || 0
    const result = await Api.service_test(s)
    testResult.value = result
  } catch (err) {
    testResult.value = {
      success: false,
      message: 'Test failed',
      details: err.message || 'An error occurred while testing the connection'
    }
  } finally {
    testing.value = false
  }
}
</script>

<style scoped>
.response-preview {
  background: rgba(0, 0, 0, 0.05);
  padding: 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  max-height: 300px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
