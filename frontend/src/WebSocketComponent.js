import React, { useState } from "react";
import useWebSocket from "react-use-websocket";

const WebSocketComponent = () => {
  const [alerts, setAlerts] = useState([]);

  // Connect to Go WebSocket server
  const { lastJsonMessage } = useWebSocket("ws://localhost:8080/ws", {
    onMessage: () => {
      if (lastJsonMessage) {
        setAlerts((prev) => [lastJsonMessage, ...prev]); // Add new alert to the list
      }
    },
    shouldReconnect: () => true, // Auto-reconnect
  });

  return (
    <div className="p-5">
      <h2 className="text-xl font-bold">🚨 Fraud Alerts</h2>
      {alerts.length === 0 ? (
        <p>No fraud detected yet.</p>
      ) : (
        <ul className="mt-3 space-y-2">
          {alerts.map((alert, index) => (
            <li key={index} className="p-3 bg-red-200 border-l-4 border-red-600">
              <strong>User ID:</strong> {alert.user_id} <br />
              <strong>Amount:</strong> ${alert.amount} <br />
              <strong>Status:</strong> {alert.status}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default WebSocketComponent;