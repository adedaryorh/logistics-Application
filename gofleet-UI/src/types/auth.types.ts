export interface User {
  id: string;
  email: string;
  name: string;
  role: 'customer' | 'driver' | 'merchant';
  phone?: string;
  avatar?: string;
  isVerified: boolean;
  isActive?: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterData {
  name: string;
  email: string;
  password: string;
  role: 'customer' | 'driver' | 'merchant';
}

export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

export interface RefreshTokenResponse {
  access_token: string;
  refresh_token?: string;
}

export interface PasswordResetRequest {
  email: string;
}

export interface PasswordResetConfirm {
  email: string;
  token: string;
  new_password: string;
}

export interface EmailVerificationRequest {
  token: string;
}

export interface TokenPayload {
  sub: string;
  role: 'customer' | 'driver' | 'merchant';
  iat: number;
  exp: number;
  [key: string]: unknown;
}

export interface AuthHeaders {
  Authorization: string;
  'X-Request-ID'?: string;
}
