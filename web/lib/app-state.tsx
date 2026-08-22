"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { NOTIFICATIONS, DEFAULT_APPLICANT, INSTITUTIONS, CATEGORY_META, EMPLOYER_RESPONSES, type Notification, type Region, type Applicant, type Application, type Institution, type CategoryKey, type Bi, type EmployerResponse } from "./data";
import { REGION_CENTROIDS } from "./geo";
import type { Locale } from "./i18n";

// RBAC (SRS §3): уровни доступа. "Родитель"/"соискатель" — НЕ роли, а факты о
// пользователе (children_.length>0 / applicant.visibility) — см. ChildLink и Applicant ниже.
export type Role = "guest" | "user" | "institution" | "moderator" | "admin";

// Минимальная привязка «родитель–ребёнок–учреждение» (SRS §7, сущность Child) —
// только age/status/institution, без лишних данных о ребёнке. `name` хранится
// только локально для удобства самого родителя в его кабинете и никогда не
// попадает в отзыв/публичные данные — это не то же самое, что PII в сущности Child.
// "transferred" — ребёнок ушёл из этого учреждения не выпустившись (перевёлся в
// другое). Для верификации отзыва (FR-15/30) считается тем же, что alumnus —
// связь была реальной, просто не завершилась выпуском.
export type ChildStatus = "current" | "alumnus" | "transferred";

export interface ChildLink {
  id: number;
  name: string;
  age: string;
  instId: number | null;
  status: ChildStatus;
}

// Данные формы регистрации учреждения — минимум, из которого собирается полный
// Institution (остальные поля — дефолты, см. registerInstitution ниже).
export interface RegisterInstitutionInput {
  name: string;
  tk: CategoryKey;
  region: Region;
  city: string;
  street: string;
  area: string;
  price: number;
  founded: number;
  students: number;
  age: string;
  phone: string;
  email: string;
  website: string;
  description: string;
}

export interface PlatformSettings {
  tierPrices: { pro: number; enterprise: number };
  maintenanceMode: boolean;
}
const DEFAULT_PLATFORM_SETTINGS: PlatformSettings = { tierPrices: { pro: 30, enterprise: 100 }, maintenanceMode: false };

interface AppStateValue {
  savedIds: number[];
  isSaved: (id: number) => boolean;
  toggleSaved: (id: number) => void;

  notifications: Notification[];
  unreadNotifications: number;
  markNotificationRead: (id: string) => void;
  markAllNotificationsRead: () => void;

  unreadMessages: number;
  setUnreadMessages: (n: number | ((prev: number) => number)) => void;

  role: Role;
  setRole: (r: Role) => void;

  locale: Locale;
  setLocale: (l: Locale) => void;

  region: Region | null;
  setRegion: (r: Region | null) => void;

  children_: ChildLink[];
  addChild: (c: Omit<ChildLink, "id">) => void;
  removeChild: (id: number) => void;

  applicant: Applicant;
  setApplicant: (a: Applicant | ((prev: Applicant) => Applicant)) => void;

  applications: Application[];
  hasApplied: (vacancyId: string) => boolean;
  addApplication: (vacancyId: string) => void;

  myInstitution: Institution | null;
  registerInstitution: (input: RegisterInstitutionInput) => void;
  setInstitutionStatus: (status: Institution["status"]) => void;

  platformSettings: PlatformSettings;
  setPlatformSettings: (patch: Partial<PlatformSettings>) => void;

  employerResponses: EmployerResponse[];
  hasResponded: (applicantId: string, instId: number) => boolean;
  addEmployerResponse: (applicantId: string, instId: number, message: string) => void;
}

const AppStateContext = createContext<AppStateValue | null>(null);
const LS_KEY = "eduhub_app_state_v2";

interface Stored {
  savedIds: number[];
  role: Role;
  locale: Locale;
  region: Region | null;
  children_: ChildLink[];
  applicant: Applicant;
  applications: Application[];
  myInstitution: Institution | null;
  platformSettings: PlatformSettings;
  employerResponses: EmployerResponse[];
}

// Демо-сид для нового посетителя (localStorage ещё пуст) — чтобы кабинет не
// выглядел пустым при первом знакомстве с прототипом. Реальный пользователь
// продолжит с этих же значений и может их менять как обычно.
const SEED_CHILDREN: ChildLink[] = [
  { id: 1001, name: "Амир Юсупов", age: "10 лет", instId: 1, status: "current" },
  { id: 1002, name: "Амир Юсупов", age: "10 лет", instId: 4, status: "current" }, // тот же ребёнок — школа + учебный центр
  { id: 1003, name: "Зарина Юсупова", age: "5 лет", instId: 2, status: "current" },
  { id: 1004, name: "Давлат Юсупов", age: "17 лет", instId: 3, status: "alumnus" },
  { id: 1005, name: "Нилуфар Юсупова", age: "8 лет", instId: 9, status: "current" },
];
const SEED_APPLICATIONS: Application[] = [
  { id: "app-seed-1", applicantId: "applicant-me", vacancyId: "v1", status: "viewed", createdAt: "14 июл 2026" },
  { id: "app-seed-2", applicantId: "applicant-me", vacancyId: "v4", status: "sent", createdAt: "16 июл 2026" },
];

function readStored(): Stored {
  const empty: Stored = { savedIds: [1, 2, 3, 5, 7], role: "guest", locale: "ru", region: null, children_: SEED_CHILDREN, applicant: DEFAULT_APPLICANT, applications: SEED_APPLICATIONS, myInstitution: null, platformSettings: DEFAULT_PLATFORM_SETTINGS, employerResponses: EMPLOYER_RESPONSES };
  if (typeof window === "undefined") return empty;
  try {
    const raw = localStorage.getItem(LS_KEY);
    if (!raw) return empty;
    const parsed = JSON.parse(raw);
    return {
      savedIds: Array.isArray(parsed.savedIds) ? parsed.savedIds : [],
      role: typeof parsed.role === "string" ? parsed.role : "guest",
      locale: parsed.locale === "tg" ? "tg" : "ru",
      region: typeof parsed.region === "string" ? parsed.region : null,
      children_: Array.isArray(parsed.children_) ? parsed.children_ : [],
      applicant: parsed.applicant && typeof parsed.applicant === "object" ? { ...DEFAULT_APPLICANT, ...parsed.applicant } : DEFAULT_APPLICANT,
      applications: Array.isArray(parsed.applications) ? parsed.applications : [],
      myInstitution: parsed.myInstitution && typeof parsed.myInstitution === "object" ? parsed.myInstitution : null,
      platformSettings: parsed.platformSettings && typeof parsed.platformSettings === "object" ? { ...DEFAULT_PLATFORM_SETTINGS, ...parsed.platformSettings } : DEFAULT_PLATFORM_SETTINGS,
      employerResponses: Array.isArray(parsed.employerResponses) ? parsed.employerResponses : EMPLOYER_RESPONSES,
    };
  } catch {
    return empty;
  }
}

export function AppStateProvider({ children }: { children: React.ReactNode }) {
  const [savedIds, setSavedIds] = useState<number[]>(() => readStored().savedIds);
  const [notifications, setNotifications] = useState<Notification[]>(NOTIFICATIONS);
  const [unreadMessages, setUnreadMessages] = useState<number>(2);
  const [role, setRole] = useState<Role>(() => readStored().role);
  const [locale, setLocale] = useState<Locale>(() => readStored().locale);
  const [region, setRegion] = useState<Region | null>(() => readStored().region);
  const [children_, setChildren_] = useState<ChildLink[]>(() => readStored().children_);
  const [applicant, setApplicant] = useState<Applicant>(() => readStored().applicant);
  const [applications, setApplications] = useState<Application[]>(() => readStored().applications);
  const [myInstitution, setMyInstitution] = useState<Institution | null>(() => readStored().myInstitution);
  const [platformSettings, setPlatformSettingsState] = useState<PlatformSettings>(() => readStored().platformSettings);
  const [employerResponses, setEmployerResponses] = useState<EmployerResponse[]>(() => readStored().employerResponses);

  useEffect(() => {
    localStorage.setItem(LS_KEY, JSON.stringify({ savedIds, role, locale, region, children_, applicant, applications, myInstitution, platformSettings, employerResponses }));
  }, [savedIds, role, locale, region, children_, applicant, applications, myInstitution, platformSettings, employerResponses]);

  const isSaved = useCallback((id: number) => savedIds.includes(id), [savedIds]);
  const toggleSaved = useCallback((id: number) => {
    setSavedIds((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]));
  }, []);

  const markNotificationRead = useCallback((id: string) => {
    setNotifications((prev) => prev.map((n) => (n.id === id ? { ...n, read: true } : n)));
  }, []);
  const markAllNotificationsRead = useCallback(() => {
    setNotifications((prev) => prev.map((n) => ({ ...n, read: true })));
  }, []);

  const unreadNotifications = notifications.filter((n) => !n.read).length;

  const addChild = useCallback((c: Omit<ChildLink, "id">) => {
    setChildren_((prev) => [...prev, { ...c, id: Date.now() }]);
  }, []);
  const removeChild = useCallback((id: number) => {
    setChildren_((prev) => prev.filter((c) => c.id !== id));
  }, []);

  const hasApplied = useCallback((vacancyId: string) => applications.some((a) => a.vacancyId === vacancyId), [applications]);
  // идемпотентно: повторный отклик на ту же вакансию не создаёт дубль
  const addApplication = useCallback((vacancyId: string) => {
    setApplications((prev) => {
      if (prev.some((a) => a.vacancyId === vacancyId)) return prev;
      return [...prev, { id: `app-${Date.now()}`, applicantId: applicant.id, vacancyId, status: "sent", createdAt: new Date().toLocaleDateString("ru-RU") }];
    });
  }, [applicant.id]);

  const bi = (v: string): Bi => ({ ru: v, tg: v }); // без сервиса перевода форма собирает одну строку на оба языка — как ListEditor в кабинете

  const registerInstitution = useCallback((input: RegisterInstitutionInput) => {
    const meta = CATEGORY_META[input.tk];
    const maxId = INSTITUTIONS.reduce((m, i) => Math.max(m, i.id), 0);
    const newInst: Institution = {
      id: maxId + 1,
      name: bi(input.name),
      type: input.tk,
      tk: input.tk,
      region: input.region,
      city: bi(input.city),
      street: bi(input.street),
      area: input.area,
      price: input.price,
      score: 0,
      rev: 0,
      transport: false,
      food: false,
      ver: false,
      status: "pending",
      color: meta.color,
      age: input.age,
      tag: null,
      coverPhoto: meta.heroPhoto,
      gallery: [],
      founded: input.founded,
      students: input.students,
      description: bi(input.description),
      achievements: [],
      metrics: [],
      staff: [],
      address: bi(`${input.street}, ${input.area}, ${input.city}`),
      phone: input.phone,
      email: input.email,
      website: input.website,
      geo: REGION_CENTROIDS[input.region],
    };
    setMyInstitution(newInst);
    setRole("institution");
  }, []);

  const setInstitutionStatus = useCallback((status: Institution["status"]) => {
    setMyInstitution((prev) => (prev ? { ...prev, status } : prev));
  }, []);

  const setPlatformSettings = useCallback((patch: Partial<PlatformSettings>) => {
    setPlatformSettingsState((prev) => ({ ...prev, ...patch, tierPrices: { ...prev.tierPrices, ...patch.tierPrices } }));
  }, []);

  const hasResponded = useCallback((applicantId: string, instId: number) =>
    employerResponses.some((r) => r.applicantId === applicantId && r.instId === instId), [employerResponses]);
  // идемпотентно: повторный отклик того же учреждения тому же кандидату не создаёт дубль
  const addEmployerResponse = useCallback((applicantId: string, instId: number, message: string) => {
    setEmployerResponses((prev) => {
      if (prev.some((r) => r.applicantId === applicantId && r.instId === instId)) return prev;
      return [...prev, { id: `er-${Date.now()}`, applicantId, instId, message: bi(message), date: new Date().toLocaleDateString("ru-RU") }];
    });
  }, []);

  const value: AppStateValue = {
    savedIds,
    isSaved,
    toggleSaved,
    notifications,
    unreadNotifications,
    markNotificationRead,
    markAllNotificationsRead,
    unreadMessages,
    setUnreadMessages,
    role,
    setRole,
    locale,
    setLocale,
    region,
    setRegion,
    children_,
    addChild,
    removeChild,
    applicant,
    setApplicant,
    applications,
    hasApplied,
    addApplication,
    myInstitution,
    registerInstitution,
    setInstitutionStatus,
    platformSettings,
    setPlatformSettings,
    employerResponses,
    hasResponded,
    addEmployerResponse,
  };

  return <AppStateContext.Provider value={value}>{children}</AppStateContext.Provider>;
}

export function useAppState() {
  const ctx = useContext(AppStateContext);
  if (!ctx) throw new Error("useAppState must be used within AppStateProvider");
  return ctx;
}

// Публичные списки (поиск/каталог/карта) должны видеть только одобренные
// учреждения — плюс собственное учреждение пользователя, если оно уже
// approved и ещё не попало в статичный сид-массив INSTITUTIONS (произойдёт
// после модерации в волне C).
export function useVisibleInstitutions(): Institution[] {
  const { myInstitution } = useAppState();
  return useMemo(() => {
    const base = INSTITUTIONS.filter((i) => i.status === "approved");
    if (myInstitution && myInstitution.status === "approved" && !INSTITUTIONS.some((i) => i.id === myInstitution.id)) {
      return [...base, myInstitution];
    }
    return base;
  }, [myInstitution]);
}
