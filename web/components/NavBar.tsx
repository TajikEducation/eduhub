"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { ChevronDown, LocateFixed, LogOut, MapPin } from "lucide-react";
import { C, FH, FB, REGION_LABEL, REGION_ORDER } from "@/lib/data";
import { useAppState, type Role } from "@/lib/app-state";
import { useT } from "@/lib/i18n";
import { detectRegionByGPS } from "@/lib/geo";
import { NotificationBell } from "./NotificationBell";
import { MessagesLink } from "./MessagesLink";

const GUIDE_LINKS: { href: string; key: string }[] = [
  { href: "/guide/parents", key: "nav.guide.parents" },
  { href: "/guide/students", key: "nav.guide.students" },
  { href: "/guide/applicants", key: "nav.guide.applicants" },
  { href: "/guide/institutions", key: "nav.guide.institutions" },
];

// RBAC-роль (SRS §3) — не то же самое, что "родитель"/"соискатель": это факты о
// пользователе (наличие ChildLink / опубликованного профиля соискателя), не роли.
// Переключатель ниже — тестовый инструмент прототипа, не выбор личности.
const ROLE_OPTIONS: { value: Role; key: string; href: string }[] = [
  { value: "user", key: "nav.role.user", href: "/account" },
  { value: "institution", key: "nav.role.institution", href: "/dashboard" },
  { value: "moderator", key: "nav.role.moderator", href: "/moderator" },
  { value: "admin", key: "nav.role.admin", href: "/admin" },
];

function useOutsideClose<T extends HTMLElement>() {
  const ref = useRef<T>(null);
  const [open, setOpen] = useState(false);
  useEffect(() => {
    function onClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    }
    document.addEventListener("mousedown", onClick);
    return () => document.removeEventListener("mousedown", onClick);
  }, []);
  return { ref, open, setOpen };
}

function Dropdown({ label, active, children, width = 210 }: { label: string; active?: boolean; children: React.ReactNode; width?: number }) {
  const { ref, open, setOpen } = useOutsideClose<HTMLDivElement>();
  return (
    <div ref={ref} style={{ position: "relative" }}>
      <button
        onClick={() => setOpen((o) => !o)}
        style={{ display: "flex", alignItems: "center", gap: 4, padding: "7px 10px", borderRadius: 8, fontSize: 13.5, fontWeight: 600, fontFamily: FH, color: active ? C.text : C.muted, background: active ? C.s3 : "transparent", border: "none", cursor: "pointer", whiteSpace: "nowrap" }}
      >
        {label} <ChevronDown size={12} style={{ transform: open ? "rotate(180deg)" : "none", transition: "transform .15s" }} />
      </button>
      {open && (
        <div onClick={() => setOpen(false)} style={{ position: "absolute", top: "calc(100% + 10px)", left: 0, width, background: C.s2, border: `1px solid ${C.border}`, borderRadius: 12, boxShadow: "0 20px 50px rgba(0,0,0,0.45)", zIndex: 120, overflow: "hidden", animation: "eh-pop .16s ease-out" }}>
          {children}
        </div>
      )}
    </div>
  );
}

export function NavBar() {
  const pathname = usePathname();
  const router = useRouter();
  const { role, setRole, region, setRegion, locale, setLocale } = useAppState();
  const t = useT();
  const { ref: roleMenuRef, open: roleMenuOpen, setOpen: setRoleMenuOpen } = useOutsideClose<HTMLDivElement>();
  const [gpsBusy, setGpsBusy] = useState(false);

  const currentRoleLabel = ROLE_OPTIONS.find((r) => r.value === role);
  const cabinetHref = currentRoleLabel?.href;

  function useGPS() {
    setGpsBusy(true);
    detectRegionByGPS()
      .then((r) => setRegion(r))
      .catch(() => {})
      .finally(() => setGpsBusy(false));
  }

  function chooseRole(opt: (typeof ROLE_OPTIONS)[number]) {
    setRole(opt.value);
    setRoleMenuOpen(false);
    router.push(opt.href);
  }

  function logout() {
    setRole("guest");
    setRoleMenuOpen(false);
    router.push("/");
  }

  return (
    <header style={{position:"sticky",top:0,zIndex:50,background:`${C.bg}EB`,backdropFilter:"blur(20px)",borderBottom:`1px solid ${C.border}`}}>
      <div style={{maxWidth:1320,margin:"0 auto",padding:"0 18px",display:"flex",alignItems:"center",gap:4,height:64}}>
        <Link href="/" style={{display:"flex",alignItems:"center",gap:9,flexShrink:0,marginRight:6}}>
          <svg width="32" height="32" viewBox="0 0 40 40" fill="none">
            <path d="M20 3 L24 15 L37 20 L24 25 L20 37 L16 25 L3 20 L16 15 Z" fill={C.teal}/>
            <path d="M20 8 L22.5 16 L31 20 L22.5 24 L20 32 L17.5 24 L9 20 L17.5 16 Z" fill={C.gold}/>
            <circle cx="20" cy="20" r="3.4" fill={C.red}/>
          </svg>
          <span style={{fontFamily:FH,fontWeight:900,fontSize:19,color:C.text,letterSpacing:"-.01em"}}>
            Edu<span style={{color:C.gold}}>Hub</span>
          </span>
        </Link>

        <nav style={{display:"flex",gap:2,alignItems:"center"}}>
          <Link href="/" style={{padding:"7px 10px",borderRadius:8,fontSize:13.5,fontWeight:600,fontFamily:FH,color:pathname==="/"?C.text:C.muted,background:pathname==="/"?C.s3:"transparent",whiteSpace:"nowrap"}}>
            {t("nav.home")}
          </Link>
          <Link href="/search" style={{padding:"7px 10px",borderRadius:8,fontSize:13.5,fontWeight:600,fontFamily:FH,color:pathname==="/search"?C.text:C.muted,background:pathname==="/search"?C.s3:"transparent",whiteSpace:"nowrap"}}>
            {t("nav.search")}
          </Link>
          <Link href="/vacancies" style={{padding:"7px 10px",borderRadius:8,fontSize:13.5,fontWeight:600,fontFamily:FH,color:pathname.startsWith("/vacancies")?C.text:C.muted,background:pathname.startsWith("/vacancies")?C.s3:"transparent",whiteSpace:"nowrap"}}>
            {t("nav.vacancies")}
          </Link>

          <Dropdown label={t("nav.guides")} active={pathname.startsWith("/guide")}>
            {GUIDE_LINKS.map((g) => (
              <Link key={g.href} href={g.href} style={{display:"block",padding:"11px 14px",color:C.text,fontFamily:FB,fontSize:13.5,textDecoration:"none"}}>
                {t(g.key)}
              </Link>
            ))}
          </Dropdown>

          <Link href="/company" style={{padding:"7px 10px",borderRadius:8,fontSize:13.5,fontWeight:600,fontFamily:FH,color:pathname==="/company"?C.text:C.muted,background:pathname==="/company"?C.s3:"transparent",whiteSpace:"nowrap"}}>
            {t("nav.about")}
          </Link>
        </nav>

        <div style={{marginLeft:"auto",display:"flex",gap:6,alignItems:"center"}}>
          <Dropdown label={region ? t(REGION_LABEL[region]) : t("nav.allRegions")} width={220}>
            <button onClick={useGPS} disabled={gpsBusy} style={{display:"flex",alignItems:"center",gap:8,width:"100%",padding:"11px 14px",background:"transparent",border:"none",borderBottom:`1px solid ${C.border}`,color:C.teal,fontFamily:FH,fontWeight:700,fontSize:13,cursor:gpsBusy?"default":"pointer",textAlign:"left"}}>
              <LocateFixed size={13}/> {gpsBusy?t("nav.gpsLocating"):t("nav.gps")}
            </button>
            <button onClick={()=>setRegion(null)} style={{display:"flex",alignItems:"center",gap:8,width:"100%",padding:"11px 14px",background:!region?C.s3:"transparent",border:"none",color:C.text,fontFamily:FB,fontSize:13.5,cursor:"pointer",textAlign:"left"}}>
              <MapPin size={13}/> {t("nav.allRegions")}
            </button>
            {REGION_ORDER.map((r) => (
              <button key={r} onClick={()=>setRegion(r)} style={{display:"block",width:"100%",padding:"11px 14px",background:region===r?C.s3:"transparent",border:"none",color:C.text,fontFamily:FB,fontSize:13.5,cursor:"pointer",textAlign:"left"}}>
                {t(REGION_LABEL[r])}
              </button>
            ))}
          </Dropdown>

          <div style={{display:"flex",borderRadius:8,overflow:"hidden",border:`1px solid ${C.border}`}}>
            <button onClick={()=>setLocale("tg")} style={{padding:"7px 11px",fontSize:12.5,fontWeight:700,color:locale==="tg"?C.bg:C.muted,fontFamily:FH,background:locale==="tg"?C.teal:"none",border:"none",cursor:"pointer"}}>ТҶ</button>
            <button onClick={()=>setLocale("ru")} style={{padding:"7px 11px",fontSize:12.5,fontWeight:700,color:locale==="ru"?C.bg:C.muted,fontFamily:FH,background:locale==="ru"?C.teal:"none",border:"none",cursor:"pointer"}}>РУ</button>
          </div>

          <MessagesLink/>
          <NotificationBell/>

          {role!=="guest" && cabinetHref && (
            <Link href={cabinetHref} style={{padding:"8px 14px",borderRadius:8,fontFamily:FH,fontWeight:700,fontSize:13,color:C.teal,border:`1px solid ${C.teal}40`,background:`${C.teal}10`,textDecoration:"none"}}>
              {t("nav.cabinet")}
            </Link>
          )}

          <div ref={roleMenuRef} style={{position:"relative"}}>
            <button
              onClick={()=>setRoleMenuOpen(o=>!o)}
              style={{display:"flex",alignItems:"center",gap:6,padding:"8px 14px",borderRadius:8,fontFamily:FH,fontWeight:700,fontSize:13,background:role==="guest"?C.teal:C.s3,color:role==="guest"?C.bg:C.text,border:role==="guest"?"none":`1px solid ${C.border}`,cursor:"pointer"}}
            >
              {role==="guest" ? t("nav.login") : t(currentRoleLabel?.key ?? "nav.login")} <ChevronDown size={13}/>
            </button>
            {roleMenuOpen && (
              <div style={{position:"absolute",top:"calc(100% + 10px)",right:0,width:230,background:C.s2,border:`1px solid ${C.border}`,borderRadius:12,boxShadow:"0 20px 50px rgba(0,0,0,0.45)",zIndex:120,overflow:"hidden",animation:"eh-pop .16s ease-out"}}>
                {ROLE_OPTIONS.map(opt=>(
                  <button key={opt.value} onClick={()=>chooseRole(opt)}
                    style={{display:"block",width:"100%",textAlign:"left",padding:"11px 14px",background:role===opt.value?C.s3:"transparent",border:"none",color:C.text,fontFamily:FB,fontSize:13.5,cursor:"pointer"}}>
                    {t(opt.key)}
                  </button>
                ))}
                {role!=="guest" && (
                  <button onClick={logout}
                    style={{display:"flex",alignItems:"center",gap:6,width:"100%",textAlign:"left",padding:"11px 14px",background:"transparent",border:"none",borderTop:`1px solid ${C.border}`,color:C.red,fontFamily:FB,fontSize:13.5,cursor:"pointer"}}>
                    <LogOut size={13}/> {t("nav.logout")}
                  </button>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}
