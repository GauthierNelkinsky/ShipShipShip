# 🚢 ShipShipShip

A modern, self-hostable changelog and roadmap platform that helps you share product updates with your community and gather feedback through feature voting.

![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.21-blue.svg)
![Node Version](https://img.shields.io/badge/node-18+-green.svg)
![Docker](https://img.shields.io/badge/docker-ready-blue.svg)

![demo](https://github.com/user-attachments/assets/7382c4b7-f936-4698-a8b0-7054b2f8b091)


## 🔗 Links

- **📋 [Website](https://shipshipship.io/)**
- **🔗 [Live Demo](https://demo.shipshipship.io/admin)** (Login: `demo` / `demo`)
- **🐳 [Docker Hub](https://hub.docker.com/r/nelkinsky/shipshipship)**


## ✨ Features

- 📋 **Rich Changelog Management** - Create, edit, and organize events with Markdown support and rich text editor
- 🗳️ **Community Voting** - Let users vote on upcoming features and gather feedback
- 📊 **Kanban Roadmap** - Drag-and-drop board with multiple event statuses (Backlog, Vote, Doing, Released, etc.)
- 📧 **Newsletter System** - Users can subscribe to receive email updates about new features and releases
- 🎨 **Modern Interface** - Responsive design with dark/light themes and real-time updates
- 🛠️ **Self-Hostable** - Complete control over your data with Docker deployment
- 🔌 **RESTful API** - Full API access for integrations and custom workflows
- 📮 **Email Notifications** - Configure SMTP settings through the admin interface to send newsletters


## 🏗️ Tech Stack

**Frontend:** SvelteKit, TailwindCSS, Shadcn/ui
**Backend:** Go (Gin), SQLite, GORM, JWT
**Deployment:** Docker, Multi-stage builds

## 🚀 Quick Start

```bash
# Option 1: Clone and start with Docker
git clone https://github.com/GauthierNelkinsky/ShipShipShip.git
cd ShipShipShip
docker-compose up -d

# Option 2: Run directly from Docker Hub
docker run -d \
  -p 8080:8080 \
  -e ADMIN_USERNAME=admin \
  -e ADMIN_PASSWORD=admin \
  -e JWT_SECRET=your-secret-key-change-in-production \
  -v changelog_data:/app/data \
  nelkinsky/shipshipship:latest

# Access the public page at http://localhost:8080
# Access the admin interface at http://localhost:8080/admin
# Default credentials: admin/admin (change immediately!)
# Use the ADMIN_USERNAME and ADMIN_PASSWORD from your .env file
```

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ADMIN_USERNAME` | `admin` | Admin login username |
| `ADMIN_PASSWORD` | `admin` | Admin login password |
| `JWT_SECRET` | `your-secret-key-change-in-production` | JWT signing secret |
| `PORT` | `8080` | Server port |
| `DB_PATH` | `./data/changelog.db` | SQLite database path |

### Docker Compose Example

```yaml
version: "3.8"
services:
  changelog:
    image: nelkinsky/shipshipship:latest
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

## 📊 Event Statuses

- **📝 Backlog** - Ideas and planned features
- **🗳️ Proposed** - Features users can vote on
- **🔄 Upcoming** - Currently in development
- **🚀 Released** - Published features (main timeline)
- **📦 Archived** - Internal events (hidden from public)

## 🛠️ Development

```bash
# Development mode
./start-dev.sh

# Manual setup
cd backend && go run main.go &
cd frontend && npm run dev
```

## 📧 Newsletter Setup

To enable the newsletter system:

1. **Access Admin Interface**: Go to `/admin` and log in with your admin credentials
2. **Configure Mail Settings**: Navigate to Mail Settings and configure your SMTP server:
   - SMTP Host (e.g., `smtp.gmail.com`)
   - SMTP Port (usually `587` for TLS, `465` for SSL)
   - SMTP Username and Password
   - SMTP Encryption (None, TLS, or SSL)
   - From Email and From Name
3. **Test Configuration**: Use the test email feature to verify your settings
4. **User Subscription**: Users can subscribe to updates using the newsletter subscription component
5. **Send Newsletters**: Publish events and send automated newsletters, or manage email templates and send custom newsletters

**Supported SMTP Providers**: Gmail, Outlook, SendGrid, Mailgun, or any standard SMTP server

**Email Templates**: Customize email templates for different event types (upcoming features, new releases, proposed features, welcome emails) through the admin interface.

## 🔒 Security Checklist

- [ ] Change default admin credentials
- [ ] Set strong JWT secret
- [ ] Use HTTPS in production
- [ ] Regular database backups
- [ ] Update dependencies regularly
- [ ] Configure SMTP settings in admin interface
- [ ] Test email configuration before going live
- [ ] Use app passwords for Gmail/Outlook when available

## 📝 License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

---

**Shipped with ShipShipShip** 🚢
