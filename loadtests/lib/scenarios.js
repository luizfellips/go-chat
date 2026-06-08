export const scenarios = {
  smoke: {
    executor: 'shared-iterations',
    vus: 1,
    iterations: 1,
    maxDuration: '1m',
  },
  load_100: {
    executor: 'ramping-vus',
    startVUs: 0,
    stages: [
      { duration: '30s', target: 100 },
      { duration: '2m', target: 100 },
      { duration: '30s', target: 0 },
    ],
    gracefulRampDown: '15s',
  },
  load_1000: {
    executor: 'ramping-vus',
    startVUs: 0,
    stages: [
      { duration: '2m', target: 1000 },
      { duration: '3m', target: 1000 },
      { duration: '1m', target: 0 },
    ],
    gracefulRampDown: '30s',
  },
  ramp_up: {
    executor: 'ramping-vus',
    startVUs: 0,
    stages: [
      { duration: '1m', target: 50 },
      { duration: '1m', target: 100 },
      { duration: '1m', target: 200 },
      { duration: '2m', target: 200 },
      { duration: '1m', target: 0 },
    ],
    gracefulRampDown: '30s',
  },
};
