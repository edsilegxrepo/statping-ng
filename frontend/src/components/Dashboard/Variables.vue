<template>
  <div>
    <h2>Notifier Variables</h2>
    You can insert dynamic fields within the notifier payloads for some notifiers.

    <p class="mt-2">
      Checkout the
      <a href="https://github.com/statping-ng/statping-ng/blob/master/types/services/struct.go">Service struct</a> and the
      <a href="https://github.com/statping-ng/statping-ng/blob/master/types/failures/struct.go">Failures struct</a> and
      create variables in golang template format.
    </p>

    <p class="mt-2">For example, if you have <b v-pre>{{.Service.Name}}</b> it will return the service name.</p>

    <h2 class="mt-3">Service Variables</h2>

    <table class="table">
      <thead>
        <tr>
          <th scope="col">Variable</th>
          <th scope="col">True Value</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td><kbd v-pre>{{.Service.Id}}</kbd></td>
          <td>1</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Service.Name}}</kbd></td>
          <td>Example Service</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Service.Domain}}</kbd></td>
          <td>https://example.com</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Service.Port}}</kbd></td>
          <td>8080</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Service.Downtime.Human}}</kbd></td>
          <td>2 minutes</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Service.Uptime.Human}}</kbd></td>
          <td>13 hours</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Service.Online}}</kbd></td>
          <td>true/false</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Service.LastStatusCode}}</kbd></td>
          <td>404</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Service.FailuresLast24Hours}}</kbd></td>
          <td>38</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Service.LastOnline}}</kbd></td>
          <td>2020-11-05T13:15:30Z</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Service.LastOffline}}</kbd></td>
          <td>2020-10-01T13:15:30Z</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Service.Online24Hours}}</kbd></td>
          <td>0.99</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Service.Online7Days}}</kbd></td>
          <td>0.97</td>
        </tr>
      </tbody>
      <small>Additional variables within the Service struct</small>
    </table>

    <h2 class="mt-3">Failure Variables</h2>

    <table class="table">
      <thead>
        <tr>
          <th scope="col">Variable</th>
          <th scope="col">True Value</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td><kbd v-pre>{{.Failure.Issue}}</kbd></td>
          <td>Received 404 status code</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Failure.ErrorCode}}</kbd></td>
          <td>404</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Failure.Service}}</kbd></td>
          <td>1</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Failure.PingTime}}</kbd></td>
          <td>12482 (microseconds)</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Failure.DowntimeAgo}}</kbd></td>
          <td>35 minutes ago</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Failure.CreatedAt}}</kbd></td>
          <td>2020-05-02 09:14:43.66381 +0000 UTC</td>
        </tr>
      </tbody>
      <small>Additional variables within the Failures struct</small>
    </table>

    <h2 class="mt-3">Core Variables</h2>

    <table class="table">
      <thead>
        <tr>
          <th scope="col">Variable</th>
          <th scope="col">True Value</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td><kbd v-pre>{{.Core.Domain}}</kbd></td>
          <td>http://localhost:8080</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Core.Name}}</kbd></td>
          <td>Statping Demo</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Core.Description}}</kbd></td>
          <td>Statping will monitor your stuff!</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Core.Version}}</kbd></td>
          <td>v0.90.34</td>
        </tr>
        <tr>
          <td><kbd v-pre>{{.Core.Started}}</kbd></td>
          <td>2020-05-02 09:14:43.66381 +0000 UTC</td>
        </tr>
      </tbody>
      <small>Additional variables within the Core struct</small>
    </table>
  </div>
</template>

<script setup></script>

<style scoped></style>
