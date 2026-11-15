package model

// Meta dipakai di response list (pagination).
type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ErrorDetail menjelaskan satu kesalahan, biasanya satu field yang gagal validasi.
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// WebResponse adalah bentuk baku SEMUA response API.
type WebResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    any           `json:"data,omitempty"`
	Meta    *Meta         `json:"meta,omitempty"`
	Errors  []ErrorDetail `json:"errors,omitempty"`
}
