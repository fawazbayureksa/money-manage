package jobs

import (
	"log"
	"my-api/services"
	"time"
)

type DebtJobs struct {
	debtService services.DebtService
}

func NewDebtJobs(debtService services.DebtService) *DebtJobs {
	return &DebtJobs{debtService: debtService}
}
func (j *DebtJobs) StartScheduler() {
	go j.runMonthlyInterestJob()
	go j.runDailyReminderJob()
}

func (j *DebtJobs) runMonthlyInterestJob() {
	for {
		now := time.Now()
		nextRun := time.Date(now.Year(), now.Month()+1, 1, 0, 5, 0, 0, now.Location())
		sleepDuration := time.Until(nextRun)

		log.Printf("[DebtJobs] Monthly interest job sleeping until %s", nextRun.Format(time.RFC3339))
		time.Sleep(sleepDuration)

		log.Printf("[DebtJobs] Running monthly interest processing...")
		if err := j.debtService.ProcessMonthlyInterest(); err != nil {
			log.Printf("[DebtJobs] Error processing monthly interest: %v", err)
		} else {
			log.Printf("[DebtJobs] Monthly interest processing completed successfully")
		}
	}
}

func (j *DebtJobs) runDailyReminderJob() {
	for {
		now := time.Now()
		// Next run at 08:00 today or tomorrow
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, now.Location())
		if now.After(nextRun) {
			nextRun = nextRun.Add(24 * time.Hour)
		}
		sleepDuration := time.Until(nextRun)

		log.Printf("[DebtJobs] Daily reminder job sleeping until %s", nextRun.Format(time.RFC3339))
		time.Sleep(sleepDuration)

		log.Printf("[DebtJobs] Sending payment reminders...")
		if err := j.debtService.SendPaymentReminders(); err != nil {
			log.Printf("[DebtJobs] Error sending reminders: %v", err)
		} else {
			log.Printf("[DebtJobs] Payment reminders sent successfully")
		}
	}
}
