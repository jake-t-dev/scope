# Scope

Scope is a developer-focused news aggregator that creates a personalized feed based on your GitHub activity. By analyzing your GitHub interests using Google's Gemini AI, Scope delivers relevant tech news articles tailored to your specific domain expertise.

## Features

- **GitHub Interest Analysis**: Scans your GitHub profile to identify your preferred languages, frameworks, and tools.
- **AI-Powered Curation**: Uses Google Gemini AI to synthesize your interests into optimized search queries.
- **Tailored News Feed**: Fetches real-time, relevant tech news from Google News RSS.
- **Modern Stack**: Built with Go 1.25, React 19, and TypeScript.

## Tech Stack

- **Backend**: Go, Google Generative AI SDK, Google News RSS
- **Frontend**: React, TypeScript, Vite, Tailwind CSS
- **Infrastructure**: Docker

## Prerequisites

Before running the application, ensure you have the following installed:

- [Go 1.25+](https://go.dev/dl/)
- [Bun](https://bun.sh/) (or Node.js/npm)
- [Docker](https://www.docker.com/) (optional, for containerized execution)
- A [Google Gemini API Key](https://aistudio.google.com/app/apikey)

## Configuration

1. Clone the repository:
   ```bash
   git clone https://github.com/jake-t-dev/scope.git
   cd scope
   ```

2. Create a `.env` file in the root directory:

3. Add your environment variables to `.env`:
   ```env
   PORT=8080
   GEMINI_API_KEY=your_gemini_api_key_here
   APP_ENV=development
   ```

## Development

You can run the application locally using the provided Makefile.

### Run Locally (Recommended)

This command runs the Go backend and the React frontend (with live reload for both):

```bash
make run
```

### Run with Docker

Build and run the entire stack in containers:

```bash
make docker-run
```

This will start:
- Backend API on `http://localhost:8080`
- Frontend on `http://localhost:5173`

To stop the containers:

```bash
make docker-down
```

### Other Commands

- **Build Binary**: `make build`
- **Run Tests**: `make test`
- **Clean Build**: `make clean`
- **Live Reload (Backend)**: `make watch`

## Project Structure

```
├── cmd/api/        # Application entry point
├── internal/       # Private application code
│   ├── ai/         # Gemini AI integration
│   ├── github/     # GitHub client
│   ├── news/       # News fetching logic
│   └── server/     # HTTP server and routes
├── frontend/       # React application
└── docker-compose.yml
```
