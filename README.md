# Forum Backend API

A high-performance REST API backend for a forum application built with Go, featuring user authentication, posts, comments, reactions, and comprehensive user profiles.

## 🚀 Features

### Core Functionality
- **User Management**: Registration, authentication, and profile management
- **Posts**: Create, read, update, delete posts with category support
- **Comments**: Threaded commenting system with full CRUD operations
- **Reactions**: Like/dislike system for both posts and comments
- **Categories**: Organize posts by predefined categories
- **User Profiles**: Comprehensive statistics and activity tracking

### Technical Features
- **Session-based Authentication**: Secure session management with IP validation
- **Rate Limiting**: Configurable request throttling
- **Pagination**: Efficient data loading with customizable page sizes
- **Sorting**: Multiple sorting options (newest, oldest, likes, comments)
- **Security**: CORS, security headers, input validation, and SQL injection protection
- **Database**: SQLite with optimized indexes and WAL mode
- **Docker Support**: Production-ready containerization

## 📋 Prerequisites

- **Go**: Version 1.24 or higher
- **Docker**: For containerized deployment (optional)
- **SQLite**: Embedded database (no separate installation needed)

## 🛠️ Installation & Setup

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd forum-backend
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env file with your configuration
   ```

4. **Run the application**
   ```bash
   go run main.go
   ```

The API server will start on `http://localhost:8080` by default.

### Docker Deployment

1. **Create Docker network**
   ```bash
   docker network create forum-network
   ```

2. **Build the image**
   ```bash
   docker build -t api-image:latest .
   ```

3. **Run the container**
   ```bash
   docker run -d \
     -p 8080:8080 \
     -v api_db_data:/app/DBPath \
     --name api-container \
     --network forum-network \
     api-image:latest
   ```

## ⚙️ Configuration

The application uses environment variables for configuration. Key settings include:

### Server Configuration
```env
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
```

### Database Configuration
```env
DB_PATH=./DBPath/forum.db
DB_MAX_CONNECTIONS=10
```

### Security Configuration
```env
SESSION_DURATION=24h
BCRYPT_COST=16
SESSION_NAME=forum_session
```

### Content Limits
```env
MIN_POST_CONTENT_LENGTH=10
MAX_POST_CONTENT_LENGTH=500
MIN_COMMENT_LENGTH=5
MAX_COMMENT_LENGTH=150
```

### CORS Configuration
```env
ALLOWED_ORIGINS=http://localhost:3000,http://frontend:3000
ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With,Cookie
```

See `.env.example` for all available configuration options.

## 📚 API Documentation

### Authentication Endpoints

#### Register User
```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "SecurePass123!",
  "confirm_password": "SecurePass123!"
}
```

#### Login
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

#### Logout
```http
POST /api/auth/logout
Cookie: forum_session=<session_id>
```

#### Get Current User
```http
POST /api/auth/me
Cookie: forum_session=<session_id>
```

### Post Endpoints

#### Get All Posts
```http
GET /api/posts?limit=20&offset=0&sort=newest
```

#### Get Single Post
```http
GET /api/posts/view/{post_id}
```

#### Get Posts by Category
```http
GET /api/posts/by-category/{category_id}?sort=likes&limit=10
```

#### Create Post
```http
POST /api/posts/create
Cookie: forum_session=<session_id>
Content-Type: application/json

{
  "content": "This is my post content...",
  "category_names": ["Programming", "Web Development"]
}
```

#### Update Post
```http
PUT /api/posts/edit/{post_id}
Cookie: forum_session=<session_id>
Content-Type: application/json

{
  "content": "Updated post content...",
  "category_names": ["Programming"]
}
```

#### Delete Post
```http
DELETE /api/posts/remove/{post_id}
Cookie: forum_session=<session_id>
```

### Comment Endpoints

#### Get Comments for Post
```http
GET /api/comments/for-post/{post_id}?sort=oldest&limit=20&offset=0
```

#### Get Single Comment
```http
GET /api/comments/view/{comment_id}
```

#### Create Comment
```http
POST /api/comments/create-on-post/{post_id}
Cookie: forum_session=<session_id>
Content-Type: application/json

{
  "content": "This is my comment..."
}
```

#### Update Comment
```http
PUT /api/comments/edit/{comment_id}
Cookie: forum_session=<session_id>
Content-Type: application/json

{
  "content": "Updated comment content..."
}
```

#### Delete Comment
```http
DELETE /api/comments/remove/{comment_id}
Cookie: forum_session=<session_id>
```

### Reaction Endpoints

#### Toggle Post Reaction
```http
POST /api/reactions/posts/toggle
Cookie: forum_session=<session_id>
Content-Type: application/json

{
  "post_id": "post_uuid",
  "reaction_type": 1
}
```

#### Toggle Comment Reaction
```http
POST /api/reactions/comments/toggle
Cookie: forum_session=<session_id>
Content-Type: application/json

{
  "comment_id": "comment_uuid",
  "reaction_type": 2
}
```

**Reaction Types:**
- `1` = Like
- `2` = Dislike

### User Profile Endpoints

#### Get User Profile
```http
GET /api/users/profile/{user_id}
Cookie: forum_session=<session_id>
```

#### Get User's Posts
```http
GET /api/users/posts/{user_id}?sort=newest&limit=20&offset=0
Cookie: forum_session=<session_id>
```

#### Get User's Liked Posts
```http
GET /api/users/liked-posts/{user_id}?sort=newest&limit=20&offset=0
Cookie: forum_session=<session_id>
```

#### Get User's Commented Posts
```http
GET /api/users/commented-posts/{user_id}?sort=newest&limit=20&offset=0
Cookie: forum_session=<session_id>
```

### Category Endpoints

#### Get All Categories
```http
GET /api/categories
```

### Query Parameters

#### Pagination
- `limit` - Number of items per page (default: 20, max: 50)
- `offset` - Number of items to skip (default: 0)

#### Sorting Options

**For Posts:**
- `newest` - Sort by creation date (newest first) - **default**
- `oldest` - Sort by creation date (oldest first)
- `likes` - Sort by like count (most liked first)
- `comments` - Sort by comment count (most commented first)

**For Comments:**
- `oldest` - Sort by creation date (oldest first) - **default**
- `newest` - Sort by creation date (newest first)
- `likes` - Sort by like count (most liked first)

## 🗄️ Database Schema

### Core Tables
- **users** - User accounts and authentication
- **sessions** - User session management
- **posts** - Forum posts
- **comments** - Post comments
- **categories** - Post categories
- **post_categories** - Post-category relationships

### Reaction Tables
- **post_reactions** - Like/dislike reactions on posts
- **comment_reactions** - Like/dislike reactions on comments

### Key Features
- **UUIDs** for all primary keys
- **Foreign key constraints** with CASCADE delete
- **Optimized indexes** for query performance
- **WAL mode** for better concurrency
- **Timestamp tracking** for created_at and updated_at

## 📊 Response Format

All API responses follow a consistent JSON format:

### Success Response
```json
{
  "success": true,
  "data": {
    // Response data here
  }
}
```

### Error Response
```json
{
  "success": false,
  "error": "Error message description"
}
```

### Paginated Response
```json
{
  "success": true,
  "data": {
    "posts": [...],
    "pagination": {
      "current_page": 1,
      "total_pages": 5,
      "total_count": 100,
      "per_page": 20,
      "has_next": true,
      "has_previous": false
    }
  }
}
```

## 🔒 Security Features

### Authentication & Authorization
- **Session-based authentication** with secure cookie handling
- **IP validation** with smart network change detection
- **Password hashing** using bcrypt with configurable cost
- **Session expiration** and automatic cleanup

### Input Validation
- **Email format validation** using Go's mail package
- **Password complexity requirements** (uppercase, lowercase, number, special character)
- **Username format validation** (alphanumeric and underscore only)
- **Content length validation** for posts and comments
- **Profanity filtering** for user-generated content

### Security Headers
- **X-Content-Type-Options**: nosniff
- **X-Frame-Options**: DENY
- **X-XSS-Protection**: enabled
- **Content-Security-Policy**: restricted to trusted origins
- **Strict-Transport-Security**: for HTTPS connections

### Rate Limiting
- **Configurable request limits** per IP address
- **Time-window based** throttling
- **Automatic cleanup** of expired rate limit data

## 🏗️ Architecture

### Project Structure
```
forum-backend/
├── main.go                 # Application entry point
├── config/                 # Configuration management
├── database/              # Database initialization & schemas
├── internal/
│   ├── handlers/          # HTTP request handlers
│   ├── middleware/        # HTTP middleware
│   ├── models/           # Data models
│   ├── repository/       # Data access layer
│   ├── routes/           # Route definitions
│   └── utils/            # Utility functions
├── queries/              # SQL query definitions
├── Dockerfile           # Container configuration
├── .env.example        # Environment configuration template
└── docker.md           # Docker deployment guide
```

### Design Patterns
- **Repository Pattern** - Clean separation of data access logic
- **Middleware Pattern** - Modular request processing
- **Factory Pattern** - Configuration and dependency injection
- **Transaction Pattern** - Atomic database operations

### Key Components

#### Repository Layer
- **UserRepository** - User management and authentication
- **PostsRepository** - Post CRUD operations
- **CommentRepository** - Comment management
- **CategoryRepository** - Category operations
- **PostReactionRepository** - Post like/dislike handling
- **CommentReactionRepository** - Comment reaction handling
- **SessionRepository** - Session management

#### Middleware Stack
- **Authentication** - Session validation and user context
- **CORS** - Cross-origin request handling
- **Rate Limiting** - Request throttling
- **Security Headers** - Security-focused HTTP headers

## 🐳 Docker Configuration

### Multi-stage Build
The Dockerfile uses a multi-stage build for optimal image size:

1. **Builder Stage** - Compiles the Go application with static linking
2. **Runtime Stage** - Minimal Alpine Linux image with only necessary dependencies

### Security Features
- **Non-root user** execution
- **Static linking** for security
- **Minimal attack surface** with Alpine Linux
- **Health checks** for container monitoring

### Volume Persistence
- Database files are stored in `/app/DBPath` volume
- Automatic database initialization on first run
- Data persistence across container restarts

## 🧪 Testing

### Manual Testing
Use the provided endpoints with tools like:
- **Postman** - For comprehensive API testing
- **curl** - For command-line testing
- **Thunder Client** (VS Code) - For integrated testing

### Example curl Commands

```bash
# Register a new user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "TestPass123!",
    "confirm_password": "TestPass123!"
  }'

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123!"
  }'

# Get all posts
curl -X GET http://localhost:8080/api/posts \
  -b cookies.txt

# Create a post
curl -X POST http://localhost:8080/api/posts/create \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{
    "content": "This is my first post!",
    "category_names": ["General discussion"]
  }'
```

## 🚦 Monitoring & Logging

### Health Checks
- **Container health check** endpoint monitoring
- **Database connection** validation
- **Session cleanup** monitoring

### Logging
- **Structured logging** for security events
- **IP change detection** with detailed logging
- **Error tracking** for debugging

## 🔧 Development

### Adding New Features

1. **Models** - Define data structures in `internal/models/`
2. **Repository** - Add data access methods in `internal/repository/`
3. **Handlers** - Create HTTP handlers in `internal/handlers/`
4. **Routes** - Register routes in `internal/routes/routes.go`
5. **Middleware** - Add middleware if needed in `internal/middleware/`

### Database Migrations
Database schema is automatically created on first run. For schema changes:
1. Update `database/sql_statements.go`
2. Handle data migration in `database/initialization_db.go`

## 📈 Performance Optimization

### Database
- **Optimized indexes** for common queries
- **WAL mode** for better concurrency
- **Connection pooling** with configurable limits
- **Query optimization** with prepared statements

### Application
- **Efficient pagination** to handle large datasets
- **Minimal data transfer** with lightweight response objects
- **Memory management** with proper context handling
- **Concurrent safety** with mutex protection for rate limiting

## 🤝 Contributing

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Code Style
- Follow Go conventions and `gofmt` formatting
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small
## 📞 Support

For support and questions:
- **Issues**: Create an issue on GitHub
- **Documentation**: Check this README and code comments
- **Configuration**: Review `.env.example` for all options

## 🔄 Version History

### v1.0.0
- Initial release with core forum functionality
- User authentication and session management
- Posts and comments with reactions
- Category support and user profiles
- Docker containerization
- Comprehensive security features