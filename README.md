# SIF Performance Tracker

A web-based tool to track the daily Net Asset Value (NAV) and performance returns of Strategic Investment Funds (SIFs). It compares performance across various timeframes (1 Day, 1 Week, 1 Month, 3 Months, 6 Months) and visualizes the performance trajectory.

<img width="1231" height="898" alt="image" src="https://github.com/user-attachments/assets/6861ae76-6969-43fe-b2b1-f5d3d3389594" />


## Features

- **Daily Data Updates:** Fetches the latest scheme NAV data.
- **Performance Table:** Sortable table showing the NAV and percentage returns over different periods.
- **Visual Trajectory:** Interactive line chart mapping the relative returns of different schemes over time (1M, 3M, 6M, All Time).
- **Sentiment Indicators:** Emoji-based sentiment analysis for 1-month returns to quickly gauge performance.

## Project Structure

- `backend/`: Go backend responsible for data fetching, transformation, and serving the application.
- `frontend/`: HTML, CSS, and Vanilla JavaScript frontend displaying the data using Chart.js and PapaParse.

## Prerequisites

- [Go](https://golang.org/) (for running the backend to fetch data and serve the application)

## Running Locally

1. Navigate to the `backend` directory:
   ```bash
   cd backend
   ```

2. To update the data and serve the website (defaults to port `8080`):
   ```bash
   go run . -serve
   ```
   Or to just update the data without serving:
   ```bash
   go run . -update
   ```

3. Open your browser and visit [http://localhost:8080](http://localhost:8080).

## Architecture Details

- **Backend:** Defines API routes, downloads JSON/CSV data for the SIP schemes, performs any required backend parsing/transformation, and serves the static frontend alongside the data.
- **Frontend:** Provides a responsive dashboard. It asynchronously loads data, processes historical CSV records to calculate relative performance figures, and renders the UI.
