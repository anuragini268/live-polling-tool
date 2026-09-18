# Live Polling Tool

A full-stack web application for creating polls, voting, and viewing live poll results.

## Features

- Create polls with multiple options
- View active polls
- Submit votes
- Real-time percentage calculation
- Live results display
- Responsive React UI
- REST API using Go and Gin
- CORS enabled

## Tech Stack

### Frontend
- React
- Vite
- JavaScript
- CSS

### Backend
- Go
- Gin Framework
- REST API

### Database
- In-memory storage for the current demo version

## API Endpoints

POST /api/polls  
Create a new poll.

GET /api/polls  
Get all polls.

POST /api/polls/:id/vote  
Submit a vote for a poll.

## How to Run

### Backend

```bash
go run .\backend
