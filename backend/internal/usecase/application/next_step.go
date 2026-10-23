package application

import "smartcredit/internal/domain"

var nextSteps = map[domain.Status]string{
	domain.StatusDraft:        "Заполните анкету и нажмите «Отправить заявку»",
	domain.StatusSubmitted:    "Проверяем полноту данных",
	domain.StatusEnriching:    "Собираем данные из ЕГРЮЛ, ФНС и БКИ",
	domain.StatusScoring:      "Оцениваем заявку",
	domain.StatusManualReview: "Заявку смотрит кредитный инспектор",
	domain.StatusCounterOffer: "Выберите один из предложенных вариантов",
	domain.StatusApproved:     "Подпишите договор",
	domain.StatusSigned:       "Зачисляем деньги на расчётный счёт",
	domain.StatusDisbursed:    "Деньги зачислены",
	domain.StatusRejected:     "Заявка отклонена, причина указана ниже",
	domain.StatusCancelled:    "Заявка отозвана",
}

func nextStep(s domain.Status, open *domain.InfoRequest) string {
	if s == domain.StatusInfoRequested && open != nil {
		return open.Description
	}
	return nextSteps[s]
}
