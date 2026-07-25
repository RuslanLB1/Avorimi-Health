package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// uploads.go — приём вложений (фото/видео/голосовые/файлы), которые
// пользователь или гость прикрепляет в чате поддержки на сайте. Хранится в
// памяти процесса (как и весь остальной Store — это прототип без базы
// данных), отдаётся обратно по сгенерированному ID: один раз для показа в
// собственном чате отправителя, второй раз — Telegram сам подтягивает файл
// по этому же публичному URL при пересылке оператору (см. telegram.go).
const maxUploadSize = 20 << 20 // 20 МБ

type uploadedFile struct {
	Data        []byte
	ContentType string
	Name        string
}

func (s *Store) SaveUpload(id string, f uploadedFile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.uploads[id] = f
}

func (s *Store) GetUpload(id string) (uploadedFile, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, ok := s.uploads[id]
	return f, ok
}

// attachmentKind классифицирует MIME-тип для выбора, как отрисовать вложение
// в чате (превью картинки/видео/плеер) и каким методом Telegram его отправить.
func attachmentKind(contentType string) string {
	switch {
	case strings.HasPrefix(contentType, "image/"):
		return "image"
	case strings.HasPrefix(contentType, "video/"):
		return "video"
	case strings.HasPrefix(contentType, "audio/"):
		return "audio"
	default:
		return "file"
	}
}

type uploadResponse struct {
	UploadID string `json:"uploadId"`
	URL      string `json:"url"`
	Kind     string `json:"kind"`
	Name     string `json:"name"`
}

// apiUploadHandler принимает multipart/form-data с одним полем "file" —
// доступен и залогиненным, и гостям (без него гость, застрявший на
// регистрации, не смог бы приложить скриншот ошибки).
func apiUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeErr(w, http.StatusRequestEntityTooLarge, "err.uploadTooLarge")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, http.StatusBadRequest, "err.uploadMissing")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "err.generic")
		return
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	id := newToken()[:20]
	store.SaveUpload(id, uploadedFile{Data: data, ContentType: contentType, Name: header.Filename})

	writeJSON(w, http.StatusCreated, uploadResponse{
		UploadID: id,
		URL:      "/api/support/uploads/" + id,
		Kind:     attachmentKind(contentType),
		Name:     header.Filename,
	})
}

func apiUploadFileHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("uploadId")
	f, ok := store.GetUpload(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", f.ContentType)
	w.Header().Set("Cache-Control", "private, max-age=86400")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, f.Name))
	w.Write(f.Data)
}
