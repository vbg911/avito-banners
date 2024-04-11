package handlers

import (
	"database/sql"
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type AdminsHandler struct {
	Logger *zap.SugaredLogger
	DB     *sql.DB
}

type requestBody struct {
	TagIds    []int                  `json:"tag_ids"`
	FeatureId int                    `json:"feature_id"`
	Content   map[string]interface{} `json:"content"`
	IsActive  bool                   `json:"is_active"`
}

func (h *AdminsHandler) NewBanner(w http.ResponseWriter, r *http.Request) {
	//todo проверка токена на валидность
	token := r.Header.Get("token")
	h.Logger.Infof(token)
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	var body requestBody
	err = json.Unmarshal(reqBody, &body)
	if err != nil {
		h.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	h.Logger.Info(body)
}
