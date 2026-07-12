package main

import (
	"net/http"
	"strings"
)

const langCookieName = "avorimi_lang"

// currentLang читает выбранный язык из cookie; при отсутствии/некорректном значении — русский по умолчанию.
func currentLang(r *http.Request) string {
	cookie, err := r.Cookie(langCookieName)
	if err != nil || !validLangs[cookie.Value] {
		return "ru"
	}
	return cookie.Value
}

// setLangHandler сохраняет выбор языка в cookie и возвращает пользователя на исходную страницу —
// полноценный запрос-ответ без JS-подмены контента, поэтому переключение работает без сдвигов/глюков.
func setLangHandler(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if !validLangs[lang] {
		lang = "ru"
	}
	http.SetCookie(w, &http.Cookie{
		Name:   langCookieName,
		Value:  lang,
		Path:   "/",
		MaxAge: 365 * 24 * 60 * 60,
	})

	redirect := r.URL.Query().Get("redirect")
	if redirect == "" || redirect[0] != '/' || len(redirect) > 1 && redirect[1] == '/' {
		redirect = "/"
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

// translations хранит все статические строки интерфейса на трёх языках.
// Ключ верхнего уровня — язык (ru/kz/en), внутри — плоский словарь "ключ" -> "текст".
var translations = map[string]map[string]string{
	"ru": {
		"nav.clinics":       "Клиники рядом",
		"nav.catalog":       "Все специалисты",
		"nav.subscriptions": "Подписки",
		"nav.myBookings":    "Мои записи",
		"nav.login":         "Войти",
		"nav.register":      "Регистрация",

		"profile.subscription":  "Подписка",
		"profile.visitsLeft":    "Осталось визитов:",
		"profile.of":            "из",
		"profile.activeUntil":   "Действует до",
		"profile.noSub":         "Подписки пока нет.",
		"profile.getSub":        "Оформить подписку",
		"profile.notifications": "Уведомления",
		"profile.resultReady":   "— результат готов",
		"profile.account":       "Аккаунт",
		"profile.myResults":     "Мои анализы",
		"profile.subPlans":      "Тарифы подписки",
		"profile.nearbyClinics": "Клиники рядом",
		"profile.logout":        "Выйти",

		"footer.contacts":     "Контакты",
		"footer.support":      "Поддержка",
		"footer.email":        "Email",
		"footer.supportHours": "Круглосуточно (24/7)",
		"footer.social":       "Социальные сети",
		"footer.followUs":     "Следите за Avorimi Health",
		"footer.ecosystem":    "Экосистема Avorimi",
		"footer.healthDesc":   "поиск клиник, запись к врачам и медицинские подписки.",
		"footer.heartsDesc":   "благотворительный фонд.",
		"footer.floralDesc":   "цветы премиум-класса для особых моментов.",
		"footer.soon":         "Скоро",
		"footer.nav":          "Навигация",
		"footer.about":        "О сервисе",
		"footer.faq":          "Частые вопросы",
		"footer.contactsLink": "Контакты",
		"footer.bottom":       "© 2026 Avorimi. Все права защищены. Создано с ❤️ в Казахстане. Версия 1.0",

		"home.hero.title1":    "Найдите клинику",
		"home.hero.title2":    "рядом и запишитесь",
		"home.hero.title3":    "здесь и сейчас",
		"home.hero.subtitle":  "Avorimi Health — современная платформа для поиска клиник, записи к врачам и оформления медицинских подписок.",
		"home.hero.cta1":      "📍 Найти клинику рядом",
		"home.hero.cta2":      "Посмотреть подписки",
		"home.hero.card1":     "Клиники рядом с вами",
		"home.hero.card2":     "Выбор клиники и специалиста",
		"home.hero.card3":     "Ближайшее свободное время",
		"home.hero.card4":     "Оплата картой онлайн",
		"home.hero.card5":     "Проверенные клиники",
		"home.flow.title":     "Запишитесь к врачу всего за несколько минут",
		"home.flow.subtitle":  "Без звонков, очередей и долгого ожидания.",
		"home.flow.step1":     "Найдите клинику",
		"home.flow.step2":     "Выберите специалиста",
		"home.flow.step3":     "Выберите время",
		"home.flow.step4":     "Оплатите онлайн",
		"home.flow.step5":     "Посетите приём",
		"home.flow.cta":       "Попробовать сейчас",
		"home.why.title":      "Почему выбирают Avorimi",
		"home.why.1":          "Запись за 2 минуты",
		"home.why.2":          "Проверенные клиники",
		"home.why.3":          "Онлайн-оплата",
		"home.why.4":          "Все записи в одном месте",
		"home.why.5":          "Защита персональных данных",
		"home.why.6":          "Экономия благодаря подпискам",
		"home.why.7":          "Анализы в личном кабинете",
		"home.why.7sub":       "Результаты сохраняются в кабинете — уведомим, как только они будут готовы.",
		"home.specialties.title": "Популярные медицинские направления",
		"home.specialties.all":   "Все направления →",
		"home.map.title":  "Найдите клинику рядом с вами",
		"home.map.point1": "клиник поблизости",
		"home.map.point2": "докторов и специалистов",
		"home.map.point3": "Сотни свободных окон каждый день",
		"home.map.cta":    "Открыть карту",
		"home.doctors.title":    "Популярные специалисты",
		"home.doctors.nextSlot": "Ближайшая запись:",
		"home.doctors.tbd":      "Уточняйте время",
		"home.doctors.cta":      "Записаться",
		"home.stats.clinics":     "клиник",
		"home.stats.specialists": "специалистов",
		"home.stats.bookings":    "записей",
		"home.stats.rating":      "средняя оценка",
		"home.reviews.title": "Отзывы пользователей",
		"home.reviews.text1": "«Очень быстро записалась к врачу.»",
		"home.reviews.name1": "— Алия",
		"home.reviews.text2": "«Лучший сервис записи в клиники.»",
		"home.reviews.name2": "— Руслан",
		"home.reviews.text3": "«Теперь не приходится звонить.»",
		"home.reviews.name3": "— Максим",
		"home.subs.title":      "Подписки на визиты",
		"home.subs.all":        "Все планы →",
		"home.subs.intro":      "Выберите подписку, которая подходит именно вам. Подписка помогает экономить на медицинских услугах и получать быстрый доступ к специалистам.",
		"home.subs.visitsFree": "визитов в месяц бесплатно",
		"home.subs.cta":        "Оформить",
		"home.faq.title": "Частые вопросы",
		"home.faq.q1": "Как работает подписка?",
		"home.faq.a1": "Вы оплачиваете тариф раз в месяц и получаете определённое число бесплатных визитов к врачам или на процедуры — записываетесь как обычно и отмечаете оплату подпиской.",
		"home.faq.q2": "Можно ли вернуть деньги?",
		"home.faq.a2": "Да, если визит не состоялся не по вашей вине — напишите в поддержку, и мы разберёмся индивидуально.",
		"home.faq.q3": "Как записаться?",
		"home.faq.a3": "Найдите клинику рядом, выберите направление и врача, затем удобное время — запись подтверждается сразу.",
		"home.faq.q4": "Как отменить запись?",
		"home.faq.a4": "В разделе «Мои записи» можно посмотреть все свои визиты; для отмены обратитесь в поддержку — скоро это можно будет сделать в один клик.",
		"home.faq.q5": "Какие клиники подключены?",
		"home.faq.a5": "Мы сотрудничаем с проверенными многопрофильными клиниками вашего города — полный список смотрите в разделе «Клиники рядом».",
		"home.faq.q6": "Что входит в тариф?",
		"home.faq.a6": "Каждый тариф даёт определённое число бесплатных визитов в месяц к любым врачам или на процедуры из каталога.",
		"home.faq.q7": "Как изменить подписку?",
		"home.faq.a7": "Оформите новый тариф в разделе «Подписки» — он заменит текущий при следующей оплате.",
		"home.cta.title":    "Заботьтесь о здоровье вместе с Avorimi",
		"home.cta.subtitle": "Выберите клинику, найдите врача и запишитесь онлайн уже сегодня.",
		"home.cta.button":   "📍 Найти клинику",

		"clinics.titleNear":     "Клиники рядом с вами",
		"clinics.titleAll":      "Клиники",
		"clinics.byDistance":    "📍 По расстоянию",
		"clinics.locating":      "Уточняем ближайшие к вам клиники…",
		"clinics.directions":    "направлений",
		"clinics.kmAway":        "км от вас",

		"clinic.backAll":     "← Все клиники",
		"clinic.available":   "Доступные обследования",
		"clinic.doctorsFrom": "врача · от",
		"clinic.none":        "В этой клинике пока нет доступных направлений.",

		"clinicCategory.choose": "Выберите врача — время и цена у каждого своё.",

		"item.freeTime": "Свободное время",
		"item.noSlots":  "Свободных слотов пока нет, загляните позже.",
		"item.in":        "в",

		"catalog.title":          "Врачи и процедуры",
		"catalog.all":            "Все",
		"catalog.doctors":        "Врачи",
		"catalog.procedures":     "Процедуры",
		"catalog.allCategories":  "Все категории",
		"catalog.sort":           "Сортировка",
		"catalog.sortDefault":    "По умолчанию",
		"catalog.sortPriceAsc":   "Сначала дешевле",
		"catalog.sortPriceDesc":  "Сначала дороже",
		"catalog.sortRating":     "По рейтингу",
		"catalog.todayFilter":    "Есть время сегодня",
		"catalog.empty":          "Ничего не найдено по выбранным фильтрам.",

		"book.title":      "Оформление записи",
		"book.bookingFor": "Запись оформляется на",
		"book.useSub":     "У меня есть подписка — использовать бесплатный визит",
		"book.continue":   "Продолжить",

		"pay.title":      "Оплата приёма",
		"pay.recordNum":  "Запись №",
		"pay.amountDue":  "Сумма к оплате",
		"pay.cardNumber": "Номер карты",
		"pay.expiry":     "Срок действия",
		"pay.demoNotice": "Это демонстрационная оплата — реальное списание средств не производится.",
		"pay.payButton":  "Оплатить",

		"success.title":       "Вы записаны!",
		"success.paidBySub":   "Визит оплачен подпиской — с вас ничего не списано.",
		"success.paidOk":      "Оплата прошла успешно.",
		"success.recordFor":   "на имя",
		"success.freeBySub":   "Бесплатно по подписке",
		"success.paid":        "оплачено",
		"success.myBookings":  "Мои записи",
		"success.bookMore":    "Записаться ещё",

		"subscriptions.title":  "Подписки на визиты",
		"subscriptions.intro":  "Оплачивайте посещения врачей и процедуры одной подпиской — без отдельной оплаты за каждый визит.",
		"subscriptions.popular": "Популярный",
		"subscriptions.getSub": "Оформить подписку",

		"subscribe.title":   "Оформление подписки",
		"subscribe.perMonth": "визитов в месяц",

		"subscribeSuccess.title":      "Подписка активна!",
		"subscribeSuccess.paidOk":     "Оплата прошла успешно.",
		"subscribeSuccess.visitsAvail": "Доступно визитов:",
		"subscribeSuccess.of":         "из",
		"subscribeSuccess.validUntil": "Действует до",
		"subscribeSuccess.bookVisit":  "Записаться на визит",

		"account.title":         "Мои записи",
		"account.noSub":         "Подписки нет",
		"account.getSubToFree":  "Оформите подписку, чтобы получать визиты бесплатно.",
		"account.getSub":        "Оформить",
		"account.history":       "История записей",
		"account.pendingPayment": "ожидает оплаты",
		"account.noBookings":    "Записей пока нет.",

		"results.title":      "Мои анализы",
		"results.intro":      "Результаты УЗИ, анализов крови, ЭКГ и других обследований сохраняются здесь. Как только результат готов, мы показываем уведомление в профиле.",
		"results.ready":      "✅ Готово",
		"results.pending":    "⏳ Ожидается",
		"results.pendingNote": "Результат появится здесь после обработки.",
		"results.empty":      "Пока нет записей на анализы или диагностику. Найдите нужное обследование в",
		"results.catalog":    "каталоге",

		"login.title":         "Вход",
		"login.phone":         "Телефон",
		"login.password":      "Пароль",
		"login.submit":        "Войти",
		"login.noAccount":     "Нет аккаунта?",
		"login.registerLink":  "Зарегистрироваться",

		"register.title":          "Регистрация",
		"register.intro":          "Аккаунт нужен, чтобы записываться на приём, оплачивать визиты и пользоваться подпиской.",
		"register.fullName":       "ФИО",
		"register.iin":            "ИИН",
		"register.idNotice":       "ℹ️ На приёме в клинике могут попросить показать удостоверение личности для сверки данных.",
		"register.phone":          "Телефон",
		"register.password":       "Пароль",
		"register.notShorter":     "Не короче 6 символов",
		"register.confirmPassword": "Повторите пароль",
		"register.submit":         "Зарегистрироваться",
		"register.haveAccount":    "Уже есть аккаунт?",
		"register.loginLink":      "Войти",

		"err.fillFullName":       "Заполните ФИО",
		"err.iinInvalid":         "ИИН должен содержать ровно 12 цифр",
		"err.passwordTooShort":   "Пароль должен быть не короче 6 символов",
		"err.passwordMismatch":   "Пароли не совпадают",
		"err.phoneTaken":         "Пользователь с таким телефоном уже зарегистрирован",
		"err.phoneDigits":        "Введите 10 цифр номера телефона",
		"err.invalidLogin":       "Неверный телефон или пароль",
		"err.slotUnavailable":    "Это время уже занято, выберите другое",
		"err.itemNotFound":       "Услуга не найдена",
		"err.noActiveSubscription": "Нет активной подписки с доступными визитами",
		"err.planNotFound":       "План не найден",
		"err.generic":            "Что-то пошло не так, попробуйте ещё раз",
	},
	"kz": {
		"nav.clinics":       "Жақын клиникалар",
		"nav.catalog":       "Барлық мамандар",
		"nav.subscriptions": "Жазылымдар",
		"nav.myBookings":    "Жазбаларым",
		"nav.login":         "Кіру",
		"nav.register":      "Тіркелу",

		"profile.subscription":  "Жазылым",
		"profile.visitsLeft":    "Қалған қабылдаулар:",
		"profile.of":            "барлығы",
		"profile.activeUntil":   "Мерзімі",
		"profile.noSub":         "Жазылым әлі жоқ.",
		"profile.getSub":        "Жазылым рәсімдеу",
		"profile.notifications": "Хабарламалар",
		"profile.resultReady":   "— нәтиже дайын",
		"profile.account":       "Аккаунт",
		"profile.myResults":     "Менің талдауларым",
		"profile.subPlans":      "Жазылым тарифтері",
		"profile.nearbyClinics": "Жақын клиникалар",
		"profile.logout":        "Шығу",

		"footer.contacts":     "Байланыс",
		"footer.support":      "Қолдау қызметі",
		"footer.email":        "Email",
		"footer.supportHours": "Тәулік бойы (24/7)",
		"footer.social":       "Әлеуметтік желілер",
		"footer.followUs":     "Avorimi Health-ті бақылаңыз",
		"footer.ecosystem":    "Avorimi экожүйесі",
		"footer.healthDesc":   "клиника іздеу, дәрігерге жазылу және медициналық жазылымдар.",
		"footer.heartsDesc":   "қайырымдылық қоры.",
		"footer.floralDesc":   "ерекше сәттерге арналған премиум-класс гүлдер.",
		"footer.soon":         "Жақында",
		"footer.nav":          "Шарлау",
		"footer.about":        "Қызмет туралы",
		"footer.faq":          "Жиі қойылатын сұрақтар",
		"footer.contactsLink": "Байланыс",
		"footer.bottom":       "© 2026 Avorimi. Барлық құқықтар қорғалған. Қазақстанда ❤️ жасалды. Нұсқа 1.0",

		"home.hero.title1":   "Жақын клиниканы",
		"home.hero.title2":   "тауып, жазылыңыз",
		"home.hero.title3":   "дәл қазір",
		"home.hero.subtitle": "Avorimi Health — клиника іздеуге, дәрігерге жазылуға және медициналық жазылым рәсімдеуге арналған заманауи платформа.",
		"home.hero.cta1":     "📍 Жақын клиниканы табу",
		"home.hero.cta2":     "Жазылымдарды қарау",
		"home.hero.card1":    "Сізге жақын клиникалар",
		"home.hero.card2":    "Клиника мен маманды таңдау",
		"home.hero.card3":    "Ең жақын бос уақыт",
		"home.hero.card4":    "Картамен онлайн төлем",
		"home.hero.card5":    "Тексерілген клиникалар",
		"home.flow.title":    "Дәрігерге бірнеше минутта жазылыңыз",
		"home.flow.subtitle": "Қоңыраусыз, кезексіз және ұзақ күтусіз.",
		"home.flow.step1":    "Клиника табыңыз",
		"home.flow.step2":    "Маманды таңдаңыз",
		"home.flow.step3":    "Уақытты таңдаңыз",
		"home.flow.step4":    "Онлайн төлеңіз",
		"home.flow.step5":    "Қабылдауға келіңіз",
		"home.flow.cta":      "Қазір байқап көру",
		"home.why.title":     "Avorimi неге таңдайды",
		"home.why.1":         "2 минутта жазылу",
		"home.why.2":         "Тексерілген клиникалар",
		"home.why.3":         "Онлайн төлем",
		"home.why.4":         "Барлық жазбалар бір жерде",
		"home.why.5":         "Жеке деректерді қорғау",
		"home.why.6":         "Жазылым арқылы үнемдеу",
		"home.why.7":         "Талдаулар жеке кабинетте",
		"home.why.7sub":      "Нәтижелер кабинетте сақталады — дайын болғанда хабарлаймыз.",
		"home.specialties.title": "Танымал медициналық бағыттар",
		"home.specialties.all":   "Барлық бағыттар →",
		"home.map.title":  "Сізге жақын клиниканы табыңыз",
		"home.map.point1": "жақын клиника",
		"home.map.point2": "дәрігер мен маман",
		"home.map.point3": "Күн сайын жүздеген бос уақыт",
		"home.map.cta":    "Картаны ашу",
		"home.doctors.title":    "Танымал мамандар",
		"home.doctors.nextSlot": "Ең жақын жазылу:",
		"home.doctors.tbd":      "Уақытын нақтылаңыз",
		"home.doctors.cta":      "Жазылу",
		"home.stats.clinics":     "клиника",
		"home.stats.specialists": "маман",
		"home.stats.bookings":    "жазба",
		"home.stats.rating":      "орташа баға",
		"home.reviews.title": "Пайдаланушы пікірлері",
		"home.reviews.text1": "«Дәрігерге өте тез жазылдым.»",
		"home.reviews.name1": "— Әлия",
		"home.reviews.text2": "«Клиникаға жазылудың ең жақсы сервисі.»",
		"home.reviews.name2": "— Руслан",
		"home.reviews.text3": "«Енді қоңырау шалудың қажеті жоқ.»",
		"home.reviews.name3": "— Максим",
		"home.subs.title":      "Қабылдауларға арналған жазылымдар",
		"home.subs.all":        "Барлық тарифтер →",
		"home.subs.intro":      "Өзіңізге қолайлы жазылымды таңдаңыз. Жазылым медициналық қызметтерге үнемдеуге және мамандарға жылдам қол жеткізуге көмектеседі.",
		"home.subs.visitsFree": "айына тегін қабылдау",
		"home.subs.cta":        "Рәсімдеу",
		"home.faq.title": "Жиі қойылатын сұрақтар",
		"home.faq.q1": "Жазылым қалай жұмыс істейді?",
		"home.faq.a1": "Айына бір рет тариф төлейсіз және дәрігерлерге немесе процедураларға белгілі бір санда тегін қабылдау аласыз — әдеттегідей жазыласыз да, төлемді жазылыммен белгілейсіз.",
		"home.faq.q2": "Ақшаны қайтаруға бола ма?",
		"home.faq.a2": "Иә, қабылдау сіздің кінәңізден болмай өтпей қалса — қолдау қызметіне жазыңыз, жеке қарастырамыз.",
		"home.faq.q3": "Қалай жазылуға болады?",
		"home.faq.a3": "Жақын клиниканы тауып, бағыт пен дәрігерді, содан кейін қолайлы уақытты таңдаңыз — жазылу бірден расталады.",
		"home.faq.q4": "Жазбаны қалай болдырмауға болады?",
		"home.faq.a4": "«Жазбаларым» бөлімінде барлық қабылдауларыңызды көре аласыз; болдырмау үшін қолдау қызметіне хабарласыңыз — жақында бұны бір батырмамен жасауға болады.",
		"home.faq.q5": "Қандай клиникалар қосылған?",
		"home.faq.a5": "Біз қалаңыздың тексерілген көппрофильді клиникаларымен жұмыс істейміз — толық тізімді «Жақын клиникалар» бөлімінен қараңыз.",
		"home.faq.q6": "Тарифке не кіреді?",
		"home.faq.a6": "Әр тариф каталогтағы кез келген дәрігерлерге немесе процедураларға айына белгілі бір санда тегін қабылдау береді.",
		"home.faq.q7": "Жазылымды қалай өзгертуге болады?",
		"home.faq.a7": "«Жазылымдар» бөлімінде жаңа тариф рәсімдеңіз — ол келесі төлемде ағымдағысын алмастырады.",
		"home.cta.title":    "Денсаулығыңызды Avorimi-мен бірге қорғаңыз",
		"home.cta.subtitle": "Клиниканы таңдаңыз, дәрігерді тауып, бүгін онлайн жазылыңыз.",
		"home.cta.button":   "📍 Клиника табу",

		"clinics.titleNear":  "Сізге жақын клиникалар",
		"clinics.titleAll":   "Клиникалар",
		"clinics.byDistance": "📍 Қашықтық бойынша",
		"clinics.locating":   "Сізге жақын клиникаларды анықтап жатырмыз…",
		"clinics.directions": "бағыт",
		"clinics.kmAway":     "км қашықтықта",

		"clinic.backAll":     "← Барлық клиникалар",
		"clinic.available":   "Қолжетімді тексерулер",
		"clinic.doctorsFrom": "дәрігер · бастап",
		"clinic.none":        "Бұл клиникада әзірге қолжетімді бағыттар жоқ.",

		"clinicCategory.choose": "Дәрігерді таңдаңыз — әрқайсысының уақыты мен бағасы әртүрлі.",

		"item.freeTime": "Бос уақыт",
		"item.noSlots":  "Әзірге бос уақыт жоқ, кейінірек қараңыз.",
		"item.in":       "мына жерде:",

		"catalog.title":         "Дәрігерлер мен процедуралар",
		"catalog.all":           "Барлығы",
		"catalog.doctors":       "Дәрігерлер",
		"catalog.procedures":    "Процедуралар",
		"catalog.allCategories": "Барлық санаттар",
		"catalog.sort":          "Сұрыптау",
		"catalog.sortDefault":   "Әдепкі бойынша",
		"catalog.sortPriceAsc":  "Алдымен арзан",
		"catalog.sortPriceDesc": "Алдымен қымбат",
		"catalog.sortRating":    "Рейтинг бойынша",
		"catalog.todayFilter":   "Бүгін уақыт бар",
		"catalog.empty":         "Таңдалған сүзгілер бойынша ештеңе табылмады.",

		"book.title":      "Жазылуды рәсімдеу",
		"book.bookingFor": "Жазылу мына атпен рәсімделеді:",
		"book.useSub":     "Менде жазылым бар — тегін қабылдауды пайдалану",
		"book.continue":   "Жалғастыру",

		"pay.title":      "Қабылдауды төлеу",
		"pay.recordNum":  "Жазба №",
		"pay.amountDue":  "Төленетін сома",
		"pay.cardNumber": "Карта нөірі",
		"pay.expiry":     "Жарамдылық мерзімі",
		"pay.demoNotice": "Бұл демо-төлем — нақты ақша есептен шығарылмайды.",
		"pay.payButton":  "Төлеу",

		"success.title":      "Сіз жазылдыңыз!",
		"success.paidBySub":  "Қабылдау жазылым арқылы төленді — сізден ештеңе алынған жоқ.",
		"success.paidOk":     "Төлем сәтті өтті.",
		"success.recordFor":  "мына атқа:",
		"success.freeBySub":  "Жазылым бойынша тегін",
		"success.paid":       "төленді",
		"success.myBookings": "Жазбаларым",
		"success.bookMore":   "Тағы жазылу",

		"subscriptions.title":   "Қабылдауларға арналған жазылымдар",
		"subscriptions.intro":   "Дәрігерге баруды және процедураларды бір жазылыммен төлеңіз — әр қабылдау үшін бөлек төлемсіз.",
		"subscriptions.popular": "Танымал",
		"subscriptions.getSub":  "Жазылым рәсімдеу",

		"subscribe.title":    "Жазылымды рәсімдеу",
		"subscribe.perMonth": "айына қабылдау",

		"subscribeSuccess.title":       "Жазылым белсенді!",
		"subscribeSuccess.paidOk":      "Төлем сәтті өтті.",
		"subscribeSuccess.visitsAvail": "Қолжетімді қабылдау:",
		"subscribeSuccess.of":          "барлығы",
		"subscribeSuccess.validUntil":  "Мерзімі",
		"subscribeSuccess.bookVisit":   "Қабылдауға жазылу",

		"account.title":          "Жазбаларым",
		"account.noSub":          "Жазылым жоқ",
		"account.getSubToFree":   "Тегін қабылдау алу үшін жазылым рәсімдеңіз.",
		"account.getSub":         "Рәсімдеу",
		"account.history":        "Жазбалар тарихы",
		"account.pendingPayment": "төлемді күтуде",
		"account.noBookings":     "Әзірге жазбалар жоқ.",

		"results.title":       "Менің талдауларым",
		"results.intro":       "УЗИ, қан талдаулары, ЭКГ және басқа тексерулердің нәтижелері осында сақталады. Нәтиже дайын болғанда, профильде хабарлама көрсетеміз.",
		"results.ready":       "✅ Дайын",
		"results.pending":     "⏳ Күтілуде",
		"results.pendingNote": "Нәтиже өңдеуден кейін осында пайда болады.",
		"results.empty":       "Талдау немесе диагностикаға жазба әзірге жоқ. Қажетті тексеруді",
		"results.catalog":     "каталогтан",

		"login.title":        "Кіру",
		"login.phone":        "Телефон",
		"login.password":     "Құпия сөз",
		"login.submit":       "Кіру",
		"login.noAccount":    "Аккаунтыңыз жоқ па?",
		"login.registerLink": "Тіркелу",

		"register.title":           "Тіркелу",
		"register.intro":           "Қабылдауға жазылу, қабылдауларды төлеу және жазылымды пайдалану үшін аккаунт қажет.",
		"register.fullName":        "Аты-жөні",
		"register.iin":             "ЖСН",
		"register.idNotice":        "ℹ️ Клиникада қабылдау кезінде деректерді салыстыру үшін жеке куәлікті сұрауы мүмкін.",
		"register.phone":           "Телефон",
		"register.password":        "Құпия сөз",
		"register.notShorter":      "6 таңбадан қысқа болмауы керек",
		"register.confirmPassword": "Құпия сөзді қайталаңыз",
		"register.submit":          "Тіркелу",
		"register.haveAccount":     "Аккаунтыңыз бар ма?",
		"register.loginLink":       "Кіру",

		"err.fillFullName":         "Аты-жөнін толтырыңыз",
		"err.iinInvalid":           "ЖСН дәл 12 саннан тұруы керек",
		"err.passwordTooShort":     "Құпия сөз 6 таңбадан қысқа болмауы керек",
		"err.passwordMismatch":     "Құпия сөздер сәйкес келмейді",
		"err.phoneTaken":           "Бұл телефон нөірімен пайдаланушы тіркелген",
		"err.phoneDigits":          "Телефон нөмірінің 10 санын енгізіңіз",
		"err.invalidLogin":         "Телефон немесе құпия сөз қате",
		"err.slotUnavailable":      "Бұл уақыт бос емес, басқасын таңдаңыз",
		"err.itemNotFound":         "Қызмет табылмады",
		"err.noActiveSubscription": "Қолжетімді қабылдауы бар белсенді жазылым жоқ",
		"err.planNotFound":         "Тариф табылмады",
		"err.generic":              "Бірдеңе дұрыс болмады, қайталап көріңіз",
	},
	"en": {
		"nav.clinics":       "Nearby clinics",
		"nav.catalog":       "All specialists",
		"nav.subscriptions": "Subscriptions",
		"nav.myBookings":    "My bookings",
		"nav.login":         "Log in",
		"nav.register":      "Sign up",

		"profile.subscription":  "Subscription",
		"profile.visitsLeft":    "Visits left:",
		"profile.of":            "of",
		"profile.activeUntil":   "Valid until",
		"profile.noSub":         "No subscription yet.",
		"profile.getSub":        "Get a subscription",
		"profile.notifications": "Notifications",
		"profile.resultReady":   "— result ready",
		"profile.account":       "Account",
		"profile.myResults":     "My test results",
		"profile.subPlans":      "Subscription plans",
		"profile.nearbyClinics": "Nearby clinics",
		"profile.logout":        "Log out",

		"footer.contacts":     "Contacts",
		"footer.support":      "Support",
		"footer.email":        "Email",
		"footer.supportHours": "Available 24/7",
		"footer.social":       "Social media",
		"footer.followUs":     "Follow Avorimi Health",
		"footer.ecosystem":    "Avorimi ecosystem",
		"footer.healthDesc":   "clinic search, doctor booking, and medical subscriptions.",
		"footer.heartsDesc":   "charitable foundation.",
		"footer.floralDesc":   "premium flowers for special moments.",
		"footer.soon":         "Coming soon",
		"footer.nav":          "Navigation",
		"footer.about":        "About",
		"footer.faq":          "FAQ",
		"footer.contactsLink": "Contacts",
		"footer.bottom":       "© 2026 Avorimi. All rights reserved. Made with ❤️ in Kazakhstan. Version 1.0",

		"home.hero.title1":   "Find a clinic",
		"home.hero.title2":   "nearby and book",
		"home.hero.title3":   "right now",
		"home.hero.subtitle": "Avorimi Health is a modern platform for finding clinics, booking doctors, and managing medical subscriptions.",
		"home.hero.cta1":     "📍 Find a clinic nearby",
		"home.hero.cta2":     "View subscriptions",
		"home.hero.card1":    "Clinics near you",
		"home.hero.card2":    "Choose a clinic and specialist",
		"home.hero.card3":    "Earliest available time",
		"home.hero.card4":    "Pay online by card",
		"home.hero.card5":    "Verified clinics",
		"home.flow.title":    "Book a doctor in just a few minutes",
		"home.flow.subtitle": "No calls, no queues, no long waits.",
		"home.flow.step1":    "Find a clinic",
		"home.flow.step2":    "Choose a specialist",
		"home.flow.step3":    "Pick a time",
		"home.flow.step4":    "Pay online",
		"home.flow.step5":    "Attend your visit",
		"home.flow.cta":      "Try it now",
		"home.why.title":     "Why choose Avorimi",
		"home.why.1":         "Book in 2 minutes",
		"home.why.2":         "Verified clinics",
		"home.why.3":         "Online payment",
		"home.why.4":         "All bookings in one place",
		"home.why.5":         "Personal data protection",
		"home.why.6":         "Save with subscriptions",
		"home.why.7":         "Test results in your account",
		"home.why.7sub":      "Results are saved in your account — we'll notify you as soon as they're ready.",
		"home.specialties.title": "Popular medical specialties",
		"home.specialties.all":   "All specialties →",
		"home.map.title":  "Find a clinic near you",
		"home.map.point1": "clinics nearby",
		"home.map.point2": "doctors and specialists",
		"home.map.point3": "Hundreds of open slots every day",
		"home.map.cta":    "Open the map",
		"home.doctors.title":    "Popular specialists",
		"home.doctors.nextSlot": "Next available:",
		"home.doctors.tbd":      "Check availability",
		"home.doctors.cta":      "Book now",
		"home.stats.clinics":     "clinics",
		"home.stats.specialists": "specialists",
		"home.stats.bookings":    "bookings",
		"home.stats.rating":      "average rating",
		"home.reviews.title": "User reviews",
		"home.reviews.text1": "\"I booked a doctor incredibly fast.\"",
		"home.reviews.name1": "— Aliya",
		"home.reviews.text2": "\"The best clinic booking service.\"",
		"home.reviews.name2": "— Ruslan",
		"home.reviews.text3": "\"No more phone calls needed.\"",
		"home.reviews.name3": "— Maxim",
		"home.subs.title":      "Visit subscriptions",
		"home.subs.all":        "All plans →",
		"home.subs.intro":      "Choose the subscription that fits you. A subscription helps you save on medical services and get fast access to specialists.",
		"home.subs.visitsFree": "free visits per month",
		"home.subs.cta":        "Get plan",
		"home.faq.title": "Frequently asked questions",
		"home.faq.q1": "How does the subscription work?",
		"home.faq.a1": "You pay for a plan once a month and get a set number of free visits to doctors or procedures — book as usual and mark the payment as covered by your subscription.",
		"home.faq.q2": "Can I get a refund?",
		"home.faq.a2": "Yes, if a visit didn't happen through no fault of yours — contact support and we'll sort it out individually.",
		"home.faq.q3": "How do I book an appointment?",
		"home.faq.a3": "Find a nearby clinic, choose a specialty and doctor, then a convenient time — the booking is confirmed instantly.",
		"home.faq.q4": "How do I cancel a booking?",
		"home.faq.a4": "You can view all your visits under \"My bookings\"; to cancel, contact support — one-click cancellation is coming soon.",
		"home.faq.q5": "Which clinics are connected?",
		"home.faq.a5": "We partner with verified multidisciplinary clinics in your city — see the full list under \"Nearby clinics\".",
		"home.faq.q6": "What's included in a plan?",
		"home.faq.a6": "Each plan gives you a set number of free monthly visits to any doctors or procedures in the catalog.",
		"home.faq.q7": "How do I change my subscription?",
		"home.faq.a7": "Get a new plan under \"Subscriptions\" — it will replace your current one at the next payment.",
		"home.cta.title":    "Take care of your health with Avorimi",
		"home.cta.subtitle": "Choose a clinic, find a doctor, and book online today.",
		"home.cta.button":   "📍 Find a clinic",

		"clinics.titleNear":  "Clinics near you",
		"clinics.titleAll":   "Clinics",
		"clinics.byDistance": "📍 By distance",
		"clinics.locating":   "Finding the clinics closest to you…",
		"clinics.directions": "specialties",
		"clinics.kmAway":     "km away",

		"clinic.backAll":     "← All clinics",
		"clinic.available":   "Available specialties",
		"clinic.doctorsFrom": "doctors · from",
		"clinic.none":        "This clinic has no available specialties yet.",

		"clinicCategory.choose": "Choose a doctor — each has their own time and price.",

		"item.freeTime": "Available times",
		"item.noSlots":  "No available slots yet, check back later.",
		"item.in":       "at",

		"catalog.title":         "Doctors and procedures",
		"catalog.all":           "All",
		"catalog.doctors":       "Doctors",
		"catalog.procedures":    "Procedures",
		"catalog.allCategories": "All categories",
		"catalog.sort":          "Sort by",
		"catalog.sortDefault":   "Default",
		"catalog.sortPriceAsc":  "Price: low to high",
		"catalog.sortPriceDesc": "Price: high to low",
		"catalog.sortRating":    "By rating",
		"catalog.todayFilter":   "Available today",
		"catalog.empty":         "Nothing found for the selected filters.",

		"book.title":      "Confirm booking",
		"book.bookingFor": "Booking for",
		"book.useSub":     "I have a subscription — use a free visit",
		"book.continue":   "Continue",

		"pay.title":      "Payment",
		"pay.recordNum":  "Booking #",
		"pay.amountDue":  "Amount due",
		"pay.cardNumber": "Card number",
		"pay.expiry":     "Expiry date",
		"pay.demoNotice": "This is a demo payment — no real money is charged.",
		"pay.payButton":  "Pay",

		"success.title":      "You're booked!",
		"success.paidBySub":  "This visit was covered by your subscription — nothing was charged.",
		"success.paidOk":     "Payment successful.",
		"success.recordFor":  "for",
		"success.freeBySub":  "Free with subscription",
		"success.paid":       "paid",
		"success.myBookings": "My bookings",
		"success.bookMore":   "Book another",

		"subscriptions.title":   "Visit subscriptions",
		"subscriptions.intro":   "Pay for doctor visits and procedures with a single subscription — no separate payment for each visit.",
		"subscriptions.popular": "Popular",
		"subscriptions.getSub":  "Get subscription",

		"subscribe.title":    "Get subscription",
		"subscribe.perMonth": "visits per month",

		"subscribeSuccess.title":       "Subscription active!",
		"subscribeSuccess.paidOk":      "Payment successful.",
		"subscribeSuccess.visitsAvail": "Visits available:",
		"subscribeSuccess.of":          "of",
		"subscribeSuccess.validUntil":  "Valid until",
		"subscribeSuccess.bookVisit":   "Book a visit",

		"account.title":          "My bookings",
		"account.noSub":          "No subscription",
		"account.getSubToFree":   "Get a subscription to receive free visits.",
		"account.getSub":         "Get plan",
		"account.history":        "Booking history",
		"account.pendingPayment": "payment pending",
		"account.noBookings":     "No bookings yet.",

		"results.title":       "My test results",
		"results.intro":       "Results of ultrasound, blood tests, ECGs and other exams are saved here. As soon as a result is ready, we show a notification in your profile.",
		"results.ready":       "✅ Ready",
		"results.pending":     "⏳ Pending",
		"results.pendingNote": "The result will appear here once processed.",
		"results.empty":       "No lab or diagnostic bookings yet. Find what you need in the",
		"results.catalog":     "catalog",

		"login.title":        "Log in",
		"login.phone":        "Phone",
		"login.password":     "Password",
		"login.submit":       "Log in",
		"login.noAccount":    "No account?",
		"login.registerLink": "Sign up",

		"register.title":           "Sign up",
		"register.intro":           "An account is needed to book appointments, pay for visits, and use a subscription.",
		"register.fullName":        "Full name",
		"register.iin":             "National ID (IIN)",
		"register.idNotice":        "ℹ️ At the clinic you may be asked to show an ID to verify your details.",
		"register.phone":           "Phone",
		"register.password":        "Password",
		"register.notShorter":      "At least 6 characters",
		"register.confirmPassword": "Confirm password",
		"register.submit":          "Sign up",
		"register.haveAccount":     "Already have an account?",
		"register.loginLink":       "Log in",

		"err.fillFullName":         "Please fill in your full name",
		"err.iinInvalid":           "IIN must contain exactly 12 digits",
		"err.passwordTooShort":     "Password must be at least 6 characters",
		"err.passwordMismatch":     "Passwords don't match",
		"err.phoneTaken":           "A user with this phone number is already registered",
		"err.phoneDigits":          "Enter 10 digits of your phone number",
		"err.invalidLogin":         "Incorrect phone number or password",
		"err.slotUnavailable":      "This time is already taken, please choose another",
		"err.itemNotFound":         "Service not found",
		"err.noActiveSubscription": "No active subscription with visits available",
		"err.planNotFound":         "Plan not found",
		"err.generic":              "Something went wrong, please try again",
	},
}

// specialtyNames переводит канонические (русские) названия направлений —
// сама ссылка/фильтр всегда использует русское значение как ключ данных,
// переводится только то, что видит пользователь.
var specialtyNames = map[string]map[string]string{
	"kz": {
		"Терапевт":                    "Терапевт",
		"Кардиолог":                   "Кардиолог",
		"Дерматолог":                  "Дерматолог",
		"Невролог":                    "Невролог",
		"Стоматолог":                  "Стоматолог",
		"Уролог":                      "Уролог",
		"Гинеколог":                   "Гинеколог",
		"Педиатр":                     "Педиатр",
		"Офтальмолог":                 "Офтальмолог",
		"Отоларинголог (ЛОР)":        "Отоларинголог (ЛОР)",
		"Эндокринолог":                "Эндокринолог",
		"Гастроэнтеролог":             "Гастроэнтеролог",
		"Аллерголог":                  "Аллерголог",
		"Психотерапевт":               "Психотерапевт",
		"Хирург":                      "Хирург",
		"Ортопед":                     "Ортопед",
		"Флеболог":                    "Флеболог",
		"Ревматолог":                  "Ревматолог",
		"Маммолог":                    "Маммолог",
		"Диетолог":                    "Диетолог",
		"УЗИ брюшной полости":         "Іш қуысының УДЗ",
		"УЗИ малого таза":             "Кіші жамбас УДЗ",
		"ЭКГ с расшифровкой":          "Түсіндірмесі бар ЭКГ",
		"Общий анализ крови":          "Жалпы қан талдауы",
		"Биохимический анализ крови": "Қанның биохимиялық талдауы",
		"Массаж спины":                "Арқа массажы",
		"Рентген":                     "Рентген",
		"Флюорография":                "Флюорография",
		"Вакцинация":                  "Вакцинация",
	},
	"en": {
		"Терапевт":                    "General practitioner",
		"Кардиолог":                   "Cardiologist",
		"Дерматолог":                  "Dermatologist",
		"Невролог":                    "Neurologist",
		"Стоматолог":                  "Dentist",
		"Уролог":                      "Urologist",
		"Гинеколог":                   "Gynecologist",
		"Педиатр":                     "Pediatrician",
		"Офтальмолог":                 "Ophthalmologist",
		"Отоларинголог (ЛОР)":        "ENT specialist",
		"Эндокринолог":                "Endocrinologist",
		"Гастроэнтеролог":             "Gastroenterologist",
		"Аллерголог":                  "Allergist",
		"Психотерапевт":               "Psychotherapist",
		"Хирург":                      "Surgeon",
		"Ортопед":                     "Orthopedist",
		"Флеболог":                    "Phlebologist",
		"Ревматолог":                  "Rheumatologist",
		"Маммолог":                    "Mammologist",
		"Диетолог":                    "Nutritionist",
		"УЗИ брюшной полости":         "Abdominal ultrasound",
		"УЗИ малого таза":             "Pelvic ultrasound",
		"ЭКГ с расшифровкой":          "ECG with report",
		"Общий анализ крови":          "Complete blood count",
		"Биохимический анализ крови": "Blood chemistry panel",
		"Массаж спины":                "Back massage",
		"Рентген":                     "X-ray",
		"Флюорография":                "Fluorography",
		"Вакцинация":                  "Vaccination",
	},
}

// labResultTexts — мок-текст результата анализа/диагностики по направлению, на трёх языках.
var labResultTexts = map[string]map[string]string{
	"ru": {
		"Общий анализ крови":         "Показатели крови в пределах нормы.",
		"Биохимический анализ крови": "Показатели крови в пределах нормы.",
		"УЗИ брюшной полости":        "Патологических изменений не выявлено.",
		"УЗИ малого таза":            "Патологических изменений не выявлено.",
		"ЭКГ с расшифровкой":         "Ритм синусовый, без отклонений.",
		"Рентген":                    "Без признаков патологии.",
		"Флюорография":               "Без признаков патологии.",
	},
	"kz": {
		"Общий анализ крови":         "Қан көрсеткіштері қалыпты шекте.",
		"Биохимический анализ крови": "Қан көрсеткіштері қалыпты шекте.",
		"УЗИ брюшной полости":        "Патологиялық өзгерістер анықталған жоқ.",
		"УЗИ малого таза":            "Патологиялық өзгерістер анықталған жоқ.",
		"ЭКГ с расшифровкой":         "Синус ырғағы, ауытқусыз.",
		"Рентген":                    "Патология белгілері жоқ.",
		"Флюорография":               "Патология белгілері жоқ.",
	},
	"en": {
		"Общий анализ крови":         "Blood values are within normal range.",
		"Биохимический анализ крови": "Blood values are within normal range.",
		"УЗИ брюшной полости":        "No pathological changes detected.",
		"УЗИ малого таза":            "No pathological changes detected.",
		"ЭКГ с расшифровкой":         "Sinus rhythm, no abnormalities.",
		"Рентген":                    "No signs of pathology.",
		"Флюорография":               "No signs of pathology.",
	},
}

// isLabCategory сообщает, положен ли по направлению результат анализа/диагностики.
func isLabCategory(category string) bool {
	_, ok := labResultTexts["ru"][category]
	return ok
}

// labResultText возвращает локализованный мок-текст результата анализа.
func labResultText(lang, category string) string {
	if m, ok := labResultTexts[lang]; ok {
		if v, ok := m[category]; ok {
			return v
		}
	}
	return labResultTexts["ru"][category]
}

// validLangs перечисляет поддерживаемые языки; используется для валидации cookie/параметра.
var validLangs = map[string]bool{"ru": true, "kz": true, "en": true}

// t возвращает перевод строки по ключу для данного языка, с откатом на русский,
// а если и там нет — возвращает сам ключ (чтобы не падать на пропущенных переводах).
func t(lang, key string) string {
	if key == "" {
		return ""
	}
	if m, ok := translations[lang]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if v, ok := translations["ru"][key]; ok {
		return v
	}
	return key
}

// specialty переводит название направления/специальности для отображения.
// Значение, используемое в ссылках и фильтрах, всегда остаётся русским (canonical).
func specialty(lang, ruName string) string {
	if lang == "ru" {
		return ruName
	}
	if m, ok := specialtyNames[lang]; ok {
		if v, ok := m[ruName]; ok {
			return v
		}
	}
	return ruName
}

// duration локализует строку длительности вида "30 мин" -> "30 min" и т.п.
func duration(lang, s string) string {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return s
	}
	num := parts[0]
	switch lang {
	case "en":
		return num + " min"
	case "kz":
		return num + " мин"
	default:
		return s
	}
}
