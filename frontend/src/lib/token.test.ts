import { describe, it, expect, beforeEach } from "vitest";
import {
  clearTokens,
  getAccessToken,
  getRefreshToken,
  isAuthenticated,
  setTokens,
} from "@/lib/token";

describe("token storage", () => {
  beforeEach(() => {
    sessionStorage.clear();
  });

  it("stores and retrieves tokens", () => {
    setTokens("access-1", "refresh-1");
    expect(getAccessToken()).toBe("access-1");
    expect(getRefreshToken()).toBe("refresh-1");
    expect(isAuthenticated()).toBe(true);
  });

  it("clears tokens on logout", () => {
    setTokens("access-1", "refresh-1");
    clearTokens();
    expect(getAccessToken()).toBeNull();
    expect(getRefreshToken()).toBeNull();
    expect(isAuthenticated()).toBe(false);
  });
});
