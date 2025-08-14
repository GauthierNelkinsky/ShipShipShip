# 🚢 ShipShipShip

A modern, self-hostable changelog and roadmap platform that helps you share product updates with your community and gather feedback through feature voting.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.21-blue.svg)
![Node Version](https://img.shields.io/badge/node-18+-green.svg)
![Docker](https://img.shields.io/badge/docker-ready-blue.svg)

## ✨ Features

### 📋 **Event Management**
- Create, edit, and organize changelog events with different statuses
- Rich text editor with Markdown support, tables, links, and images
- Drag-and-drop Kanban board for easy organization
- Event tagging and categorization

### 🗳️ **Community Engagement**
- **Feature Voting**: Let users vote on upcoming features
- **Roadmap Visibility**: Show what's coming next
- **Timeline View**: Beautiful changelog with release history
- **Public Comments**: Gather feedback from your community

### 🎨 **Modern Interface**
- **Responsive Design**: Works perfectly on desktop and mobile
- **Dark/Light Theme**: Built-in theme switcher with user preference
- **Clean Admin Panel**: Intuitive management interface
- **Real-time Updates**: Dynamic content without page refreshes

### 🛠️ **Developer Friendly**
- **Self-hostable**: Complete control over your data
- **Docker Ready**: One-command deployment
- **SQLite Database**: No complex database setup required
- **RESTful API**: Full API access for integrations
- **JWT Authentication**: Secure admin access

## 🏗️ Tech Stack

### Frontend
- **SvelteKit** - Modern reactive framework
- **TailwindCSS** - Utility-first CSS framework
- **Shadcn/ui** - Beautiful, accessible components
- **TipTap Editor** - Rich text editing experience
- **Lucide Icons** - Consistent iconography

### Backend
- **Go** - Fast, reliable backend with Gin framework
- **SQLite** - Lightweight, embedded database
- **GORM** - Go ORM for database operations
- **JWT** - Secure authentication
- **CORS** - Cross-origin resource sharing

### Deployment
- **Docker** - Containerized deployment
- **Multi-stage builds** - Optimized image sizes
- **Health checks** - Production-ready monitoring

## 🚀 Quick Start

### Option 1: Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/GauthierNelkinsky/ShipShipShip.git
cd ShipShipShip

# Start with Docker Compose
docker-compose up -d

# Access your changelog
open http://localhost:8080
```

**Default admin credentials**: `admin` / `admin` (change these immediately!)

### Option 2: Manual Setup

#### Prerequisites
- Node.js 18+ and npm
- Go 1.21+
- SQLite

#### Backend Setup
```bash
cd backend
go mod download
go build -o main .
./main
```

#### Frontend Setup
```bash
cd frontend
npm install
npm run build
```

#### Development Mode
```bash
# Use the development script
./start-dev.sh

# Or start services individually
cd backend && go run main.go &
cd frontend && npm run dev
```

## ⚙️ Configuration

### Environment Variables

Create a `.env` file or set these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `ADMIN_USERNAME` | `admin` | Admin login username |
| `ADMIN_PASSWORD` | `admin` | Admin login password |
| `JWT_SECRET` | `your-secret-key-change-in-production` | JWT signing secret |
| `PORT` | `8080` | Server port |
| `DB_PATH` | `./data/changelog.db` | SQLite database path |
| `GIN_MODE` | `debug` | Gin mode (`release` for production) |

### Docker Compose Configuration

```yaml
version: "3.8"
services:
  changelog:
    image: shipshipship:latest
    ports:
      - "8080:8080"
    environment:
      - ADMIN_USERNAME=youradmin
      - ADMIN_PASSWORD=securerpassword
      - JWT_SECRET=your-jwt-secret-change-this
      - GIN_MODE=release
    volumes:
      - changelog_data:/app/data
    restart: unless-stopped

volumes:
  changelog_data:
```

## 📊 Event Management

### Event Statuses

Events can have different statuses that control how they're displayed:

- **📝 Backlogs**: Ideas and planned features (simple list view)
- **🗳️ Vote**: Features users can vote on (voting cards with counters)
- **🔄 Doing**: Currently in development (progress indicators)
- **🚀 Release**: Released features (main timeline with dates)
- **📅 Upcoming**: Planned releases (upcoming cards with timeframes)
- **📦 Archived**: Internal events (hidden from public view)

### Rich Content Editor

- **Markdown Support**: Full Markdown syntax
- **Rich Text Tools**: Bold, italic, lists, headers
- **Tables**: Create and edit tables
- **Links**: Add external and internal links
- **Images**: Upload and embed images
- **Code Blocks**: Syntax highlighting

## 🔌 API Reference

### Public Endpoints

```bash
# Get all public events
GET /api/events

# Get specific event
GET /api/events/:id

# Vote for an event
POST /api/events/:id/vote

# Admin login
POST /api/auth/login
```

### Admin Endpoints (JWT Required)

```bash
# Validate admin token
GET /admin/validate

# Get all events (including archived)
GET /admin/events

# Create new event
POST /admin/events

# Update event
PUT /admin/events/:id

# Delete event
DELETE /admin/events/:id

# Get/Update settings
GET /admin/settings
PUT /admin/settings
```

### Authentication

Include JWT token in Authorization header:
```bash
Authorization: Bearer <your-jwt-token>
```

## 🎨 Customization

### Branding
- **Logo**: Upload light and dark theme logos
- **Title**: Customize your changelog title
- **Colors**: Adjust primary color scheme
- **Website Link**: Link to your main website

### Themes
- **Automatic**: Follows system preference
- **Light Mode**: Clean, bright interface
- **Dark Mode**: Easy on the eyes for night browsing

## 🐳 Production Deployment

### Docker Production Setup

```dockerfile
# Use official image or build your own
FROM shipshipship:latest

# Set production environment
ENV GIN_MODE=release
ENV ADMIN_USERNAME=youradmin
ENV ADMIN_PASSWORD=secure-password-here
ENV JWT_SECRET=your-very-secure-jwt-secret

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/events
```

### Reverse Proxy (Nginx)

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 🔒 Security

### Production Checklist
- [ ] Change default admin credentials
- [ ] Set a strong JWT secret
- [ ] Use HTTPS in production
- [ ] Regular database backups
- [ ] Update dependencies regularly
- [ ] Monitor logs for suspicious activity

### Database Backups

```bash
# Backup SQLite database
cp /app/data/changelog.db /backups/changelog-$(date +%Y%m%d).db

# Restore from backup
cp /backups/changelog-20240101.db /app/data/changelog.db
```

## 🛠️ Development

### Project Structure

```
ShipShipShip/
├── backend/              # Go backend
│   ├── handlers/        # API route handlers
│   ├── middleware/      # Authentication & CORS
│   ├── models/         # Database models
│   ├── database/       # Database setup
│   └── main.go         # Application entry point
├── frontend/            # SvelteKit frontend
│   ├── src/
│   │   ├── routes/     # Page components
│   │   └── lib/        # Shared components & utilities
│   └── static/         # Static assets
├── docker-compose.yml   # Development environment
├── Dockerfile          # Production build
└── README.md           # This file
```

### Development Commands

```bash
# Start development servers
./start-dev.sh

# Backend only
cd backend && go run main.go

# Frontend only
cd frontend && npm run dev

# Build for production
docker build -t shipshipship .

# Run tests
cd backend && go test ./...
cd frontend && npm run check
```

### Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with ❤️ using modern web technologies
- Inspired by the need for simple, effective changelog management
- Community-driven development and feedback

## 🔗 Links

- **GitHub**: [https://github.com/GauthierNelkinsky/ShipShipShip](https://github.com/GauthierNelkinsky/ShipShipShip)
- **Issues**: [Report bugs or request features](https://github.com/GauthierNelkinsky/ShipShipShip/issues)
- **Discussions**: [Community discussions](https://github.com/GauthierNelkinsky/ShipShipShip/discussions)

---

**Shipped with ShipShipShip** 🚢