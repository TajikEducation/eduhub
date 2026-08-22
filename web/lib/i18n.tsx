"use client";

import { useAppState } from "./app-state";
import type { Bi } from "./data";

export type Locale = "ru" | "tg";

/**
 * Only strings actually reused across multiple places live here (keyed).
 * One-off copy is passed straight to `t()` as a literal {ru,tg} object at
 * the call site — no need to register every microcopy string here.
 */
export const UI_STRINGS: Record<string, Bi> = {
  // nav
  "nav.home": { ru: "Главная", tg: "Асосӣ" },
  "nav.search": { ru: "Каталог", tg: "Каталог" },
  "nav.vacancies": { ru: "Вакансии", tg: "Ҷойҳои холӣ" },
  "nav.about": { ru: "О нас", tg: "Дар бораи мо" },
  "nav.categories": { ru: "Типы учреждений", tg: "Намудҳои муассиса" },
  "nav.guides": { ru: "Руководства", tg: "Дастурҳо" },
  "nav.gps": { ru: "Определить по GPS", tg: "Тавассути GPS муайян кардан" },
  "nav.gpsLocating": { ru: "Определяем…", tg: "Муайян мекунем…" },
  "nav.guide.parents": { ru: "Родителям", tg: "Ба волидайн" },
  "nav.guide.students": { ru: "Ученикам", tg: "Ба хонандагон" },
  "nav.guide.institutions": { ru: "Учреждениям", tg: "Ба муассисаҳо" },
  "nav.cabinet": { ru: "Кабинет", tg: "Кабинет" },
  "nav.login": { ru: "Войти", tg: "Ворид шудан" },
  "nav.logout": { ru: "Выйти", tg: "Баромадан" },
  "nav.allRegions": { ru: "Все регионы", tg: "Ҳама минтақаҳо" },
  "nav.role.user": { ru: "Обычный пользователь", tg: "Истифодабарандаи оддӣ" },
  "nav.role.institution": { ru: "Учреждение", tg: "Муассиса" },
  "nav.role.moderator": { ru: "Модератор", tg: "Модератор" },
  "nav.role.admin": { ru: "Администратор платформы", tg: "Маъмури платформа" },
  "nav.guide.applicants": { ru: "Соискателям", tg: "Ба ҷуяндагони кор" },

  // common
  "common.openProfile": { ru: "Открыть профиль", tg: "Профилро кушоед" },
  "common.open": { ru: "Открыть", tg: "Кушодан" },
  "common.save": { ru: "Сохранить изменения", tg: "Тағйиротро нигоҳ доред" },
  "common.saved": { ru: "Изменения сохранены", tg: "Тағйирот нигоҳ дошта шуд" },
  "common.cancel": { ru: "Отмена", tg: "Бекор кардан" },
  "common.add": { ru: "Добавить", tg: "Илова кардан" },
  "common.edit": { ru: "Редактировать", tg: "Таҳрир кардан" },
  "common.delete": { ru: "Удалить", tg: "Нест кардан" },
  "common.reply": { ru: "Ответить", tg: "Ҷавоб додан" },
  "common.send": { ru: "Отправить", tg: "Фиристодан" },
  "common.readMore": { ru: "Читать полностью…", tg: "Пурра хондан…" },
  "common.collapse": { ru: "Свернуть", tg: "Пӯшидан" },
  "common.reset": { ru: "Сбросить", tg: "Тоза кардан" },
  "common.clearAll": { ru: "Очистить всё", tg: "Ҳамаро тоза кардан" },
  "common.back": { ru: "Назад", tg: "Бозгашт" },
  "common.reviews": { ru: "отзывов", tg: "шарҳ" },
  "common.perMonth": { ru: "сом/мес", tg: "сомонӣ/моҳ" },
  "common.verified": { ru: "МОН РТ", tg: "ВМО ҶТ" },
  "common.months": {
    ru: "янв,фев,мар,апр,май,июн,июл,авг,сен,окт,ноя,дек",
    tg: "янв,фев,мар,апр,май,июн,июл,авг,сен,окт,ноя,дек",
  },
  "common.messageEllipsis": { ru: "Сообщение…", tg: "Паём…" },
  "common.writeInChat": { ru: "Написать в чат", tg: "Дар чат нависед" },
  "common.write": { ru: "Написать", tg: "Навиштан" },

  // dashboard
  "dash.overview": { ru: "Обзор", tg: "Хулоса" },
  "dash.settings": { ru: "Настройки", tg: "Танзимот" },

  // tabs (institution profile)
  "tab.about": { ru: "О нас", tg: "Дар бораи мо" },
  "tab.staff": { ru: "Персонал", tg: "Кормандон" },
  "tab.gallery": { ru: "Галерея", tg: "Галерея" },
  "tab.achievements": { ru: "Достижения", tg: "Дастовардҳо" },
  "tab.menu": { ru: "Питание", tg: "Ғизо" },
  "tab.reviews": { ru: "Отзывы", tg: "Шарҳҳо" },
  "tab.news": { ru: "Новости", tg: "Хабарҳо" },
  "tab.contacts": { ru: "Контакты", tg: "Тамос" },
  "tab.alumni": { ru: "Выпускники", tg: "Хатмкунандагон" },
  "tab.vacancies": { ru: "Вакансии", tg: "Ҷойҳои холӣ" },

  // empty states
  "empty.staff": { ru: "Персонал не добавлен", tg: "Кормандон илова нашудаанд" },
  "empty.reviews": { ru: "Отзывов пока нет", tg: "Ҳанӯз шарҳе нест" },
  "empty.news": { ru: "Новостей пока нет", tg: "Ҳанӯз хабаре нест" },
  "empty.achievements": { ru: "Достижения не добавлены", tg: "Дастовардҳо илова нашудаанд" },
  "empty.notifications": { ru: "Уведомлений пока нет", tg: "Ҳанӯз огоҳиномае нест" },
  "empty.notFound": { ru: "Ничего не найдено", tg: "Ҳеҷ чиз ёфт нашуд" },
  "empty.vacancies": { ru: "Вакансий пока нет", tg: "Ҷойҳои холӣ ҳанӯз нестанд" },

  // geo
  "geo.region": { ru: "Регион", tg: "Минтақа" },
  "geo.city": { ru: "Город", tg: "Шаҳр" },
  "geo.district": { ru: "Район", tg: "Ноҳия" },
  "geo.street": { ru: "Улица", tg: "Кӯча" },

  // metric labels (used as keys directly via inst.metrics[].label)
  "Качество обучения": { ru: "Качество обучения", tg: "Сифати таълим" },
  "Условия и помещения": { ru: "Условия и помещения", tg: "Шароит ва биноҳо" },
  "Безопасность": { ru: "Безопасность", tg: "Бехатарӣ" },
  "Питание": { ru: "Питание", tg: "Ғизо" },
  "Развозка": { ru: "Развозка", tg: "Расонидан" },
  "Цена/качество": { ru: "Цена/качество", tg: "Нарх/сифат" },
  "Участие родителей": { ru: "Участие родителей", tg: "Иштироки волидайн" },
  "Инклюзивность": { ru: "Инклюзивность", tg: "Фарогирӣ" },

  // work hours (fallback, same across institutions)
  "hours.weekdays": { ru: "Пн–Пт", tg: "Дш–Ҷм" },
  "hours.saturday": { ru: "Суббота", tg: "Шанбе" },
  "hours.sunday": { ru: "Воскресенье", tg: "Якшанбе" },
  "hours.dayoff": { ru: "Выходной", tg: "Рӯзи истироҳат" },

  // fallback menu (shown when an institution has no custom foodMenu)
  "menu.mon": { ru: "Понедельник", tg: "Душанбе" },
  "menu.tue": { ru: "Вторник", tg: "Сешанбе" },
  "menu.wed": { ru: "Среда", tg: "Чоршанбе" },
  "menu.thu": { ru: "Четверг", tg: "Панҷшанбе" },
  "menu.fri": { ru: "Пятница", tg: "Ҷумъа" },
  "menu.breakfast": { ru: "Завтрак", tg: "Ноништа" },
  "menu.lunch": { ru: "Обед", tg: "Хӯроки нисфирӯзӣ" },
  "menu.snack": { ru: "Полдник", tg: "Насри" },
  "menu.breakfastTxt": { ru: "Каша + чай", tg: "Оши сершир + чой" },
  "menu.lunchTxt": { ru: "Суп + плов", tg: "Шӯрбо + палов" },
  "menu.snackTxt": { ru: "Фрукты", tg: "Мева" },

  // achievement category groupings
  "ach.institution": { ru: "Достижения учреждения", tg: "Дастовардҳои муассиса" },
  "ach.student": { ru: "Достижения учеников", tg: "Дастовардҳои хонандагон" },
  "ach.teacher": { ru: "Достижения педагогов", tg: "Дастовардҳои омӯзгорон" },

  // review connection gate (SRS FR-15/FR-30 — верификация через привязку ребёнка)
  "review.needConnection": { ru: "Отзыв могут оставить только родители, чей ребёнок связан с этим учреждением.", tg: "Шарҳро танҳо волидайне гузошта метавонанд, ки фарзандашон бо ин муассиса алоқаманд аст." },
  "review.linkChild": { ru: "Указать ребёнка в личном кабинете", tg: "Дар кабинети шахсӣ фарзандро нишон додан" },
  "review.overall": { ru: "Общая оценка", tg: "Баҳои умумӣ" },

  // children tab (UserCabinet)
  "children.institution": { ru: "Учреждение", tg: "Муассиса" },
  "children.noInstitution": { ru: "Без привязки к учреждению", tg: "Бе алоқаманд ба муассиса" },
  "children.status": { ru: "Статус", tg: "Мақом" },
  "children.statusCurrent": { ru: "Учится сейчас", tg: "Ҳоло мехонад" },
  "children.statusAlumnus": { ru: "Выпускник", tg: "Хатмкарда" },
  "children.statusTransferred": { ru: "Перешёл в другое учреждение", tg: "Ба муассисаи дигар гузашт" },
  "children.multiInstHint": { ru: "Один ребёнок может учиться сразу в нескольких учреждениях (например, школа + учебный центр) — просто добавьте отдельную запись с тем же именем для каждого.", tg: "Як фарзанд метавонад дар якчанд муассиса якбора хонад (масалан, мактаб + марказ) — барои ҳар кадом сабти алоҳидаро бо номи якхела илова кунед." },

  // профиль «Ищу работу» — вкладка единого кабинета пользователя, не отдельный кабинет
  "applicant.position": { ru: "Желаемая должность", tg: "Мансаби дилхоҳ" },
  "applicant.about": { ru: "О себе", tg: "Дар бораи худ" },
  "applicant.cv": { ru: "Резюме (файл)", tg: "Резюме (файл)" },
  "applicant.education": { ru: "Образование", tg: "Таҳсилот" },
  "applicant.experience": { ru: "Опыт работы", tg: "Таҷрибаи корӣ" },
  "applicant.skills": { ru: "Навыки", tg: "Малакаҳо" },
  "applicant.publish": { ru: "Опубликовать профиль", tg: "Профилро нашр кунед" },
  "applicant.unpublish": { ru: "Снять с публикации", tg: "Аз нашр гирифтан" },
  "applicant.publishedNote": { ru: "Профиль опубликован — виден учреждениям и другим пользователям.", tg: "Профил нашр шудааст — ба муассисаҳо ва дигар корбарон намоён аст." },
  "applicant.draftNote": { ru: "Черновик — не виден другим, пока вы не опубликуете профиль.", tg: "Лоиҳа — то нашр кардани профил ба дигарон намоён нест." },
  "applicant.visibility.draft": { ru: "Черновик", tg: "Лоиҳа" },
  "applicant.visibility.draftDesc": { ru: "Виден только вам", tg: "Танҳо ба худи шумо намоён аст" },
  "applicant.visibility.on_response": { ru: "По отклику", tg: "Аз рӯи ҷавоб" },
  "applicant.visibility.on_responseDesc": { ru: "Виден только учреждению, которому вы откликнулись", tg: "Танҳо ба муассисае, ки ба он ҷавоб додаед, намоён аст" },
  "applicant.visibility.public": { ru: "Публично", tg: "Оммавӣ" },
  "applicant.visibility.publicDesc": { ru: "Виден всем в разделе «Ищу работу»", tg: "Дар бахши «Ҷустуҷӯи кор» ба ҳама намоён аст" },
  "applicant.hideContacts": { ru: "Скрывать телефон и email, пока учреждение не ответит в чате", tg: "Телефон ва email-ро то ҷавоби муассиса дар чат пинҳон доштан" },
  "tab.candidates": { ru: "Ищу работу", tg: "Ҷустуҷӯи кор" },
  "empty.candidates": { ru: "Пока нет опубликованных резюме", tg: "Ҳанӯз резюмеи нашршуда нест" },
  "applicant.myApplications": { ru: "Мои отклики", tg: "Ҷавобҳои ман" },
  "applicant.employerResponses": { ru: "Кто откликнулся мне", tg: "Кӣ ба ман ҷавоб дод" },
  "empty.employerResponses": { ru: "Пока никто из учреждений вам не написал", tg: "То ҳол ҳеҷ муассиса ба шумо нанавиштааст" },
  "empty.applications": { ru: "Вы ещё не откликались на вакансии", tg: "Шумо ҳанӯз ба ҷойҳои холӣ ҷавоб надодаед" },
  "application.status.sent": { ru: "Отправлен", tg: "Фиристода шуд" },
  "application.status.viewed": { ru: "Просмотрен учреждением", tg: "Аз ҷониби муассиса дида шуд" },
  "application.status.closed": { ru: "Закрыт", tg: "Пӯшида шуд" },
};

export function translate(x: string | Bi, locale: Locale): string {
  if (typeof x === "string") {
    const entry = UI_STRINGS[x];
    return entry ? entry[locale] : x;
  }
  return x[locale];
}

export function useT() {
  const { locale } = useAppState();
  return (x: string | Bi) => translate(x, locale);
}

export function useLocale() {
  const { locale, setLocale } = useAppState();
  return { locale, setLocale };
}
