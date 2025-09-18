Real-Time Financial Transaction Monitoring System

This project is a full-stack application designed to monitor financial transactions in real-time, detect fraudulent activity based on a set of rules, and visualize the data on a live dashboard.
Key Features

    Automated Fraud Detection: The Go backend analyzes transactions against multiple rules:

        High-Value Transactions: Flags any transaction over a configurable threshold.

        Velocity Checks: Flags users making too many transactions in a short period.

        Geographic Inconsistency: Flags transactions from the same user in geographically impossible locations within a short time frame (e.g., Tokyo and New York within the same minute).

    Live World Map Visualization: Transaction locations are plotted on a dynamic world map, with fraudulent alerts highlighted in red.


    Transaction Simulator: The frontend includes a tool to simulate a realistic stream of transactions to demonstrate the system's capabilities.

Tech Stack

    Backend: Go

        Web Server: gorilla/mux for routing.

        WebSockets: gorilla/websocket for real-time communication.

        Database: PostgreSQL (lib/pq driver).

    Frontend: React

        Styling: Tailwind CSS.

        Map Visualization: react-simple-maps.

    Database: PostgreSQL (running in Docker).

How to Run
1. Start the Database

docker run --name fraud-db -e POSTGRES_PASSWORD=mysecretpassword -e POSTGRES_USER=user -e POSTGRES_DB=fraud_detection -p 5432:5432 -d postgres

2. Create the Database Table

Connect to the running container:

docker exec -it fraud-db psql -U user -d fraud_detection

And run the SQL schema from backend/database.go.
3. Run the Backend Server

Navigate to the backend directory:

cd backend
go run .

The server will start on localhost:8080.
4. Run the Frontend Dashboard

Navigate to the frontend directory in a new terminal:

cd frontend
npm install
npm start

The dashboard will be available at http://localhost:3000.