# Jarvis (J.I.I. / PCAI)

Jarvis (J.I.I. PWA Console) is a Progressive Web Application (PWA) and Go-based backend server designed for real-time interactions, voice inputs, image analysis, and offline-capable messaging. The system uses WebSockets for low-latency communication and includes Service Workers for offline support and Web Push notifications.

## Features

* **Real-time WebSocket Chat**: Instant bidirectional communication between the web client and the Go server.
* **Progressive Web App (PWA)**: Installable on mobile devices with standalone UI, utilizing Service Workers (`sw.js`).
* **Offline Message Queuing**: Messages sent while offline are temporarily stored using IndexedDB (`OfflineManager`) and automatically resent when the network connection is restored.
* **Speech-to-Text**: Built-in voice input using the Web Speech API (`SpeechRecognition`).
* **Image Upload & Analysis**: Supports selecting and uploading images (Base64 encoded) for AI analysis.
* **Web Push Notifications**: Server-side push notification support using `webpush-go` to alert users of completed tasks.
* **Smart Idle Screen**: Avatar automatically zooms to fullscreen and plays time-contextual videos (morning/afternoon/evening) after 3 minutes of inactivity.
* **Dockerized Deployment**: Includes a `Makefile` and `Dockerfile` for quick deployment.

## Tech Stack

### Backend
* **Language**: Go 1.25+
* **Framework / Server**: [SherryServer](https://github.com/asccclass/sherryserver)
* **WebSocket**: [gorilla/websocket](https://github.com/gorilla/websocket)
* **Web Push**: [webpush-go](https://github.com/SherClockHolmes/webpush-go)
* **Environment Configuration**: [godotenv](https://github.com/joho/godotenv)

### Frontend
* **Core**: HTML5, CSS3, Vanilla JavaScript
* **Storage**: IndexedDB (Offline queuing)
* **APIs**: WebSockets, Web Speech API, Service Workers, Push API
* **Markdown**: [marked.js](https://marked.js.org/) for rendering server responses

## Project Structure

```text
.
├── server.go        # Go server initialization and SherryServer logic
├── router.go        # HTTP router configuration & static file serving
├── hub.go           # WebSocket Hub for broadcasting and client management
├── websocket.go     # WebSocket handler and message parsing logic
├── subscribe.go     # Web Push subscription handling and payload sending
├── go.mod           # Go dependencies
├── makefile         # Build scripts and Docker commands
├── clean.sh         # Cleanup scripts
├── envfile.example  # Example environment variables
└── www/
    └── html/        # Webroot containing the PWA assets
        ├── index.html     # Main PWA Chat Console Interface
        ├── manifest.json  # PWA Manifest file
        ├── sw.js          # Service Worker for offline capability & push
        ├── js/
        │   └── app.js     # IndexedDB Offline queue & Notification logic
        └── icons/         # App icons for various PWA resolutions
```

## Getting Started

### Prerequisites

* Go 1.25 or later (if building locally)
* Docker (for containerized deployment)
* `make` utility

### Setup Instructions

1. **Clone the repository**:
   ```bash
   git clone <your-repo-url>
   cd jarvis
   ```

2. **Environment Variables**:
   Copy the example environment file and configure it:
   ```bash
   cp envfile.example envfile
   ```
   *Modify `PORT`, `DocumentRoot` (default: `www/html`), or anything else required in `envfile`.*

3. **Install Dependencies**:
   ```bash
   make init
   ```

### Running the Server Locally

To compile and run the application natively on Linux/macOS:
```bash
make build
./app
```

### Running with Docker

Use the provided `Makefile` to build and run the system within a Docker container:
```bash
# Builds the Docker image and starts the container in the background
make run

# View logs
make log

# Stop the container
make stop

# Remove the container
make rm
```

## Usage

Once the server is running, navigate to `http://localhost:<PORT>` (or the domain you configured) in your browser.
* **Chat**: Type in the preview box or click the microphone to transcribe your voice.
* **Offline Mode**: If you lose connection, messages will be cached. Once reconnected, they will automatically sync to the server.
* **Clear Chat**: Click the broom icon (🧹) to clear the UI and purge the offline database.

---
*Note: For the Service Worker, Push Notifications, and Voice Recognition to work fully, the site MUST be served over HTTPS (or `localhost`).*
