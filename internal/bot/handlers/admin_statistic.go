package handlers

import (
	"TgDonation/internal/database/models"
	"fmt"
	tele "gopkg.in/telebot.v4"
	"gorm.io/gorm"
	"log"
)

func (h *Handler) onStatistic(c tele.Context) error {
	h.tryDeleteMessage(c)
	if !h.isAdminFromDB(int(c.Sender().ID)) {
		return c.Send("❌ У вас нет доступа к этой функции.")
	}

	// ---- Общие метрики ----
	var totalDonations float64
	var donationsCount int64

	// 1) Сумма и кол-во пожертвований
	if err := h.DB.Model(&models.Donation{}).
		Select("COALESCE(SUM(amount),0)").
		Scan(&totalDonations).Error; err != nil {
		return c.Send("⚠️ Ошибка при загрузке статистики (сумма).")
	}
	if err := h.DB.Model(&models.Donation{}).Count(&donationsCount).Error; err != nil {
		return c.Send("⚠️ Ошибка при загрузке статистики (кол-во).")
	}

	// 2) Средний размер пожертвования
	var avgDonation float64
	if donationsCount > 0 {
		avgDonation = totalDonations / float64(donationsCount)
	}

	// 3) Сколько целей сейчас активно
	var activeGoalsCount int64
	if err := h.DB.Model(&models.Goal{}).
		Where("status = ?", "active").
		Count(&activeGoalsCount).Error; err != nil {
		return c.Send("⚠️ Ошибка при загрузке статистики (цели).")
	}

	// 4) Топ цель (по TargetSum)
	var topGoal models.Goal
	if err := h.DB.Order("target_sum DESC").First(&topGoal).Error; err == nil {
		// если есть хотя бы одна цель, err будет nil, иначе - record not found
	}

	var totalDonation models.TotalDonation
	err := h.DB.First(&totalDonation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Если записи нет, устанавливаем баланс в 0
			totalDonation.Total = 0
		} else {
			log.Printf("Ошибка при получении TotalDonation: %v", err)
			return c.Send("⚠️ Ошибка при загрузке баланса копилки.")
		}
	}
	total := totalDonation.Total

	// ---- Расширенные метрики ----

	// (A) Число уникальных доноров
	var uniqueDonorsCount int64
	if err := h.DB.Model(&models.Donation{}).
		Distinct("user_id").
		Count(&uniqueDonorsCount).Error; err != nil {
		return c.Send("⚠️ Ошибка при загрузке статистики (уникальные доноры).")
	}

	// (B) Топ-донор (кто пожертвовал всего больше всех)
	type DonorStat struct {
		UserID int
		Total  float64
	}
	var topDonor DonorStat
	if err := h.DB.Model(&models.Donation{}).
		Select("user_id, SUM(amount) as total").
		Group("user_id").
		Order("total DESC").
		Limit(1).
		Scan(&topDonor).Error; err != nil {
		// Если нет донатов — будет пустой
	}

	var topDonorUser models.User
	if topDonor.UserID != 0 {
		_ = h.DB.First(&topDonorUser, topDonor.UserID)
	}

	// (C) Распределение по банкам
	type BankStat struct {
		BankName string
		Total    float64
	}
	var bankStats []BankStat
	if err := h.DB.Model(&models.Donation{}).
		Select("bank_name as bank_name, SUM(amount) as total").
		Group("bank_name").
		Scan(&bankStats).Error; err != nil {
		// обработка ошибки
	}

	// ---- Формируем отчёт ----
	report := "<b>📊 <u>Статистика:</u></b>\n\n"
	// 1) Общие
	report += fmt.Sprintf("💰 <b>Общая сумма пожертвований:</b> %s ₽\n", formatFloatNoTrailingZeros(totalDonations))
	report += fmt.Sprintf("💳 <b>Баланс Копилки:</b> %s ₽\n", formatFloatNoTrailingZeros(total))
	report += fmt.Sprintf("📈 <b>Количество пожертвований:</b> %d\n", donationsCount)
	report += fmt.Sprintf("💲 <b>Средний размер пожертвования:</b> %s ₽\n", formatFloatNoTrailingZeros(avgDonation))
	report += fmt.Sprintf("🎯 <b>Активных целей:</b> %d\n\n", activeGoalsCount)

	// 2) Топ цель
	if topGoal.ID != 0 {
		report += fmt.Sprintf(
			"🏆 <b>Топ цель по целевой сумме:</b> %s\n   🎯 Целевая сумма: %s ₽\n\n",
			topGoal.Title,
			formatFloatNoTrailingZeros(topGoal.TargetSum),
		)
	} else {
		report += "🏆 <b>Топ цель:</b> Нет данных\n\n"
	}

	// 3) Уникальные доноры
	report += fmt.Sprintf("👥 <b>Уникальных доноров:</b> %d\n", uniqueDonorsCount)

	// 4) Топ-донор
	if topDonor.UserID != 0 {
		donorName := topDonorUser.Username
		if donorName == "" {
			donorName = fmt.Sprintf("UserID:%d", topDonorUser.TgID)
		}
		report += fmt.Sprintf("🤝 <b>Топ-донор:</b> %s (Сумма: %s ₽)\n\n", donorName, formatFloatNoTrailingZeros(topDonor.Total))
	} else {
		report += "🤝 <b>Топ-донор:</b> Нет данных\n\n"
	}

	// 5) Распределение по банкам
	if len(bankStats) > 0 {
		report += "<b>🏦 Распределение по банкам:</b>\n"
		for _, bs := range bankStats {
			report += fmt.Sprintf(" - %s: %s ₽\n", bs.BankName, formatFloatNoTrailingZeros(bs.Total))
		}
		report += "\n"
	}

	// Заключение
	back := backButton("back_to_panel")
	return c.Send(report, back, tele.ModeHTML)
}
