import React, { useState, useEffect, useCallback } from 'react';

// --- Reusable Components ---

const Card = ({ children, className = '' }) => (
  <div className={`bg-gray-800 shadow-lg rounded-xl p-6 ${className}`}>
    {children}
  </div>
);

const CardHeader = ({ title, subtitle }) => (
  <div className="mb-4 border-b border-gray-700 pb-4">
    <h2 className="text-2xl font-bold text-white">{title}</h2>
    <p className="text-gray-400">{subtitle}</p>
  </div>
);

const StatusIndicator = ({ connected }) => (
  <div className="flex items-center space-x-2">
    <div className={`w-3 h-3 rounded-full ${connected ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`}></div>
    <span className="text-gray-300 font-medium">
      {connected ? 'Connected to Real-Time Service' : 'Disconnected'}
    </span>
  </div>
);

// --- Main App Component ---

function App() {
  const [transactions, setTransactions] = useState([]);
  const [alerts, setAlerts] = useState([]);
  const [isConnected, setIsConnected] = useState(false);
  const [formState, setFormState] = useState({ userId: '', amount: '', description: '' });
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Effect to handle WebSocket connection
  useEffect(() => {
    const socket = new WebSocket('ws://localhost:8080/ws');

    socket.onopen = () => {
      console.log('WebSocket connection established');
      setIsConnected(true);
    };

    socket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      if (message.type === 'new_transaction') {
        setTransactions(prev => [message.payload, ...prev]);
      } else if (message.type === 'fraud_alert') {
        setAlerts(prev => [message.payload, ...prev]);
      }
    };

    socket.onclose = () => {
      console.log('WebSocket connection closed');
      setIsConnected(false);
    };

    socket.onerror = (error) => {
      console.error('WebSocket error:', error);
      setIsConnected(false);
    };

    // Cleanup on component unmount
    return () => {
      socket.close();
    };
  }, []);

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormState(prev => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!formState.userId || !formState.amount) {
      alert('User ID and Amount are required.');
      return;
    }
    setIsSubmitting(true);
    try {
      const response = await fetch('http://localhost:8080/transaction', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          userId: formState.userId,
          amount: parseFloat(formState.amount),
          description: formState.description,
        }),
      });
      if (!response.ok) {
        throw new Error('Failed to submit transaction');
      }
      // The websocket will push the update, so no need to manually add it here.
      // Clear part of the form for the next transaction.
      setFormState(prev => ({ ...prev, amount: '', description: '' }));

    } catch (error) {
      console.error('Submission error:', error);
      alert('Could not submit transaction. Check the console.');
    } finally {
      setIsSubmitting(false);
    }
  };


  return (
    <div className="bg-gray-900 min-h-screen text-gray-200 font-sans p-4 sm:p-8">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <header className="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-8">
          <div>
            <h1 className="text-4xl font-extrabold text-white tracking-tight">Financial Monitoring Dashboard</h1>
            <p className="text-gray-400 mt-1">Real-time fraud detection and transaction streaming.</p>
          </div>
          <div className="mt-4 sm:mt-0">
            <StatusIndicator connected={isConnected} />
          </div>
        </header>

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">

          {/* Left Column: Transaction Simulation */}
          <div className="lg:col-span-1">
            <Card>
              <CardHeader title="Simulate New Transaction" subtitle="Create a test transaction." />
              <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                  <label htmlFor="userId" className="block text-sm font-medium text-gray-300 mb-1">User ID</label>
                  <input type="text" name="userId" id="userId" value={formState.userId} onChange={handleInputChange} className="w-full bg-gray-700 border border-gray-600 rounded-md px-3 py-2 text-white focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500" placeholder="e.g., user-123" />
                </div>
                <div>
                  <label htmlFor="amount" className="block text-sm font-medium text-gray-300 mb-1">Amount ($)</label>
                  <input type="number" name="amount" id="amount" value={formState.amount} onChange={handleInputChange} className="w-full bg-gray-700 border border-gray-600 rounded-md px-3 py-2 text-white focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500" placeholder="e.g., 99.99" step="0.01" />
                </div>
                <div>
                  <label htmlFor="description" className="block text-sm font-medium text-gray-300 mb-1">Description</label>
                  <input type="text" name="description" id="description" value={formState.description} onChange={handleInputChange} className="w-full bg-gray-700 border border-gray-600 rounded-md px-3 py-2 text-white focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500" placeholder="e.g., Online Purchase" />
                </div>
                <button type="submit" disabled={isSubmitting} className="w-full bg-indigo-600 hover:bg-indigo-700 text-white font-bold py-2 px-4 rounded-md transition duration-300 ease-in-out disabled:bg-indigo-900 disabled:cursor-not-allowed">
                  {isSubmitting ? 'Submitting...' : 'Submit Transaction'}
                </button>
              </form>
            </Card>
          </div>

          {/* Right Column: Feeds */}
          <div className="lg:col-span-2 space-y-8">
            {/* Fraud Alerts */}
            <Card>
              <CardHeader title="Fraud Alerts" subtitle="High-priority events requiring attention." />
              <div className="h-64 overflow-y-auto pr-2 space-y-3">
                {alerts.length === 0 ? <p className="text-gray-500">No alerts yet.</p> :
                  alerts.map(alert => (
                    <div key={alert.alertId} className="bg-red-900/50 border border-red-700 p-3 rounded-lg animate-fade-in">
                      <p className="font-bold text-red-300">{alert.reason}</p>
                      <p className="text-sm text-gray-300">User <span className="font-mono bg-gray-700 px-1 rounded">{alert.userId}</span> flagged for high transaction frequency.</p>
                      <p className="text-xs text-gray-400 mt-1">{new Date(alert.timestamp).toLocaleString()}</p>
                    </div>
                  ))}
              </div>
            </Card>

            {/* Live Transactions */}
            <Card>
              <CardHeader title="Live Transaction Feed" subtitle="All incoming financial transactions." />
              <div className="h-96 overflow-y-auto pr-2 space-y-3">
                {transactions.length === 0 ? <p className="text-gray-500">Waiting for transactions...</p> :
                  transactions.map(tx => (
                    <div key={tx.id} className="bg-gray-700/50 p-3 rounded-lg flex justify-between items-center animate-fade-in-down">
                      <div>
                        <p className="font-semibold text-white">{tx.description || 'N/A'}</p>
                        <p className="text-sm text-gray-400">User ID: <span className="font-mono">{tx.userId}</span></p>
                      </div>
                      <div className="text-right">
                        <p className="text-lg font-bold text-green-400">${parseFloat(tx.amount).toFixed(2)}</p>
                        <p className="text-xs text-gray-500">{new Date(tx.timestamp).toLocaleTimeString()}</p>
                      </div>
                    </div>
                  ))}
              </div>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;
