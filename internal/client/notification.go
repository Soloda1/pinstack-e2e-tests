package client

import (
	"log/slog"
	"net/url"
	"strconv"

	"github.com/Soloda1/pinstack-system-tests/internal/fixtures"
)

type NotificationClient struct {
	client *Client
}

func NewNotificationClient(client *Client) *NotificationClient {
	return &NotificationClient{
		client: client,
	}
}

func (nc *NotificationClient) GetNotificationByID(notificationID int64) (*fixtures.Notification, error) {
	nc.client.log.Debug("Getting notification by ID", slog.Int64("notification_id", notificationID))

	var response fixtures.Notification
	err := nc.client.Get("/v1/notification/"+strconv.FormatInt(notificationID, 10), nil, &response)
	if err != nil {
		nc.client.log.Error("Failed to get notification by ID",
			slog.Int64("notification_id", notificationID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	nc.client.log.Debug("Got notification by ID successfully",
		slog.Int64("notification_id", notificationID),
		slog.String("type", response.Type),
		slog.Bool("is_read", response.IsRead),
	)
	return &response, nil
}

func (nc *NotificationClient) SendNotification(req fixtures.SendNotificationRequest) (*fixtures.SendNotificationResponse, error) {
	nc.client.log.Info("Sending notification",
		slog.Int64("user_id", req.UserID),
		slog.String("type", req.Type),
	)

	var response fixtures.SendNotificationResponse
	err := nc.client.Post("/v1/notification/send", req, &response)
	if err != nil {
		nc.client.log.Error("Failed to send notification",
			slog.Int64("user_id", req.UserID),
			slog.String("type", req.Type),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	nc.client.log.Info("Notification sent successfully",
		slog.String("message", response.Message),
	)
	return &response, nil
}

func (nc *NotificationClient) ReadNotification(notificationID int64) (*fixtures.ReadNotificationResponse, error) {
	nc.client.log.Debug("Marking notification as read", slog.Int64("notification_id", notificationID))

	var response fixtures.ReadNotificationResponse
	err := nc.client.Put("/v1/notification/"+strconv.FormatInt(notificationID, 10)+"/read", nil, &response)
	if err != nil {
		nc.client.log.Error("Failed to mark notification as read",
			slog.Int64("notification_id", notificationID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	nc.client.log.Debug("Notification marked as read successfully",
		slog.Int64("notification_id", notificationID),
		slog.Bool("success", response.Success),
	)
	return &response, nil
}

func (nc *NotificationClient) RemoveNotification(notificationID int64) (*fixtures.RemoveNotificationResponse, error) {
	nc.client.log.Info("Removing notification", slog.Int64("notification_id", notificationID))

	var response fixtures.RemoveNotificationResponse
	err := nc.client.Delete("/v1/notification/"+strconv.FormatInt(notificationID, 10), &response)
	if err != nil {
		nc.client.log.Error("Failed to remove notification",
			slog.Int64("notification_id", notificationID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	nc.client.log.Info("Notification removed successfully",
		slog.Int64("notification_id", notificationID),
		slog.Bool("success", response.Success),
	)
	return &response, nil
}

func (nc *NotificationClient) ReadAllUserNotifications(userID int64) (*fixtures.ReadAllUserNotificationsResponse, error) {
	nc.client.log.Info("Marking all notifications as read for user", slog.Int64("user_id", userID))

	var response fixtures.ReadAllUserNotificationsResponse
	err := nc.client.Put("/v1/notification/read-all", nil, &response)
	if err != nil {
		nc.client.log.Error("Failed to mark all notifications as read",
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	nc.client.log.Info("All notifications marked as read successfully",
		slog.Int64("user_id", userID),
		slog.Bool("success", response.Success),
	)
	return &response, nil
}

func (nc *NotificationClient) GetUnreadCount(userID int64) (*fixtures.GetUnreadCountResponse, error) {
	nc.client.log.Debug("Getting unread notification count", slog.Int64("user_id", userID))

	var response fixtures.GetUnreadCountResponse
	err := nc.client.Get("/v1/notification/unread-count", nil, &response)
	if err != nil {
		nc.client.log.Error("Failed to get unread notification count",
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	nc.client.log.Debug("Got unread notification count successfully",
		slog.Int64("user_id", userID),
		slog.Int("count", response.Count),
	)
	return &response, nil
}

func (nc *NotificationClient) GetUserNotificationFeed(userID int64, page, limit int) (*fixtures.GetUserNotificationFeedResponse, error) {
	nc.client.log.Debug("Getting user notification feed",
		slog.Int64("user_id", userID),
		slog.Int("page", page),
		slog.Int("limit", limit),
	)

	queryParams := url.Values{}
	queryParams.Add("page", strconv.Itoa(page))
	queryParams.Add("limit", strconv.Itoa(limit))

	var response fixtures.GetUserNotificationFeedResponse
	err := nc.client.Get("/v1/notification/feed", queryParams, &response)
	if err != nil {
		nc.client.log.Error("Failed to get user notification feed",
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	nc.client.log.Debug("Got user notification feed successfully",
		slog.Int64("user_id", userID),
		slog.Int("total", response.Total),
		slog.Int("notifications_count", len(response.Notifications)),
	)
	return &response, nil
}
