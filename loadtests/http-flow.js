import { sleep } from 'k6';
import { scenarios } from './lib/scenarios.js';
import { USERS } from './lib/config.js';
import {
  login,
  searchUser,
  createConversation,
  listConversations,
  sendMessage,
  listMessages,
} from './lib/api.js';

const scenarioName = __ENV.SCENARIO || 'smoke';

export const options = {
  scenarios: {
    [scenarioName]: scenarios[scenarioName],
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    'http_req_duration{name:auth_login}': ['p(95)<1000'],
    'http_req_duration{name:conversations_create}': ['p(95)<1000'],
    'http_req_duration{name:messages_list}': ['p(95)<1000'],
  },
};

export function setup() {
  const loginRes = login(USERS.alice.email, USERS.alice.password);
  if (loginRes.status !== 200) {
    throw new Error(`setup login failed: ${loginRes.status} ${loginRes.body}`);
  }

  const token = loginRes.json('access_token');
  const bobRes = searchUser(token, USERS.bob.username);
  if (bobRes.status !== 200) {
    throw new Error(`setup search failed: ${bobRes.status} ${bobRes.body}`);
  }

  return { bobId: bobRes.json('id') };
}

export default function (data) {
  const loginRes = login(USERS.alice.email, USERS.alice.password);
  if (loginRes.status !== 200) {
    return;
  }

  const token = loginRes.json('access_token');

  const convRes = createConversation(token, data.bobId);
  if (convRes.status !== 201) {
    return;
  }

  const conversationId = convRes.json('id');

  listConversations(token);

  sendMessage(token, conversationId, `load test message from VU ${__VU} iter ${__ITER}`);

  listMessages(token, conversationId);

  sleep(1);
}
