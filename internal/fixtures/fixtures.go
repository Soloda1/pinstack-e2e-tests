package fixtures

import (
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

func init() {
	gofakeit.Seed(0)
}

// ========= Auth Data Generators =========

func GenerateRegisterRequest() *RegisterRequest {
	return &RegisterRequest{
		Username:  gofakeit.Username(),
		Email:     gofakeit.Email(),
		Password:  gofakeit.Password(true, true, true, true, false, 10),
		FullName:  gofakeit.Name(),
		Bio:       gofakeit.HipsterSentence(10),
		AvatarURL: gofakeit.ImageURL(300, 300),
	}
}

func GenerateLoginRequest(username, password string) *LoginRequest {
	if username == "" {
		username = gofakeit.Username()
	}

	if password == "" {
		password = gofakeit.Password(true, true, true, true, false, 10)
	}

	return &LoginRequest{
		Login:    username,
		Password: password,
	}
}

func GenerateRefreshTokenRequest(token string) *RefreshTokenRequest {
	if token == "" {
		token = gofakeit.UUID()
	}

	return &RefreshTokenRequest{
		RefreshToken: token,
	}
}

func GenerateLogoutRequest(token string) *LogoutRequest {
	if token == "" {
		token = gofakeit.UUID()
	}

	return &LogoutRequest{
		RefreshToken: token,
	}
}

func GenerateUpdatePasswordRequest() *UpdatePasswordRequest {
	return &UpdatePasswordRequest{
		OldPassword: gofakeit.Password(true, true, true, true, false, 10),
		NewPassword: gofakeit.Password(true, true, true, true, false, 12),
	}
}

// ========= User Data Generators =========

func GenerateUser() *User {
	return &User{
		ID:        rand.Intn(10000) + 1,
		Username:  gofakeit.Username(),
		Email:     gofakeit.Email(),
		FullName:  gofakeit.Name(),
		Bio:       gofakeit.HipsterSentence(10),
		AvatarURL: gofakeit.ImageURL(300, 300),
		CreatedAt: time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour),
		UpdatedAt: time.Now(),
	}
}

func GenerateCreateUserRequest() *CreateUserRequest {
	return &CreateUserRequest{
		Username:  gofakeit.Username(),
		Email:     gofakeit.Email(),
		Password:  gofakeit.Password(true, true, true, true, false, 10),
		FullName:  gofakeit.Name(),
		Bio:       gofakeit.HipsterSentence(10),
		AvatarURL: gofakeit.ImageURL(300, 300),
	}
}

func GenerateUpdateUserRequest(id int) *UpdateUserRequest {
	return &UpdateUserRequest{
		ID:       id,
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		FullName: gofakeit.Name(),
		Bio:      gofakeit.HipsterSentence(10),
	}
}

func GenerateUpdateAvatarRequest() *UpdateAvatarRequest {
	return &UpdateAvatarRequest{
		AvatarURL: gofakeit.ImageURL(300, 300),
	}
}

// ========= Post Data Generators =========

func GenerateMediaItemInput() MediaItemInput {
	mediaTypes := []string{"image", "video"}

	return MediaItemInput{
		Type:     mediaTypes[rand.Intn(len(mediaTypes))],
		URL:      gofakeit.ImageURL(800, 600),
		Position: rand.Intn(9),
	}
}

func GeneratePostMedia(id int) PostMedia {
	mediaTypes := []string{"image", "video"}

	return PostMedia{
		ID:       id,
		Type:     mediaTypes[rand.Intn(len(mediaTypes))],
		URL:      gofakeit.ImageURL(800, 600),
		Position: rand.Intn(9),
	}
}

func GenerateTag() Tag {
	return Tag{
		ID:   rand.Intn(100) + 1,
		Name: gofakeit.Word(),
	}
}

func GeneratePostAuthor() PostAuthor {
	return PostAuthor{
		ID:        rand.Intn(10000) + 1,
		Username:  gofakeit.Username(),
		FullName:  gofakeit.Name(),
		AvatarURL: gofakeit.ImageURL(300, 300),
	}
}

func GeneratePost() *Post {
	var medialist []PostMedia
	for i := 0; i < rand.Intn(5); i++ {
		medialist = append(medialist, GeneratePostMedia(i+1))
	}

	var tags []Tag
	for i := 0; i < rand.Intn(4)+1; i++ {
		tags = append(tags, GenerateTag())
	}

	createdAt := time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour)

	return &Post{
		ID:        rand.Intn(10000) + 1,
		Title:     gofakeit.Sentence(5),
		Content:   gofakeit.Paragraph(3, 5, 10, "\n"),
		Author:    GeneratePostAuthor(),
		Media:     medialist,
		Tags:      tags,
		CreatedAt: createdAt,
		UpdatedAt: createdAt.Add(time.Duration(rand.Intn(48)) * time.Hour),
	}
}

func GenerateCreatePostRequest() *CreatePostRequest {
	var medialist []MediaItemInput
	for i := 0; i < rand.Intn(5); i++ {
		medialist = append(medialist, GenerateMediaItemInput())
	}

	var tags []string
	for i := 0; i < rand.Intn(4)+1; i++ {
		tags = append(tags, gofakeit.Word())
	}

	return &CreatePostRequest{
		Title:      gofakeit.Sentence(5),
		Content:    gofakeit.Paragraph(3, 5, 10, "\n"),
		MediaItems: medialist,
		Tags:       tags,
	}
}

func GenerateUpdatePostRequest() *UpdatePostRequest {
	var medialist []MediaItemInput
	for i := 0; i < rand.Intn(5); i++ {
		medialist = append(medialist, GenerateMediaItemInput())
	}

	var tags []string
	for i := 0; i < rand.Intn(4)+1; i++ {
		tags = append(tags, gofakeit.Word())
	}

	return &UpdatePostRequest{
		Title:      gofakeit.Sentence(5),
		Content:    gofakeit.Paragraph(3, 5, 10, "\n"),
		MediaItems: medialist,
		Tags:       tags,
	}
}

// ========= Relation Data Generators =========

func GenerateFollowRequest() *FollowRequest {
	return &FollowRequest{
		FolloweeID: rand.Intn(10000) + 1,
	}
}

func GenerateUnfollowRequest() *UnfollowRequest {
	return &UnfollowRequest{
		FolloweeID: rand.Intn(10000) + 1,
	}
}

// ========= Notification Data Generators =========

func GenerateNotification() *Notification {
	notificationTypes := []string{"follow", "like", "comment", "mention", "system"}

	return &Notification{
		ID:        rand.Intn(10000) + 1,
		UserID:    rand.Intn(10000) + 1,
		Type:      notificationTypes[rand.Intn(len(notificationTypes))],
		Payload:   map[string]interface{}{"data": gofakeit.Sentence(5)},
		IsRead:    rand.Intn(2) == 1,
		CreatedAt: time.Now().Add(-time.Duration(rand.Intn(48)) * time.Hour),
	}
}

func GenerateSendNotificationRequest() *SendNotificationRequest {
	notificationTypes := []string{"follow", "like", "comment", "mention", "system"}

	return &SendNotificationRequest{
		UserID:  rand.Intn(10000) + 1,
		Type:    notificationTypes[rand.Intn(len(notificationTypes))],
		Payload: map[string]interface{}{"data": gofakeit.Sentence(5)},
	}
}

// ========= Test Fixtures Sets =========

func GenerateTestUsers(count int) []*User {
	var users []*User
	for i := 0; i < count; i++ {
		users = append(users, GenerateUser())
	}
	return users
}

func GenerateTestPosts(count int) []*Post {
	var posts []*Post
	for i := 0; i < count; i++ {
		posts = append(posts, GeneratePost())
	}
	return posts
}

func GenerateTestNotifications(userID int, count int) []*Notification {
	var notifications []*Notification
	for i := 0; i < count; i++ {
		notification := GenerateNotification()
		notification.UserID = userID
		notifications = append(notifications, notification)
	}
	return notifications
}

func CreateUserJourney() (*RegisterRequest, *User, []*Post, []*Notification, []*User) {
	registerData := GenerateRegisterRequest()

	user := &User{
		ID:        1,
		Username:  registerData.Username,
		Email:     registerData.Email,
		FullName:  registerData.FullName,
		Bio:       registerData.Bio,
		AvatarURL: registerData.AvatarURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	var posts []*Post
	for i := 0; i < 3; i++ {
		post := GeneratePost()
		post.Author.ID = user.ID
		post.Author.Username = user.Username
		post.Author.FullName = user.FullName
		post.Author.AvatarURL = user.AvatarURL
		posts = append(posts, post)
	}

	notifications := GenerateTestNotifications(user.ID, 5)

	otherUsers := GenerateTestUsers(5)

	return registerData, user, posts, notifications, otherUsers
}
