# 🚢 ShipShipShip

A modern, self-hostable changelog and roadmap platform with emoji reactions, custom themes, and automated newsletters.

![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.21-blue.svg)
![Node Version](https://img.shields.io/badge/node-20+-green.svg)
![Docker](https://img.shields.io/badge/docker-ready-blue.svg)

<img width="1280" height="832" alt="SSSBanner" src="https://github.com/user-attachments/assets/cc286bff-7838-4b35-b065-44ef3c99bc50" />

## 🔗 Links

- **📋 [Website](https://shipshipship.io/)**
- **🔗 [Live Demo](https://demo.shipshipship.io/admin)** (Login: `demo` / `demo`)
- **🐳 [Docker Hub](https://hub.docker.com/r/nelkinsky/shipshipship)**

## ✨ Features

- 📋 **Rich Event Management** - TipTap editor with markdown support, tags, and media uploads
- 😊 **Emoji Reactions** - 8 reaction types (👍❤️🔥🎉👀💡🤔👎) for community feedback
- 🗳️ **Voting System** - Let users vote on proposed features
- 📊 **Kanban Board** - Drag-and-drop interface with customizable statuses
- 🎨 **Theme System** - Install custom themes with manifest-based configuration
- 📧 **Newsletter Automation** - Auto-send emails when events change status
- 📮 **Email Templates** - Customizable templates for different event types
- 🔧 **Admin Dashboard** - Full-featured SvelteKit admin panel
- 🔌 **RESTful API** - Complete API for integrations

## 🏗️ Tech Stack

**Admin:** SvelteKit 2 · Svelte 5 · TailwindCSS · shadcn-svelte · TipTap  
**Backend:** Go 1.21 · Gin · SQLite · GORM  
**Deploy:** Docker (AMD64 & ARM64)

## 🚀 Quick Start

> 🔑 Admin Panel: `/admin`
> Visit http://localhost:8080/admin (or https://your-domain/admin) right after installation to complete setup.

### Docker (Recommended)

```bash
docker run -d \
  -p 8080:8080 \
  -e ADMIN_USERNAME=admin \
  -e ADMIN_PASSWORD=changeme \
  -e JWT_SECRET=your-secret-key \
  -e BASE_URL=https://changelog.yourdomain.com \
  -v shipshipship_data:/app/data \
  nelkinsky/shipshipship:latest
```

🔑 **Admin Panel:** http://localhost:8080/admin

### Docker Compose

```yaml
version: "3.8"
services:
  shipshipship:
    image: nelkinsky/shipshipship:latest
    ports:
      - "8080:8080"
    environment:
      - ADMIN_USERNAME=admin
      - ADMIN_PASSWORD=changeme
      - JWT_SECRET=your-secret-key
      - BASE_URL=https://changelog.yourdomain.com
      - GIN_MODE=release
    volumes:
      - shipshipship_data:/app/data
    restart: unless-stopped

volumes:
  shipshipship_data:
```

🔑 **Admin Panel:** http://localhost:8080/admin

### Local Development

```bash
git clone https://github.com/GauthierNelkinsky/ShipShipShip.git
cd ShipShipShip

# Start backend + admin dev server
./start-dev.sh

# Or backend only
./quick-start.sh
```

**Dev URLs:**
- Backend: http://localhost:8080
- Admin Dev: http://localhost:5173

## ☁️ One-Click Deploy

[![Deploy on RepoCloud](https://d16t0pc4846x52.cloudfront.net/deploylobe.svg)](https://repocloud.io/details/ShipShipShip/)

## ⚙️ Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `ADMIN_USERNAME` | `admin` | Admin username |
| `ADMIN_PASSWORD` | `admin` | Admin password |
| `JWT_SECRET` | `your-secret-key-change-in-production` | JWT signing key |
| `BASE_URL` | _(auto-detected)_ | Base URL of your instance (e.g., `https://changelog.yourdomain.com`) - used for email unsubscribe links |
| `PORT` | `8080` | Server port |
| `GIN_MODE` | `debug` | `debug` or `release` |
| `DB_PATH` | `./data/changelog.db` | Database path |

## 🎨 Theme System

ShipShipShip separates the admin interface from the public-facing changelog through installable themes:

1. **Install Theme:** Upload via `/admin/customization/theme`
2. **Configure:** Customize theme settings (colors, layout, etc.)
3. **Map Statuses:** Connect your event statuses to theme categories
4. **Publish:** Your themed changelog appears at the root URL

Without a theme, the root URL shows the admin interface for initial setup.

## 📧 Newsletter Setup

1. Go to `/admin/newsletter/settings`
2. Configure SMTP settings (Gmail, Outlook, SendGrid, etc.)
3. Test configuration
4. Enable automation for status-based triggers
5. Customize email templates

**Automation:** Automatically send newsletters when events move to specific statuses (e.g., "Released").

## 🛠️ Development

```bash
# Full dev mode (hot reload)
./start-dev.sh

# Rebuild everything
./start-dev.sh --rebuild

# Backend only
./quick-start.sh
```

### Project Structure

```
admin/              # SvelteKit admin panel (SPA)
  ├── src/routes/admin/
  │   ├── events/          # Kanban board
  │   ├── newsletter/      # Email management
  │   └── customization/   # Themes & branding
  └── build/        # Static output (served by backend)

backend/            # Go API server
  ├── handlers/     # API endpoints
  ├── models/       # Database models
  ├── services/     # Business logic (email, automation)
  └── main.go       # Server entry point

data/               # SQLite + uploads + themes
```

## 📖 API Examples

**Public:**
```bash
# Get public events
curl http://localhost:8080/api/events

# Add reaction
curl -X POST http://localhost:8080/api/events/1/reactions \
  -H "Content-Type: application/json" \
  -d '{"reaction_type":"thumbs_up"}'

# Subscribe to newsletter
curl -X POST http://localhost:8080/api/newsletter/subscribe \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com"}'
```

**Admin (requires JWT):**
```bash
# Create event
curl -X POST http://localhost:8080/api/admin/events \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"New Feature","status":"Proposed","content":"..."}'
```

## 🤝 Contributing

Contributions welcome! Fork the repo, create a feature branch, and submit a PR.

## 📝 License

Apache 2.0 - see [LICENSE](LICENSE)

## 💬 Support

- **Issues:** [GitHub Issues](https://github.com/GauthierNelkinsky/ShipShipShip/issues)
- **Website:** [shipshipship.io](https://shipshipship.io/)
- **Demo:** [demo.shipshipship.io](https://demo.shipshipship.io/)

---

**Built with ❤️ and shipped with ShipShipShip** 🚢
