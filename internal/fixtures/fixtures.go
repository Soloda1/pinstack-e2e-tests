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
	MinTagItems      = 2
	MaxMediaPosition = 8

	// Time constants
	MaxDaysAgo            = 30
	MaxHoursAfterCreation = 48

	// String lengths
	BioSentences   = 10
	TitleSentences = 3

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

// random int in [0;n)
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
		ID:        int64(safeRandIntn(MaxTestID) + 1),
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

func GenerateUpdateUserRequest(id int64, username, email, fullName, bio string) *UpdateUserRequest {
	if username == "" {
		username = gofakeit.Username()
	}

	if email == "" {
		email = gofakeit.Email()
	}

	if fullName == "" {
		fullName = gofakeit.Name()
	}

	if bio == "" {
		bio = gofakeit.HipsterSentence(BioSentences)
	}

	return &UpdateUserRequest{
		ID:       id,
		Username: username,
		Email:    email,
		FullName: fullName,
		Bio:      bio,
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
		Position: safeRandIntn(MaxMediaPosition) + 1,
	}
}

func GeneratePostMedia(id int64) PostMedia {
	return PostMedia{
		ID:       id,
		Type:     MediaTypes[safeRandIntn(len(MediaTypes))],
		URL:      gofakeit.ImageURL(PostImageWidth, PostImageHeight),
		Position: safeRandIntn(MaxMediaPosition) + 1,
	}
}

func GenerateTag() Tag {
	timestamp := strconv.Itoa(time.Now().Nanosecond())
	prefix := "test" + timestamp
	return Tag{
		ID:   int64(safeRandIntn(MaxTestID) + 1),
		Name: prefix + gofakeit.Generate("????????????????????"),
	}
}

func GeneratePostAuthor() PostAuthor {
	return PostAuthor{
		ID:        int64(safeRandIntn(MaxTestID) + 1),
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
		medialist = append(medialist, GeneratePostMedia(int64(i+1)))
	}

	var tags []Tag
	for i := 0; i < safeRandIntn(MaxTagItems)+MinTagItems; i++ {
		tags = append(tags, GenerateTag())
	}

	createdAt := time.Now().Add(-time.Duration(safeRandIntn(MaxDaysAgo)) * 24 * time.Hour)

	return &Post{
		ID:        int64(safeRandIntn(MaxTestID) + 1),
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
		tags = append(tags, GenerateTag().Name)
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
		tags = append(tags, GenerateTag().Name)
	}

	return &UpdatePostRequest{
		Title:      gofakeit.Sentence(TitleSentences),
		Content:    gofakeit.Paragraph(ParagraphMinSentences, ParagraphMaxSentences, ParagraphMaxWordsPerSentence, ParagraphBreak),
		MediaItems: medialist,
		Tags:       tags,
	}
}

// ========= Relation Data Generators =========

func GenerateFollowRequest(followeeId int64) *FollowRequest {
	if followeeId == 0 {
		followeeId = int64(safeRandIntn(MaxTestID) + 1)
	}
	return &FollowRequest{
		FolloweeID: followeeId,
	}
}

func GenerateUnfollowRequest(followeeId int64) *UnfollowRequest {
	if followeeId == 0 {
		followeeId = int64(safeRandIntn(MaxTestID) + 1)
	}
	return &UnfollowRequest{
		FolloweeID: followeeId,
	}
}

func GenerateGetFollowersResponse(userID int64, page, limit int) *GetFollowersResponse {
	followers := GenerateTestUsers(safeRandIntn(limit) + 1)
	total := len(followers) + safeRandIntn(50)
	totalPages := (total + limit - 1) / limit

	var followersUsers []User
	for _, user := range followers {
		followersUsers = append(followersUsers, *user)
	}

	return &GetFollowersResponse{
		Followers:  followersUsers,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

func GenerateGetFolloweesResponse(userID int64, page, limit int) *GetFolloweesResponse {
	followees := GenerateTestUsers(safeRandIntn(limit) + 1)
	total := len(followees) + safeRandIntn(50)
	totalPages := (total + limit - 1) / limit

	var followeesUsers []User
	for _, user := range followees {
		followeesUsers = append(followeesUsers, *user)
	}

	return &GetFolloweesResponse{
		Followees:  followeesUsers,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// ========= Notification Data Generators =========

const (
	// Notification types
	NotificationTypeFollowCreated = "follow_created"
	NotificationTypeLike          = "like"
	NotificationTypeComment       = "comment"
	NotificationTypeMention       = "mention"
	NotificationTypeSystem        = "system"

	// Payload constants
	PayloadDataKey = "data"

	// Random constants
	RandomBoolModulo = 2
)

var NotificationTypes = []string{
	NotificationTypeFollowCreated,
	NotificationTypeLike,
	NotificationTypeComment,
	NotificationTypeMention,
	NotificationTypeSystem,
}

func GenerateNotification(userId int64) *Notification {
	if userId == 0 {
		userId = int64(safeRandIntn(MaxTestID) + 1)
	}
	return &Notification{
		ID:        int64(safeRandIntn(MaxTestID) + 1),
		UserID:    userId,
		Type:      NotificationTypes[safeRandIntn(len(NotificationTypes))],
		Payload:   map[string]interface{}{PayloadDataKey: gofakeit.Sentence(TitleSentences)},
		IsRead:    safeRandIntn(RandomBoolModulo) == 1,
		CreatedAt: time.Now().Add(-time.Duration(safeRandIntn(MaxHoursAfterCreation)) * time.Hour),
	}
}

func GenerateSendNotificationRequest(userId int64) *SendNotificationRequest {
	if userId == 0 {
		userId = int64(safeRandIntn(MaxTestID) + 1)
	}
	return &SendNotificationRequest{
		UserID:  userId,
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

func GenerateTestNotifications(userID int64, count int) []*Notification {
	var notifications []*Notification
	for i := 0; i < count; i++ {
		notification := GenerateNotification(userID)
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
