package notification

import (
	"log"
	"strconv"
	"time"

	"github.com/shumipro/meetapp/server/models"
	"github.com/shumipro/meetapp/server/oauth"
	"golang.org/x/net/context"
)

func SendDiscussion(ctx context.Context, discussion models.DiscussionInfo, appInfo models.AppInfo) {
	notificationType := models.NotificationDiscussion
	nowTime := time.Now()
	id := strconv.FormatInt(nowTime.UnixNano(), 10)

	notification := models.Notification{}
	notification.NotificationID = id
	notification.SourceID = discussion.ID
	notification.NotificationType = notificationType
	notification.DetailURL = generateURL(notificationType, appInfo.ID)
	notification.Message = generateMessage(notificationType, discussion.Message)
	notification.IsRead = false
	notification.CreatedAt = nowTime

	a, _ := oauth.FromContext(ctx)
	// TODO: あとで共通化する
	// ディスカッションの結果として同期する必要ないので非同期処理する
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
			}
		}()

		for _, m := range appInfo.Members {
			// 自分は通知しない
			if m.UserID == a.UserID {
				continue
			}

			sendNotification(ctx, m.UserID, notification)
		}
	}()
}

func SendStar(ctx context.Context, user models.User, appInfo models.AppInfo) {
	notificationType := models.NotificationStar
	nowTime := time.Now()
	id := strconv.FormatInt(nowTime.UnixNano(), 10)

	notification := models.Notification{}
	notification.NotificationID = id
	notification.SourceID = user.ID
	notification.NotificationType = notificationType
	notification.DetailURL = generateURL(notificationType, user.ID)
	notification.Message = generateMessage(notificationType, user.Name)
	notification.IsRead = false
	notification.CreatedAt = nowTime

	a, _ := oauth.FromContext(ctx)
	// ディスカッションの結果として同期する必要ないので非同期処理する
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
			}
		}()

		for _, m := range appInfo.Members {
			// 自分は通知しない
			if m.UserID == a.UserID {
				continue
			}

			sendNotification(ctx, m.UserID, notification)
		}
	}()
}

func sendNotification(ctx context.Context, userID string, notification models.Notification) {
	err := models.NotificationTable.AddNotification(ctx, userID, notification)
	if err != nil {
		// 非同期処理なのでpanicしない
		log.Println("ERROR!", err)
	} else {
		log.Println("OK: AddNotification", userID, notification)
	}
}

func generateURL(notification models.NotificationType, sourceID string) string {
	switch notification {
	case models.NotificationDiscussion:
		return "/app/detail/" + sourceID
	case models.NotificationStar:
		return "/mypage/other/" + sourceID
	default:
		return ""
	}
}

func generateMessage(notification models.NotificationType, message string) string {
	switch notification {
	case models.NotificationDiscussion:
		return "新着メッセージ: " + message
	case models.NotificationStar:
		return message + "さんが「いいね」しました。"
	default:
		return ""
	}
}
