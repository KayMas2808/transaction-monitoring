import React, { useState, useEffect, useRef } from 'react';
import { ComposableMap, Geographies, Geography, Marker } from 'react-simple-maps';

// world map rendering
const geoUrl = "https://cdn.jsdelivr.net/npm/world-atlas@2/countries-110m.json";

// city locations for simulation
const locations = [
  { name: "New York", coordinates: [-74.006, 40.7128] },
  { name: "London", coordinates: [-0.1278, 51.5074] },
  { name: "Tokyo", coordinates: [139.6917, 35.6895] },
  { name: "Sydney", coordinates: [151.2093, -33.8688] },
  { name: "Cairo", coordinates: [31.2357, 30.0444] },
  { name: "Moscow", coordinates: [37.6173, 55.7558] },
  { name: "Rio de Janeiro", coordinates: [-43.1729, -22.9068] },
  { name: "Beijing", coordinates: [116.4074, 39.9042] },
];

function App() {
  const [transactions, setTransactions] = useState([]);
  const [alerts, setAlerts] = useState([]);
  const [isSimulating, setIsSimulating] = useState(false);
  const simulationInterval = useRef(null);

  const [totalTransactions, setTotalTransactions] = useState(0);
  const [totalFraud, setTotalFraud] = useState(0);

  const ws = useRef(null);

  useEffect(() => {
    ws.current = new WebSocket('ws://localhost:8080/ws');
    ws.current.onopen = () => console.log('WebSocket connected');
    ws.current.onclose = () => console.log('WebSocket disconnected');

    ws.current.onmessage = (event) => {
      const message = JSON.parse(event.data);
      if (message.type === 'new_transaction') {
        const newTransaction = { ...message.payload, isNew: true };
        setTransactions(prev => [newTransaction, ...prev.slice(0, 99)]);
        setTotalTransactions(prev => prev + 1);
        
        const loc = locations.find(l => l.name === newTransaction.location);
        if(loc) {
          const alert = { ...loc, type: 'transaction' };
          setAlerts(prev => [alert, ...prev.slice(0, 49)]);
        }

      } else if (message.type === 'fraud_alert') {
          if (message.payload && message.payload.transaction) {
              const fraudulentTx = { ...message.payload.transaction, is_fraud: true, isNew: true };
              setTransactions(prev => [fraudulentTx, ...prev.slice(0, 99)]);
              
              setTotalFraud(prev => prev + 1);

              const loc = locations.find(l => l.name === fraudulentTx.location);
              if(loc) {
                const alert = { ...loc, type: 'fraud' };
                setAlerts(prev => [alert, ...prev.slice(0, 49)]);
              }
          }
      }
    };

    return () => {
      ws.current.close();
    };
  }, []);

  useEffect(() => {
    if (transactions.some(t => t.isNew)) {
      const timer = setTimeout(() => {
        setTransactions(prev => prev.map(t => ({ ...t, isNew: false })));
      }, 1500);
      return () => clearTimeout(timer);
    }
  }, [transactions]);


  const simulateTransaction = () => {
    const shouldBeImpossibleTravel = Math.random() < 0.1;
    const shouldBeHighValue = Math.random() < 0.05;

    const userId = `user_${Math.floor(Math.random() * 10)}`;
    const location = locations[Math.floor(Math.random() * locations.length)];
    
    let amount = (Math.random() * 500).toFixed(2);
    if(shouldBeHighValue) {
      amount = (1500 + Math.random() * 500).toFixed(2);
    }

    const transaction = {
      user_id: userId,
      amount: parseFloat(amount),
      card_number: `4242-XXXX-XXXX-${Math.floor(1000 + Math.random() * 9000)}`,
      merchant_details: `Merchant ${String.fromCharCode(65 + Math.floor(Math.random() * 26))}`,
      location: location.name,
    };

    fetch('http://localhost:8080/transaction', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(transaction),
    }).catch(err => console.error("Simulation Error:", err));

    if (shouldBeImpossibleTravel) {
        setTimeout(() => {
            let distantLocation;
            do {
                distantLocation = locations[Math.floor(Math.random() * locations.length)];
            } while (distantLocation.name === location.name);

            const secondTransaction = {
                ...transaction,
                location: distantLocation.name,
                amount: parseFloat((Math.random() * 500).toFixed(2)),
            };

            fetch('http://localhost:8080/transaction', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(secondTransaction),
            }).catch(err => console.error("Simulation Error (Impossible Travel):", err));
        }, 500);
    }
  };

  const toggleSimulation = () => {
    if (isSimulating) {
      clearInterval(simulationInterval.current);
      setIsSimulating(false);
    } else {
      setIsSimulating(true);
      simulationInterval.current = setInterval(simulateTransaction, 2000);
    }
  };

  const fraudRate = totalTransactions > 0 ? ((totalFraud / totalTransactions) * 100).toFixed(2) : 0;

  return (
    <div className="bg-gray-900 text-white min-h-screen font-sans">
      <div className="container mx-auto p-4 md:p-8">
        <header className="mb-8">
          <h1 className="text-4xl font-bold text-cyan-400">Real-Time Transaction Monitor</h1>
          <p className="text-gray-400">Live feed of transactions and fraud alerts.</p>
        </header>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="lg:col-span-2 bg-gray-800 p-6 rounded-lg shadow-2xl">
              <div className="flex justify-between items-center mb-4">
                  <h2 className="text-2xl font-semibold">Live Transaction Map</h2>
                  <div className="flex items-center space-x-4">
                      <button onClick={simulateTransaction} className="bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded-lg transition-colors">
                          Single Transaction
                      </button>
                      <button onClick={toggleSimulation} className={`font-bold py-2 px-4 rounded-lg transition-colors ${isSimulating ? 'bg-red-600 hover:bg-red-700' : 'bg-green-600 hover:bg-green-700'}`}>
                          {isSimulating ? 'Stop Simulation' : 'Start Simulation'}
                      </button>
                  </div>
              </div>
              <div className="w-full h-96 bg-gray-900 rounded-lg overflow-hidden">
                <ComposableMap projectionConfig={{ scale: 140 }} style={{ width: "100%", height: "100%" }}>
                  <Geographies geography={geoUrl}>
                    {({ geographies }) =>
                      geographies.map(geo => <Geography key={geo.rsmKey} geography={geo} fill="#334155" stroke="#1e293b" />)
                    }
                  </Geographies>
                  {alerts.map((alert, i) => (
                    <Marker key={i} coordinates={alert.coordinates}>
                      <circle r={5} className={alert.type === 'fraud' ? 'text-red-500 animate-ping' : 'text-cyan-400'} fillOpacity={0.5} />
                      <circle r={3} className={alert.type === 'fraud' ? 'text-red-500' : 'text-cyan-400'} fill="currentColor" />
                    </Marker>
                  ))}
                </ComposableMap>
              </div>
          </div>

          <div className="bg-gray-800 p-6 rounded-lg shadow-2xl">
            <h2 className="text-2xl font-semibold mb-4 border-b border-gray-700 pb-2">Statistics</h2>
            <div className="grid grid-cols-3 gap-4 text-center mb-6">
              <div>
                <p className="text-gray-400 text-sm">Total Transactions</p>
                <p className="text-2xl font-bold text-cyan-400">{totalTransactions}</p>
              </div>
              <div>
                <p className="text-gray-400 text-sm">Fraud Alerts</p>
                <p className="text-2xl font-bold text-red-500">{totalFraud}</p>
              </div>
              <div>
                <p className="text-gray-400 text-sm">Fraud Rate</p>
                <p className="text-2xl font-bold text-yellow-400">{fraudRate}%</p>
              </div>
            </div>
            <h2 className="text-2xl font-semibold mb-4 border-b border-gray-700 pb-2">Transaction Feed</h2>
            <div className="space-y-3 h-96 overflow-y-auto pr-2">
              {transactions.map(tx => (
                <div key={tx.id || Math.random()} className={`p-3 rounded-lg transition-all duration-1000 ${tx.isNew ? 'bg-gray-600 scale-105' : 'bg-gray-700'} ${tx.is_fraud ? 'border-2 border-red-500' : ''}`}>
                  <div className="flex justify-between items-center">
                    <span className="font-bold">{tx.user_id}</span>
                    <span className={`font-bold ${tx.is_fraud ? 'text-red-400' : 'text-green-400'}`}>${typeof tx.amount === 'number' ? tx.amount.toFixed(2) : '0.00'}</span>
                  </div>
                  <div className="flex justify-between items-center text-sm text-gray-400 mt-1">
                    <span>{tx.location || 'N/A'}</span>
                    <span>{new Date(tx.created_at).toLocaleTimeString()}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;

