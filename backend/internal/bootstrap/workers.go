package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/moh-sso-dashboard/core"
	reportschedulerfeature "github.com/moh-sso-dashboard/internal/adapters/modules/report_scheduler"
	"github.com/moh-sso-dashboard/internal/config"
	dataqualityfeature "github.com/moh-sso-dashboard/internal/features/data_quality"
	platformfeature "github.com/moh-sso-dashboard/internal/features/platform"
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
	Statuses                       map[string]*core.WorkerStatus
	PrimaryDB                      *sql.DB
	Config                         *config.Config
	ProcessRepository              processRepo.ProcessRepository
	ImportService                  *importSvc.Service
	FileStorage                    storage.Storage
	EmailRepository                emailRepo.EmailRepository
	SMTPService                    worker.EmailSender
	NotificationDeliveryRepository notificationDeliveryRepo.NotificationDeliveryRepository
	EmailService                   service.EmailService
	SMSService                     service.SMSService
	AuditService                   *service.AuditService
	Logger                         *logger.Logger
	DQADB                          *sql.DB
	DWHDB                          *sql.DB
}

func enabledDeliveryWorkers(cfg *config.Config) []string {
	if cfg.CoreOnly() {
		return []string{}
	}
	names := []string{"email-outbox"}
	if cfg == nil || cfg.NotificationDelivery.WorkerEnabled {
		names = append(names, "notification-email")
	}
	return names
}
func schedulerWorker(cfg *config.Config, services services, log *logger.Logger) *reportschedulerfeature.Worker {
	if !cfg.ModuleEnabled("report-scheduler") {
		return nil
	}
	return reportschedulerfeature.NewWorker(services.ReportScheduler, cfg.ReportSchedulerWorkerInterval, log.Error)
}

type workerTask struct {
	name string
	run  func(context.Context) error
}

// startWorkerGroup cancels and joins every started task before infrastructure
// cleanup. The stop callback is safe to retry after a shutdown deadline.
func startWorkerGroup(ctx context.Context, tasks []workerTask, logError func(...interface{})) core.Stop {
	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var failures []error
	for _, task := range tasks {
		task := task
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := task.run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				mu.Lock()
				failures = append(failures, fmt.Errorf("%s: %w", task.name, err))
				mu.Unlock()
				if logError != nil {
					logError("Background worker stopped: ", task.name, err)
				}
			}
		}()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	return func(shutdown context.Context) error {
		cancel()
		select {
		case <-done:
			mu.Lock()
			defer mu.Unlock()
			return errors.Join(failures...)
		case <-shutdown.Done():
			return shutdown.Err()
		}
	}
}
func startBackgroundWorkers(ctx context.Context, deps workerDependencies) (core.Stop, error) {
	if deps.Config.CoreOnly() {
		return startWorkerGroup(ctx, nil, nil), nil
	}
	heartbeat := func(name string) func(context.Context) {
		return func(ctx context.Context) {
			if deps.PrimaryDB == nil {
				return
			}
			timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if err := (&platformfeature.PostgresRepository{DB: deps.PrimaryDB}).RecordWorkerHeartbeat(timeout, name); err != nil {
				deps.Logger.Error("platform worker heartbeat unavailable", "worker", name)
			}
		}
	}
	// Prepare every worker before launching any goroutine so constructor failures
	// cannot leave partially started workers accessing released resources.
	tasks := []workerTask{}
	if deps.DQADB != nil && deps.DWHDB != nil {
		dqa := dataqualityfeature.NewScheduleWorker(dataqualityfeature.NewDQAStore(deps.DQADB), deps.DWHDB, time.Minute, deps.Logger.Error)
		tasks = append(tasks, workerTask{"DQA scheduled run", dqa.Start})
	}
	document, err := worker.NewDocumentWorker(deps.ProcessRepository, deps.ImportService, 3*time.Second, deps.FileStorage, deps.Logger)
	if err != nil {
		return nil, fmt.Errorf("document worker: %w", err)
	}
	tasks = append(tasks, workerTask{"document", document.Start})
	email, err := worker.NewEmailWorker(deps.EmailRepository, deps.SMTPService, 3*time.Second, 20, 3, deps.Logger)
	if err != nil {
		return nil, fmt.Errorf("email worker: %w", err)
	}
	email.SetStatus(deps.Statuses["email-outbox"])
	email.SetHeartbeat(heartbeat("email-outbox"))
	tasks = append(tasks, workerTask{"email-outbox", email.Start})
	if deps.Config == nil || deps.Config.NotificationDelivery.WorkerEnabled {
		interval := 3 * time.Second
		batch := int32(20)
		if deps.Config != nil {
			interval = deps.Config.NotificationDelivery.WorkerInterval
			batch = deps.Config.NotificationDelivery.BatchSize
		}
		notification, err := worker.NewNotificationEmailDeliveryWorker(deps.NotificationDeliveryRepository, deps.EmailService, interval, batch, 3, deps.Logger)
		if err != nil {
			return nil, fmt.Errorf("notification email worker: %w", err)
		}
		notification.SetStatus(deps.Statuses["notification-email"])
		notification.SetHeartbeat(heartbeat("notification-email"))
		tasks = append(tasks, workerTask{"notification-email", notification.Start})
		if deps.SMSService != nil && deps.SMSService.Enabled() {
			sms, err := worker.NewNotificationSMSDeliveryWorker(deps.NotificationDeliveryRepository, deps.SMSService, interval, batch, 3, deps.Logger, deps.AuditService)
			if err != nil {
				return nil, fmt.Errorf("notification SMS worker: %w", err)
			}
			tasks = append(tasks, workerTask{"notification-sms", sms.Start})
		}
	}
	return startWorkerGroup(ctx, tasks, deps.Logger.Error), nil
}
