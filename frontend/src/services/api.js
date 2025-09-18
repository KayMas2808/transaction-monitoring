import axios from "axios";

const API_BASE_URL = process.env.REACT_APP_API_URL || "http://localhost:3000";

const api = axios.create({
  baseURL: API_BASE_URL,
});

// auth endpoints
export const auth = {
  login: (credentials) => api.post("/auth/login", credentials),
  register: (data) => api.post("/auth/register", data),
};

// transactions endpoints
export const transactions = {
  list: () => api.get("/transactions"),
  create: (data) => api.post("/transactions", data),
};

// axios instance
export default api;
