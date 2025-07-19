package client

import "time"

// ========= Auth Types =========

type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FullName  string `json:"full_name,omitempty"`
	Bio       string `json:"bio,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type RegisterResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginRequest struct {
	Login    string `json:"login"` // Can be email or username
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type UpdatePasswordResponse struct {
	Message string `json:"message"`
}

// ========= User Types =========

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Bio       string    `json:"bio"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FullName  string `json:"full_name,omitempty"`
	Bio       string `json:"bio,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type CreateUserResponse User

type UpdateUserRequest struct {
	ID       int    `json:"id"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	FullName string `json:"full_name,omitempty"`
	Bio      string `json:"bio,omitempty"`
}

type UpdateUserResponse User

type UpdateAvatarRequest struct {
	AvatarURL string `json:"avatar_url"`
}

type SearchUsersResponse struct {
	Users []User `json:"users"`
	Total int    `json:"total"`
}

// ========= Post Types =========

type MediaItemInput struct {
	Type     string `json:"type"`
	URL      string `json:"url"`
	Position int    `json:"position,omitempty"`
}

type PostMedia struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	URL      string `json:"url"`
	Position int    `json:"position"`
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type PostAuthor struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
}

type Post struct {
	ID        int         `json:"id"`
	Title     string      `json:"title"`
	Content   string      `json:"content"`
	Author    PostAuthor  `json:"author"`
	Media     []PostMedia `json:"media"`
	Tags      []Tag       `json:"tags"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type CreatePostRequest struct {
	Title      string           `json:"title"`
	Content    string           `json:"content,omitempty"`
	MediaItems []MediaItemInput `json:"media_items,omitempty"`
	Tags       []string         `json:"tags,omitempty"`
}

type CreatePostResponse struct {
	ID              int         `json:"id"`
	Title           string      `json:"title"`
	Content         string      `json:"content"`
	AuthorID        int         `json:"author_id"`
	AuthorUsername  string      `json:"author_username"`
	AuthorFullName  string      `json:"author_full_name"`
	AuthorEmail     string      `json:"author_email"`
	AuthorBio       string      `json:"author_bio"`
	AuthorAvatarURL string      `json:"author_avatar_url"`
	Media           []PostMedia `json:"media"`
	Tags            []Tag       `json:"tags"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type UpdatePostRequest struct {
	Title      string           `json:"title,omitempty"`
	Content    string           `json:"content,omitempty"`
	MediaItems []MediaItemInput `json:"media_items,omitempty"`
	Tags       []string         `json:"tags,omitempty"`
}

type UpdatePostResponse Post

type ListPostsResponse struct {
	Posts []Post `json:"posts"`
	Total int    `json:"total"`
}

// ========= Relation Types =========

type FollowRequest struct {
	FolloweeID int `json:"followee_id"`
}

type FollowResponse struct {
	Message string `json:"message"`
}

type UnfollowRequest struct {
	FolloweeID int `json:"followee_id"`
}

type UnfollowResponse struct {
	Message string `json:"message"`
}

// ========= Notification Types =========

type Notification struct {
	ID        int         `json:"id"`
	UserID    int         `json:"user_id"`
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	IsRead    bool        `json:"is_read"`
	CreatedAt time.Time   `json:"created_at"`
}

type SendNotificationRequest struct {
	UserID  int         `json:"user_id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

type SendNotificationResponse struct {
	NotificationID int    `json:"notification_id"`
	Message        string `json:"message"`
}

type ReadNotificationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type RemoveNotificationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ReadAllUserNotificationsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type GetUnreadCountResponse struct {
	Count int `json:"count"`
}

type GetUserNotificationFeedResponse struct {
	Notifications []Notification `json:"notifications"`
	Page          int            `json:"page"`
	Limit         int            `json:"limit"`
	Total         int            `json:"total"`
	TotalPages    int            `json:"total_pages"`
}

type ErrorBody struct {
	Status  int    `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}
