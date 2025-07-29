// frontend/src/App.js
import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from 'react-hot-toast'; // Required for toast notifications

// Import the new Login component
import Login from './components/Login';
// Import the existing Dashboard component
import Dashboard from './components/Dashboard';
// Import the new ProtectedRoute component
import ProtectedRoute from './components/ProtectedRoute';
// Import the auth helpers to check initial auth status for redirection
import { auth } from './services/api';

function App() {
  return (
    <Router>
      {/* Toaster for global notifications */}
      <Toaster position="top-right" reverseOrder={false} />

      <Routes>
        {/* Public Route for Login page */}
        <Route path="/login" element={<Login />} />

        {/* Protected Route for Dashboard */}
        {/* The Dashboard component also includes the WebSocketComponent implicitly in its structure */}
        <Route
          path="/dashboard/*" // Use /* to match any nested routes within Dashboard if you add them later
          element={
            <ProtectedRoute>
              <Dashboard />
            </ProtectedRoute>
          }
        />

        {/* Default Route: Redirect based on authentication status */}
        <Route
          path="/"
          element={
            // If authenticated, go to dashboard; otherwise, go to login
            <Navigate to={auth.isAuthenticated() ? '/dashboard' : '/login'} replace />
          }
        />

        {/* Fallback Route: Redirect any unmatched paths to the root */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Router>
  );
}

export default App;