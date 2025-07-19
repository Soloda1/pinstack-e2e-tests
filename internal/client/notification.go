package client

import (
	"github.com/Soloda1/pinstack-e2e-tests/internal/custom_errors"
	"log/slog"
	"net/url"
	"strconv"
)

type NotificationClient struct {
	client *Client
}

func NewNotificationClient(client *Client) *NotificationClient {
	return &NotificationClient{
		client: client,
	}
}

func (nc *NotificationClient) GetNotificationByID(notificationID int) (*Notification, error) {
	nc.client.log.Debug("Getting notification by ID", slog.Int("notification_id", notificationID))

	var response Notification
	err := nc.client.Get("/v1/notification/"+strconv.Itoa(notificationID), nil, &response)
	if err != nil {
		nc.client.log.Error("Failed to get notification by ID",
			slog.Int("notification_id", notificationID),
			slog.String("error", err.Error()),
		)
		return nil, custom_errors.ErrNotificationGetFailed
	}

	nc.client.log.Debug("Got notification by ID successfully",
		slog.Int("notification_id", notificationID),
		slog.String("type", response.Type),
		slog.Bool("is_read", response.IsRead),
	)
	return &response, nil
}

func (nc *NotificationClient) SendNotification(req SendNotificationRequest) (*SendNotificationResponse, error) {
	nc.client.log.Info("Sending notification",
		slog.Int("user_id", req.UserID),
		slog.String("type", req.Type),
	)

	var response SendNotificationResponse
	err := nc.client.Post("/v1/notification/send", req, &response)
	if err != nil {
		nc.client.log.Error("Failed to send notification",
			slog.Int("user_id", req.UserID),
			slog.String("type", req.Type),
			slog.String("error", err.Error()),
		)
		return nil, custom_errors.ErrNotificationSendFailed
	}

	nc.client.log.Info("Notification sent successfully",
		slog.Int("notification_id", response.NotificationID),
		slog.String("message", response.Message),
	)
	return &response, nil
}

func (nc *NotificationClient) ReadNotification(notificationID int) (*ReadNotificationResponse, error) {
	nc.client.log.Debug("Marking notification as read", slog.Int("notification_id", notificationID))

	var response ReadNotificationResponse
	err := nc.client.Put("/v1/notification/"+strconv.Itoa(notificationID)+"/read", nil, &response)
	if err != nil {
		nc.client.log.Error("Failed to mark notification as read",
			slog.Int("notification_id", notificationID),
			slog.String("error", err.Error()),
		)
		return nil, custom_errors.ErrNotificationMarkReadFailed
	}

	nc.client.log.Debug("Notification marked as read successfully",
		slog.Int("notification_id", notificationID),
		slog.Bool("success", response.Success),
	)
	return &response, nil
}

func (nc *NotificationClient) RemoveNotification(notificationID int) (*RemoveNotificationResponse, error) {
	nc.client.log.Info("Removing notification", slog.Int("notification_id", notificationID))

	var response RemoveNotificationResponse
	err := nc.client.Delete("/v1/notification/"+strconv.Itoa(notificationID), &response)
	if err != nil {
		nc.client.log.Error("Failed to remove notification",
			slog.Int("notification_id", notificationID),
			slog.String("error", err.Error()),
		)
		return nil, custom_errors.ErrNotificationRemoveFailed
	}

	nc.client.log.Info("Notification removed successfully",
		slog.Int("notification_id", notificationID),
		slog.Bool("success", response.Success),
	)
	return &response, nil
}

func (nc *NotificationClient) ReadAllUserNotifications(userID int) (*ReadAllUserNotificationsResponse, error) {
	nc.client.log.Info("Marking all notifications as read for user", slog.Int("user_id", userID))

	var response ReadAllUserNotificationsResponse
	err := nc.client.Put("/v1/notification/read-all/"+strconv.Itoa(userID), nil, &response)
	if err != nil {
		nc.client.log.Error("Failed to mark all notifications as read",
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
		)
		return nil, custom_errors.ErrNotificationReadAllFailed
	}

	nc.client.log.Info("All notifications marked as read successfully",
		slog.Int("user_id", userID),
		slog.Bool("success", response.Success),
	)
	return &response, nil
}

func (nc *NotificationClient) GetUnreadCount(userID int) (*GetUnreadCountResponse, error) {
	nc.client.log.Debug("Getting unread notification count", slog.Int("user_id", userID))

	var response GetUnreadCountResponse
	err := nc.client.Get("/v1/notification/unread-count/"+strconv.Itoa(userID), nil, &response)
	if err != nil {
		nc.client.log.Error("Failed to get unread notification count",
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
		)
		return nil, custom_errors.ErrNotificationCountFailed
	}

	nc.client.log.Debug("Got unread notification count successfully",
		slog.Int("user_id", userID),
		slog.Int("count", response.Count),
	)
	return &response, nil
}

func (nc *NotificationClient) GetUserNotificationFeed(userID, page, limit int) (*GetUserNotificationFeedResponse, error) {
	nc.client.log.Debug("Getting user notification feed",
		slog.Int("user_id", userID),
		slog.Int("page", page),
		slog.Int("limit", limit),
	)

	queryParams := url.Values{}

	if page > 0 {
		queryParams.Set("page", strconv.Itoa(page))
	}

	if limit > 0 {
		queryParams.Set("limit", strconv.Itoa(limit))
	}

	var response GetUserNotificationFeedResponse
	err := nc.client.Get("/v1/notification/feed/"+strconv.Itoa(userID), queryParams, &response)
	if err != nil {
		nc.client.log.Error("Failed to get notification feed",
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
		)
		return nil, custom_errors.ErrNotificationFeedFailed
	}

	nc.client.log.Debug("Got user notification feed successfully",
		slog.Int("user_id", userID),
		slog.Int("total", response.Total),
		slog.Int("notifications_count", len(response.Notifications)),
	)
	return &response, nil
}
