import http from 'k6/http';
import { check } from 'k6';
import { BASE_URL } from './config.js';

const JSON_HEADERS = { 'Content-Type': 'application/json' };

export function authHeaders(token) {
  return {
    ...JSON_HEADERS,
    Authorization: `Bearer ${token}`,
  };
}

export function login(email, password) {
  const res = http.post(
    `${BASE_URL}/auth/login`,
    JSON.stringify({ email, password }),
    { headers: JSON_HEADERS, tags: { name: 'auth_login' } },
  );

  check(res, {
    'login status 200': (r) => r.status === 200,
    'login returns access_token': (r) => !!r.json('access_token'),
  });

  return res;
}

export function searchUser(token, username) {
  const res = http.get(`${BASE_URL}/users/search?username=${username}`, {
    headers: authHeaders(token),
    tags: { name: 'users_search' },
  });

  check(res, {
    'search status 200': (r) => r.status === 200,
    'search returns id': (r) => !!r.json('id'),
  });

  return res;
}

export function createConversation(token, participantId) {
  const res = http.post(
    `${BASE_URL}/conversations`,
    JSON.stringify({ participant_id: participantId }),
    { headers: authHeaders(token), tags: { name: 'conversations_create' } },
  );

  check(res, {
    'create conversation status 201': (r) => r.status === 201,
    'create conversation returns id': (r) => !!r.json('id'),
  });

  return res;
}

export function listConversations(token) {
  const res = http.get(`${BASE_URL}/conversations`, {
    headers: authHeaders(token),
    tags: { name: 'conversations_list' },
  });

  check(res, {
    'list conversations status 200': (r) => r.status === 200,
  });

  return res;
}

export function sendMessage(token, conversationId, content) {
  const res = http.post(
    `${BASE_URL}/conversations/${conversationId}/messages`,
    JSON.stringify({ content }),
    { headers: authHeaders(token), tags: { name: 'messages_send' } },
  );

  check(res, {
    'send message status 201': (r) => r.status === 201,
    'send message returns id': (r) => !!r.json('id'),
  });

  return res;
}

export function listMessages(token, conversationId, limit = 50) {
  const res = http.get(`${BASE_URL}/conversations/${conversationId}/messages?limit=${limit}`, {
    headers: authHeaders(token),
    tags: { name: 'messages_list' },
  });

  check(res, {
    'list messages status 200': (r) => r.status === 200,
    'list messages returns array': (r) => Array.isArray(r.json('messages')),
  });

  return res;
}
