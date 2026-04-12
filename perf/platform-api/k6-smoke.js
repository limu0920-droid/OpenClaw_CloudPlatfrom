import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 5,
  duration: '1m',
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<1000'],
  },
};

const baseUrl = __ENV.BASE_URL || 'http://127.0.0.1:18080';

export default function () {
  const health = http.get(`${baseUrl}/healthz`);
  check(health, { 'healthz 200': (r) => r.status === 200 });

  const ready = http.get(`${baseUrl}/readyz`);
  check(ready, { 'readyz 200': (r) => r.status === 200 });

  const bootstrap = http.get(`${baseUrl}/api/v1/bootstrap`);
  check(bootstrap, { 'bootstrap 200': (r) => r.status === 200 });

  const portalInstances = http.get(`${baseUrl}/api/v1/portal/instances`);
  check(portalInstances, { 'portal instances 200': (r) => r.status === 200 });

  const adminOrders = http.get(`${baseUrl}/api/v1/admin/orders`);
  check(adminOrders, { 'admin orders 200': (r) => r.status === 200 });

  sleep(1);
}
