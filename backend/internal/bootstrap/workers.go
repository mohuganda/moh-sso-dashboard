package bootstrap

import (
	"context"
	"errors"
	"time"

	"github.com/moh-sso-dashboard/internal/config"
	logger "github.com/moh-sso-dashboard/internal/log"
	emailRepo "github.com/moh-sso-dashboard/internal/repository/email"
	notificationDeliveryRepo "github.com/moh-sso-dashboard/internal/repository/notification_delivery"
	processRepo "github.com/moh-sso-dashboard/internal/repository/processes"
	"github.com/moh-sso-dashboard/internal/service"
	importSvc "github.com/moh-sso-dashboard/internal/service/import"
	"github.com/moh-sso-dashboard/internal/storage"
	"github.com/moh-sso-dashboard/internal/worker"
)

type workerDependencies struct {
	Config                         *config.Config
	ProcessRepository              processRepo.ProcessRepository
	ImportService                  *importSvc.Service
	FileStorage                    storage.Storage
	EmailRepository                emailRepo.EmailRepository
	SMTPService                    worker.EmailSender
	NotificationDeliveryRepository notificationDeliveryRepo.NotificationDeliveryRepository
	EmailService                   service.EmailService
	SMSService                     service.SMSService
	Logger                         *logger.Logger
}

func startBackgroundWorkers(ctx context.Context, deps workerDependencies) {
	documentWorker, err := worker.NewDocumentWorker(
		deps.ProcessRepository,
		deps.ImportService,
		3*time.Second,
		deps.FileStorage,
		deps.Logger,
	)
	if err != nil {
		deps.Logger.Fatal("Failed to initialize document worker: ", err)
	}

	go func() {
		deps.Logger.Info("Background document worker started")

		if err := documentWorker.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			deps.Logger.Error("Document worker stopped with error: ", err)
		}
	}()

	emailWorker, err := worker.NewEmailWorker(
		deps.EmailRepository,
		deps.SMTPService,
		3*time.Second,
		20,
		3,
		deps.Logger,
	)
	if err != nil {
		deps.Logger.Fatal("Failed to initialize email worker: ", err)
	}

	go func() {
		deps.Logger.Info("Background email worker started")

		if err := emailWorker.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			deps.Logger.Error("Email worker stopped with error: ", err)
		}
	}()

	if deps.Config != nil && !deps.Config.NotificationDelivery.WorkerEnabled {
		deps.Logger.Info("Notification delivery workers disabled")
		return
	}

	notificationWorkerInterval := 3 * time.Second
	notificationWorkerBatchSize := int32(20)
	if deps.Config != nil {
		notificationWorkerInterval = deps.Config.NotificationDelivery.WorkerInterval
		notificationWorkerBatchSize = deps.Config.NotificationDelivery.BatchSize
	}

	notificationEmailDeliveryWorker, err := worker.NewNotificationEmailDeliveryWorker(
		deps.NotificationDeliveryRepository,
		deps.EmailService,
		notificationWorkerInterval,
		notificationWorkerBatchSize,
		3,
		deps.Logger,
	)
	if err != nil {
		deps.Logger.Fatal("Failed to initialize notification email delivery worker: ", err)
	}

	go func() {
		deps.Logger.Info("Background notification email delivery worker started")

		if err := notificationEmailDeliveryWorker.Start(ctx); err != nil &&
			!errors.Is(err, context.Canceled) {
			deps.Logger.Error("Notification email delivery worker stopped with error: ", err)
		}
	}()

	if deps.SMSService != nil && deps.SMSService.Enabled() {
		notificationSMSDeliveryWorker, err := worker.NewNotificationSMSDeliveryWorker(
			deps.NotificationDeliveryRepository,
			deps.SMSService,
			notificationWorkerInterval,
			notificationWorkerBatchSize,
			3,
			deps.Logger,
		)
		if err != nil {
			deps.Logger.Fatal("Failed to initialize notification SMS delivery worker: ", err)
		}

		go func() {
			deps.Logger.Info("Background notification SMS delivery worker started")

			if err := notificationSMSDeliveryWorker.Start(ctx); err != nil &&
				!errors.Is(err, context.Canceled) {
				deps.Logger.Error("Notification SMS delivery worker stopped with error: ", err)
			}
		}()
	} else {
		deps.Logger.Info("Notification SMS delivery worker disabled")
	}
}
