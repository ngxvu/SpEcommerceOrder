package ginext

// BodyMeta represents a body meta data like pagination or extra response information
// it should always be rendered as a map of key: value
type BodyMeta map[string]interface{}

// GeneralBody defines a general response body
type GeneralBody struct {
	Data  interface{} `json:"data,omitempty"`
	Meta  BodyMeta    `json:"meta,omitempty"`
	Error interface{} `json:"error,omitempty"`
}

func NewBody(data interface{}, meta BodyMeta) *GeneralBody {
	return &GeneralBody{
		Data: data,
		Meta: meta,
	}
}

func NewBodyPaginated(data interface{}, pager *Pager) *GeneralBody {
	return &GeneralBody{
		Data: data,
		Meta: BodyMeta{
			"page":        pager.GetPage(),
			"total_pages": pager.GetTotalPages(),
			"page_size":   pager.GetPageSize(),
			"total":       pager.TotalRows,
			"metadata":    pager.Metadata,
		},
	}
}
