import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    orders: {
      executor: 'constant-arrival-rate',
      rate: 1000, // 1000 iterations per second (RPS)
      timeUnit: '1s', // per second
      duration: '30s', // run for 30 seconds
      preAllocatedVUs: 200, // k6 will allocate this many VUs before test
      maxVUs: 1000, // upper limit, in case it needs more
    },
  },
};

export default function () {
  const url = 'http://localhost:8080/orders'; // 👈 your order-service endpoint

  const payload = JSON.stringify({
    productId: '7404fc75-1c7a-4f4e-8489-19e93f1dd994',
    qty: 1,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.post(url, payload, params);

  check(res, {
    'status is 201': (r) => r.status === 201,
  });

  sleep(1);
}
