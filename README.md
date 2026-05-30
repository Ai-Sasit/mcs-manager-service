# MCS Manager Service 🚀

A modern, high-performance Minecraft Server Management Dashboard. This project provides a comprehensive web interface to manage, monitor, and control multiple Minecraft server instances with ease.

## ✨ Features

- **Real-time Monitoring**: Live resource usage (CPU, RAM, Disk) and server status.
- **Server Management**: Create, start, stop, and restart Minecraft servers.
- **Console Access**: Built-in terminal to interact with server processes via WebSockets.
- **File Manager**: Integrated tools for managing server files and configurations.
- **Port Manager**: Monitor network ports and manage associated processes.
- **Modern UI**: Neo-glassmorphism design with a responsive sidebar and dark mode.

## 🛠️ Tech Stack

### Backend
- **Core**: [Go (Golang)](https://golang.org/)
- **Web Framework**: [Fiber v3](https://docs.gofiber.io/)
- **Database**: MongoDB for app metadata.
- **Real-time**: WebSockets for live console and status updates.
- **Logging**: Custom buffered logger with daily rotation.

### Frontend
- **Framework**: [Vue.js 3](https://vuejs.org/) (Composition API)
- **State Management**: [Pinia](https://pinia.vuejs.org/)
- **Build Tool**: [Vite](https://vitejs.dev/)
- **Styling**: Vanilla CSS with modern design tokens.
- **Icons**: Lucide Vue & Material Design Icons.

---

## 📂 Project Structure

```text
mc-manage/
├── backend/            # Go backend service
│   ├── src/            # Business logic, controllers, and services
│   ├── logs/           # Application log files
│   └── main.go         # Entry point
├── frontend/           # Vue.js frontend (Vite)
│   ├── src/            # Components, views, and stores
│   └── public/         # Static assets
└── database/           # MongoDB Docker Compose and docs
```

---

## 🚀 Getting Started

### Backend Setup
1. Navigate to the backend directory: `cd backend`
2. Install dependencies: `go mod tidy`
3. Configure environment: Copy `.env.example` to `.env`
4. Start MongoDB using `database/docker-compose.yml` or provide `MONGO_URI`
5. Run in development: `go run main.go`

### Frontend Setup
1. Navigate to the frontend directory: `cd frontend`
2. Install dependencies: `npm install`
3. Run in development: `npm run dev`

---

## 📊 Monitoring & Logs

### How to tail logs in server

There are two primary ways to monitor logs depending on how the application is running:

#### 1. Using Journalctl (Recommended for Systemd)
If you have deployed the application using the provided `mc-manage.service` file, use:
```bash
journalctl -u mc-manage.service -f
```

#### 2. Tailing Application Log Files
The application automatically creates daily log files in the `backend/logs` directory.
To tail today's log file:
```bash
tail -f backend/logs/$(date +%F)-mc-manage.log
```

#### 3. View Port Usage
To see which processes are using specific ports (useful for troubleshooting):
```bash
# Check if port 8080 is in use
sudo lsof -i :8080
```

---

## 📄 License
[Private / Proprietary] - Pukky Company
