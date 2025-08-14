# Chessload Changelog

A lightweight, self-hostable changelog and roadmap web application built with SvelteKit and Go.

## Features

- **📋 Event Management**: Create, edit, and organize changelog events with different statuses
- **🗳️ Community Voting**: Let users vote on upcoming features
- **📅 Timeline View**: Beautiful changelog timeline with releases and roadmap
- **🎨 Dark/Light Theme**: Built-in theme switcher
- **👤 Admin Panel**: Simple and intuitive admin interface
- **🐳 Docker Ready**: Easy deployment with Docker
- **💾 SQLite Database**: Lightweight and portable database
- **📱 Responsive Design**: Works great on desktop and mobile

## Tech Stack

- **Frontend**: SvelteKit + TailwindCSS + Shadcn/ui-inspired components
- **Backend**: Go with Gin framework
- **Database**: SQLite
- **Authentication**: JWT-based admin authentication
- **Deployment**: Docker with multi-stage builds

## Quick Start

### Using Docker (Recommended)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd chessload-changelog
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your preferred settings
   ```

3. **Run with Docker Compose**
   ```bash
   docker-compose up -d
   ```

4. **Access the application**
   - Public changelog: http://localhost:8080
   - Admin panel: http://localhost:8080/admin
   - Default credentials: admin/admin (change these!)

### Manual Installation

#### Prerequisites
- Node.js 18+ and npm
- Go 1.21+
- SQLite

#### Backend Setup
```bash
cd backend
go mod download
go run main.go
```

#### Frontend Setup
```bash
cd frontend
npm install
npm run build
```

The backend serves both API and static files.

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ADMIN_USERNAME` | `admin` | Admin login username |
| `ADMIN_PASSWORD` | `admin` | Admin login password |
| `JWT_SECRET` | `your-secret-key-change-in-production` | JWT signing secret |
| `PORT` | `8080` | Server port |
| `DB_PATH` | `./data/changelog.db` | SQLite database path |
| `GIN_MODE` | `debug` | Gin mode (release/debug) |

### Event Statuses

Events can have the following statuses:

- **Backlogs**: Ideas and planned features (displayed as simple list)
- **Vote**: Features users can vote on (displayed as voting cards)
- **Doing**: Currently in development (displayed in "Now" section)
- **Release**: Released features (displayed in main changelog timeline)
- **Upcoming**: Planned for specific timeframes (displayed as upcoming cards)
- **Archived**: Internal events not shown to public

## API Endpoints

### Public Endpoints
- `GET /api/events` - Get all public events
- `GET /api/events/:id` - Get specific event
- `POST /api/events/:id/vote` - Vote for an event
- `POST /api/auth/login` - Admin login

### Admin Endpoints (require JWT token)
- `GET /admin/validate` - Validate admin token
- `GET /admin/events` - Get all events (including archived)
- `POST /admin/events` - Create new event
- `PUT /admin/events/:id` - Update event
- `DELETE /admin/events/:id` - Delete event

## Development

### Frontend Development
```bash
cd frontend
npm run dev
```

### Backend Development
```bash
cd backend
go run main.go
```

### Building for Production
```bash
# Build everything with Docker
docker build -t chessload-changelog .

# Or build separately
cd frontend && npm run build
cd ../backend && go build -o main .
```

## Deployment

### Docker Compose (Recommended)
```yaml
version: '3.8'
services:
  changelog:
    image: chessload-changelog:latest
    ports:
      - "8080:8080"
    environment:
      - ADMIN_USERNAME=your-admin
      - ADMIN_PASSWORD=secure-password
      - JWT_SECRET=your-jwt-secret
    volumes:
      - changelog_data:/app/data
    restart: unless-stopped

volumes:
  changelog_data:
```

### Environment Setup
1. Copy `.env.example` to `.env`
2. Update credentials and secrets
3. Run `docker-compose up -d`

## Usage

### Adding Events

1. Go to `/admin` and log in
2. Navigate to "Manage Events"
3. Click "New Event"
4. Fill in the details:
   - **Title**: Event name
   - **Description**: Brief description
   - **Status**: Choose appropriate status
   - **Tags**: Categorize your event
   - **Content**: Detailed description in Markdown
   - **Date**: Set date in YYYY-MM-DD format (displays as "10 Aug. 2025")

### Managing Votes

Events with "Vote" status will display voting buttons on the public page. Users can vote to show interest in features.

### Organizing Timeline

- Use "Release" status for completed features
- Use "Doing" for current development
- Use "Upcoming" for planned features with timeframes
- Use "Backlogs" for ideas and future considerations

## Security

- Change default admin credentials
- Use a strong JWT secret in production
- Consider running behind a reverse proxy (nginx, Caddy)
- Regular database backups (SQLite file in `/app/data/`)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is open source and available under the MIT License.
