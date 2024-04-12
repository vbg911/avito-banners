package storage

import (
	"database/sql"
	"encoding/json"
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

func NewBanner(db *sql.DB, tagId []int, featureId int, content map[string]interface{}, isActive bool) (int, error) {
	query := `
		INSERT INTO public.banners(
	banner_id, tag_ids, feature_id, content, is_active, created_at, updated_at)
	VALUES (DEFAULT, $1, $2, $3, $4, DEFAULT, DEFAULT) RETURNING banner_id;
	`

	jsonData, err := json.Marshal(content)
	if err != nil {
		return 0, err
	}

	jsonString := string(jsonData)

	rows, err := db.Query(query, pq.Array(tagId), featureId, jsonString, isActive)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	// Обработка результатов запроса
	var id int
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
	}
	return id, nil
}

func DeleteBanner(db *sql.DB, Id int) (int, error) {
	result, err := db.Exec(
		"DELETE FROM public.banners WHERE banner_id = $1",
		Id,
	)

	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rows), nil
}

func UpdateBanner(db *sql.DB, id int, tagId []int, featureId int, content map[string]interface{}, isActive bool) (int, error) {
	query := `
		UPDATE public.banners
	SET tag_ids=$1, feature_id=$2, content=$3, is_active=$4, updated_at=CURRENT_TIMESTAMP
	WHERE banner_id=$5;
	`

	jsonData, err := json.Marshal(content)
	if err != nil {
		return 0, err
	}

	jsonString := string(jsonData)

	result, err := db.Exec(query, pq.Array(tagId), featureId, jsonString, isActive, id)
	if err != nil {
		return 0, err
	}

	// Обработка результатов запроса
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rows), nil
}

func GetBannersAdmin(db *sql.DB, tagsId int, featureId int, limit int, offset int) (*sql.Rows, error) {
	query1 := `SELECT * FROM public.banners WHERE feature_id = $1 LIMIT $2 OFFSET $3;`
	query2 := `SELECT * FROM public.banners WHERE tag_ids @> $1 LIMIT $2 OFFSET $3;`
	query3 := `SELECT * FROM public.banners WHERE (tag_ids @> $1) AND (feature_id = $2) LIMIT $3 OFFSET $4;`
	query4 := `SELECT * FROM public.banners LIMIT $1 OFFSET $2;`

	var (
		rows *sql.Rows
		err  error
	)
	if tagsId == -1 && featureId != -1 {
		rows, err = db.Query(query1, featureId, limit, offset)
	} else if tagsId == -1 && featureId == -1 {
		rows, err = db.Query(query4, limit, offset)
	} else if tagsId != -1 && featureId == -1 {
		rows, err = db.Query(query2, pq.Array([]int{tagsId}), limit, offset)
	} else if tagsId != -1 && featureId != -1 {
		rows, err = db.Query(query3, pq.Array([]int{tagsId}), featureId, limit, offset)
	}
	if err != nil {
		return nil, err
	}

	return rows, nil
}
