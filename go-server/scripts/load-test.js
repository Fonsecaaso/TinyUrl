import http from 'k6/http';
import { sleep, check } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 30 },
    { duration: '1m', target: 60 },
    { duration: '1m', target: 90 },
    { duration: '1m', target: 120 },
    { duration: '1m', target: 150 },
    { duration: '1m', target: 50 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<800'], // adjust according to my SLO
    http_req_failed: ['rate<0.01'],   // <1% error
  },
};

export default function () {
  const res = http.get('https://fonsecaaso.com/api/tL8AQC', {
    headers: {
      'User-Agent': 'k6-prod-test',
      'X-K6-Test': 'true', 
    },
    redirects: 5, 
  });

  check(res, {
    'status = 200 or 3xx': (r) => r.status === 200 || (r.status >= 300 && r.status < 400),
    'response time < 800ms': (r) => r.timings.duration < 800,
  });

  sleep(0);
}
