package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/zhavkk/news-service/src/news/internal/logger"
	"github.com/zhavkk/news-service/src/news/internal/models"
	"github.com/zhavkk/news-service/src/news/internal/storage"
)

type NewsRepository struct {
	storage *storage.Storage
}

func NewNewsRepository(storage *storage.Storage) *NewsRepository {
	return &NewsRepository{
		storage: storage,
	}
}

func (r *NewsRepository) Create(ctx context.Context, news *models.News) error {
	const op = "NewsRepository.Create"
	logger.Log.Debug(op, "title", news.Title)

	newsQuery := `
    INSERT INTO news (title, category, start_time, end_time) 
    VALUES ($1, $2, $3, $4)
    RETURNING id, created_at
    `

	contentBlockQuery := `
    INSERT INTO content_blocks (news_id, type, content, position) 
    VALUES ($1, $2, $3, $4)
    RETURNING id, created_at
    `

	tx, ok := storage.GetTxFromContext(ctx)
	if !ok {
		logger.Log.Error(op, "No transaction found in context", nil)
		return ErrNoTransactionInContext
	}

	var newsID int64
	err := tx.QueryRow(ctx, newsQuery,
		news.Title,
		news.Category,
		news.StartTime,
		news.EndTime,
	).Scan(&newsID, &news.CreatedAt)

	if err != nil {
		logger.Log.Error(op, "Failed to create news", err)
		return fmt.Errorf("%w: %v", ErrFailedToCreateNews, err)
	}

	news.ID = newsID
	logger.Log.Debug(op, "News created successfully", newsID, "title", news.Title)

	for i := range news.Content {
		block := &news.Content[i]
		block.NewsID = newsID

		err := tx.QueryRow(ctx, contentBlockQuery,
			newsID,
			block.Type,
			block.Content,
			block.Position,
		).Scan(&block.ID, &block.CreatedAt)

		if err != nil {
			logger.Log.Error(op, "Failed to create content block", err, "newsID", newsID)
			return fmt.Errorf("%w: %v", ErrFailedToCreateContentBlock, err)
		}

		logger.Log.Debug(op, "Content block created successfully", block.ID, "content", block.Content)
	}

	return nil
}
func (r *NewsRepository) GetByID(ctx context.Context, id int64) (*models.News, error) {
	const op = "NewsRepository.GetByID"
	logger.Log.Debug(op, "Getting news by ID", id)

	newsQuery := `
    SELECT id, title, category, start_time, end_time, created_at 
    FROM news 
    WHERE id = $1
    `

	news := &models.News{}
	err := r.storage.GetPool().QueryRow(ctx, newsQuery, id).Scan(
		&news.ID,
		&news.Title,
		&news.Category,
		&news.StartTime,
		&news.EndTime,
		&news.CreatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			logger.Log.Debug(op, "News not found", id)
			return nil, ErrNotFound
		}
		logger.Log.Error(op, "Failed to get news by ID", err, "id", id)
		return nil, fmt.Errorf("%w: %v", ErrFailedToGetNews, err)
	}

	blocksQuery := `
    SELECT id, type, content, position, created_at 
    FROM content_blocks 
    WHERE news_id = $1 
    ORDER BY position
    `

	rows, err := r.storage.GetPool().Query(ctx, blocksQuery, id)
	if err != nil {
		logger.Log.Error(op, "Failed to query content blocks", err, "newsID", id)
		return nil, fmt.Errorf("%w: %v", ErrFailedToGetContentBlocks, err)
	}
	defer rows.Close()

	news.Content = make([]models.ContentBlock, 0)
	for rows.Next() {
		var block models.ContentBlock
		var blockType string

		err := rows.Scan(
			&block.ID,
			&blockType,
			&block.Content,
			&block.Position,
			&block.CreatedAt,
		)

		if err != nil {
			logger.Log.Error(op, "Failed to scan content block", err, "newsID", id)
			return nil, fmt.Errorf("%w: %v", ErrFailedToGetContentBlocks, err)
		}

		block.NewsID = id
		block.Type = models.BlockType(blockType)
		news.Content = append(news.Content, block)
	}

	if err = rows.Err(); err != nil {
		logger.Log.Error(op, "Error iterating content blocks", err, "newsID", id)
		return nil, fmt.Errorf("%w: %v", ErrFailedToGetContentBlocks, err)
	}

	return news, nil
}

func (r *NewsRepository) Update(ctx context.Context, news *models.News) error {
	const op = "NewsRepository.Update"
	logger.Log.Debug(op, "Updating news", news.ID, "title", news.Title)

	newsQuery := `
    UPDATE news
    SET title = $1, category = $2, start_time = $3, end_time = $4
    WHERE id = $5 
    `

	deleteBlocksQuery := `
    DELETE FROM content_blocks
    WHERE news_id = $1
    `

	insertBlockQuery := `
    INSERT INTO content_blocks (news_id, type, content, position)
    VALUES ($1, $2, $3, $4)
    RETURNING id, created_at
    `

	tx, ok := storage.GetTxFromContext(ctx)
	if !ok {
		logger.Log.Error(op, "No transaction found in context", nil)
		return ErrNoTransactionInContext
	}

	result, err := tx.Exec(ctx, newsQuery,
		news.Title,
		news.Category,
		news.StartTime,
		news.EndTime,
		news.ID,
	)
	if err != nil {
		logger.Log.Error(op, "Failed to update news", err, "id", news.ID)
		return fmt.Errorf("%w: %v", ErrFailedToUpdateNews, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		logger.Log.Warn(op, "News not found for update", news.ID)
		return ErrNotFound
	}

	_, err = tx.Exec(ctx, deleteBlocksQuery, news.ID)
	if err != nil {
		logger.Log.Error(op, "Failed to delete content blocks", err, "newsID", news.ID)
		return fmt.Errorf("%w: %v", ErrFailedToDeleteContentBlocks, err)
	}

	for i := range news.Content {
		block := &news.Content[i]
		block.NewsID = news.ID

		err := tx.QueryRow(ctx, insertBlockQuery,
			news.ID,
			block.Type,
			block.Content,
			block.Position,
		).Scan(&block.ID, &block.CreatedAt)

		if err != nil {
			logger.Log.Error(op, "Failed to create content block", err, "newsID", news.ID)
			return fmt.Errorf("%w: %v", ErrFailedToCreateContentBlock, err)
		}

		logger.Log.Debug(op, "Content block created successfully", block.ID, "content", block.Content)
	}

	logger.Log.Debug(op, "News updated successfully", news.ID)
	return nil
}

func (r *NewsRepository) Delete(ctx context.Context, id int64) error {
	const op = "NewsRepository.Delete"
	logger.Log.Debug(op, "Deleting news", id)

	query := `
    DELETE FROM news
    WHERE id = $1
    `

	tx, ok := storage.GetTxFromContext(ctx)
	if !ok {
		logger.Log.Error(op, "No transaction found in context", nil)
		return ErrNoTransactionInContext
	}

	result, err := tx.Exec(ctx, query, id)
	if err != nil {
		logger.Log.Error(op, "Failed to delete news", "error", "id", id)
		return fmt.Errorf("%w: %v", ErrFailedToDeleteNews, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		logger.Log.Warn(op, "News not found for deletion", id)
		return ErrNotFound
	}

	logger.Log.Debug(op, "News deleted successfully", id)
	return nil
}

func (r *NewsRepository) List(
	ctx context.Context,
	offset int,
	limit int,
	search string,
	category string,
	sortBy string,
	sortDir string,
) ([]*models.News, int64, error) {
	const op = "NewsRepository.List"
	logger.Log.Debug(op, "offset", offset, "limit", limit, "search", search, "category", category)

	countQuery := `
    SELECT COUNT(*) FROM news n 
    WHERE 1=1
    `

	query := `
    SELECT n.id, n.title, n.category, n.created_at, n.start_time, n.end_time  
    FROM news n
    WHERE 1=1
    `

	args := []interface{}{}
	paramCount := 1

	if search != "" {
		searchCondition := fmt.Sprintf(" AND n.title ILIKE $%d", paramCount)
		query += searchCondition
		countQuery += searchCondition
		args = append(args, "%"+search+"%")
		paramCount++
	}

	if category != "" {
		categoryCondition := fmt.Sprintf(" AND n.category = $%d", paramCount)
		query += categoryCondition
		countQuery += categoryCondition
		args = append(args, category)
		paramCount++
	}

	allowedSortFields := map[string]string{
		"created_at": "n.created_at",
		"title":      "n.title",
		"category":   "n.category",
		"start_time": "n.start_time",
		"end_time":   "n.end_time",
	}

	sortField, ok := allowedSortFields[sortBy]
	if !ok {
		sortField = "n.created_at"
	}

	if sortDir != "asc" && sortDir != "desc" {
		sortDir = "desc"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortField, sortDir)

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramCount, paramCount+1)
	args = append(args, limit, offset)

	var totalCount int64
	err := r.storage.GetPool().QueryRow(ctx, countQuery, args[:paramCount-1]...).Scan(&totalCount)
	if err != nil {
		logger.Log.Error(op, "Failed to get total count", err)
		return nil, 0, fmt.Errorf("%w: %v", ErrFailedToGetNews, err)
	}

	rows, err := r.storage.GetPool().Query(ctx, query, args...)
	if err != nil {
		logger.Log.Error(op, "Failed to query news", err)
		return nil, 0, fmt.Errorf("%w: %v", ErrFailedToGetNews, err)
	}
	defer rows.Close()

	newsList := make([]*models.News, 0)
	for rows.Next() {
		news := &models.News{}
		err := rows.Scan(
			&news.ID,
			&news.Title,
			&news.Category,
			&news.CreatedAt,
			&news.StartTime,
			&news.EndTime,
		)
		if err != nil {
			logger.Log.Error(op, "Failed to scan news row", err)
			return nil, 0, fmt.Errorf("%w: %v", ErrFailedToGetNews, err)
		}

		newsList = append(newsList, news)
	}

	if err = rows.Err(); err != nil {
		logger.Log.Error(op, "Error iterating rows", err)
		return nil, 0, fmt.Errorf("%w: %v", ErrFailedToGetNews, err)
	}

	if len(newsList) > 0 {
		if err = r.loadContentBlocks(ctx, newsList); err != nil {
			logger.Log.Error(op, "Failed to load content blocks", err)
			return nil, 0, err // ошибка уже обёрнута в loadContentBlocks
		}
	}

	return newsList, totalCount, nil
}

func (r *NewsRepository) loadContentBlocks(ctx context.Context, newsList []*models.News) error {
	const op = "NewsRepository.loadContentBlocks"

	if len(newsList) == 0 {
		logger.Log.Debug(op, "Empty news list, skipping content block loading", nil)
		return nil
	}

	newsIDs := make([]int64, len(newsList))
	newsMap := make(map[int64]*models.News)
	for i, news := range newsList {
		newsIDs[i] = news.ID
		newsMap[news.ID] = news
		news.Content = make([]models.ContentBlock, 0)
	}

	placeholders := make([]string, len(newsIDs))
	args := make([]interface{}, len(newsIDs))
	for i, id := range newsIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
        SELECT id, news_id, type, content, position, created_at
        FROM content_blocks
        WHERE news_id IN (%s)
        ORDER BY news_id, position
    `, strings.Join(placeholders, ", "))

	rows, err := r.storage.GetPool().Query(ctx, query, args...)
	if err != nil {
		logger.Log.Error(op, "Failed to query content blocks", err, "newsIDs", newsIDs)
		return fmt.Errorf("%w: %v", ErrFailedToGetContentBlocks, err)
	}
	defer rows.Close()

	contentFound := false
	for rows.Next() {
		contentFound = true
		var block models.ContentBlock
		var blockType string
		var newsID int64

		err := rows.Scan(
			&block.ID,
			&newsID,
			&blockType,
			&block.Content,
			&block.Position,
			&block.CreatedAt,
		)

		if err != nil {
			logger.Log.Error(op, "Failed to scan content block", err)
			return fmt.Errorf("%w: %v", ErrFailedToGetContentBlocks, err)
		}

		block.NewsID = newsID
		block.Type = models.BlockType(blockType)

		if news, ok := newsMap[newsID]; ok {
			news.Content = append(news.Content, block)
		}
	}

	if err = rows.Err(); err != nil {
		logger.Log.Error(op, "Error iterating content blocks", err)
		return fmt.Errorf("%w: %v", ErrFailedToGetContentBlocks, err)
	}

	if !contentFound {
		logger.Log.Debug(op, "No content blocks found for the news items", newsIDs)
	}

	return nil
}
