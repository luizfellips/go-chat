export interface User {
  id: string;
  email: string;
  username: string;
  avatar_url?: string | null;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface LoginResponse extends AuthTokens {
  user: User;
}
