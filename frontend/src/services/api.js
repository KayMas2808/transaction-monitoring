import axios from "axios";

// base instance
const API = axios.create({
  baseURL: process.env.REACT_APP_API_URL || "http://localhost:8080/api",
});

// token to every request if present
API.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Auth helper
export const auth = {
  login: (credentials) => API.post("/auth/login", credentials),
  register: (userData) => API.post("/auth/register", userData),
  logout: () => API.post("/auth/logout"),
  me: () => API.get("/auth/me"),
};

export const transactions = {
  getAll: () => API.get("/transactions"),
  getById: (id) => API.get(`/transactions/${id}`),
  create: (data) => API.post("/transactions", data),
};

export default API;
