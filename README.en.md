# Video Feed System

A fully functional backend service for a short video feed system, built on the Go language Gin framework, providing core features such as user management, video publishing, and social interaction.

## Project Introduction

This project is a complete video feed system, implementing core functionality similar to Douyin or TikTok. The system adopts a separated frontend and backend architecture. The backend is developed using the Go language, the frontend is built using Vue.js, and it supports containerized deployment.

## Core Features

### User System
- User registration and login (supports JWT authentication)
- Personal information management (avatar, bio modification)
- Password change and security management
- Automatic token refresh mechanism

### Video Features
- Video upload and publishing (supports thumbnail generation)
- View video details
- Video search
- Popular video recommendations
- Video playback record statistics
- Batch delete videos

### Social Interaction
- Like/unlike videos
- Post and reply to comments
- Follow/unfollow users
- Manage followers and following lists
- Query user relationships

### Notification System
- Like notifications
- Comment notifications
- Follow notifications
- Unread message count

### Message Queue
- RabbitMQ asynchronous processing of like events
- Asynchronous notification push
- Reliable message delivery mechanism

## Tech Stack

### Backend Technologies
- **Language**: Go 1.26
- **Web Framework**: Gin
- **ORM**: GORM
- **Cache**: Redis
- **Message Queue**: RabbitMQ
- **Authentication**: JWT
- **Database**: MySQL
- **Configuration Management**: YAML

### Frontend Technologies
- **Framework**: Vue.js 3
- **Build Tool**: Vite
- **Language**: TypeScript
- **Routing**: Vue Router
- **State Management**: Pinia
- **UI Library**: Native CSS + SVG Icons

### Deployment & Operations
- **Containerization**: Docker
- **Orchestration**: Docker Compose
- **Reverse Proxy**: Nginx

## Project Structure

```
feed-system_video/
├── backend/                    # Backend Service
│   ├── cmd/                    # Entry Files
│   │   ├── api/               # API Service Entry
│   │   └── worker/            # Message Queue Consumer
│   ├── internal/              # Internal Modules
│   │   ├── account/          # User Account Module
│   │   ├── apierror/         # API Error Handling
│   │   ├── auth/             # JWT Authentication
│   │   ├── config/           # Configuration Loading
│   │   ├── db/               # Database Connection
│   │   ├── feed/             # Feed Stream Module
│   │   ├── http/             # HTTP Routing
│   │   ├── middleware/       # Middleware
│   │   │   ├── cors/         # CORS Handling
│   │   │   ├── feedcache/    # Feed Cache
│   │   │   ├── jwt/          # JWT Middleware
│   │   │   ├── rabbitmq/     # RabbitMQ Connection
│   │   │   ├── ratelimit/    # Rate Limiting
│   │   │   ├── redis/        # Redis Operations
│   │   │   └── storage/      # File Storage
│   │   ├── notification/     # Notification Module
│   │   ├── social/           # Social Relationship Module
│   │   ├── video/            # Video Module
│   │   └── worker/           # Background Tasks
│   └── configs/              # Configuration Files
├── frontend/                  # Frontend Application
│   ├── src/                  # Source Code
│   │   ├── api/              # API Calls
│   │   ├── composables/      # Composables
│   │   ├── router/           # Routing Configuration
│   │   ├── stores/           # State Management
│   │   ├── views/            # Page Components
│   │   └── main.ts           # Application Entry
│   └── public/               # Static Assets
├── sql/                      # Database Scripts
├── docker-compose.yml        # Docker Compose Configuration
└── README.md                 # Project Documentation
```

## Quick Start

### Environment Requirements

- Docker >= 20.10
- Docker Compose >= 2.0
- Go >= 1.20 (Local Development)
- Node.js >= 18 (Local Development)

### Deploy with Docker

1. Clone the project and enter the directory:
```bash
git clone https://gitee.com/huanghu6/feed-system_video.git
cd feed-system_video
```

2. Configure database connection:
```bash
# Edit backend configuration file
vim backend/configs/config.yaml
```

3. Start services:
```bash
docker-compose up -d
```

4. Access services:
- Frontend Page: http://localhost:3000
- API Service: http://localhost:8080

### Local Development

#### Starting the Backend

1. Enter the backend directory:
```bash
cd backend
```

2. Install dependencies:
```bash
go mod tidy
```

3. Configure environment variables and database connection

4. Start services:
```bash
# Start API Service
go run cmd/api/main.go

# Start Message Queue Consumer
go run cmd/worker/main.go
```

#### Starting the Frontend

1. Enter the frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
npm install
```

3. Start development server:
```bash
npm run dev
```

## Configuration Instructions

### Backend Configuration (`backend/configs/config.yaml`)

Main configuration items include:

- **server**: Service port, run mode
- **mysql**: Database connection info
- **redis**: Redis connection info
- **jwt**: Token secret and expiration time
- **rabbitmq**: Message queue connection
- **storage**: File storage path

### Frontend Configuration

The frontend API address is configured in `frontend/src/api/client.ts`, and the `baseURL` value can be modified according to the environment.

## API Endpoints

### Authentication Module

| Method | Path | Description |
|------|------|------|
| POST | /account/register | User Registration |
| POST | /account/login | User Login |
| GET | /account/profile | Get Personal Info |
| PUT | /account/avatar | Update Avatar |
| PUT | /account/bio | Update Bio |
| PUT | /account/password | Change Password |

### Video Module

| Method | Path | Description |
|------|------|------|
| POST | /video/publish | Publish Video |
| GET | /video/list/author | Get User Video List |
| GET | /video/detail/:id | Get Video Details |
| POST | /video/like | Like Video |
| POST | /video/unlike | Unlike Video |
| POST | /video/comment | Post Comment |
| GET | /video/comments | Get Comment List |
| GET | /video/hot | Get Popular Videos |
| GET | /video/search | Search Videos |
| DELETE | /video/:id | Delete Video |

### Social Module

| Method | Path | Description |
|------|------|------|
| POST | /social/follow | Follow User |
| POST | /social/unfollow | Unfollow User |
| GET | /social/followers | Get Followers List |
| GET | /social/vloggers | Get Following List |
| GET | /social/counts | Get Follow Statistics |
| GET | /social/isfollowing | Check Following Status |

### Notification Module

| Method | Path | Description |
|------|------|------|
| GET | /notification/list | Get Notification List |
| GET | /notification/unread | Get Unread Count |

## Database Design

Database scripts are located at `sql/feed_system_video.sql`, containing the following main tables:

- `accounts`: User accounts table
- `videos`: Video information table
- `likes`: Like records table
- `comments`: Comments table
- `socials`: Social relationship table
- `notifications`: Notifications table

## License

This project is open source under the MIT License.