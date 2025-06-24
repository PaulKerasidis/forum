# Sorting and Pagination Guide

## 📋 **Overview**

This forum API implements a flexible and efficient sorting and pagination system that works across both **posts** and **comments** with a unified approach.

---

## 🔄 **Pagination System**

### **How It Works**
- **Limit**: Number of items per page (default: 20, max: 50)
- **Offset**: Number of items to skip (starts at 0)
- **Metadata**: Rich pagination information in responses

### **URL Parameters**
```
?limit=20&offset=0    # First page, 20 items
?limit=20&offset=20   # Second page, 20 items  
?limit=10&offset=50   # Sixth page, 10 items
```

### **Pagination Response Format**
```json
{
  "success": true,
  "data": {
    "posts": [...],
    "pagination": {
      "current_page": 1,
      "total_pages": 10,
      "total_count": 200,
      "per_page": 20,
      "has_next": true,
      "has_previous": false
    }
  }
}
```

### **Implementation Details**
- **Default limit**: 20 items
- **Maximum limit**: 50 items (prevents abuse)
- **Minimum limit**: 1 item
- **Offset validation**: Cannot be negative
- **Auto-correction**: Invalid values are corrected to defaults

---

## 📊 **Sorting System**

### **Posts Sorting Options**
| Sort Option | Description | SQL Order |
|-------------|-------------|-----------|
| `newest` | Latest posts first (default) | `ORDER BY p.created_at DESC` |
| `oldest` | Oldest posts first | `ORDER BY p.created_at ASC` |
| `likes` | Most liked posts first | `ORDER BY like_count DESC, p.created_at DESC` |
| `comments` | Most commented posts first | `ORDER BY comment_count DESC, p.created_at DESC` |

### **Comments Sorting Options**
| Sort Option | Description | SQL Order |
|-------------|-------------|-----------|
| `oldest` | Oldest comments first (default) | `ORDER BY c.created_at ASC` |
| `newest` | Latest comments first | `ORDER BY c.created_at DESC` |
| `likes` | Most liked comments first | `ORDER BY like_count DESC, c.created_at ASC` |

### **URL Parameter**
```
?sort=newest     # Posts: newest first
?sort=likes      # Posts: most liked first
?sort=oldest     # Comments: oldest first (conversation flow)
```

---

## 🎯 **Smart Defaults**

### **Content-Aware Defaults**
```go
// Posts: Show newest content first
DefaultSortOptions(ContentTypePosts) → {SortBy: "newest"}

// Comments: Show oldest first (natural conversation flow)
DefaultSortOptions(ContentTypeComments) → {SortBy: "oldest"}
```

### **Why Different Defaults?**
- **Posts**: Users want to see latest content first
- **Comments**: Users want to read conversations chronologically

---

## 🔧 **Implementation Architecture**

### **1. Unified Utilities**
**File**: `api/utils/sort_options.go`
```go
// Content-aware parsing
func ParseSortOptions(r *http.Request, contentType ContentType) SortOptions

// Validation
func IsValidSortOption(sort string, contentType ContentType) bool

// SQL generation
func BuildOrderClause(sortBy string, contentType ContentType) string
```

### **2. Pagination Utilities**
**File**: `api/utils/pagination.go`
```go
// Parse and validate pagination parameters
func ParsePaginationParams(r *http.Request) (limit, offset int)

// Create rich pagination metadata
func NewPaginationInfo(totalCount, limit, offset int) PaginationInfo
```

### **3. Dynamic Query Building**
**File**: `api/queries/post_queries.go`
```go
// Build queries with dynamic sorting
func BuildPostsQuery(joins, whereClause, orderClause string) string

// Pre-built queries with sort support
func GetAllPostsWithSortQuery(orderClause string) string
```

---

## 📱 **Frontend Usage Examples**

### **Basic Pagination**
```javascript
// Load first page
const response = await fetch('/api/posts?limit=20&offset=0');

// Load next page
const nextPage = await fetch('/api/posts?limit=20&offset=20');

// Use pagination metadata
const { current_page, total_pages, has_next } = response.data.pagination;
```

### **Sorting Options**
```javascript
// Sort posts by most liked
const popular = await fetch('/api/posts?sort=likes&limit=20&offset=0');

// Sort comments chronologically  
const comments = await fetch('/api/comments/for-post/123?sort=oldest&limit=10&offset=0');

// Sort posts by most active discussions
const active = await fetch('/api/posts?sort=comments&limit=20&offset=0');
```

### **Combined Parameters**
```javascript
// Popular posts, second page
const url = '/api/posts?sort=likes&limit=20&offset=20';

// Recent comments, small pages
const commentsUrl = '/api/comments/for-post/123?sort=newest&limit=5&offset=0';
```

---

## 🎯 **Supported Endpoints**

### **Posts Endpoints**
- `GET /api/posts` - All posts with pagination & sorting
- `GET /api/posts/by-category/{id}` - Category posts with pagination & sorting  
- `GET /api/users/posts/{id}` - User posts with pagination & sorting
- `GET /api/users/liked-posts/{id}` - User liked posts with pagination & sorting
- `GET /api/users/commented-posts/{id}` - User commented posts with pagination & sorting

### **Comments Endpoints**
- `GET /api/comments/for-post/{id}` - Post comments with pagination & sorting

### **URL Pattern**
```
/{endpoint}?limit={limit}&offset={offset}&sort={sort_option}
```

---

## ⚡ **Performance Features**

### **Efficient SQL**
- **Single queries** with proper JOINs instead of N+1 queries
- **Indexed columns** for sorting (created_at, reaction counts)
- **Optimized aggregation** with subqueries for counts

### **Smart Aggregation**
```sql
-- Efficient like/dislike counting
LEFT JOIN (
    SELECT post_id, COUNT(*) as count 
    FROM post_reactions 
    WHERE reaction_type = 1
    GROUP BY post_id
) like_counts ON p.post_id = like_counts.post_id
```

### **Pagination Metadata**
- **Single count query** for total items
- **Calculated pagination** info (pages, has_next, etc.)
- **No over-fetching** - only get requested items

---

## 🛡️ **Error Handling & Validation**

### **Parameter Validation**
```go
// Auto-correction of invalid values
if limit <= 0 { limit = 20 }     // Default page size
if limit > 50 { limit = 50 }     // Prevent abuse  
if offset < 0 { offset = 0 }     // No negative offset
```

### **Sort Validation**
```go
// Content-aware validation
if !IsValidSortOption(sort, ContentTypePosts) {
    sort = "newest"  // Fallback to default
}
```

### **Graceful Degradation**
- Invalid parameters → Use sensible defaults
- Unknown sort options → Use default sorting
- Database errors → Return appropriate HTTP status codes

---

## 📈 **Example API Responses**

### **Posts with Pagination**
```json
{
  "success": true,
  "data": {
    "posts": [
      {
        "post_id": "123",
        "post_content": "Hello world!",
        "username": "john_doe",
        "like_count": 15,
        "comment_count": 3,
        "created_at": "2024-06-10T15:30:00Z"
      }
    ],
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

### **Comments with Sorting**
```json
{
  "success": true,
  "data": {
    "comments": [
      {
        "comment_id": "456",
        "comment_content": "Great post!",
        "username": "jane_doe",
        "like_count": 5,
        "created_at": "2024-06-10T16:00:00Z"
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 2,
      "total_count": 15,
      "per_page": 10,
      "has_next": true,
      "has_previous": false
    }
  }
}
```

---

## 🎯 **Best Practices**

### **For Frontend Developers**
1. **Always use pagination** for better UX and performance
2. **Respect the max limit** (50) to avoid being rate-limited
3. **Use appropriate sort defaults** for each content type
4. **Handle pagination metadata** for navigation UI
5. **Implement optimistic updates** for better perceived performance

### **For API Consumers**
1. **Cache responses** when appropriate
2. **Use reasonable page sizes** (10-50 items)
3. **Implement infinite scroll** or traditional pagination
4. **Show loading states** during API calls
5. **Handle edge cases** (empty results, last page)

---

## 🚀 **Advanced Features**

### **Content-Type Awareness**
The system automatically adapts defaults and validation rules based on whether you're requesting posts or comments.

### **Flexible Query Building**
Dynamic SQL generation allows for complex filtering while maintaining performance.

### **Rich Metadata**
Pagination responses include everything needed for building sophisticated frontend navigation.

### **Unified API Design**
Consistent parameter naming and response formats across all endpoints.