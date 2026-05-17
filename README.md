# YuwanaDev MDM - Backend Server

The core backend service for the YuwanaDev Mobile Device Management (MDM) platform. It orchestrates real-time WebSocket communication between Android Agents and the Dashboard.

## Features
- **WebSocket Hub**: Manages bidirectional communication streams with thousands of devices.
- **Device Management**: Tracks device states, groups, and command histories.
- **WebRTC Signaling**: Facilitates the handshake required for remote screen mirroring.
- **RESTful API**: Serves data to the frontend dashboard.

## Tech Stack
- **Language**: Go 1.21+
- **Framework**: Fiber
- **Database**: PostgreSQL
- **Migrations**: Goose

## Build & Run
Make sure you have a PostgreSQL database running, then:
```bash
# Copy example env
cp .env.example .env

# Run migrations
make migrate-up

# Start the server
make run
```

