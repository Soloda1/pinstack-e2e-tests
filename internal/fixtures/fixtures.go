package fixtures

import (
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

const (
	// General constants
	MaxTestID = 10000

	// Password constants
	MinPasswordLength        = 6
	DefaultNewPasswordLength = 10
	UpdatedPasswordLength    = 12

	// Image dimensions
	AvatarSize      = 300
	PostImageWidth  = 800
	PostImageHeight = 600

	// Collection sizes
	MaxMediaItems    = 5
	MaxTagItems      = 4
	MinTagItems      = 1
	MaxMediaPosition = 9

	// Time constants
	MaxDaysAgo            = 30
	MaxHoursAfterCreation = 48

	// String lengths
	BioSentences   = 10
	TitleSentences = 5

	// Test data limits
	DefaultUserJourneyPostCount         = 3
	DefaultUserJourneyNotificationCount = 5
	DefaultOtherUsersCount              = 5
)

// Thread-safe random number generator
var (
	globalRand     *rand.Rand
	globalRandLock sync.Mutex
	seedOnce       sync.Once
)

func GetSafeRandom() *rand.Rand {
	globalRandLock.Lock()
	defer globalRandLock.Unlock()
	return globalRand
}

func safeRandIntn(n int) int {
	globalRandLock.Lock()
	defer globalRandLock.Unlock()
	return globalRand.Intn(n)
}

func initializeSeed() {
	seedOnce.Do(func() {
		seed := time.Now().UnixNano()
		if envSeed, ok := os.LookupEnv("TEST_SEED"); ok {
			if parsedSeed, err := strconv.ParseInt(envSeed, 10, 64); err == nil {
				seed = parsedSeed
			}
		}
		gofakeit.Seed(seed)

		globalRand = rand.New(rand.NewSource(seed))
	})
}

func init() {
	initializeSeed()
}

// ========= Auth Data Generators =========

func GenerateRegisterRequest() *RegisterRequest {
	return &RegisterRequest{
		Username:  gofakeit.Username(),
		Email:     gofakeit.Email(),
		Password:  gofakeit.Password(true, true, true, true, false, DefaultNewPasswordLength),
		FullName:  gofakeit.Name(),
		Bio:       gofakeit.HipsterSentence(BioSentences),
		AvatarURL: gofakeit.ImageURL(AvatarSize, AvatarSize),
	}
}

func GenerateLoginRequest(username, password string) *LoginRequest {
	if username == "" {
		username = gofakeit.Username()
	}

	if password == "" {
		password = gofakeit.Password(true, true, true, true, false, DefaultNewPasswordLength)
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
		OldPassword: gofakeit.Password(true, true, true, true, false, DefaultNewPasswordLength),
		NewPassword: gofakeit.Password(true, true, true, true, false, UpdatedPasswordLength),
	}
}

// ========= User Data Generators =========

func GenerateUser() *User {
	return &User{
		ID:        safeRandIntn(MaxTestID) + 1,
		Username:  gofakeit.Username(),
		Email:     gofakeit.Email(),
		FullName:  gofakeit.Name(),
		Bio:       gofakeit.HipsterSentence(BioSentences),
		AvatarURL: gofakeit.ImageURL(AvatarSize, AvatarSize),
		CreatedAt: time.Now().Add(-time.Duration(safeRandIntn(MaxDaysAgo)) * 24 * time.Hour),
		UpdatedAt: time.Now(),
	}
}

func GenerateCreateUserRequest() *CreateUserRequest {
	return &CreateUserRequest{
		Username:  gofakeit.Username(),
		Email:     gofakeit.Email(),
		Password:  gofakeit.Password(true, true, true, true, false, DefaultNewPasswordLength),
		FullName:  gofakeit.Name(),
		Bio:       gofakeit.HipsterSentence(BioSentences),
		AvatarURL: gofakeit.ImageURL(AvatarSize, AvatarSize),
	}
}

func GenerateUpdateUserRequest(id int) *UpdateUserRequest {
	return &UpdateUserRequest{
		ID:       id,
		Username: gofakeit.Username(),
		Email:    gofakeit.Email(),
		FullName: gofakeit.Name(),
		Bio:      gofakeit.HipsterSentence(BioSentences),
	}
}

func GenerateUpdateAvatarRequest() *UpdateAvatarRequest {
	return &UpdateAvatarRequest{
		AvatarURL: gofakeit.ImageURL(AvatarSize, AvatarSize),
	}
}

// ========= Post Data Generators =========

const (
	// Media types
	MediaTypeImage = "image"
	MediaTypeVideo = "video"
)

var MediaTypes = []string{MediaTypeImage, MediaTypeVideo}

func GenerateMediaItemInput() MediaItemInput {
	return MediaItemInput{
		Type:     MediaTypes[safeRandIntn(len(MediaTypes))],
		URL:      gofakeit.ImageURL(PostImageWidth, PostImageHeight),
		Position: safeRandIntn(MaxMediaPosition),
	}
}

func GeneratePostMedia(id int) PostMedia {
	return PostMedia{
		ID:       id,
		Type:     MediaTypes[safeRandIntn(len(MediaTypes))],
		URL:      gofakeit.ImageURL(PostImageWidth, PostImageHeight),
		Position: safeRandIntn(MaxMediaPosition),
	}
}

func GenerateTag() Tag {
	return Tag{
		ID:   safeRandIntn(MaxTestID) + 1,
		Name: gofakeit.Word(),
	}
}

func GeneratePostAuthor() PostAuthor {
	return PostAuthor{
		ID:        safeRandIntn(MaxTestID) + 1,
		Username:  gofakeit.Username(),
		FullName:  gofakeit.Name(),
		AvatarURL: gofakeit.ImageURL(AvatarSize, AvatarSize),
	}
}

const (
	// Paragraph generation constants
	ParagraphMinSentences        = 3
	ParagraphMaxSentences        = 5
	ParagraphMaxWordsPerSentence = 10
	ParagraphBreak               = "\n"
)

func GeneratePost() *Post {
	var medialist []PostMedia
	for i := 0; i < safeRandIntn(MaxMediaItems); i++ {
		medialist = append(medialist, GeneratePostMedia(i+1))
	}

	var tags []Tag
	for i := 0; i < safeRandIntn(MaxTagItems)+MinTagItems; i++ {
		tags = append(tags, GenerateTag())
	}

	createdAt := time.Now().Add(-time.Duration(safeRandIntn(MaxDaysAgo)) * 24 * time.Hour)

	return &Post{
		ID:        safeRandIntn(MaxTestID) + 1,
		Title:     gofakeit.Sentence(TitleSentences),
		Content:   gofakeit.Paragraph(ParagraphMinSentences, ParagraphMaxSentences, ParagraphMaxWordsPerSentence, ParagraphBreak),
		Author:    GeneratePostAuthor(),
		Media:     medialist,
		Tags:      tags,
		CreatedAt: createdAt,
		UpdatedAt: createdAt.Add(time.Duration(safeRandIntn(MaxHoursAfterCreation)) * time.Hour),
	}
}

func GenerateCreatePostRequest() *CreatePostRequest {
	var medialist []MediaItemInput
	for i := 0; i < safeRandIntn(MaxMediaItems); i++ {
		medialist = append(medialist, GenerateMediaItemInput())
	}

	var tags []string
	for i := 0; i < safeRandIntn(MaxTagItems)+MinTagItems; i++ {
		tags = append(tags, gofakeit.Word())
	}

	return &CreatePostRequest{
		Title:      gofakeit.Sentence(TitleSentences),
		Content:    gofakeit.Paragraph(ParagraphMinSentences, ParagraphMaxSentences, ParagraphMaxWordsPerSentence, ParagraphBreak),
		MediaItems: medialist,
		Tags:       tags,
	}
}

func GenerateUpdatePostRequest() *UpdatePostRequest {
	var medialist []MediaItemInput
	for i := 0; i < safeRandIntn(MaxMediaItems); i++ {
		medialist = append(medialist, GenerateMediaItemInput())
	}

	var tags []string
	for i := 0; i < safeRandIntn(MaxTagItems)+MinTagItems; i++ {
		tags = append(tags, gofakeit.Word())
	}

	return &UpdatePostRequest{
		Title:      gofakeit.Sentence(TitleSentences),
		Content:    gofakeit.Paragraph(ParagraphMinSentences, ParagraphMaxSentences, ParagraphMaxWordsPerSentence, ParagraphBreak),
		MediaItems: medialist,
		Tags:       tags,
	}
}

// ========= Relation Data Generators =========

func GenerateFollowRequest() *FollowRequest {
	return &FollowRequest{
		FolloweeID: safeRandIntn(MaxTestID) + 1,
	}
}

func GenerateUnfollowRequest() *UnfollowRequest {
	return &UnfollowRequest{
		FolloweeID: safeRandIntn(MaxTestID) + 1,
	}
}

// ========= Notification Data Generators =========

const (
	// Notification types
	NotificationTypeFollow  = "follow"
	NotificationTypeLike    = "like"
	NotificationTypeComment = "comment"
	NotificationTypeMention = "mention"
	NotificationTypeSystem  = "system"

	// Payload constants
	PayloadDataKey = "data"

	// Random constants
	RandomBoolModulo = 2
)

var NotificationTypes = []string{
	NotificationTypeFollow,
	NotificationTypeLike,
	NotificationTypeComment,
	NotificationTypeMention,
	NotificationTypeSystem,
}

func GenerateNotification() *Notification {
	return &Notification{
		ID:        safeRandIntn(MaxTestID) + 1,
		UserID:    safeRandIntn(MaxTestID) + 1,
		Type:      NotificationTypes[safeRandIntn(len(NotificationTypes))],
		Payload:   map[string]interface{}{PayloadDataKey: gofakeit.Sentence(TitleSentences)},
		IsRead:    safeRandIntn(RandomBoolModulo) == 1,
		CreatedAt: time.Now().Add(-time.Duration(safeRandIntn(MaxHoursAfterCreation)) * time.Hour),
	}
}

func GenerateSendNotificationRequest() *SendNotificationRequest {
	return &SendNotificationRequest{
		UserID:  safeRandIntn(MaxTestID) + 1,
		Type:    NotificationTypes[safeRandIntn(len(NotificationTypes))],
		Payload: map[string]interface{}{PayloadDataKey: gofakeit.Sentence(TitleSentences)},
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

const (
	// Default IDs
	DefaultUserID = 1
)

// ========= User Journey Generators =========

func CreateUserJourney() *UserJourney {
	registerData := GenerateRegisterRequest()

	user := &User{
		ID:        DefaultUserID,
		Username:  registerData.Username,
		Email:     registerData.Email,
		FullName:  registerData.FullName,
		Bio:       registerData.Bio,
		AvatarURL: registerData.AvatarURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	var posts []*Post
	for i := 0; i < DefaultUserJourneyPostCount; i++ {
		post := GeneratePost()
		post.Author.ID = user.ID
		post.Author.Username = user.Username
		post.Author.FullName = user.FullName
		post.Author.AvatarURL = user.AvatarURL
		posts = append(posts, post)
	}

	notifications := GenerateTestNotifications(user.ID, DefaultUserJourneyNotificationCount)

	otherUsers := GenerateTestUsers(DefaultOtherUsersCount)

	return &UserJourney{
		RegisterRequest: registerData,
		User:            user,
		Posts:           posts,
		Notifications:   notifications,
		OtherUsers:      otherUsers,
	}

}
