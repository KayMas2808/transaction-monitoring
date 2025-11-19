# Real-Time Transaction Monitoring System

A financial fraud detection system that monitors transactions in real-time and flags suspicious activity using multiple detection algorithms.

## What It Does

This system watches financial transactions as they happen and automatically detects fraud using four different methods:

**Velocity Check** - Flags users making too many transactions too quickly (more than 5 per minute). Uses Redis for lightning-fast counting.

**High Value Check** - Catches transactions over $1500.

**Geographic Inconsistency** - Detects impossible travel, like transactions in London and New York within 60 seconds.

**Z-Score Analysis** - Uses statistics to find transactions that don't match a user's normal spending pattern. If a transaction is more than 3 standard deviations from their average, it gets flagged.

## Tech Stack

Backend: Go with Gorilla Mux and WebSockets
Frontend: React with Tailwind CSS
Database: PostgreSQL for storing transactions
Cache: Redis for high-speed fraud checks
Infrastructure: Docker Compose

## Running It

Clone the repo and run:

```bash
docker-compose up --build
```

Then open http://localhost:3000 in your browser.

Click "Start Simulation" to generate fake transactions and watch the fraud detection in action. The map shows where transactions are happening, and fraudulent ones appear in red with the reason they were flagged.

## How It Works

When a transaction comes in, it gets saved to PostgreSQL and immediately broadcast to all connected browsers via WebSocket. At the same time, the fraud detector runs all four checks. If any check fails, the transaction is marked as fraud in the database and a fraud alert is sent to all browsers.

Redis handles the velocity tracking and stores recent transaction amounts for Z-Score calculations. This keeps everything fast even under heavy load.

The frontend maintains "sticky" locations for users during simulation, so they don't teleport around the world with every transaction. Users stay in one city and only occasionally travel, making the fraud detection more realistic.

## Documentation

For detailed technical documentation including architecture diagrams and code explanations, see TECHNICAL_DOCUMENTATION.md.