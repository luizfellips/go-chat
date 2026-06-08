import http from 'k6/http';
import { check, sleep } from 'k6';
import { scenarios } from './lib/scenarios.js';
import { BASE_URL, USERS } from './lib/config.js';

const scenarioName = __ENV.SCENARIO || 'smoke';

export const options = {
  scenarios: {
    [scenarioName]: scenarios[scenarioName],
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    'http_req_duration{name:auth_login}': ['p(95)<1000'],
  },
};

export default function () {
  const res = http.post(
    `${BASE_URL}/auth/login`,
    JSON.stringify({
      email: USERS.alice.email,
      password: USERS.alice.password,
    }),
    {
      headers: { 'Content-Type': 'application/json' },
      tags: { name: 'auth_login' },
    },
  );

  check(res, {
    'login status 200': (r) => r.status === 200,
    'login returns access_token': (r) => !!r.json('access_token'),
  });

  sleep(1);
}
