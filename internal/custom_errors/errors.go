package custom_errors

import "errors"

// Ошибки пользователя
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUsernameExists    = errors.New("username already exists")
	ErrEmailExists       = errors.New("email already exists")
	ErrInvalidUsername   = errors.New("invalid username")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrPasswordMismatch  = errors.New("passwords do not match")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserCreateFailed  = errors.New("failed to create user")
)

// Ошибки аутентификации
var (
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidRefreshToken   = errors.New("invalid refresh token")
	ErrUnauthenticated       = errors.New("unauthenticated")
	ErrTokenExpired          = errors.New("token expired")
	ErrInvalidToken          = errors.New("invalid token")
	ErrTokenGenerationFailed = errors.New("token generation failed")
	ErrRegistrationFailed    = errors.New("failed to register user")
	ErrLoginFailed           = errors.New("failed to login")
	ErrRefreshTokenFailed    = errors.New("failed to refresh token")
	ErrLogoutFailed          = errors.New("failed to logout")
	ErrPasswordUpdateFailed  = errors.New("failed to update password")
)

// Ошибки валидации
var (
	ErrValidationFailed = errors.New("validation failed")
	ErrInvalidInput     = errors.New("invalid input")
	ErrRequiredField    = errors.New("required field is missing")
	ErrForbidden        = errors.New("forbidden")
)

// Ошибки базы данных
var (
	ErrDatabaseConnection  = errors.New("database connection error")
	ErrDatabaseQuery       = errors.New("database query error")
	ErrDatabaseTransaction = errors.New("database transaction error")
)

// Ошибки внешних сервисов
var (
	ErrExternalServiceUnavailable = errors.New("external service unavailable")
	ErrExternalServiceTimeout     = errors.New("external service timeout")
	ErrExternalServiceError       = errors.New("external service error")
)

// Ошибки файловой системы
var (
	ErrFileNotFound     = errors.New("file not found")
	ErrFileAccessDenied = errors.New("file access denied")
	ErrFileTooLarge     = errors.New("file too large")
)

// Ошибки конфигурации
var (
	ErrConfigNotFound   = errors.New("configuration not found")
	ErrConfigInvalid    = errors.New("invalid configuration")
	ErrConfigLoadFailed = errors.New("failed to load configuration")
)

// Ошибки кэша
var (
	ErrCacheMiss     = errors.New("cache miss")
	ErrCacheDisabled = errors.New("cache disabled")
	ErrCacheError    = errors.New("cache error")
)

// Ошибки rate limiting
var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrTooManyRequests   = errors.New("too many requests")
)

// Ошибки бизнес-логики
var (
	ErrOperationNotAllowed = errors.New("operation not allowed")
	ErrResourceLocked      = errors.New("resource is locked")
	ErrInsufficientRights  = errors.New("insufficient rights")
)

// Ошибки поиска
var (
	ErrSearchFailed       = errors.New("search failed")
	ErrInvalidSearchQuery = errors.New("invalid search query")
)

// Ошибки аватара
var (
	ErrInvalidAvatarFormat = errors.New("invalid avatar format")
	ErrAvatarUploadFailed  = errors.New("avatar upload failed")
	ErrAvatarDeleteFailed  = errors.New("avatar delete failed")
)

// Ошибки постов
var (
	ErrPostNotFound      = errors.New("post not found")
	ErrNoUpdateRows      = errors.New("no post update rows")
	ErrPostValidation    = errors.New("post validation failed")
	ErrInvalidTagName    = errors.New("invalid tag name")
	ErrTagNotFound       = errors.New("tag not found")
	ErrTagsNotFound      = errors.New("tags not found")
	ErrTagAlreadyExists  = errors.New("tag already exists")
	ErrMediaNotFound     = errors.New("media not found")
	ErrMediaAttachFailed = errors.New("failed to attach media to post")
	ErrMediaDetachFailed = errors.New("failed to detach media from post")
)

// Follower relation errors
var (
	ErrSelfFollow               = errors.New("cannot follow yourself")
	ErrFollowRelationExists     = errors.New("follow relation already exists")
	ErrFollowRelationNotFound   = errors.New("follow relation not found")
	ErrFollowRelationCreateFail = errors.New("failed to create follow relation")
	ErrFollowRelationDeleteFail = errors.New("failed to delete follow relation")
	ErrAlreadyFollowing         = errors.New("already following this user")
	ErrUnexpectedEventType      = errors.New("unexpected event type in outbox")
)

// Notification errors
var (
	ErrNotificationNotFound       = errors.New("notification not found")
	ErrNotificationCreateFailed   = errors.New("failed to create notification")
	ErrNotificationInvalidType    = errors.New("invalid notification type")
	ErrNotificationInvalidPayload = errors.New("invalid notification payload")
	ErrNotificationAccessDenied   = errors.New("access to notification denied")
	ErrNotificationLimitExceeded  = errors.New("notification limit exceeded")
	ErrNotificationAlreadyExists  = errors.New("notification already exists")
)

// HTTP Client errors
var (
	ErrInvalidURL            = errors.New("invalid URL")
	ErrRequestCreationFailed = errors.New("failed to create request")
	ErrRequestFailed         = errors.New("request failed")
	ErrResponseReadFailed    = errors.New("failed to read response body")
	ErrJSONMarshalFailed     = errors.New("failed to marshal JSON")
	ErrJSONUnmarshalFailed   = errors.New("failed to unmarshal JSON")
	ErrAPIError              = errors.New("API error")
	ErrStatusCode            = errors.New("unexpected status code")
)
