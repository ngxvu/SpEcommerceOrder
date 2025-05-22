package utils

import "context"

type MetaData struct {
	TraceID string `json:"trace_id"`
	Success bool   `json:"success"`
}

func NewMetaData(ctx context.Context) *MetaData {
	// Extract x-request-id from context
	requestID, ok := ctx.Value("x-request-id").(string)
	if !ok {
		return nil
	}
	return &MetaData{
		TraceID: requestID,
		Success: true,
	}
}
