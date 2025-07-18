import React, { useState, useEffect } from 'react';
import { LineChart, Line, AreaChart, Area, BarChart, Bar, PieChart, Pie, Cell, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { ShieldExclamationIcon, CurrencyDollarIcon, ChartBarIcon, UsersIcon, BellIcon, CheckCircleIcon, XCircleIcon, ClockIcon } from '@heroicons/react/24/outline';
import { api } from '../services/api';
import { formatCurrency, formatDate } from '../utils/helpers';

const Dashboard = () => {
  const [stats, setStats] = useState(null);
  const [transactions, setTransactions] = useState([]);
  const [alerts, setAlerts] = useState([]);
  const [chartData, setChartData] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDashboardData();
    const interval = setInterval(fetchDashboardData, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, []);

  const fetchDashboardData = async () => {
    try {
      const [statsRes, transactionsRes, alertsRes] = await Promise.all([
        api.get('/transactions/stats'),
        api.get('/transactions?limit=10'),
        api.get('/alerts?limit=10')
      ]);

      setStats(statsRes.data);
      setTransactions(transactionsRes.data.transactions);
      setAlerts(alertsRes.data.alerts);
      
      // Generate chart data (last 7 days)
      const chartData = generateChartData(transactionsRes.data.transactions);
      setChartData(chartData);
      
      setLoading(false);
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
      setLoading(false);
    }
  };

  const generateChartData = (transactions) => {
    const last7Days = Array.from({ length: 7 }, (_, i) => {
      const date = new Date();
      date.setDate(date.getDate() - (6 - i));
      return {
        date: date.toISOString().split('T')[0],
        day: date.toLocaleDateString('en-US', { weekday: 'short' }),
        transactions: 0,
        fraudulent: 0,
        amount: 0
      };
    });

    transactions.forEach(transaction => {
      const transactionDate = new Date(transaction.created_at).toISOString().split('T')[0];
      const dayData = last7Days.find(day => day.date === transactionDate);
      if (dayData) {
        dayData.transactions += 1;
        dayData.amount += transaction.amount;
        if (transaction.fraud_score >= 0.7) {
          dayData.fraudulent += 1;
        }
      }
    });

    return last7Days;
  };

  const StatCard = ({ title, value, icon: Icon, trend, color = 'primary' }) => (
    <div className={`bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow`}>
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-600">{title}</p>
          <p className="text-2xl font-bold text-gray-900">{value}</p>
          {trend && (
            <p className={`text-sm ${trend > 0 ? 'text-green-600' : 'text-red-600'}`}>
              {trend > 0 ? '↗' : '↘'} {Math.abs(trend)}% from last week
            </p>
          )}
        </div>
        <div className={`p-3 rounded-full bg-${color}-100`}>
          <Icon className={`h-6 w-6 text-${color}-600`} />
        </div>
      </div>
    </div>
  );

  const AlertCard = ({ alert }) => {
    const severityColors = {
      low: 'bg-blue-100 text-blue-800',
      medium: 'bg-yellow-100 text-yellow-800',
      high: 'bg-orange-100 text-orange-800',
      critical: 'bg-red-100 text-red-800'
    };

    return (
      <div className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-sm transition-shadow">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <div className="flex items-center space-x-2">
              <span className={`px-2 py-1 rounded-full text-xs font-medium ${severityColors[alert.severity]}`}>
                {alert.severity.toUpperCase()}
              </span>
              <span className="text-xs text-gray-500">
                {formatDate(alert.created_at)}
              </span>
            </div>
            <p className="mt-2 text-sm text-gray-900">{alert.message}</p>
            <p className="text-xs text-gray-600">
              Transaction ID: {alert.transaction_id} | Amount: {formatCurrency(alert.amount)}
            </p>
          </div>
          <div className="flex space-x-1">
            <button className="p-1 text-green-600 hover:bg-green-50 rounded">
              <CheckCircleIcon className="h-4 w-4" />
            </button>
            <button className="p-1 text-red-600 hover:bg-red-50 rounded">
              <XCircleIcon className="h-4 w-4" />
            </button>
          </div>
        </div>
      </div>
    );
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Transaction Monitoring Dashboard</h1>
          <p className="text-gray-600">Real-time fraud detection and transaction analytics</p>
        </div>

        {/* Stats Grid */}
        {stats && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            <StatCard
              title="Total Transactions"
              value={stats.total_transactions.toLocaleString()}
              icon={ChartBarIcon}
              trend={12}
            />
            <StatCard
              title="Fraud Rate"
              value={`${(stats.fraud_rate * 100).toFixed(2)}%`}
              icon={ShieldExclamationIcon}
              trend={-5}
              color="danger"
            />
            <StatCard
              title="Total Volume"
              value={formatCurrency(stats.total_amount)}
              icon={CurrencyDollarIcon}
              trend={8}
              color="success"
            />
            <StatCard
              title="Avg Transaction"
              value={formatCurrency(stats.avg_amount)}
              icon={UsersIcon}
              trend={3}
            />
          </div>
        )}

        {/* Charts Section */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          {/* Transaction Volume Chart */}
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Transaction Volume (7 Days)</h3>
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="day" />
                <YAxis />
                <Tooltip />
                <Area type="monotone" dataKey="transactions" stroke="#3b82f6" fill="#3b82f6" fillOpacity={0.6} />
              </AreaChart>
            </ResponsiveContainer>
          </div>

          {/* Fraud Detection Chart */}
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Fraud vs Normal Transactions</h3>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="day" />
                <YAxis />
                <Tooltip />
                <Bar dataKey="transactions" fill="#22c55e" name="Normal" />
                <Bar dataKey="fraudulent" fill="#ef4444" name="Fraudulent" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Recent Activity */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Recent Transactions */}
          <div className="bg-white rounded-lg shadow-sm border border-gray-200">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-semibold text-gray-900">Recent Transactions</h3>
            </div>
            <div className="p-6">
              <div className="space-y-4">
                {transactions.slice(0, 5).map((transaction) => (
                  <div key={transaction.id} className="flex items-center justify-between py-3 border-b border-gray-100 last:border-b-0">
                    <div className="flex items-center space-x-3">
                      <div className={`w-2 h-2 rounded-full ${
                        transaction.fraud_score >= 0.7 ? 'bg-red-500' : 
                        transaction.fraud_score >= 0.3 ? 'bg-yellow-500' : 'bg-green-500'
                      }`}></div>
                      <div>
                        <p className="text-sm font-medium text-gray-900">
                          {formatCurrency(transaction.amount)} {transaction.currency}
                        </p>
                        <p className="text-xs text-gray-500">User {transaction.user_id}</p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="text-sm text-gray-900">{transaction.status}</p>
                      <p className="text-xs text-gray-500">{formatDate(transaction.created_at)}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Recent Alerts */}
          <div className="bg-white rounded-lg shadow-sm border border-gray-200">
            <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
              <h3 className="text-lg font-semibold text-gray-900">Recent Alerts</h3>
              <BellIcon className="h-5 w-5 text-gray-400" />
            </div>
            <div className="p-6">
              <div className="space-y-4">
                {alerts.slice(0, 4).map((alert) => (
                  <AlertCard key={alert.id} alert={alert} />
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard; 