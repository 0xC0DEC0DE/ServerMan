================================================================================
                              CCS SERVER DASHBOARD
                         Comprehensive Server Management System
================================================================================

┌─────────────────────────────────────────────────────────────────────────────┐
│                                OVERVIEW                                     │
└─────────────────────────────────────────────────────────────────────────────┘

ServerCon is a full-stack web application for managing and monitoring CCS 
(Calloway Compute Service) servers. The system provides a modern dashboard 
interface for server management, user administration, and real-time monitoring
capabilities.

┌─────────────────────────────────────────────────────────────────────────────┐
│                            SYSTEM ARCHITECTURE                             │
└─────────────────────────────────────────────────────────────────────────────┘

    ┌─────────────────┐    HTTP/REST API    ┌─────────────────┐
    │   React Frontend │ ◄─────────────────► │  Go Backend     │
    │   (Port 3000)    │                     │  (Port 8080)    │
    └─────────────────┘                     └─────────────────┘
                                                     │
                                                     ▼
                                            ┌─────────────────┐
                                            │  Database       │
                                            │  (ClickHouse)   │
                                            └─────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                               FEATURES                                     │
└─────────────────────────────────────────────────────────────────────────────┘

🖥️  SERVER MANAGEMENT
   • Real-time server status monitoring
   • Server configuration and control
   • Snapshot management and creation
   • Server credentials access
   • VNC/Remote desktop integration

👥 USER ADMINISTRATION
   • User account management
   • Group-based access control
   • Role-based permissions
   • Admin panel for user operations

🔐 AUTHENTICATION & SECURITY
   • Session-based authentication
   • Secure credential storage
   • Group-based authorization
   • Automatic session cleanup

📊 MONITORING & SYNC
   • Automated server data synchronization
   • API key status monitoring
   • Background service management
   • Real-time status updates

┌─────────────────────────────────────────────────────────────────────────────┐
│                            TECHNOLOGY STACK                                │
└─────────────────────────────────────────────────────────────────────────────┘

BACKEND (Go)
├── Framework: Gin Web Framework
├── Database: ClickHouse
├── Authentication: Custom session management
├── Sync Service: Background data synchronization
└── API: RESTful endpoints

FRONTEND (React)
├── Build Tool: Rsbuild
├── UI Framework: Bulma CSS
├── Router: React Router v7
├── Animation: Framer Motion
├── Remote Desktop: React VNC
└── Code Quality: Biome

┌─────────────────────────────────────────────────────────────────────────────┐
│                              INSTALLATION                                  │
└─────────────────────────────────────────────────────────────────────────────┘

PREREQUISITES
• Go 1.19+
• Node.js 18+
• pnpm package manager
• ClickHouse database

BACKEND SETUP
1. Clone the repository
   git clone <repository-url>
   cd servercon

2. Install Go dependencies
   go mod download

3. Set up environment variables
   cp .env.example .env
   # Edit .env with your configuration

4. Run database migrations
   go run main.go

5. Start the backend server
   go run main.go
   # Server runs on http://localhost:8080

FRONTEND SETUP
1. Navigate to frontend directory
   cd frontend

2. Install dependencies
   pnpm install

3. Start development server
   pnpm dev
   # Frontend runs on http://localhost:3000

4. Build for production
   pnpm build

┌─────────────────────────────────────────────────────────────────────────────┐
│                            API ENDPOINTS                                   │
└─────────────────────────────────────────────────────────────────────────────┘

AUTHENTICATION
POST   /login                    - Initiate login process
GET    /callback                 - Authentication callback
GET    /logout                   - User logout

SERVER MANAGEMENT
GET    /api/servers              - List all servers
GET    /api/server/:id           - Get server details
GET    /api/server/:id/snapshots - Get server snapshots
GET    /api/server/:id/credentials - Get server credentials

USER MANAGEMENT (Admin)
GET    /api/admin/users          - List all users
POST   /api/admin/users          - Add new user
PUT    /api/admin/users/:email   - Update user groups
DELETE /api/admin/users/:email   - Delete user

SYSTEM ADMINISTRATION
POST   /api/admin/sync           - Trigger manual sync
GET    /api/admin/api-keys/status - Check API key status
GET    /api/user                 - Get current user info
GET    /api/os_options           - Get OS options
GET    /api/apps                 - Get available applications

┌─────────────────────────────────────────────────────────────────────────────┐
│                            CONFIGURATION                                   │
└─────────────────────────────────────────────────────────────────────────────┘

ENVIRONMENT VARIABLES
• Database connection settings
• Authentication configuration
• API endpoint URLs
• Sync service intervals
• Session security settings

DATABASE SCHEMA
• users: User account information
• server_snapshots_cache: Server snapshot data
• sessions: User session management
• Additional tables for server and application data

┌─────────────────────────────────────────────────────────────────────────────┐
│                              DEVELOPMENT                                   │
└─────────────────────────────────────────────────────────────────────────────┘

BACKEND DEVELOPMENT
• Follow Go best practices
• Use structured logging
• Implement proper error handling
• Write unit tests for critical functions

FRONTEND DEVELOPMENT
• Use modern React patterns with hooks
• Follow Biome linting rules
• Implement responsive design with Bulma
• Use TypeScript for type safety (recommended)

CODE QUALITY
• Backend: gofmt, golint, go vet
• Frontend: Biome for formatting and linting
• Version control: Git with proper commit messages

┌─────────────────────────────────────────────────────────────────────────────┐
│                              DEPLOYMENT                                    │
└─────────────────────────────────────────────────────────────────────────────┘

PRODUCTION BUILD
1. Build frontend assets
   cd frontend && pnpm build

2. Compile Go binary
   go build -o servercon main.go

3. Deploy with proper environment configuration
4. Set up reverse proxy (nginx recommended)
5. Configure SSL certificates
6. Set up monitoring and logging

┌─────────────────────────────────────────────────────────────────────────────┐
│                               SECURITY                                     │
└─────────────────────────────────────────────────────────────────────────────┘

• All API endpoints require authentication
• Group-based access control for sensitive operations
• Secure session management with automatic cleanup
• HTTPS recommended for production
• Regular security updates for dependencies

┌─────────────────────────────────────────────────────────────────────────────┐
│                              MONITORING                                    │
└─────────────────────────────────────────────────────────────────────────────┘

• Automatic sync service runs every 5 minutes
• Session cleanup runs hourly
• API key status monitoring
• Server health checks
• Real-time dashboard updates

┌─────────────────────────────────────────────────────────────────────────────┐
│                              SUPPORT                                       │
└─────────────────────────────────────────────────────────────────────────────┘

For issues, feature requests, or contributions:
• Check the project documentation
• Review existing issues
• Follow the contribution guidelines
• Contact the development team

================================================================================
                            SERVER MANAGEMENT MADE SIMPLE
================================================================================