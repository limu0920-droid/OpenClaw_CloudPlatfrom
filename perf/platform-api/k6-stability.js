import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    {
      duration: '5m',
      target: 10 // 逐步增加到10个虚拟用户
    },
    {
      duration: '1h', // 持续1小时的稳定负载
      target: 10
    },
    {
      duration: '5m',
      target: 0 // 逐步减少到0个虚拟用户
    }
  ],
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<2000'],
  },
};

const baseUrl = __ENV.BASE_URL || 'http://127.0.0.1:18080';

// 测试端点列表
const endpoints = [
  '/healthz',
  '/readyz',
  '/api/v1/bootstrap',
  '/api/v1/portal/instances',
  '/api/v1/admin/orders',
  '/api/v1/portal/overview',
  '/api/v1/runtime/clusters',
  '/api/v1/portal/profile',
  '/api/v1/auth/providers',
  '/api/v1/search/config'
];

export default function () {
  // 随机选择一个端点进行测试
  const randomEndpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
  const response = http.get(`${baseUrl}${randomEndpoint}`);
  
  check(response, {
    'status 200': (r) => r.status === 200
  });
  
  // 随机休眠时间，模拟真实用户的访问模式
  sleep(Math.random() * 3 + 1); // 1-4秒的随机休眠
}
