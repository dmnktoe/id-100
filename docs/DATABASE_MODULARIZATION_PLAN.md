# Database Modularization Plan

## Overview

This document outlines the plan to modularize database queries in the ID-100 application by implementing the Repository Pattern. This will separate data access logic from business logic in handlers.

## Current State

### Problems

1. **app.go** is 722 lines long with embedded SQL queries
2. **admin.go** mixes business logic with direct database access
3. No clear separation between data access and presentation layers
4. Difficult to test database operations in isolation
5. Code duplication across handlers

### Current Structure

```
internal/handlers/
├── app.go       (722 lines - deriven, contributions, uploads)
├── admin.go     (admin operations with embedded queries)
└── middleware/
    └── token.go (token validation with queries)
```

## Proposed Solution: Repository Pattern

### New Structure

```
internal/
├── repository/
│   ├── deriven.go        (All deriven queries)
│   ├── contribution.go   (All contribution queries)
│   ├── upload_token.go   (Token management queries)
│   ├── stats.go          (Statistics queries)
│   └── repository.go     (Interface definitions)
├── handlers/
│   ├── app.go            (Business logic only)
│   └── admin.go          (Business logic only)
└── models/
    └── models.go         (Existing models)
```

## Implementation Plan

### Phase 1: Create Repository Package

#### 1. repository/deriven.go

```go
package repository

import (
    "context"
    "id-100/internal/models"
    "github.com/jackc/pgx/v5/pgxpool"
)

type DerivenRepository struct {
    db *pgxpool.Pool
}

func NewDerivenRepository(db *pgxpool.Pool) *DerivenRepository {
    return &DerivenRepository{db: db}
}

// GetAll retrieves all deriven with pagination
func (r *DerivenRepository) GetAll(ctx context.Context, limit, offset int) ([]models.Derive, error)

// GetByID retrieves a single derive by ID
func (r *DerivenRepository) GetByID(ctx context.Context, id int) (*models.Derive, error)

// GetByNumberWithContributions retrieves derive with all contributions
func (r *DerivenRepository) GetByNumberWithContributions(ctx context.Context, number int, cityFilter string) (*models.Derive, []models.Contribution, error)

// CountAll counts total deriven
func (r *DerivenRepository) CountAll(ctx context.Context) (int, error)

// GetByCity filters deriven by city with pagination
func (r *DerivenRepository) GetByCity(ctx context.Context, city string, limit, offset int) ([]models.Derive, error)

// CountByCity counts deriven with contributions from specific city
func (r *DerivenRepository) CountByCity(ctx context.Context, city string) (int, error)
```

#### 2. repository/contribution.go

```go
package repository

import (
    "context"
    "id-100/internal/models"
    "github.com/jackc/pgx/v5/pgxpool"
)

type ContributionRepository struct {
    db *pgxpool.Pool
}

func NewContributionRepository(db *pgxpool.Pool) *ContributionRepository {
    return &ContributionRepository{db: db}
}

// Create inserts a new contribution
func (r *ContributionRepository) Create(ctx context.Context, contrib *models.Contribution) (int, error)

// GetByDeriveID retrieves all contributions for a derive
func (r *ContributionRepository) GetByDeriveID(ctx context.Context, deriveID int, cityFilter string) ([]models.Contribution, error)

// GetByID retrieves a single contribution
func (r *ContributionRepository) GetByID(ctx context.Context, id int) (*models.Contribution, error)

// Delete removes a contribution by ID
func (r *ContributionRepository) Delete(ctx context.Context, id int) error

// GetRecent gets recent contributions for admin dashboard
func (r *ContributionRepository) GetRecent(ctx context.Context, limit int) ([]models.RecentContrib, error)

// CountTotal counts all contributions
func (r *ContributionRepository) CountTotal(ctx context.Context) (int, error)

// GetDistinctCities gets all distinct cities from contributions
func (r *ContributionRepository) GetDistinctCities(ctx context.Context) ([]string, error)
```

#### 3. repository/upload_token.go

```go
package repository

import (
    "context"
    "id-100/internal/models"
    "github.com/jackc/pgx/v5/pgxpool"
)

type UploadTokenRepository struct {
    db *pgxpool.Pool
}

func NewUploadTokenRepository(db *pgxpool.Pool) *UploadTokenRepository {
    return &UploadTokenRepository{db: db}
}

// GetByToken retrieves token info by token string
func (r *UploadTokenRepository) GetByToken(ctx context.Context, token string) (*models.TokenInfo, error)

// Create creates a new upload token
func (r *UploadTokenRepository) Create(ctx context.Context, token, bagName string, maxUploads int) (int, error)

// UpdateSessionInfo updates session information
func (r *UploadTokenRepository) UpdateSessionInfo(ctx context.Context, tokenID int, playerName, playerCity string) error

// IncrementUploads increments the upload counter
func (r *UploadTokenRepository) IncrementUploads(ctx context.Context, tokenID int) error

// DecrementUploads decrements the upload counter (for deletions)
func (r *UploadTokenRepository) DecrementUploads(ctx context.Context, tokenID int) error

// GetAll retrieves all tokens for admin view
func (r *UploadTokenRepository) GetAll(ctx context.Context) ([]models.TokenInfo, error)

// LogUpload logs an upload to upload_logs table
func (r *UploadTokenRepository) LogUpload(ctx context.Context, tokenID, contributionID int) error

// GetTokenIDByContributionID gets token ID from upload log
func (r *UploadTokenRepository) GetTokenIDByContributionID(ctx context.Context, contributionID int) (int, error)
```

#### 4. repository/stats.go

```go
package repository

import (
    "context"
    "database/sql"
    "github.com/jackc/pgx/v5/pgxpool"
)

type StatsRepository struct {
    db *pgxpool.Pool
}

func NewStatsRepository(db *pgxpool.Pool) *StatsRepository {
    return &StatsRepository{db: db}
}

// GetFooterStats retrieves all footer statistics
func (r *StatsRepository) GetFooterStats(ctx context.Context) (totalDeriven, totalContribs, activeUsers, totalCities int, lastActivity sql.NullTime, err error) {
    // Implementation moved from database.GetFooterStats()
}
```

### Phase 2: Update Handlers

#### Example: app.go DerivenHandler

**Before:**
```go
func DerivenHandler(c echo.Context) error {
    // Direct SQL query
    query := `SELECT ... FROM deriven d ...`
    rows, err := database.DB.Query(context.Background(), query, ...)
    // Process rows...
}
```

**After:**
```go
func DerivenHandler(c echo.Context) error {
    // Use repository
    derives, err := derivenRepo.GetAll(c.Request().Context(), limit, offset)
    if err != nil {
        return handleError(c, err)
    }
    // Render template with data
}
```

### Phase 3: Update Database Package

Move `GetFooterStats()` from `database.go` to `repository/stats.go` and update callers.

## Queries to Extract

### From app.go

1. **DerivenHandler**
   - Get distinct cities for filter dropdown
   - Count deriven (total or filtered by city)
   - Get deriven list with pagination

2. **DeriveHandler**
   - Get derive by number
   - Get contributions for derive (with optional city filter)

3. **SessionUploadFormHandler**
   - Get token info
   - Check upload limits
   - Create contribution
   - Log upload
   - Increment upload counter
   - Get derive by ID

4. **DeleteUploadHandler**
   - Get contribution details
   - Delete contribution
   - Delete from upload logs
   - Decrement upload counter
   - Delete S3 file

### From admin.go

1. **AdminTokensHandler**
   - Get all tokens

2. **AdminCreateTokenHandler**
   - Create token

3. **AdminDeleteContributionHandler**
   - Get contribution with token info
   - Delete contribution
   - Delete from upload logs
   - Decrement upload counter
   - Delete S3 file

4. **AdminDashboardHandler**
   - Get footer stats
   - Get recent contributions

### From middleware/token.go

1. **TokenMiddleware**
   - Get token by token string
   - Update session info

## Benefits

### 1. Separation of Concerns
- Handlers focus on HTTP and business logic
- Repository handles all database operations
- Clear boundaries between layers

### 2. Testability
```go
// Can easily mock repository in tests
type MockDerivenRepo struct {
    GetAllFunc func(ctx context.Context, limit, offset int) ([]models.Derive, error)
}

func (m *MockDerivenRepo) GetAll(ctx context.Context, limit, offset int) ([]models.Derive, error) {
    return m.GetAllFunc(ctx, limit, offset)
}
```

### 3. Maintainability
- All deriven queries in one place
- Easier to optimize or modify queries
- Reduces code duplication

### 4. Reusability
- Repository methods can be used by multiple handlers
- Consistent data access patterns

### 5. Type Safety
- Strong typing for all data access operations
- Compile-time checking of query parameters

## Migration Strategy

### Step 1: Create Repository Package (No Breaking Changes)
- Create repository package alongside existing code
- Implement all repository methods
- Add tests for repository layer

### Step 2: Update Handlers Gradually
- Update one handler at a time
- Test each change thoroughly
- Keep existing functionality working

### Step 3: Clean Up
- Remove unused database package functions
- Remove duplicate query code
- Update documentation

### Step 4: Add Tests
- Unit tests for each repository
- Integration tests for handlers using repositories

## Testing Strategy

### Repository Tests
```go
func TestDerivenRepository_GetAll(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    repo := NewDerivenRepository(db)
    
    // Insert test data
    insertTestDeriven(t, db)
    
    // Test query
    derives, err := repo.GetAll(context.Background(), 10, 0)
    assert.NoError(t, err)
    assert.Len(t, derives, 10)
}
```

### Handler Tests with Mock Repository
```go
func TestDerivenHandler(t *testing.T) {
    // Create mock repository
    mockRepo := &MockDerivenRepo{
        GetAllFunc: func(ctx context.Context, limit, offset int) ([]models.Derive, error) {
            return []models.Derive{{ID: 1, Number: 1}}, nil
        },
    }
    
    // Test handler with mock
    // ...
}
```

## Timeline

### Phase 1: Foundation (1-2 days)
- Create repository package structure
- Implement repository interfaces
- Add basic tests

### Phase 2: Migration (2-3 days)
- Update app.go handlers
- Update admin.go handlers
- Update middleware

### Phase 3: Testing & Cleanup (1 day)
- Comprehensive testing
- Remove old code
- Documentation

**Total Estimated Time: 4-6 days**

## Success Criteria

- [ ] All database queries moved to repository layer
- [ ] app.go reduced to < 400 lines
- [ ] admin.go reduced to < 300 lines
- [ ] 80%+ test coverage for repository layer
- [ ] All existing functionality works correctly
- [ ] No performance regression
- [ ] Documentation updated

## Next Steps

1. Review and approve this plan
2. Create feature branch for modularization
3. Implement Phase 1 (repository package)
4. Review and test Phase 1
5. Implement Phase 2 (handler migration)
6. Review and test Phase 2
7. Implement Phase 3 (cleanup)
8. Merge to main branch

## Notes

- This refactoring can be done incrementally without breaking existing functionality
- Each phase can be reviewed and merged separately
- Repository pattern is a Go best practice for data access
- Will significantly improve code maintainability and testability
