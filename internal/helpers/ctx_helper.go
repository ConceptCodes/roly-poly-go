package helpers

import (
	"context"
	"net/http"
	"roly-poly/internal/constants"

	"github.com/google/uuid"
)

type ContextKey string

const (
	RequestIDKey ContextKey = constants.RequestIdCtxKey
	ApiKey       ContextKey = constants.ApiKeyCtxKey
	UserId       ContextKey = constants.UserIdCtxKey
)

func SetRequestId(r *http.Request, requestID string) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, RequestIDKey, requestID)
	return r.WithContext(ctx)
}

func GetRequestId(r *http.Request) string {
	return r.Context().Value(RequestIDKey).(string)
}

func SetApiKey(r *http.Request, apiKey string) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, ApiKey, apiKey)
	return r.WithContext(ctx)
}

func GetApiKey(r *http.Request) string {
	apiKey := r.Context().Value(ApiKey)
	if apiKey == nil {
		return ""
	}
	return apiKey.(string)
}

func SetUserId(r *http.Request, userId uuid.UUID) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, UserId, userId)
	return r.WithContext(ctx)
}

func GetUserId(r *http.Request) uuid.UUID {
	userId := r.Context().Value(UserId)
	if userId == nil {
		return uuid.Nil
	}
	return userId.(uuid.UUID)
}
