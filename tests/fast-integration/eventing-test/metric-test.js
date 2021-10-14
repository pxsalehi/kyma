const axios = require("axios");
const https = require("https");
const { expect } = require("chai");
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;

const {
  kubectlPortForward,
  debug,
  retryPromise,
} = require("../utils");


const prometheusPort = 9090;

function prometheusPortForward() {
  return kubectlPortForward("kyma-system", "prometheus-monitoring-prometheus-0", prometheusPort);
}

const dashboards = {
  // The delivery dashboard
  delivery_publisherProxy: {
    title: `Requests to publisher proxy`,
    query: 'sum(rate(promhttp_metric_handler_requests_total{namespace="kyma-system", pod=~"event.*publisher.*", code=~"2.*"}[5m]))',
    backends: ['nats', 'beb'],
    // The assert function receives the `data` section of the query result:
    // https://prometheus.io/docs/prometheus/latest/querying/api/#instant-queries
    assert: function (resp) {
      expect(resp.result.length).to.be.greaterThan(0, "No value found in the result");
      expect(getMetricValue(resp.result[0])).to.be.greaterThan(0);
    }
  },
  delivery_applicationConnectivityValidator: {
    title: 'Requests to application connectivity validator',
    query: 'sum by(destination_service) (rate(istio_requests_total{destination_service=~"commerce-validator.kyma-integration.svc.cluster.local", response_code=~"2.*"}[5m]))',
    backends: ['nats', 'beb'],
    assert: function (resp) {
      let foundMetric = false;
      resp.result.forEach(res => {
        if (foundMetric) return;
        foundMetric = res.metric.destination_service.includes('commerce-validator') && getMetricValue(res) > 0;
      });
      expect(foundMetric).to.be.true
    }
  },
  delivery_subscribers: {
    title: 'Requests to subscribers',
    query: `
          sum (rate(
            istio_requests_total{
              source_workload=~"eventing.*controller",
              destination_workload!~"istio-.+|dex|unknown|[a-z0-9-]+-dispatcher", 
              response_code=~"2.*"
            }[5m])
          ) by (le,source_workload_namespace,source_workload,destination_workload_namespace,destination_workload,response_code)`,
    backends: ['nats'],
    assert: function (resp) {
      let foundMetric = false;
      resp.result.forEach(res => {
        if (foundMetric === true) return;
        foundMetric = res.metric.destination_workload.startsWith('lastorder') && getMetricValue(res) > 0;
      })
      expect(foundMetric).to.be.true
    }
  },
  // The latency dashboard
  latency_connectivityValidatorToPublisherProxy: {
    title: 'Latency of Connectivity Validator -> Event Publisher',
    query: `
        histogram_quantile(
          0.99999, 
          sum(rate(
            istio_request_duration_milliseconds_bucket{
              source_workload_namespace="kyma-integration",
              source_workload=~".*connectivity-validator.*",
              destination_workload_namespace="kyma-system",
              destination_workload=~"eventing.*publisher.*"
            }[5m])
          ) by (le,source_workload_namespace,source_workload,destination_workload_namespace,destination_workload))
        `,
    backends: ['nats', 'beb'],
    assert: function(resp) {
      let foundMetric = false;
      resp.result.forEach(res => {
        if (foundMetric) return;
        foundMetric = res.metric.source_workload.toLowerCase() === 'commerce-connectivity-validator' &&
          res.metric.destination_workload.toLowerCase() === 'eventing-publisher-proxy' &&
          getMetricValue(res) > 0;
      });
      expect(foundMetric).to.be.true
    }
  },
  latency_eventPublisherToMessagingServer: {
    title: 'Latency of Event Publisher -> Messaging Server',
    query: 'histogram_quantile(0.99999, sum(rate(event_publish_to_messaging_server_latency_bucket{namespace="kyma-system"}[5m])) by (le,pod,namespace,service))',
    backends: ['nats', 'beb'],
    assert: function(resp) {
      let foundMetric = false;
      resp.result.forEach(res => {
        if (foundMetric) return;
        foundMetric = res.metric.namespace.toLowerCase() === 'kyma-system' &&
          res.metric.pod.toLowerCase().startsWith('eventing-publisher-proxy') &&
          getMetricValue(res) > 0;
      });
      expect(foundMetric).to.be.true;
    }
  },
  latency_eventDispatcherToSubscribers: {
    title: 'Latency of Event Dispatcher -> Subscribers',
    query: `
          histogram_quantile(
          0.99999, 
          sum(rate(
            istio_request_duration_milliseconds_bucket{
              source_workload=~"eventing.*controller",
              destination_workload!~"istio-.+|dex|unknown|[a-z0-9-]+-dispatcher"
            }[5m])
          ) by (le,source_workload_namespace,source_workload,destination_workload_namespace,destination_workload))
        `,
    backends: ['nats'],
    assert: function (resp) {
      let foundMetric = false;
      resp.result.forEach(res => {
        if (foundMetric) return;
        foundMetric = res.metric.source_workload === 'eventing-controller' &&
          res.metric.destination_workload.toLowerCase().startsWith('lastorder') &&
          getMetricValue(res) > 0;
      });
      expect(foundMetric).to.be.true;
    }
  },
  // The pods dashboard
  pods_memoryUsage: {
    title: 'Memory usage',
    // This is not the exact query used in Grafana, but it ensures memory usage of eventing components are visible
    query: 'sum by(container, pod) (container_memory_usage_bytes{job="kubelet", container!="POD", container !=""}) * on(pod) group_left() kube_pod_labels{label_kyma_project_io_dashboard="eventing"}',
    backends: ['nats', 'beb'],
    assert: ensureEventingPodsArePresent
  },
  pods_cpuUsage: {
    title: 'CPU usage',
    // This is not the exact query used in Grafana, but it ensures CPU usage of eventing components are visible
    query: 'sum by (container, pod) (irate(container_cpu_usage_seconds_total{job="kubelet", image!="", container!="POD"}[4m])) * on(pod) group_left() kube_pod_labels{label_kyma_project_io_dashboard="eventing"}',
    backends: ['nats', 'beb'],
    assert: ensureEventingPodsArePresent
  },
  pods_networkReceive: {
    title: 'Network receive',
    query: 'sum by (pod) (irate(container_network_receive_bytes_total{job="kubelet"}[4m])) * on(pod) kube_pod_labels{label_kyma_project_io_dashboard="eventing"}',
    backends: ['nats', 'beb'],
    assert: ensureEventingPodsArePresent
  },
  pods_networkTransmit: {
    title: 'Network transmit',
    query: 'sum by (pod) (irate(container_network_transmit_bytes_total{job="kubelet"}[4m])) * on(pod) kube_pod_labels{label_kyma_project_io_dashboard="eventing"}',
    backends: ['nats', 'beb'],
    assert: ensureEventingPodsArePresent
  }
}

// A generic assertion for the pod dashboards
function ensureEventingPodsArePresent (resp) {
  let controllerFound = false, publisherProxyFound = false, natsFound = false;
  resp.result.forEach(res => {
    if (controllerFound && publisherProxyFound && natsFound) return;
    if (res.metric.pod.startsWith('eventing-nats')) natsFound = true;
    if (res.metric.pod.startsWith('eventing-controller')) controllerFound = true;
    if (res.metric.pod.startsWith('eventing-publisher-proxy')) publisherProxyFound = true;
  });
  expect(controllerFound).to.be.true
  expect(publisherProxyFound).to.be.true
  expect(natsFound).to.be.true
}

// Given a Prometheus metric result, extracts the value of the metric
function getMetricValue(metric) {
  return parseFloat(metric.value[1])
}

function runDashboardTestCase(dashboardName, test) {
  return retryPromise(async () => {
    await axios.get('http://localhost:' + prometheusPort + '/api/v1/query?query=' + test.query).then(resp => {
      debug(dashboardName + ' result: ' + JSON.stringify(resp.data, null, 2));
      // Make sure the query was successfully processed and returned a result
      expect(resp.status).to.be.equal(200)
      expect(resp.data.status).to.equal("success")
      expect(resp.data).to.have.nested.property('data.result')
      test.assert(resp.data.data)
    }).catch(reason => {
      throw new Error(reason)
    })
  }, 30, 6000);
}

function eventingMonitoringTest(backend) {
  for (const [dashboardName, test] of Object.entries(dashboards)) {
    if (test.backends.includes(backend)) {
      it('Testing dashboard: ' + test.title, async () => {
        await runDashboardTestCase(dashboardName, test)
      });
    }
  }
}

module.exports = {
  eventingMonitoringTest,
  prometheusPortForward
}