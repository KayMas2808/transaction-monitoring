import React from "react";
import WebSocketComponent from "./WebSocketComponent";

function App() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100">
      <div className="bg-white shadow-md rounded-lg p-5 w-1/2">
        <h1 className="text-2xl font-bold text-center mb-4">Fraud Monitoring Dashboard</h1>
        <WebSocketComponent />
      </div>
    </div>
  );
}

export default App;
