import React, { useEffect, useState, useCallback } from "react";
import api from "../services/api";
import { transactions } from "../services/api";

const Dashboard = () => {
  const [dashboardData, setDashboardData] = useState(null);
  const [error, setError] = useState("");

  const fetchDashboardData = useCallback(async () => {
    try {
      const res = await api.get("/dashboard");
      setDashboardData(res.data);
    } catch (err) {
      console.error("Error fetching dashboard data:", err);
      setError("Failed to load dashboard data.");
    }
  }, []);

  useEffect(() => {
    fetchDashboardData();
  }, [fetchDashboardData]);

  if (error) {
    return <p className="error">{error}</p>;
  }

  if (!dashboardData) {
    return <p>Loading dashboard...</p>;
  }

  return (
    <div className="dashboard-container">
      <h2>Dashboard</h2>
      <pre>{JSON.stringify(dashboardData, null, 2)}</pre>
    </div>
  );
};

export default Dashboard;
