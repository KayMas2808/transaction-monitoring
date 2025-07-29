import axios from 'axios';
import toast from 'react-hot-toast';

const BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

// Create axios instance
const api = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('authToken');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('authToken');
      localStorage.removeItem('user');
      window.location.href = '/login';
      toast.error('Session expired. Please login again.');
    } else if (error.response?.status === 403) {
      toast.error('You do not have permission to perform this action.');
    } else if (error.response?.status >= 500) {
      toast.error('Server error. Please try again later.');
    } else if (error.code === 'ECONNABORTED') {
      toast.error('Request timeout. Please check your connection.');
    } else {
      toast.error(error.response?.data?.message || 'An error occurred');
    }
    return Promise.reject(error);
  }
);

// Auth API
export const authAPI = {
  login: (credentials) => api.post('/auth/login', credentials),
  getProfile: () => api.get('/auth/profile'),
  createUser: (userData) => api.post('/auth/users', userData),
};

// Transaction API
export const transactionAPI = {
  getAll: (params = {}) => api.get('/transactions', { params }),
  getById: (id) => api.get(`/transactions/${id}`),
  create: (transactionData) => api.post('/transactions', transactionData),
  review: (id, reviewData) => api.put(`/transactions/${id}/review`, reviewData),
  getStats: () => api.get('/transactions/stats'),
};

// Alert API
export const alertAPI = {
  getAll: (params = {}) => api.get('/alerts', { params }),
  resolve: (id, resolveData) => api.put(`/alerts/${id}/resolve`, resolveData),
};

// Admin API
export const adminAPI = {
  getFraudRules: () => api.get('/admin/fraud-rules'),
  createFraudRule: (ruleData) => api.post('/admin/fraud-rules', ruleData),
  updateFraudRule: (id, ruleData) => api.put(`/admin/fraud-rules/${id}`, ruleData),
  getAuditLogs: (params = {}) => api.get('/admin/audit-logs', { params }),
  getConnections: () => api.get('/admin/connections'),
};

// WebSocket connection
export const createWebSocketConnection = (onMessage, onError) => {
  const token = localStorage.getItem('authToken');
  const wsUrl = `${BASE_URL.replace('http', 'ws').replace('/api/v1', '')}/ws`;
  
  const ws = new WebSocket(wsUrl);
  
  ws.onopen = () => {
    console.log('WebSocket connected');
    // Send authentication
    ws.send(JSON.stringify({ type: 'auth', token }));
  };
  
  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      onMessage(data);
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  };
  
  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
    if (onError) onError(error);
  };
  
  ws.onclose = () => {
    console.log('WebSocket disconnected');
  };
  
  return ws;
};

// Authentication helpers
export const auth = {
  isAuthenticated: () => !!localStorage.getItem('authToken'),
  getToken: () => localStorage.getItem('authToken'),
  getUser: () => {
    const user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
  },
  setAuth: (token, user) => {
    localStorage.setItem('authToken', token);
    localStorage.setItem('user', JSON.stringify(user));
  },
  clearAuth: () => {
    localStorage.removeItem('authToken');
    localStorage.removeItem('user');
  },
};

export { api };
export default api; 