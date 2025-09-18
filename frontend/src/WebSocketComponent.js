import React, { useEffect, useState } from 'react';

const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const WS_URL = `${wsProtocol}//${window.location.host}/ws`;

const WebSocketComponent = () => {
  const [alerts, setAlerts] = useState([]);

  useEffect(() => {
    const ws = new WebSocket(WS_URL);

    ws.onmessage = (event) => {
      const newAlert = JSON.parse(event.data);
      setAlerts((prev) => [...prev, newAlert]);
    };

    return () => ws.close();
  }, []);

  return (
    <div>
      <h3>Alerts</h3>
      <ul>
        {alerts.map((a, i) => (
          <li key={i}>{a.message}</li>
        ))}
      </ul>
    </div>
  );
};

export default WebSocketComponent;
