import axios from "axios";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8082";
const ORDER_API_URL = process.env.NEXT_PUBLIC_ORDER_API_URL || "http://localhost:8084";
const SEARCH_API_URL = process.env.NEXT_PUBLIC_SEARCH_API_URL || "http://localhost:8086";
const AUTH_API_URL = process.env.NEXT_PUBLIC_AUTH_API_URL || "http://localhost:8080";

// Catalog API (products, categories, variants)
export const catalogApi = axios.create({
  baseURL: `${API_BASE_URL}/api/v1/catalog`,
  headers: { "Content-Type": "application/json" },
  timeout: 10000,
});

// Auth API
export const authApi = axios.create({
  baseURL: `${AUTH_API_URL}/api/v1/auth`,
  headers: { "Content-Type": "application/json" },
  withCredentials: true,
  timeout: 10000,
});

// Order API (cart, checkout, orders)
export const orderApi = axios.create({
  baseURL: `${ORDER_API_URL}/api/v1`,
  headers: { "Content-Type": "application/json" },
  timeout: 10000,
});

// Search API
export const searchApi = axios.create({
  baseURL: `${SEARCH_API_URL}/api/v1/search`,
  headers: { "Content-Type": "application/json" },
  timeout: 10000,
});

// Inject auth token into order and auth requests
const addAuthToken = (api: ReturnType<typeof axios.create>) => {
  api.interceptors.request.use((config) => {
    if (typeof window !== "undefined") {
      const token = localStorage.getItem("access_token");
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }
    return config;
  });
};

addAuthToken(orderApi);
addAuthToken(authApi);

// Auto-refresh on 401 — tries once per request, then rejects
const AUTH_BASE = process.env.NEXT_PUBLIC_AUTH_API_URL || "http://localhost:8080";

orderApi.interceptors.response.use(
  (res) => res,
  async (error) => {
    const originalRequest = error.config;
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      try {
        const refreshRes = await fetch(`${AUTH_BASE}/api/v1/auth/refresh`, {
          method: "POST",
          credentials: "include",
        });
        if (refreshRes.ok) {
          const data = await refreshRes.json();
          const jwt = data.jwt;
          if (typeof window !== "undefined") localStorage.setItem("access_token", jwt);
          originalRequest.headers.Authorization = `Bearer ${jwt}`;
          return orderApi(originalRequest);
        }
      } catch {}
    }
    return Promise.reject(error);
  }
);
