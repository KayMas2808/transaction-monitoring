import React, { useState, useEffect, useRef } from 'react';

// --- Main Application Component ---
function App() {
  // State management
  const [transactions, setTransactions] = useState([]);
  const [alerts, setAlerts] = useState([]);
  const [isAutomated, setIsAutomated] = useState(false);
  const simulationInterval = useRef(null);

  // WebSocket connection
  useEffect(() => {
    // Connect to the WebSocket server on the Go backend
    const ws = new WebSocket('ws://localhost:8080/ws');

    ws.onopen = () => {
      console.log('WebSocket connected');
    };

    ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      console.log('Received message:', message);

      // Distinguish between new transactions and fraud alerts
      if (message.type === 'new_transaction') {
        // Add new transaction to the top of the list, keeping the list at 20 items
        setTransactions(prev => [message.payload, ...prev.slice(0, 19)]);
      } else if (message.type === 'fraud_alert') {
        // Add new alert to the top of the list
        setAlerts(prev => [message.payload, ...prev.slice(0, 19)]);
        
        // Defensively check for the nested transaction object before updating the UI
        const fraudulentTransaction = message.payload.transaction;
        if (fraudulentTransaction) {
            // Update the specific transaction to mark it as fraudulent for visual feedback
            setTransactions(prev =>
              prev.map(t =>
                t.id === fraudulentTransaction.id
                  ? { ...t, is_fraud: true }
                  : t
              )
            );
        }
      }
    };

    ws.onclose = () => {
      console.log('WebSocket disconnected');
    };

    // Clean up the connection when the component unmounts
    return () => {
      ws.close();
    };
  }, []); // Empty dependency array ensures this runs only once on mount

  // Automated Transaction Simulation
  useEffect(() => {
    if (isAutomated) {
      // Start sending a new transaction every 2 seconds
      simulationInterval.current = setInterval(() => {
        handleSimulateTransaction(true); // true indicates it's an automated transaction
      }, 2000);
    } else {
      // Stop the simulation
      clearInterval(simulationInterval.current);
    }

    // Cleanup on unmount
    return () => clearInterval(simulationInterval.current);
  }, [isAutomated]); // This effect runs whenever 'isAutomated' changes

  // Function to send a transaction simulation request to the backend
  const handleSimulateTransaction = async (isAuto = false) => {
    try {
      // Generate random data for the transaction
      const randomUserId = `user_${Math.floor(Math.random() * 10)}`;
      const randomAmount = (Math.random() * 500).toFixed(2);
      const randomMerchant = `merchant_${Math.floor(Math.random() * 50)}`;

      const transactionData = {
        user_id: randomUserId,
        amount: parseFloat(randomAmount),
        merchant_details: randomMerchant,
        card_number: `4242-4242-4242-${Math.floor(1000 + Math.random() * 9000)}`
      };

      const response = await fetch('http://localhost:8080/transaction', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(transactionData),
      });

      if (!response.ok) {
        throw new Error('Failed to simulate transaction');
      }
      console.log('Simulated transaction:', transactionData);
    } catch (error) {
      console.error('Error simulating transaction:', error);
    }
  };
  
  const toggleAutomation = () => {
    setIsAutomated(!isAutomated);
  };


  return (
    <div className="min-h-screen bg-gray-100 text-gray-800 font-sans">
      <header className="bg-white shadow-md">
        <div className="container mx-auto px-6 py-4">
          <h1 className="text-3xl font-bold text-gray-900">Real-Time Transaction Monitor</h1>
          <p className="text-gray-600">Admin dashboard for monitoring financial events and fraud alerts.</p>
        </div>
      </header>

      <main className="container mx-auto px-6 py-8">
        {/* --- Controls Section --- */}
        <div className="bg-white p-6 rounded-lg shadow-lg mb-8">
            <h2 className="text-xl font-semibold mb-4">Controls</h2>
            <div className="flex space-x-4">
                <button
                    onClick={() => handleSimulateTransaction(false)}
                    className="bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded-lg transition duration-300 ease-in-out transform hover:scale-105"
                >
                    Simulate Single Transaction
                </button>
                <button
                    onClick={toggleAutomation}
                    className={`${
                        isAutomated ? 'bg-red-500 hover:bg-red-600' : 'bg-green-500 hover:bg-green-600'
                    } text-white font-bold py-2 px-4 rounded-lg transition duration-300 ease-in-out transform hover:scale-105`}
                >
                    {isAutomated ? 'Stop Automation' : 'Start Automated Transactions'}
                </button>
            </div>
        </div>

        {/* --- Dashboard Grids --- */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          
          {/* --- Live Transactions Column --- */}
          <div className="bg-white p-6 rounded-lg shadow-lg">
            <h2 className="text-xl font-semibold mb-4 border-b pb-2">Live Transaction Stream</h2>
            <div className="overflow-y-auto h-96">
              <ul className="space-y-3">
                {transactions.map(tx => (
                  <TransactionItem key={tx.id} tx={tx} />
                ))}
              </ul>
            </div>
          </div>
          
          {/* --- Fraud Alerts Column --- */}
          <div className="bg-white p-6 rounded-lg shadow-lg">
            <h2 className="text-xl font-semibold mb-4 text-red-600 border-b pb-2">Fraud Alerts</h2>
            <div className="overflow-y-auto h-96">
              <ul className="space-y-3">
                {alerts.map(alert => (
                  <AlertItem key={alert.id} alert={alert} />
                ))}
              </ul>
            </div>
          </div>

        </div>
      </main>
    </div>
  );
}


// --- Sub-component for a single transaction item ---
const TransactionItem = ({ tx }) => {
    const isFraud = tx.is_fraud;
    const baseClasses = "p-3 rounded-lg border border-gray-200 transition duration-300";
    // If it's fraud, apply a flashing red background. Otherwise, a calm gray.
    const conditionalClasses = isFraud 
        ? "bg-red-100 border-red-400 animate-pulse" 
        : "bg-gray-50";

    return (
        <li className={`${baseClasses} ${conditionalClasses}`}>
            <div className="flex justify-between items-center">
                <span className="font-mono text-sm text-gray-700">{tx.user_id}</span>
                <span className="font-bold text-lg text-gray-900">${parseFloat(tx.amount).toFixed(2)}</span>
            </div>
            <div className="text-xs text-gray-500 mt-1">
                <span>{tx.merchant_details}</span> | <span>ID: {tx.id}</span>
            </div>
        </li>
    );
};

// --- Sub-component for a single alert item ---
const AlertItem = ({ alert }) => (
  <li className="bg-red-50 p-3 rounded-lg border border-red-300">
    <div className="font-bold text-red-800">{alert.rule_name}</div>
    <div className="text-sm text-red-700 mt-1">
      <p>{alert.details}</p>
      {alert.transaction && <p className="font-mono text-xs mt-2">Transaction ID: {alert.transaction.id}</p>}
    </div>
  </li>
);

export default App;

