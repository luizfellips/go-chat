export const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080/api/v1';

export const USERS = {
  alice: {
    email: __ENV.ALICE_EMAIL || 'alice@example.com',
    password: __ENV.ALICE_PASSWORD || 'password123',
  },
  bob: {
    email: __ENV.BOB_EMAIL || 'bob@example.com',
    password: __ENV.BOB_PASSWORD || 'password123',
    username: __ENV.BOB_USERNAME || 'bob',
  },
};
