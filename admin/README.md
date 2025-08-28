# Admin Panel - ShipShipShip

This is the admin panel for ShipShipShip, a standalone SvelteKit application for managing your changelog and project settings.

## Features

- **Event Management**: Create, edit, and organize changelog events
- **Customization**: Brand your changelog with custom colors, logos, and styling
- **Tag Management**: Organize events with custom tags
- **Newsletter Management**: Manage subscribers and send automated newsletters
- **User Authentication**: Secure admin access with login/logout functionality

## Getting Started

### Prerequisites

- Node.js 18+ 
- npm or yarn

### Installation

1. Install dependencies:
```bash
npm install
```

2. Start the development server:
```bash
npm run dev
```

3. Open [http://localhost:5173](http://localhost:5173) in your browser

### Building for Production

```bash
npm run build
```

## Project Structure

```
src/
├── routes/
│   ├── login/           # Authentication pages
│   ├── events/          # Event management
│   ├── customization/   # Branding and styling
│   └── newsletter/      # Newsletter management
├── lib/
│   ├── components/      # Reusable UI components
│   ├── stores/          # Svelte stores for state management
│   └── api.ts          # API client for backend communication
└── app.html            # HTML template
```

## Authentication

The admin panel requires authentication. Default credentials are set in your backend configuration. The admin uses JWT tokens for session management.

## API Integration

The admin panel communicates with the ShipShipShip backend API. Make sure your backend is running and accessible before using the admin panel.

## Deployment

This admin panel can be deployed independently from the public changelog. Popular deployment options include:

- **Vercel**: Connect your repository for automatic deployments
- **Netlify**: Deploy via Git or manual uploads
- **Static hosting**: Build and upload the `build/` directory to any static host

## Configuration

The admin panel automatically detects the backend API based on your deployment setup. For custom configurations, modify the `API_BASE` constant in `src/lib/api.ts`.

## Development

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build locally
- `npm run check` - Run type checking