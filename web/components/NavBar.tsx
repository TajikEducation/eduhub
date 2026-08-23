"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { ChevronDown, LocateFixed, LogOut, MapPin, Menu, X } from "lucide-react";
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
  const [mobileOpen, setMobileOpen] = useState(false);
  const [prevPathname, setPrevPathname] = useState(pathname);
  if (pathname !== prevPathname) {
    setPrevPathname(pathname);
    setMobileOpen(false);
  }

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
      <style jsx>{`
        .eh-nav-burger { display: none; }
        @media (max-width: 880px) {
          .eh-nav-links, .eh-nav-right-desktop { display: none !important; }
          .eh-nav-burger { display: flex !important; }
        }
      `}</style>
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

        <nav className="eh-nav-links" style={{display:"flex",gap:2,alignItems:"center"}}>
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

        <div className="eh-nav-right-desktop" style={{marginLeft:"auto",display:"flex",gap:6,alignItems:"center"}}>
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
              onClick={()=> role==="guest" ? router.push("/login") : setRoleMenuOpen(o=>!o)}
              style={{display:"flex",alignItems:"center",gap:6,padding:"8px 14px",borderRadius:8,fontFamily:FH,fontWeight:700,fontSize:13,background:role==="guest"?C.teal:C.s3,color:role==="guest"?C.bg:C.text,border:role==="guest"?"none":`1px solid ${C.border}`,cursor:"pointer"}}
            >
              {role==="guest" ? t("nav.login") : t(currentRoleLabel?.key ?? "nav.login")} {role!=="guest" && <ChevronDown size={13}/>}
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

        {/* мобильный компактный кластер — видим только <880px, дублирует иконки/кнопку из eh-nav-right-desktop */}
        <div className="eh-nav-right-mobile" style={{display:"none",marginLeft:"auto",alignItems:"center",gap:6}}>
          <MessagesLink/>
          <NotificationBell/>
          <button onClick={()=>setMobileOpen(o=>!o)} aria-label={t("nav.menu")} className="eh-nav-burger"
            style={{width:36,height:36,borderRadius:8,background:C.s3,border:`1px solid ${C.border}`,color:C.text,alignItems:"center",justifyContent:"center",cursor:"pointer"}}>
            {mobileOpen ? <X size={17}/> : <Menu size={17}/>}
          </button>
        </div>
      </div>

      {mobileOpen && (
        <div style={{borderTop:`1px solid ${C.border}`,background:C.bg,padding:"14px 18px 20px",display:"flex",flexDirection:"column",gap:4,maxHeight:"calc(100vh - 64px)",overflowY:"auto"}}>
          <Link href="/" onClick={()=>setMobileOpen(false)} style={{padding:"11px 12px",borderRadius:9,fontSize:14.5,fontWeight:600,fontFamily:FH,color:pathname==="/"?C.text:C.muted,background:pathname==="/"?C.s3:"transparent"}}>{t("nav.home")}</Link>
          <Link href="/search" onClick={()=>setMobileOpen(false)} style={{padding:"11px 12px",borderRadius:9,fontSize:14.5,fontWeight:600,fontFamily:FH,color:pathname==="/search"?C.text:C.muted,background:pathname==="/search"?C.s3:"transparent"}}>{t("nav.search")}</Link>
          <Link href="/vacancies" onClick={()=>setMobileOpen(false)} style={{padding:"11px 12px",borderRadius:9,fontSize:14.5,fontWeight:600,fontFamily:FH,color:pathname.startsWith("/vacancies")?C.text:C.muted,background:pathname.startsWith("/vacancies")?C.s3:"transparent"}}>{t("nav.vacancies")}</Link>
          {GUIDE_LINKS.map((g) => (
            <Link key={g.href} href={g.href} onClick={()=>setMobileOpen(false)} style={{padding:"11px 12px",borderRadius:9,fontSize:14.5,fontWeight:600,fontFamily:FH,color:pathname===g.href?C.text:C.muted,background:pathname===g.href?C.s3:"transparent"}}>{t(g.key)}</Link>
          ))}
          <Link href="/company" onClick={()=>setMobileOpen(false)} style={{padding:"11px 12px",borderRadius:9,fontSize:14.5,fontWeight:600,fontFamily:FH,color:pathname==="/company"?C.text:C.muted,background:pathname==="/company"?C.s3:"transparent"}}>{t("nav.about")}</Link>

          <div style={{height:1,background:C.border,margin:"10px 0"}}/>

          <div style={{display:"flex",gap:8,alignItems:"center",padding:"0 12px",marginBottom:10}}>
            <button onClick={useGPS} disabled={gpsBusy} style={{display:"flex",alignItems:"center",gap:6,padding:"8px 12px",borderRadius:8,border:`1px solid ${C.border}`,background:"transparent",color:C.teal,fontFamily:FH,fontWeight:700,fontSize:12.5,cursor:gpsBusy?"default":"pointer"}}>
              <LocateFixed size={13}/> {gpsBusy?t("nav.gpsLocating"):t("nav.gps")}
            </button>
            <select value={region ?? ""} onChange={(e)=>setRegion((e.target.value || null) as typeof region)}
              style={{flex:1,padding:"8px 10px",borderRadius:8,border:`1px solid ${C.border}`,background:C.s2,color:C.text,fontFamily:FB,fontSize:13}}>
              <option value="">{t("nav.allRegions")}</option>
              {REGION_ORDER.map((r) => <option key={r} value={r}>{t(REGION_LABEL[r])}</option>)}
            </select>
          </div>

          <div style={{display:"flex",padding:"0 12px",marginBottom:14}}>
            <div style={{display:"flex",borderRadius:8,overflow:"hidden",border:`1px solid ${C.border}`}}>
              <button onClick={()=>setLocale("tg")} style={{padding:"8px 14px",fontSize:12.5,fontWeight:700,color:locale==="tg"?C.bg:C.muted,fontFamily:FH,background:locale==="tg"?C.teal:"none",border:"none",cursor:"pointer"}}>ТҶ</button>
              <button onClick={()=>setLocale("ru")} style={{padding:"8px 14px",fontSize:12.5,fontWeight:700,color:locale==="ru"?C.bg:C.muted,fontFamily:FH,background:locale==="ru"?C.teal:"none",border:"none",cursor:"pointer"}}>РУ</button>
            </div>
          </div>

          <div style={{padding:"0 12px",display:"flex",flexDirection:"column",gap:8}}>
            {role!=="guest" && cabinetHref && (
              <Link href={cabinetHref} onClick={()=>setMobileOpen(false)} style={{textAlign:"center",padding:"11px 14px",borderRadius:9,fontFamily:FH,fontWeight:700,fontSize:14,color:C.teal,border:`1px solid ${C.teal}40`,background:`${C.teal}10`,textDecoration:"none"}}>
                {t("nav.cabinet")}
              </Link>
            )}
            {role==="guest" ? (
              <Link href="/login" onClick={()=>setMobileOpen(false)} style={{textAlign:"center",padding:"11px 14px",borderRadius:9,fontFamily:FH,fontWeight:700,fontSize:14,background:C.teal,color:C.bg,textDecoration:"none"}}>
                {t("nav.login")}
              </Link>
            ) : (
              <>
                {ROLE_OPTIONS.filter(o=>o.value!==role).map(opt=>(
                  <button key={opt.value} onClick={()=>{chooseRole(opt); setMobileOpen(false);}} style={{textAlign:"left",padding:"10px 14px",borderRadius:9,background:C.s2,border:`1px solid ${C.border}`,color:C.text,fontFamily:FB,fontSize:13.5,cursor:"pointer"}}>
                    {t(opt.key)}
                  </button>
                ))}
                <button onClick={()=>{logout(); setMobileOpen(false);}} style={{display:"flex",alignItems:"center",gap:6,textAlign:"left",padding:"10px 14px",borderRadius:9,background:"transparent",border:`1px solid ${C.border}`,color:C.red,fontFamily:FB,fontSize:13.5,cursor:"pointer"}}>
                  <LogOut size={13}/> {t("nav.logout")}
                </button>
              </>
            )}
          </div>
        </div>
      )}

      <style jsx>{`
        @media (max-width: 880px) {
          .eh-nav-right-mobile { display: flex !important; }
        }
      `}</style>
    </header>
  );
}
