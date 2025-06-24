forum-frontend/
├── main.go                     # Entry point - server setup and routing
├── go.mod                      # Go modules file
├── go.sum                      # Dependencies checksum
├── config/
│   └── config.go              # Configuration management (API URLs, timeouts)
├── internal/
│   ├── api/                   # APIClient service layer
│   │   ├── client.go          # Main APIClient struct and methods
│   │   ├── posts.go           # Post-related API calls
│   │   ├── users.go           # User-related API calls
│   │   ├── comments.go        # Comment-related API calls
│   │   ├── auth.go            # Authentication API calls
│   │   └── errors.go          # Custom API error types
│   ├── handlers/              # HTTP request handlers
│   │   ├── home.go            # Home page handler
│   │   ├── posts.go           # Post CRUD handlers
│   │   ├── auth.go            # Login/register handlers
│   │   ├── profile.go         # User profile handlers
│   │   └── static.go          # Static file handling
│   ├── models/                # Data structures
│   │   ├── post.go            # Post-related structs
│   │   ├── user.go            # User-related structs
│   │   ├── comment.go         # Comment-related structs
│   │   ├── response.go        # API response structs
│   │   └── page_data.go       # Template data structs
│   ├── middleware/            # HTTP middleware (optional for now)
│   │   ├── logging.go         # Request logging
│   │   └── recovery.go        # Panic recovery
│   └── utils/                 # Helper utilities
│       ├── template.go        # Template rendering helpers
│       ├── validation.go      # Input validation
│       └── helpers.go         # General helper functions
├── templates/                 # HTML templates
│   ├── layouts/
│   │   ├── base.html          # Base layout template
│   │   └── auth.html          # Authentication layout
│   ├── pages/
│   │   ├── home.html          # Home page template
│   │   ├── post_detail.html   # Single post view
│   │   ├── post_create.html   # Create post form
│   │   ├── login.html         # Login page
│   │   ├── register.html      # Registration page
│   │   └── profile.html       # User profile page
│   ├── components/
│   │   ├── post_card.html     # Reusable post card
│   │   ├── comment.html       # Comment component
│   │   ├── pagination.html    # Pagination component
│   │   └── sidebar.html       # Categories sidebar
│   └── partials/
│       ├── header.html        # Site header
│       ├── footer.html        # Site footer
│       └── navigation.html    # Navigation menu
├── static/                    # Static assets
│   ├── css/
│   │   ├── main.css           # Main stylesheet
│   │   ├── components.css     # Component styles
│   │   └── responsive.css     # Mobile responsive styles
│   ├── js/
│   │   ├── main.js            # Main JavaScript file
│   │   ├── forms.js           # Form handling
│   │   └── interactions.js    # UI interactions (likes, etc.)
│   ├── images/
│   │   ├── logo.png
│   │   └── icons/
│   └── fonts/
├── .env                       # Environment variables (gitignored)
├── .gitignore                 # Git ignore file
└── README.md                  # Project documentation

# Key Files Breakdown:

## Core Application Files:
main.go                    # Server startup, routing, dependency injection
config/config.go           # Centralized configuration

## APIClient Service (Heart of the architecture):
internal/api/client.go     # APIClient struct, common HTTP methods
internal/api/posts.go      # GetPosts(), GetPost(), CreatePost(), etc.
internal/api/users.go      # GetUser(), LoginUser(), RegisterUser(), etc.
internal/api/comments.go   # GetComments(), CreateComment(), etc.
internal/api/auth.go       # Authentication-specific API calls
internal/api/errors.go     # Custom error types for API responses

## HTTP Layer:
internal/handlers/home.go      # Homepage logic
internal/handlers/posts.go     # Post CRUD operations
internal/handlers/auth.go      # Login/logout handlers
internal/handlers/profile.go   # User profile handlers

## Data Layer:
internal/models/post.go        # Post, Category, PaginatedResponse structs
internal/models/user.go        # User, UserProfile structs
internal/models/comment.go     # Comment struct
internal/models/response.go    # APIResponse, ErrorResponse structs
internal/models/page_data.go   # Template-specific data structs

## Templates:
templates/layouts/base.html    # Main layout with header/footer
templates/pages/*.html         # Individual page templates
templates/components/*.html    # Reusable components

## Static Assets:
static/css/main.css           # Styling
static/js/main.js             # Minimal JavaScript for interactions