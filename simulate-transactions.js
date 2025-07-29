const axios = require('axios');
const API_BASE_URL = 'http://localhost:8080/api/v1';

async function runSimulation() {
    console.log('Starting transaction simulation...');

    let authToken = '';

    // 1. Login and get token
    try {
        console.log('Attempting to log in as admin...');
        const loginResponse = await axios.post(`${API_BASE_URL}/auth/login`, {
            username: 'admin',
            password: 'admin123',
        });
        authToken = loginResponse.data.token;
        console.log('Login successful! Token obtained.');
    } catch (error) {
        console.error('Login failed:', error.response ? error.response.data : error.message);
        console.error('Please ensure the backend is running and the admin credentials in migrations.sql are correct.');
        return;
    }

    const transactions = [
        // --- Legit Transactions ---
        { user_id: 101, amount: 50.00, currency: 'USD', merchant_id: 'SHOP_ABC', location: 'New York' },
        { user_id: 102, amount: 120.75, currency: 'USD', merchant_id: 'EAT_DEF', location: 'Chicago' },
        { user_id: 101, amount: 25.50, currency: 'USD', merchant_id: 'GAS_XYZ', location: 'New York' },
        { user_id: 103, amount: 75.00, currency: 'EUR', merchant_id: 'BOOK_STORE', location: 'Berlin' },

        // --- Fraudulent: High Amount Transaction (threshold > $10,000) ---
        { user_id: 201, amount: 15000.00, currency: 'USD', merchant_id: 'LUXURY_GOODS', location: 'Los Angeles' },
        await new Promise(resolve => setTimeout(resolve, 2000)), // Short delay for next transaction

        // --- Fraudulent: Velocity Check (more than 3 transactions in 1 minute for same user_id) ---
        { user_id: 301, amount: 10.00, currency: 'USD', merchant_id: 'GAME_PURCHASE', location: 'Online' },
        await new Promise(resolve => setTimeout(resolve, 1000)), // 1 second delay
        { user_id: 301, amount: 20.00, currency: 'USD', merchant_id: 'GAME_PURCHASE', location: 'Online' },
        await new Promise(resolve => setTimeout(resolve, 1000)), // 1 second delay
        { user_id: 301, amount: 5.00, currency: 'USD', merchant_id: 'GAME_PURCHASE', location: 'Online' },
        await new Promise(resolve => setTimeout(resolve, 1000)), // 1 second delay
        { user_id: 301, amount: 15.00, currency: 'USD', merchant_id: 'GAME_PURCHASE', location: 'Online' }, // This should trigger velocity
        await new Promise(resolve => setTimeout(resolve, 65000)), // Wait >1 minute for next rule set

        // --- Fraudulent: Amount Pattern (round number, e.g., $5000.00) ---
        { user_id: 401, amount: 5000.00, currency: 'USD', merchant_id: 'INVESTMENT_FIRM', location: 'Miami' },
        await new Promise(resolve => setTimeout(resolve, 2000)), // Short delay

        // --- Fraudulent: Time Anomaly (assumes server's local time 2 AM - 6 AM) ---
        // Note: For this rule to reliably trigger, the server's time would need to be within the 2-6 AM window.
        // We'll send it regardless, the server will timestamp it.
        { user_id: 501, amount: 300.00, currency: 'GBP', merchant_id: 'LATE_NIGHT_SHOP', location: 'London' },
        await new Promise(resolve => setTimeout(resolve, 2000)), // Short delay

        // --- Fraudulent: Location Anomaly (transaction from a new location for a user) ---
        // These two transactions from user 601 in Tokyo/Kyoto will be marked as new locations.
        // If user_id 601 has previous transactions from Tokyo, only Kyoto might be new,
        // or if no history, both will be new.
        { user_id: 601, amount: 200.00, currency: 'USD', merchant_id: 'TRAVEL_AGENT', location: 'Tokyo' },
        await new Promise(resolve => setTimeout(resolve, 2000)), // Short delay
        { user_id: 601, amount: 100.00, currency: 'USD', merchant_id: 'TOUR_OPERATOR', location: 'Kyoto' },
        await new Promise(resolve => setTimeout(resolve, 2000)), // Short delay
    ];

    for (let i = 0; i < transactions.length; i++) {
        const item = transactions[i];

        if (typeof item === 'object') {
            const transaction = item;
            try {
                const config = {
                    headers: {
                        Authorization: `Bearer ${authToken}`,
                        'Content-Type': 'application/json',
                    },
                };
                const response = await axios.post(`${API_BASE_URL}/transactions`, transaction, config);
                console.log(`[${i + 1}/${transactions.length}] Transaction sent: ID ${response.data.id}, Amount: ${transaction.amount}, Fraud Score: ${response.data.fraud_score}, Status: ${response.data.status}`);
            } catch (error) {
                console.error(`[${i + 1}/${transactions.length}] Error sending transaction Amount: ${transaction.amount}:`, error.response ? error.response.data : error.message);
            }
        } else if (typeof item === 'number') { // It's a delay time
            // This means the "await new Promise..." was directly put into the array
            await new Promise(resolve => setTimeout(resolve, item));
            console.log(`Pausing for ${item / 1000} seconds...`);
        }
    }

    console.log('Simulation complete. Check your dashboard at http://localhost:3000/dashboard');
}

runSimulation();