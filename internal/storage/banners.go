package storage

import (
	"database/sql"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func GetBanners(db *sql.DB, tagId int, featureId int) ([]byte, error) {
	query := `
		SELECT content
		FROM public.banners
		WHERE tag_ids @> $1 AND feature_id = $2 and is_active = $3
	`

	// Выполнение запроса с параметрами
	rows, err := db.Query(query, pq.Array([]int{tagId}), featureId, true)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Обработка результатов запроса
	var jsonResponse []byte
	for rows.Next() {
		if err := rows.Scan(&jsonResponse); err != nil {
			return nil, err
		}
	}
	return jsonResponse, nil
}

func NewBanner(db *sql.DB, tagId int, featureId int, content map[string]interface{}, isActive bool) (int, error) {
	query := `
		INSERT INTO public.banners(
	banner_id, tag_ids, feature_id, content, is_active, created_at, updated_at)
	VALUES (DEFAULT, $2, $3, $4, $5, DEFAULT, DEFAULT);
	`

}
