# SEO Package

This package provides a centralized SEO configuration and metadata generation for the application.

## Single Source of Truth

All SEO metadata is defined in one place: `config.go`

### Adding a New Page

To add SEO metadata for a new page:

1. Add the page configuration to `config.go`:

```go
"my_new_page": {
    Path:        "/my-new-page",
    Title:       "My New Page - 🏠🆔💯",
    Description: "Description of my new page.",
    Type:        "website",
},
```

2. Add the page key to the `GetStaticPages()` method if you want it in the sitemap.

3. Use the builder in your handler:

```go
baseURL := seo.GetBaseURLFromRequest(c.Scheme(), c.Request().Host, c.Request().Header.Get("X-Forwarded-Host"))
builder := seo.NewBuilder(baseURL)
meta := builder.ForPage("my_new_page")
```

## Usage Examples

### Static Page

```go
builder := seo.NewBuilder(baseURL)
meta := builder.ForPage("leitfaden")
// Returns metadata with title, description, URL, etc.
```

### Dynamic ID Page

```go
builder := seo.NewBuilder(baseURL)
meta := builder.ForID(42, "ID Title", "ID Description", "image.jpg")
// Returns metadata customized for ID #42
```

### Default/Home Page

```go
builder := seo.NewBuilder(baseURL)
meta := builder.Default()
// Returns default homepage metadata
```

### Custom Metadata

```go
builder := seo.NewBuilder(baseURL)
meta := builder.Custom("Title", "Description", "image.jpg", "https://url.com", "article")
```

## Sitemap

The sitemap is automatically generated at `/sitemap.xml` and includes:

- All static pages defined in the config
- All dynamic ID pages from the database
- Proper priority and change frequency for SEO

The sitemap updates dynamically when the endpoint is accessed, so it always reflects the current state of the application.

## Backward Compatibility

The `internal/utils/seo.go` file provides backward-compatible wrapper functions, so existing handlers continue to work without changes:

- `GetDefaultSEOMetadata(baseURL)` → uses `seo.NewBuilder(baseURL).Default()`
- `GetPageSEOMetadata(name, baseURL)` → uses `seo.NewBuilder(baseURL).ForPage(name)`
- `GetBaseURLFromRequest()` → uses `seo.GetBaseURLFromRequest()`
