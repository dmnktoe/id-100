package app

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/chai2010/webp"
	"github.com/labstack/echo/v4"

	"id-100/internal/imgutil"
	"id-100/internal/middleware"
	"id-100/internal/repository"
	"id-100/internal/seo"
	"id-100/internal/templates"
	"id-100/internal/utils"
)

// UploadGetHandler displays the upload form
func UploadGetHandler(c echo.Context) error {
	stats := utils.GetFooterStats()

	// Get deriven list
	list, err := repository.GetDerivenForUpload(context.Background())
	if err != nil {
		return c.String(http.StatusInternalServerError, "Datenbankfehler")
	}

	// Fetch session uploads for this token/session
	tokenID, _ := c.Get("token_id").(int)
	sessionNumber, _ := c.Get("session_number").(int)
	sessionContribs, err := repository.GetSessionUploads(context.Background(), tokenID, sessionNumber)
	if err != nil {
		log.Printf("Failed to fetch session uploads: %v", err)
		sessionContribs = []map[string]interface{}{}
	}

	// Normalize image URLs
	for _, sc := range sessionContribs {
		if imgUrl, ok := sc["image_url"].(string); ok {
			sc["image_url"] = utils.EnsureFullImageURL(imgUrl)
		}
	}

	token, _ := c.Get("token").(string)
	currentPlayer, _ := c.Get("current_player").(string)

	// Build a map[string]bool of derive numbers that were uploaded in THIS session/token
	uploadedNumbers := make(map[string]bool)
	totalPoints := 0
	for _, sc := range sessionContribs {
		if num, ok := sc["number"].(int); ok {
			uploadedNumbers[strconv.Itoa(num)] = true
			// Find the points for this derive number
			for _, d := range list {
				if d.Number == num {
					totalPoints += d.Points
					break
				}
			}
		}
	}

	// Generate SEO metadata
	baseURL := seo.GetBaseURLFromRequest(c.Scheme(), c.Request().Host, c.Request().Header.Get("X-Forwarded-Host"))
	builder := seo.NewBuilder(baseURL)
	seoMeta := builder.ForPage("upload")

	return c.Render(http.StatusOK, "layout", templates.MergeTemplateData(map[string]interface{}{
		"Title":           seoMeta.Title,
		"SEO":             seoMeta,
		"Deriven":         list,
		"ContentTemplate": "upload.content",
		"CurrentPath":     c.Request().URL.Path,
		"CurrentYear":     time.Now().Year(),
		"FooterStats":     stats,
		"SessionContribs": sessionContribs,
		"SelectedNumber":  c.QueryParam("number"),
		"Token":           token,
		"CurrentPlayer":   currentPlayer,
		"UploadedNumbers": uploadedNumbers,
		"TotalPoints":     totalPoints,
	}))
}

// UploadPostHandler handles image upload
func UploadPostHandler(c echo.Context) error {
	// Get token info from middleware context
	tokenID, ok := c.Get("token_id").(int)
	if !ok {
		return c.String(http.StatusForbidden, "Token nicht gefunden")
	}

	currentPlayer, _ := c.Get("current_player").(string)
	sessionNumber, _ := c.Get("session_number").(int)

	deriveNumberStr := c.FormValue("derive_number")
	deriveNumber, err := strconv.Atoi(deriveNumberStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "Ungültige Aufgabennummer")
	}

	file, err := c.FormFile("image")
	if err != nil {
		return c.String(http.StatusBadRequest, "Kein Bild gefunden")
	}

	src, _ := file.Open()
	defer src.Close()
	// Decode and auto-orient based on EXIF so mobile uploads keep the correct rotation
	img, err := imgutil.DecodeAutoOriented(src)
	if err != nil {
		return c.String(http.StatusBadRequest, "Ungültiges Bildformat oder Korrektur fehlgeschlagen")
	}
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Lossless: false, Quality: 75}); err != nil {
		return c.String(http.StatusInternalServerError, "WebP-Kodierung fehlgeschlagen")
	}

	cfg, _ := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("S3_REGION")),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("S3_ACCESS_KEY"),
			os.Getenv("S3_SECRET_KEY"),
			""),
		),
	)
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = true
	})

	fileName := fmt.Sprintf("derive_%s_%d.webp", deriveNumberStr, time.Now().Unix())

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("S3_BUCKET")),
		Key:         aws.String(fileName),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("image/webp"),
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "S3 Fehler: "+err.Error())
	}

	// Store just the filename in DB
	relativePath := fileName

	// generate tiny LQIP (data-uri) and store it
	lqip, lqipErr := utils.GenerateLQIP(img, 24)
	if lqipErr != nil {
		log.Printf("LQIP generation failed: %v", lqipErr)
		lqip = ""
	}

	// Get derive internal ID
	internalID, err := repository.GetDeriveIDByNumber(context.Background(), deriveNumberStr)
	if err != nil {
		return c.String(http.StatusNotFound, "Aufgabe nicht gefunden")
	}

	// Get optional user comment (max 100 chars)
	userComment := c.FormValue("comment")
	runes := []rune(userComment)
	if len(runes) > 100 {
		userComment = string(runes[:100])
	}

	// Insert contribution
	currentPlayerCity, _ := c.Get("current_player_city").(string)
	contributionID, err := repository.InsertContribution(context.Background(),
		internalID, relativePath, lqip, currentPlayer, currentPlayerCity, userComment)

	if err != nil {
		log.Printf("DB Error inserting contribution: %v", err)
		return c.String(http.StatusInternalServerError, "DB Error")
	}

	// Log upload
	err = repository.InsertUploadLog(context.Background(), tokenID, deriveNumber, currentPlayer, sessionNumber, contributionID)
	if err != nil {
		log.Printf("Failed to log upload: %v", err)
	}

	// Increment total_uploads counter
	err = repository.IncrementTokenUploadCount(context.Background(), tokenID)
	if err != nil {
		log.Printf("Failed to increment upload counter: %v", err)
	}

	// Redirect back to the upload page
	redirectURL := "/upload?uploaded=1"
	originalToken := c.Request().URL.Query().Get("token")
	if originalToken == "" && c.Request().Method == "POST" {
		contentType := c.Request().Header.Get("Content-Type")
		if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			const maxFormSize = int64(2 * 1024 * 1024)
			c.Request().Body = http.MaxBytesReader(c.Response().Writer, c.Request().Body, maxFormSize)
			if formToken := c.FormValue("token"); formToken != "" {
				originalToken = formToken
			}
		}
	}
	if originalToken != "" {
		redirectURL = fmt.Sprintf("%s&token=%s", redirectURL, url.QueryEscape(originalToken))
	}
	return c.Redirect(http.StatusSeeOther, redirectURL)
}

// SetPlayerNameHandler handles the name entry form submission
func SetPlayerNameHandler(c echo.Context) error {
	// Protect against large request bodies
	const maxFormSize = int64(2 * 1024 * 1024)
	if strings.Contains(c.Request().Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		c.Request().Body = http.MaxBytesReader(c.Response().Writer, c.Request().Body, maxFormSize)
	}

	playerName := c.FormValue("player_name")
	token := c.FormValue("token")

	if playerName == "" || token == "" {
		return c.String(http.StatusBadRequest, "Name und Token erforderlich")
	}

	// Consent checkbox (required)
	consent := c.FormValue("agree_privacy")
	if consent == "" {
		bagName, _ := repository.GetBagNameByToken(context.Background(), token)
		
		// Generate SEO metadata
		baseURL := seo.GetBaseURLFromRequest(c.Scheme(), c.Request().Host, c.Request().Header.Get("X-Forwarded-Host"))
		builder := seo.NewBuilder(baseURL)
		seoMeta := builder.Custom(
			"Willkommen bei Innenstadt ID - 100",
			"Willkommen bei der urbanen Stadtrallye. Registriere dich und starte deine Entdeckungsreise.",
			"",
			baseURL+"/upload/set-name",
			"website",
		)
		
		return c.Render(http.StatusBadRequest, "layout", templates.MergeTemplateData(map[string]interface{}{
			"Title":           seoMeta.Title,
			"SEO":             seoMeta,
			"ContentTemplate": "enter_name.content",
			"Token":           token,
			"BagName":         bagName,
			"FormError":       "Bitte bestätige die Datenschutzerklärung und dass du keine erkennbaren Personen ohne Einwilligung hochlädst.",
		}))
	}

	playerCity := strings.TrimSpace(c.FormValue("player_city"))

	// Save name and city in session
	session, _ := middleware.Store.Get(c.Request(), "id-100-session")
	session.Values["player_name"] = playerName
	session.Values["player_city"] = playerCity
	session.Save(c.Request(), c.Response())

	// Update database
	err := repository.UpdatePlayerNameAndCity(context.Background(), playerName, playerCity, token)
	if err != nil {
		log.Printf("Error setting player name: %v", err)
	}

	// Redirect to upload page
	return c.Redirect(http.StatusSeeOther, "/upload?token="+token)
}
