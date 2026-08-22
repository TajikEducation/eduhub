"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams, usePathname, useRouter, useSearchParams } from "next/navigation";
import { ArrowLeft, MapPin, Phone, Globe, Star, CheckCircle, Award, Users, Utensils, Camera, MessageSquare, Newspaper, BookOpen, ChevronRight, Mail, UserSquare2, Medal, Trophy, PenLine, Briefcase, Wallet, Clock, type LucideIcon } from "lucide-react";
import { C, FH, FB, INSTITUTIONS, ALL_STAFF, REVIEWS, NEWS_ITEMS, VACANCIES, CATEGORY_META, REGION_LABEL, AMENITY_INFO, type AmenityKey, type Review } from "@/lib/data";
import { chatHref } from "@/lib/chat-window";
import { useReveal, revealStyle } from "@/lib/useReveal";
import { AchievementModal } from "@/components/AchievementModal";
import { AmenityModal } from "@/components/AmenityModal";
import { Toast } from "@/components/Toast";
import { useAppState } from "@/lib/app-state";
import { useT } from "@/lib/i18n";
import { SingleMap } from "@/components/InstMap";

type Tab = "about"|"staff"|"gallery"|"achievements"|"alumni"|"menu"|"reviews"|"news"|"vacancies";

const ACH_TIER: Record<string,{icon:LucideIcon;color:string}> = {
  gold:{icon:Medal,color:C.gold}, silver:{icon:Medal,color:C.sub}, bronze:{icon:Medal,color:C.goldD}, special:{icon:Trophy,color:C.teal},
};
const SOCIAL_META = {
  instagram: { abbr:"IG", color:"#E1306C", url:(h:string)=>`https://instagram.com/${h}` },
  telegram:  { abbr:"TG", color:"#229ED9", url:(h:string)=>`https://t.me/${h}` },
  facebook:  { abbr:"FB", color:"#1877F2", url:(h:string)=>`https://facebook.com/${h}` },
} as const;
const ACH_CATEGORIES = [
  {key:"institution" as const, labelKey:"ach.institution"},
  {key:"student"     as const, labelKey:"ach.student"},
  {key:"teacher"     as const, labelKey:"ach.teacher"},
];

function Stars({ s, size=13 }:{s:number;size?:number}) {
  return (
    <span style={{display:"flex",gap:2}}>
      {[1,2,3,4,5].map(i=>(<svg key={i} width={size} height={size} viewBox="0 0 24 24" fill={i<=Math.round(s)?C.gold:C.dim}><path d="M12 2l3 7h7l-5.5 4 2 7L12 16l-6.5 4 2-7L2 9h7z"/></svg>))}
    </span>
  );
}

function MetricBar({ label, value, t }:{label:string;value:number;t:(x:string)=>string}) {
  const pct = (value/5)*100;
  const color = pct>=90?C.ok:pct>=75?C.teal:pct>=60?C.gold:C.red;
  return (
    <div style={{marginBottom:10}}>
      <div style={{display:"flex",justifyContent:"space-between",fontSize:12.5,color:C.sub,marginBottom:5,fontFamily:FH}}>
        <span>{t(label)}</span><span style={{color,fontWeight:700}}>{value.toFixed(1)}</span>
      </div>
      <div style={{height:5,borderRadius:999,background:C.s3,overflow:"hidden"}}>
        <div style={{height:"100%",width:"100%",borderRadius:999,background:color,transform:`scaleX(${pct/100})`,transformOrigin:"left",transition:"transform .5s"}}/>
      </div>
    </div>
  );
}

export default function InstitutionProfilePage() {
  const params = useParams<{ id: string }>();
  return <InstitutionProfileInner key={params.id} />;
}

function InstitutionProfileInner() {
  const params = useParams<{ id: string }>();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const router = useRouter();
  const instId = Number(params.id);
  const t = useT();
  const { children_: children, hasApplied, addApplication, myInstitution } = useAppState();
  const { ref: staffRef, visible: staffVisible } = useReveal<HTMLDivElement>();

  const [tab, setTab] = useState<Tab>(() => (searchParams.get("tab") as Tab) ?? "about");
  const [lightbox, setLightbox] = useState<string|null>(null);
  const [showFullBio, setShowFullBio] = useState(false);
  const [amenityOpen, setAmenityOpen] = useState<AmenityKey|null>(null);
  const [addedReviews, setAddedReviews] = useState<Review[]>([]);
  const [showReviewForm, setShowReviewForm] = useState(false);
  const [reviewName, setReviewName] = useState("");
  const [metricScores, setMetricScores] = useState<number[]>(() => INSTITUTIONS.find(i=>i.id===instId)?.metrics.map(()=>5) ?? []);
  const [reviewText, setReviewText] = useState("");
  const [reviewToast, setReviewToast] = useState<string|null>(null);
  const [visitName, setVisitName] = useState("");
  const [visitPhone, setVisitPhone] = useState("");
  const [visitDate, setVisitDate] = useState("");

  const highlightReviewId = searchParams.get("review");
  const highlightVacancyId = searchParams.get("vacancy");

  useEffect(() => {
    if (!highlightReviewId || tab !== "reviews") return;
    const el = document.getElementById(`review-${highlightReviewId}`);
    el?.scrollIntoView({ behavior: "smooth", block: "center" });
  }, [highlightReviewId, tab]);

  useEffect(() => {
    if (!highlightVacancyId || tab !== "vacancies") return;
    const el = document.getElementById(`vacancy-${highlightVacancyId}`);
    el?.scrollIntoView({ behavior: "smooth", block: "center" });
  }, [highlightVacancyId, tab]);

  const seedInst = INSTITUTIONS.find(i=>i.id===instId);
  const inst = seedInst ?? (myInstitution?.id===instId ? myInstitution : INSTITUTIONS[0]);
  const staff = ALL_STAFF.filter(p=>p.instId===inst.id);
  const reviews = [...addedReviews, ...REVIEWS.filter(r=>r.instId===inst.id)];
  const news = NEWS_ITEMS.filter(n => n.status === "published" && n.instId === inst.id);
  const instVacancies = VACANCIES.filter(v => v.status === "published" && v.instId === inst.id);
  const meta = CATEGORY_META[inst.tk];
  const showAlumni = inst.tk==="cat_school" || inst.tk==="cat_uni";
  const descRu = inst.description.ru;

  // SRS FR-15/FR-30 — оставить отзыв может тот, у кого есть ребёнок, структурно
  // привязанный к этому учреждению (текущий ученик или выпускник). Это факт о
  // пользователе (ChildLink), не активная демо-роль (модель B) — иначе пользователь,
  // переключившийся в другой раздел, терял бы право на свой же отзыв.
  const hasConnection = children.some(c => c.instId === inst.id);

  function submitReview() {
    if (!reviewName.trim() || !reviewText.trim() || !hasConnection) return;
    const name = reviewName.trim();
    const text = reviewText.trim();
    const score = Math.round((metricScores.reduce((a,b)=>a+b,0) / metricScores.length) * 10) / 10;
    setAddedReviews(prev => [{
      id: `local-${prev.length}`,
      name: { ru: name, tg: name },
      init: name[0]?.toUpperCase() ?? "?",
      score,
      metrics: metricScores,
      date: new Date().toLocaleDateString("ru-RU"),
      text: { ru: text, tg: text },
      reply: null,
      instId: inst.id,
    }, ...prev]);
    setReviewName(""); setMetricScores(inst.metrics.map(()=>5)); setReviewText("");
    setShowReviewForm(false);
    setReviewToast(t({ ru: "Отзыв опубликован", tg: "Шарҳ нашр шуд" }));
  }

  const TABS: {k:Tab; label:string; icon:LucideIcon}[] = [
    {k:"about",       label:t("tab.about"),       icon:BookOpen},
    {k:"staff",       label:t("tab.staff"),       icon:Users},
    {k:"gallery",     label:t("tab.gallery"),     icon:Camera},
    {k:"achievements",label:t("tab.achievements"),icon:Award},
    ...(showAlumni ? [{k:"alumni" as Tab, label:t("tab.alumni"), icon:UserSquare2}] : []),
    {k:"menu",        label:t("tab.menu"),        icon:Utensils},
    {k:"reviews",     label:t("tab.reviews"),     icon:Star},
    {k:"news",        label:t("tab.news"),        icon:Newspaper},
    {k:"vacancies",   label:t("nav.vacancies"),   icon:Briefcase},
  ];

  return (
    <div style={{fontFamily:FB}}>
      {/* ── COVER ── */}
      <div style={{position:"relative",height:340,overflow:"hidden"}}>
        <img src={inst.coverPhoto} alt={t(inst.name)} style={{width:"100%",height:"100%",objectFit:"cover"}}/>
        <div style={{position:"absolute",inset:0,background:`linear-gradient(180deg,${C.overlay}26 0%,${C.overlay}D9 75%,${C.overlay} 100%)`}}/>
        <div style={{position:"absolute",top:20,left:28}}>
          <button onClick={()=>router.back()} style={{display:"inline-flex",alignItems:"center",gap:7,padding:"8px 14px",borderRadius:8,background:"rgba(0,0,0,.5)",border:`1px solid rgba(255,255,255,.15)`,color:"#fff",fontFamily:FH,fontWeight:700,fontSize:13,backdropFilter:"blur(8px)",cursor:"pointer"}}>
            <ArrowLeft size={14}/> {t("common.back")}
          </button>
        </div>
        <div style={{position:"absolute",bottom:24,left:28,right:28,display:"flex",alignItems:"flex-end",justifyContent:"space-between"}}>
          <div>
            <div style={{display:"flex",gap:8,marginBottom:10,flexWrap:"wrap"}}>
              <span style={{fontSize:12,fontWeight:700,padding:"4px 10px",borderRadius:7,background:inst.color,color:C.overlay,fontFamily:FH}}>{t(meta.label)}</span>
              {inst.ver && <Link href="/company" style={{fontSize:12,fontWeight:700,padding:"4px 10px",borderRadius:7,background:`${C.ok}22`,border:`1px solid ${C.ok}`,color:C.ok,fontFamily:FH,display:"flex",alignItems:"center",gap:4,textDecoration:"none"}}><CheckCircle size={11}/> {t("common.verified")}</Link>}
              {inst.tag && <span style={{fontSize:12,fontWeight:700,padding:"4px 10px",borderRadius:7,background:`${C.gold}22`,border:`1px solid ${C.gold}`,color:C.gold,fontFamily:FH}}>{t(inst.tag)}</span>}
            </div>
            <h1 style={{fontFamily:FH,fontWeight:900,fontSize:"clamp(22px,3vw,36px)",color:"#fff",lineHeight:1.1,marginBottom:6}}>{t(inst.name)}</h1>
            <div style={{display:"flex",alignItems:"center",gap:12,flexWrap:"wrap"}}>
              <div style={{display:"flex",alignItems:"center",gap:6}}>
                <Stars s={inst.score} size={15}/>
                <span style={{fontFamily:FH,fontWeight:800,fontSize:15,color:"#fff"}}>{inst.score}</span>
                <span style={{fontSize:13,color:"rgba(255,255,255,.55)"}}>({inst.rev} {t("common.reviews")})</span>
              </div>
              <span style={{fontSize:13,color:"rgba(255,255,255,.6)",display:"flex",alignItems:"center",gap:4}}><MapPin size={12}/>{inst.area}, {t(inst.city)}</span>
            </div>
          </div>
          <div style={{textAlign:"right"}}>
            <p style={{fontFamily:FH,fontWeight:900,fontSize:30,color:"#fff"}}>{inst.price} <span style={{fontSize:14,fontWeight:400,color:"rgba(255,255,255,.55)"}}>{t("common.perMonth")}</span></p>
          </div>
        </div>
      </div>

      {/* ── TABS ── */}
      <div style={{borderBottom:`1px solid ${C.border}`,background:C.bg,position:"sticky",top:64,zIndex:40}}>
        <div style={{maxWidth:1260,margin:"0 auto",padding:"0 28px",display:"flex",gap:2,overflowX:"auto"}}>
          {TABS.map(({k,label,icon:Icon})=>(
            <button key={k} onClick={()=>{setTab(k); window.scrollTo({top:0,behavior:"auto"});}}
              style={{display:"flex",alignItems:"center",gap:6,padding:"14px 14px",fontFamily:FH,fontWeight:700,fontSize:13,color:tab===k?C.teal:C.muted,borderBottom:`2px solid ${tab===k?C.teal:"transparent"}`,whiteSpace:"nowrap",background:"none",borderLeft:"none",borderRight:"none",borderTop:"none",cursor:"pointer"}}>
              <Icon size={14}/> {label}
            </button>
          ))}
        </div>
      </div>

      {/* ── CONTENT ── */}
      <div style={{maxWidth:1260,margin:"0 auto",padding:"32px 28px 80px"}}>

        {/* ABOUT */}
        {tab==="about" && (
          <div style={{display:"grid",gridTemplateColumns:"1fr 320px",gap:24,alignItems:"start"}}>
            <div style={{display:"flex",flexDirection:"column",gap:20}}>
              <div style={{borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:26}}>
                <h2 style={{fontFamily:FH,fontWeight:800,fontSize:19,color:C.text,marginBottom:14}}>{t("tab.about")}</h2>
                <p style={{fontSize:14.5,color:C.sub,lineHeight:1.78,maxHeight:showFullBio?undefined:120,overflow:showFullBio?undefined:"hidden"}}>
                  {t(inst.description)}
                </p>
                {descRu.length > 200 && (
                  <button onClick={()=>setShowFullBio(!showFullBio)} style={{marginTop:10,fontSize:13,color:C.teal,fontFamily:FH,fontWeight:600,background:"none",border:"none",cursor:"pointer"}}>
                    {showFullBio?t("common.collapse"):t("common.readMore")}
                  </button>
                )}
              </div>

              {/* key facts */}
              <div style={{display:"grid",gridTemplateColumns:staff.length>0?"repeat(4,1fr)":"repeat(3,1fr)",gap:12}}>
                {[
                  {l:t({ru:"Основана",tg:"Таъсисёфта"}),v:inst.founded},
                  {l:t({ru:"Учеников",tg:"Хонандагон"}),v:inst.students},
                  {l:t({ru:"Возраст",tg:"Синну сол"}),v:inst.age},
                  ...(staff.length>0 ? [{l:t({ru:"Учеников на педагога",tg:"Хонанда ба як омӯзгор"}),v:`${Math.round(inst.students/staff.length)}:1`}] : []),
                ].map(({l,v})=>(
                  <div key={l} style={{borderRadius:14,border:`1px solid ${C.border}`,background:C.s1,padding:"16px",textAlign:"center"}}>
                    <p style={{fontFamily:FH,fontWeight:900,fontSize:22,color:C.teal,marginBottom:4}}>{v}</p>
                    <p style={{fontSize:12.5,color:C.sub}}>{l}</p>
                  </div>
                ))}
              </div>

              {/* geo */}
              <div style={{borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:26}}>
                <h3 style={{fontFamily:FH,fontWeight:800,fontSize:17,color:C.text,marginBottom:16}}>{t({ru:"Расположение",tg:"Ҷойгиршавӣ"})}</h3>
                <div style={{display:"grid",gridTemplateColumns:"repeat(4,1fr)",gap:12}}>
                  {[
                    {l:t("geo.region"),v:t(REGION_LABEL[inst.region])},
                    {l:t("geo.city"),v:t(inst.city)},
                    {l:t("geo.district"),v:inst.area},
                    {l:t("geo.street"),v:t(inst.street)},
                  ].map(g=>(
                    <div key={g.l}>
                      <p style={{fontSize:11,fontWeight:700,color:C.dim,textTransform:"uppercase",letterSpacing:".05em",fontFamily:FH,marginBottom:4}}>{g.l}</p>
                      <p style={{fontSize:13.5,color:C.text}}>{g.v}</p>
                    </div>
                  ))}
                </div>
                {inst.geo && (
                  <div style={{marginTop:18}}>
                    <SingleMap lat={inst.geo.lat} lng={inst.geo.lng} routeLabel={t({ru:"Построить маршрут",tg:"Роҳро тартиб додан"})} />
                  </div>
                )}
              </div>

              {/* metrics */}
              <div style={{borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:26}}>
                <div style={{display:"flex",justifyContent:"space-between",alignItems:"center",marginBottom:16,gap:12,flexWrap:"wrap"}}>
                  <h3 style={{fontFamily:FH,fontWeight:800,fontSize:17,color:C.text}}>{t({ru:"Рейтинг по параметрам",tg:"Рейтинг аз рӯи меъёрҳо"})}</h3>
                  <button
                    onClick={()=>{ setTab("reviews"); setShowReviewForm(true); window.scrollTo({top:0,behavior:"auto"}); }}
                    style={{display:"flex",alignItems:"center",gap:6,padding:"8px 14px",borderRadius:10,background:`${C.teal}18`,border:`1px solid ${C.teal}44`,color:C.teal,fontFamily:FH,fontWeight:700,fontSize:12.5,cursor:"pointer"}}>
                    <PenLine size={13}/> {t({ru:"Оценить учреждение",tg:"Муассисаро баҳо додан"})}
                  </button>
                </div>
                <div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:"0 28px"}}>
                  {inst.metrics.map(m=>(
                    <MetricBar key={m.label} label={m.label} value={m.v} t={t}/>
                  ))}
                </div>
              </div>

              {/* features */}
              <div style={{borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:26}}>
                <h3 style={{fontFamily:FH,fontWeight:800,fontSize:17,color:C.text,marginBottom:16}}>{t({ru:"Услуги и возможности",tg:"Хизматрасонӣ ва имконот"})}</h3>
                <div style={{display:"grid",gridTemplateColumns:"repeat(2,1fr)",gap:10}}>
                  {([
                    {key:"transport" as AmenityKey,v:inst.transport},
                    {key:"food" as AmenityKey,v:inst.food},
                  ]).map(f=>{
                    const info = AMENITY_INFO[f.key];
                    const Icon = info.icon;
                    return (
                      <button key={f.key} onClick={()=>f.key==="food" ? setTab("menu") : setAmenityOpen(f.key)}
                        style={{display:"flex",alignItems:"center",gap:8,padding:"10px 12px",borderRadius:12,background:f.v?`${C.ok}12`:C.s2,border:`1px solid ${f.v?C.ok+"33":C.border}`,cursor:"pointer",textAlign:"left"}}>
                        <Icon size={16} style={{color:f.v?C.ok:C.dim,flexShrink:0}}/>
                        <span style={{fontSize:12.5,fontFamily:FH,fontWeight:600,color:f.v?C.text:C.dim}}>{t(info.title)}</span>
                        {f.v && <CheckCircle size={11} style={{color:C.ok,marginLeft:"auto",flexShrink:0}}/>}
                      </button>
                    );
                  })}
                </div>
              </div>
            </div>

            {/* sidebar */}
            <div style={{display:"flex",flexDirection:"column",gap:16}}>
              <div style={{borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:22}}>
                <h3 style={{fontFamily:FH,fontWeight:800,fontSize:16,color:C.text,marginBottom:14}}>{t({ru:"Контакты",tg:"Тамос"})}</h3>
                {[
                  {icon:<Phone size={15}/>,l:inst.phone},
                  {icon:<Mail size={15}/>,l:inst.email},
                  {icon:<MapPin size={15}/>,l:t(inst.address)},
                ].map((c,i)=>(
                  <div key={i} style={{display:"flex",alignItems:"flex-start",gap:10,fontSize:13.5,color:C.sub,padding:"8px 0",borderBottom:`1px solid ${C.border}`}}>
                    <span style={{color:C.teal,flexShrink:0,marginTop:1}}>{c.icon}</span>{c.l}
                  </div>
                ))}

                {inst.website && (
                  <a href={`https://${inst.website}`} target="_blank" rel="noopener noreferrer"
                    style={{display:"flex",alignItems:"center",gap:8,marginTop:12,padding:"10px 14px",borderRadius:12,background:`${C.teal}14`,border:`1px solid ${C.teal}33`,color:C.teal,textDecoration:"none",fontFamily:FH,fontWeight:700,fontSize:13.5}}>
                    <Globe size={14}/> {inst.website}
                  </a>
                )}

                {inst.socials && (inst.socials.instagram || inst.socials.telegram || inst.socials.facebook) && (
                  <div style={{display:"flex",gap:8,marginTop:10}}>
                    {(Object.keys(SOCIAL_META) as (keyof typeof SOCIAL_META)[]).map(key=>{
                      const handle = inst.socials?.[key];
                      if(!handle) return null;
                      const sm = SOCIAL_META[key];
                      return (
                        <a key={key} href={sm.url(handle)} target="_blank" rel="noopener noreferrer" aria-label={key}
                          style={{width:36,height:36,borderRadius:8,background:`${sm.color}18`,border:`1px solid ${sm.color}44`,display:"flex",alignItems:"center",justifyContent:"center",color:sm.color,fontFamily:FH,fontWeight:800,fontSize:10.5,textDecoration:"none"}}>
                          {sm.abbr}
                        </a>
                      );
                    })}
                  </div>
                )}

                <Link href={chatHref(inst.id)} style={{marginTop:14,width:"100%",padding:"11px",borderRadius:12,background:C.teal,color:C.overlay,fontFamily:FH,fontWeight:800,fontSize:14,display:"flex",alignItems:"center",justifyContent:"center",gap:7,cursor:"pointer",border:"none",textDecoration:"none"}}>
                  <MessageSquare size={15}/> {t("common.writeInChat")}
                </Link>
              </div>

              <div style={{borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:22}}>
                <h3 style={{fontFamily:FH,fontWeight:800,fontSize:16,color:C.text,marginBottom:14}}>{t({ru:"Режим работы",tg:"Тартиби корӣ"})}</h3>
                {[["hours.weekdays","07:30 – 18:00"],["hours.saturday","08:00 – 14:00"],["hours.sunday","hours.dayoff"]].map(([dKey,tVal])=>(
                  <div key={dKey} style={{display:"flex",justifyContent:"space-between",fontSize:13.5,padding:"7px 0",borderBottom:`1px solid ${C.border}`,color:C.sub}}>
                    <span>{t(dKey)}</span><span style={{color:tVal==="hours.dayoff"?C.red:C.text,fontWeight:600}}>{tVal.startsWith("hours.")?t(tVal):tVal}</span>
                  </div>
                ))}
              </div>

              <div style={{borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:22}}>
                <h3 style={{fontFamily:FH,fontWeight:800,fontSize:16,color:C.text,marginBottom:14}}>{t({ru:"Записаться на визит",tg:"Ба ташриф сабти ном шудан"})}</h3>
                <input value={visitName} onChange={e=>setVisitName(e.target.value)} placeholder={t({ru:"Ваше имя",tg:"Номи шумо"})} style={{display:"block",width:"100%",padding:"9px 12px",borderRadius:9,border:`1px solid ${C.border}`,background:C.s2,color:C.text,fontFamily:FB,fontSize:13.5,outline:"none",marginBottom:8,boxSizing:"border-box"}}/>
                <input value={visitPhone} onChange={e=>setVisitPhone(e.target.value)} placeholder={t({ru:"Телефон",tg:"Телефон"})} style={{display:"block",width:"100%",padding:"9px 12px",borderRadius:9,border:`1px solid ${C.border}`,background:C.s2,color:C.text,fontFamily:FB,fontSize:13.5,outline:"none",marginBottom:8,boxSizing:"border-box"}}/>
                <input value={visitDate} onChange={e=>setVisitDate(e.target.value)} type="date" placeholder={t({ru:"Дата",tg:"Сана"})} style={{display:"block",width:"100%",padding:"9px 12px",borderRadius:9,border:`1px solid ${C.border}`,background:C.s2,color:C.text,fontFamily:FB,fontSize:13.5,outline:"none",marginBottom:8,boxSizing:"border-box"}}/>
                <button
                  disabled={!visitName.trim() || !visitPhone.trim()}
                  onClick={()=>{
                    setReviewToast(t({ ru: "Заявка на визит отправлена", tg: "Дархости ташриф фиристода шуд" }));
                    setVisitName(""); setVisitPhone(""); setVisitDate("");
                  }}
                  style={{width:"100%",padding:11,borderRadius:10,background:C.teal,color:C.overlay,fontFamily:FH,fontWeight:800,fontSize:13.5,border:"none",cursor:"pointer",marginTop:4,opacity:(!visitName.trim()||!visitPhone.trim())?0.5:1}}>
                  {t({ru:"Отправить заявку",tg:"Аризаро фиристодан"})}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* STAFF */}
        {tab==="staff" && (
          <div>
            <h2 style={{fontFamily:FH,fontWeight:800,fontSize:22,color:C.text,marginBottom:24}}>{t({ru:"Педагогический состав",tg:"Ҳайати педагогӣ"})}</h2>
            {staff.length === 0 ? (
              <p style={{color:C.muted,fontSize:14}}>{t("empty.staff")}</p>
            ) : (
              <div ref={staffRef} style={{display:"grid",gridTemplateColumns:"repeat(3,1fr)",gap:18}}>
                {staff.map((p,i)=>(
                  <button key={p.id} onClick={()=>router.push(`/people/${p.id}`)}
                    style={{textAlign:"left",borderRadius:18,overflow:"hidden",border:`1px solid ${C.border}`,background:C.s1,cursor:"pointer",...revealStyle(staffVisible,i*70)}}>
                    <div style={{height:220,overflow:"hidden",position:"relative"}}>
                      <img src={p.photo} alt={t(p.name)} style={{width:"100%",height:"100%",objectFit:"cover",transition:"transform .3s"}}/>
                      <div style={{position:"absolute",inset:0,background:`linear-gradient(180deg,transparent 55%,${C.overlay}E6 100%)`}}/>
                      <div style={{position:"absolute",bottom:14,left:14,right:14}}>
                        <p style={{fontFamily:FH,fontWeight:800,fontSize:16,color:"#fff"}}>{t(p.name)}</p>
                        <p style={{fontSize:12.5,color:C.teal,fontFamily:FH}}>{t(p.roleLabel)}</p>
                      </div>
                    </div>
                    <div style={{padding:"14px 16px"}}>
                      {p.subject && (
                        <div style={{display:"flex",gap:6,marginBottom:10}}>
                          <span style={{fontSize:11,fontWeight:600,padding:"3px 8px",borderRadius:6,background:`${C.teal}18`,color:C.teal,fontFamily:FH}}>{t(p.subject)}</span>
                        </div>
                      )}
                      <div style={{display:"flex",justifyContent:"space-between",alignItems:"center"}}>
                        <span style={{fontSize:13,color:C.sub}}>{p.exp} {t({ru:"опыта",tg:"таҷриба"})}</span>
                        <span style={{display:"flex",alignItems:"center",gap:4,fontSize:12.5,color:C.teal,fontFamily:FH,fontWeight:600}}>{t({ru:"Подробнее",tg:"Муфассал"})} <ChevronRight size={13}/></span>
                      </div>
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>
        )}

        {/* GALLERY */}
        {tab==="gallery" && (
          <div>
            <h2 style={{fontFamily:FH,fontWeight:800,fontSize:22,color:C.text,marginBottom:24}}>{t({ru:"Фотогалерея",tg:"Галереяи расмҳо"})}</h2>
            <div style={{display:"grid",gridTemplateColumns:"repeat(3,1fr)",gap:10}}>
              {inst.gallery.map((g,i)=>(
                <div key={i} onClick={()=>setLightbox(g.url)} style={{aspectRatio:"4/3",borderRadius:12,overflow:"hidden",cursor:"pointer",position:"relative"}}>
                  <img src={g.url} alt={t(g.label)} style={{width:"100%",height:"100%",objectFit:"cover",transition:"transform .3s"}}
                    onMouseEnter={e=>(e.currentTarget.style.transform="scale(1.05)")}
                    onMouseLeave={e=>(e.currentTarget.style.transform="scale(1)")}/>
                  <div style={{position:"absolute",bottom:0,left:0,right:0,padding:"6px 10px",background:"linear-gradient(transparent,rgba(0,0,0,.6))",fontSize:12.5,color:"rgba(255,255,255,.8)",fontFamily:FH}}>{t(g.label)}</div>
                </div>
              ))}
            </div>
            {lightbox && (
              <div onClick={()=>setLightbox(null)} style={{position:"fixed",inset:0,background:"rgba(0,0,0,.9)",zIndex:999,display:"flex",alignItems:"center",justifyContent:"center",cursor:"pointer"}}>
                <img src={lightbox} alt="" style={{maxWidth:"92vw",maxHeight:"92vh",objectFit:"contain",borderRadius:12}}/>
              </div>
            )}
          </div>
        )}

        {/* ACHIEVEMENTS */}
        {tab==="achievements" && (
          <div>
            <h2 style={{fontFamily:FH,fontWeight:800,fontSize:22,color:C.text,marginBottom:24}}>{t("tab.achievements")}</h2>
            {ACH_CATEGORIES.map(cat=>{
              const items = inst.achievements.filter(a=>a.type===cat.key);
              if(!items.length) return null;
              return (
                <div key={cat.key} style={{marginBottom:32}}>
                  <h3 style={{fontFamily:FH,fontWeight:700,fontSize:17,color:C.teal,marginBottom:14}}>{t(cat.labelKey)}</h3>
                  <div style={{display:"grid",gridTemplateColumns:"repeat(3,1fr)",gap:12}}>
                    {items.map(ach=>{
                      const p = new URLSearchParams(searchParams.toString());
                      p.set("achievement", ach.id);
                      const tier = ACH_TIER[ach.category];
                      const TierIcon = tier.icon;
                      return (
                        <Link key={ach.id} href={`${pathname}?${p.toString()}`} scroll={false}
                          style={{display:"block",borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:"18px 20px",textDecoration:"none",cursor:"pointer",transition:"border-color .15s,transform .15s"}}>
                          <TierIcon size={24} style={{color:tier.color,display:"block",marginBottom:10}}/>
                          <p style={{fontFamily:FH,fontWeight:700,fontSize:15,color:C.text,marginBottom:4}}>{t(ach.title)}</p>
                          <p style={{fontSize:12.5,color:C.sub}}>{ach.year}</p>
                          <p style={{fontSize:12.5,color:C.muted,marginTop:4}}>{t(ach.desc)}</p>
                        </Link>
                      );
                    })}
                  </div>
                </div>
              );
            })}
            {!inst.achievements.length && <p style={{color:C.muted,fontSize:14}}>{t("empty.achievements")}</p>}
            <AchievementModal achievements={inst.achievements}/>
          </div>
        )}

        {/* ALUMNI */}
        {tab==="alumni" && (
          <div>
            <h2 style={{fontFamily:FH,fontWeight:800,fontSize:22,color:C.text,marginBottom:24}}>{t("tab.alumni")}</h2>
            {!inst.alumni || inst.alumni.length===0 ? (
              <p style={{color:C.muted,fontSize:14}}>{t({ru:"Выпускники не добавлены",tg:"Хатмкунандагон илова нашудаанд"})}</p>
            ) : (
              <div style={{display:"grid",gridTemplateColumns:"repeat(4,1fr)",gap:16}}>
                {inst.alumni.map(a=>(
                  <div key={a.id} style={{borderRadius:16,border:`1px solid ${C.border}`,background:C.s1,padding:18,textAlign:"center"}}>
                    <img src={a.photo} alt={t(a.name)} style={{width:64,height:64,borderRadius:"50%",objectFit:"cover",margin:"0 auto 12px"}}/>
                    <p style={{fontFamily:FH,fontWeight:700,fontSize:14,color:C.text,marginBottom:3}}>{t(a.name)}</p>
                    <p style={{fontSize:12,color:C.dim,marginBottom:8}}>{t({ru:"Выпуск",tg:"Хатм"})} {a.gradYear}</p>
                    <p style={{fontSize:12.5,color:C.sub,lineHeight:1.5}}>{t(a.now)}</p>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* MENU */}
        {tab==="menu" && (
          <div>
            <h2 style={{fontFamily:FH,fontWeight:800,fontSize:22,color:C.text,marginBottom:6}}>{t({ru:"Меню питания",tg:"Менюи ғизо"})}</h2>
            <p style={{fontSize:14,color:C.sub,marginBottom:24}}>{t({ru:"Еженедельное расписание питания",tg:"Ҷадвали ҳафтаинаи ғизо"})}</p>
            {!inst.food ? (
              <div style={{padding:48,borderRadius:16,border:`1px dashed ${C.border}`,textAlign:"center",color:C.muted}}>
                <Utensils size={28} style={{margin:"0 auto 12px",color:C.dim}}/>
                <p style={{fontFamily:FH,fontWeight:700,fontSize:16,color:C.text,marginBottom:6}}>{t({ru:"Питание не предусмотрено",tg:"Ғизо пешбинӣ нашудааст"})}</p>
                <p style={{fontSize:13.5}}>{t({ru:"В данном учреждении нет организованного питания",tg:"Дар ин муассиса ғизои муташаккил нест"})}</p>
              </div>
            ) : inst.foodMenu ? (
              <div style={{display:"grid",gridTemplateColumns:`repeat(${inst.foodMenu.length},1fr)`,gap:14}}>
                {inst.foodMenu.map(dayMenu=>(
                  <div key={dayMenu.day} style={{borderRadius:16,border:`1px solid ${C.border}`,background:C.s1,overflow:"hidden"}}>
                    <div style={{padding:"10px 14px",background:C.s2,borderBottom:`1px solid ${C.border}`}}>
                      <p style={{fontFamily:FH,fontWeight:800,fontSize:13,color:C.teal}}>{dayMenu.day}</p>
                    </div>
                    <div style={{padding:"12px 14px",display:"flex",flexDirection:"column",gap:10}}>
                      {dayMenu.meals.map((meal,mi)=>(
                        <div key={mi}>
                          <p style={{fontSize:12.5,color:C.sub,marginBottom:4}}>{meal.name}</p>
                          <p style={{fontSize:11.5,color:C.muted}}>{meal.cal}</p>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div style={{display:"grid",gridTemplateColumns:"repeat(5,1fr)",gap:14}}>
                {["menu.mon","menu.tue","menu.wed","menu.thu","menu.fri"].map(dayKey=>(
                  <div key={dayKey} style={{borderRadius:16,border:`1px solid ${C.border}`,background:C.s1,overflow:"hidden"}}>
                    <div style={{padding:"10px 14px",background:C.s2,borderBottom:`1px solid ${C.border}`}}>
                      <p style={{fontFamily:FH,fontWeight:800,fontSize:13,color:C.teal}}>{t(dayKey)}</p>
                    </div>
                    <div style={{padding:"12px 14px",display:"flex",flexDirection:"column",gap:8}}>
                      {[["menu.breakfast","menu.breakfastTxt"],["menu.lunch","menu.lunchTxt"],["menu.snack","menu.snackTxt"]].map(([mealKey,txtKey])=>(
                        <div key={mealKey}>
                          <p style={{fontSize:11,fontWeight:700,textTransform:"uppercase",color:C.dim,fontFamily:FH,marginBottom:3}}>{t(mealKey)}</p>
                          <p style={{fontSize:12.5,color:C.sub}}>{t(txtKey)}</p>
                        </div>
                      ))}
                      <p style={{fontSize:11.5,color:C.muted,marginTop:4,borderTop:`1px solid ${C.border}`,paddingTop:6}}>~650 {t({ru:"ккал",tg:"ккал"})}</p>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* REVIEWS */}
        {tab==="reviews" && (
          <div>
            <div style={{display:"grid",gridTemplateColumns:"240px 1fr",gap:24,marginBottom:32}}>
              <div style={{borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:24,textAlign:"center"}}>
                <p style={{fontFamily:FH,fontWeight:900,fontSize:52,color:C.text,lineHeight:1}}>{inst.score}</p>
                <Stars s={inst.score} size={18}/>
                <p style={{fontSize:13,color:C.sub,marginTop:8}}>{inst.rev} {t("common.reviews")}</p>
                <p style={{fontSize:11,color:C.dim,marginTop:6}}>{t({ru:"8 независимых метрик",tg:"8 меъёри мустақил"})}</p>
                <div style={{marginTop:16}}>
                  {[5,4,3,2,1].map(n=>{
                    const countN = reviews.filter(r=>Math.round(r.score)===n).length;
                    const pct = reviews.length ? Math.round((countN/reviews.length)*100) : 0;
                    return (
                      <div key={n} style={{display:"flex",alignItems:"center",gap:8,marginBottom:4}}>
                        <span style={{fontSize:12.5,color:C.sub,width:12}}>{n}</span>
                        <div style={{flex:1,height:6,borderRadius:999,background:C.s3}}>
                          <div style={{height:"100%",borderRadius:999,background:C.gold,width:`${pct}%`}}/>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
              <div style={{display:"flex",flexDirection:"column",gap:14}}>
                {hasConnection ? (
                  <div style={{display:"flex",justifyContent:"flex-end"}}>
                    <button onClick={()=>setShowReviewForm(v=>!v)} style={{display:"flex",alignItems:"center",gap:7,padding:"9px 16px",borderRadius:10,background:showReviewForm?C.s3:`${C.teal}18`,border:`1px solid ${showReviewForm?C.border:C.teal+"44"}`,color:showReviewForm?C.sub:C.teal,fontFamily:FH,fontWeight:700,fontSize:13,cursor:"pointer"}}>
                      <PenLine size={14}/> {t({ru:"Написать отзыв",tg:"Шарҳ навиштан"})}
                    </button>
                  </div>
                ) : (
                  <div style={{borderRadius:14,border:`1px dashed ${C.border}`,background:C.s1,padding:"14px 18px",display:"flex",alignItems:"center",gap:12}}>
                    <PenLine size={16} style={{color:C.dim,flexShrink:0}}/>
                    <div style={{flex:1}}>
                      <p style={{fontSize:13,color:C.sub,lineHeight:1.5}}>{t("review.needConnection")}</p>
                      <button onClick={()=>router.push("/account")} style={{marginTop:6,background:"none",border:"none",color:C.teal,fontFamily:FH,fontWeight:700,fontSize:12.5,cursor:"pointer",padding:0}}>
                        {t("review.linkChild")}
                      </button>
                    </div>
                  </div>
                )}

                {showReviewForm && hasConnection && (
                  <div style={{borderRadius:16,border:`1px solid ${C.teal}44`,background:C.s1,padding:"18px 20px",display:"flex",flexDirection:"column",gap:12}}>
                    <input value={reviewName} onChange={e=>setReviewName(e.target.value)} placeholder={t({ru:"Ваше имя",tg:"Номи шумо"})}
                      style={{padding:"10px 14px",borderRadius:10,border:`1px solid ${C.border}`,background:C.s2,color:C.text,fontFamily:FB,fontSize:14,outline:"none"}}/>
                    <div style={{display:"flex",flexDirection:"column",gap:8}}>
                      {inst.metrics.map((m,mi)=>(
                        <div key={mi} style={{display:"flex",alignItems:"center",justifyContent:"space-between",gap:8}}>
                          <span style={{fontSize:12.5,color:C.sub}}>{t(m.label)}</span>
                          <div style={{display:"flex",gap:3}}>
                            {[1,2,3,4,5].map(n=>(
                              <button key={n} type="button" onClick={()=>setMetricScores(prev=>prev.map((v,i)=>i===mi?n:v))} aria-label={`${n}`} style={{background:"none",border:"none",cursor:"pointer",padding:1}}>
                                <Star size={16} fill={n<=metricScores[mi]?C.gold:"none"} stroke={n<=metricScores[mi]?C.gold:C.dim}/>
                              </button>
                            ))}
                          </div>
                        </div>
                      ))}
                      <div style={{display:"flex",alignItems:"center",justifyContent:"space-between",gap:8,marginTop:4,paddingTop:8,borderTop:`1px solid ${C.border}`}}>
                        <span style={{fontSize:12.5,color:C.text,fontFamily:FH,fontWeight:700}}>{t("review.overall")}</span>
                        <span style={{fontFamily:FH,fontWeight:800,color:C.gold,fontSize:14}}>
                          {metricScores.length ? (Math.round((metricScores.reduce((a,b)=>a+b,0)/metricScores.length)*10)/10) : 0} ★
                        </span>
                      </div>
                    </div>
                    <textarea value={reviewText} onChange={e=>setReviewText(e.target.value)} placeholder={t({ru:"Расскажите о своём опыте",tg:"Дар бораи таҷрибаи худ нақл кунед"})} rows={3}
                      style={{padding:"10px 14px",borderRadius:10,border:`1px solid ${C.border}`,background:C.s2,color:C.text,fontFamily:FB,fontSize:14,outline:"none",resize:"vertical"}}/>
                    <button onClick={submitReview} disabled={!reviewName.trim()||!reviewText.trim()} style={{alignSelf:"flex-start",padding:"9px 18px",borderRadius:10,background:C.teal,color:C.overlay,fontFamily:FH,fontWeight:700,fontSize:13.5,border:"none",cursor:"pointer",opacity:(!reviewName.trim()||!reviewText.trim())?0.5:1}}>
                      {t({ru:"Опубликовать",tg:"Нашр кардан"})}
                    </button>
                  </div>
                )}

                {reviews.map(r=>(
                  <div key={r.id} id={`review-${r.id}`} style={{borderRadius:16,border:`1px solid ${highlightReviewId===r.id?C.teal:C.border}`,background:highlightReviewId===r.id?`${C.teal}0f`:C.s1,padding:"18px 20px",transition:"background .4s,border-color .4s"}}>
                    <div style={{display:"flex",justifyContent:"space-between",alignItems:"flex-start",marginBottom:10}}>
                      <div style={{display:"flex",gap:10}}>
                        <div style={{width:38,height:38,borderRadius:"50%",background:C.teal,display:"flex",alignItems:"center",justifyContent:"center",fontFamily:FH,fontWeight:800,color:C.overlay,fontSize:15}}>{r.init}</div>
                        <div>
                          <p style={{fontFamily:FH,fontWeight:700,fontSize:14,color:C.text}}>{t(r.name)}</p>
                          <p style={{fontSize:12.5,color:C.sub}}>{r.date}</p>
                        </div>
                      </div>
                      <Stars s={r.score}/>
                    </div>
                    <p style={{fontSize:14,color:C.sub,lineHeight:1.7}}>{t(r.text)}</p>
                    {r.reply && (
                      <div style={{marginTop:10,padding:"10px 14px",borderRadius:10,background:`${C.teal}12`}}>
                        <p style={{fontSize:12,fontWeight:700,color:C.teal,fontFamily:FH,marginBottom:4}}>{t({ru:"Ответ учреждения:",tg:"Ҷавоби муассиса:"})}</p>
                        <p style={{fontSize:13.5,color:C.sub}}>{t(r.reply)}</p>
                      </div>
                    )}
                  </div>
                ))}
                {reviews.length===0 && <p style={{color:C.muted,fontSize:14}}>{t("empty.reviews")}</p>}
              </div>
            </div>
          </div>
        )}

        {/* NEWS */}
        {tab==="news" && (
          <div>
            <h2 style={{fontFamily:FH,fontWeight:800,fontSize:22,color:C.text,marginBottom:24}}>{t("tab.news")}</h2>
            <div style={{display:"grid",gridTemplateColumns:"repeat(3,1fr)",gap:18}}>
              {news.map(n=>(
                <Link key={n.id} href={`/news/${n.id}`} style={{display:"block",borderRadius:18,overflow:"hidden",border:`1px solid ${C.border}`,background:C.s1,textDecoration:"none",cursor:"pointer",transition:"transform .18s,border-color .18s"}}>
                  {n.coverUrl && <div style={{height:160,overflow:"hidden"}}><img src={n.coverUrl} alt={t(n.title)} style={{width:"100%",height:"100%",objectFit:"cover"}}/></div>}
                  <div style={{padding:"16px 18px"}}>
                    <span style={{fontSize:11.5,fontWeight:700,color:C.teal,fontFamily:FH,textTransform:"uppercase",letterSpacing:".05em"}}>{t(n.category)}</span>
                    <h3 style={{fontFamily:FH,fontWeight:800,fontSize:15.5,color:C.text,margin:"6px 0 8px",lineHeight:1.3}}>{t(n.title)}</h3>
                    <p style={{fontSize:13,color:C.sub,lineHeight:1.6,marginBottom:10}}>{t(n.content).slice(0,100)}...</p>
                    <div style={{display:"flex",justifyContent:"space-between",alignItems:"center"}}>
                      <p style={{fontSize:12,color:C.muted}}>{n.date}</p>
                      <p style={{fontSize:12,color:C.muted}}>{n.views} {t({ru:"просм.",tg:"дидан"})}</p>
                    </div>
                  </div>
                </Link>
              ))}
              {!news.length && <p style={{color:C.muted,fontSize:14,gridColumn:"1/-1"}}>{t("empty.news")}</p>}
            </div>
          </div>
        )}

        {/* VACANCIES */}
        {tab==="vacancies" && (
          <div style={{display:"flex",flexDirection:"column",gap:16}}>
            {instVacancies.length===0 ? (
              <div style={{padding:56,borderRadius:16,border:`1px dashed ${C.border}`,textAlign:"center",color:C.muted}}>
                <Briefcase size={28} style={{color:C.dim,margin:"0 auto 12px"}}/>
                <p style={{fontFamily:FH,fontWeight:800,fontSize:17,color:C.text}}>{t("empty.vacancies")}</p>
              </div>
            ) : instVacancies.map(v=>(
              <div key={v.id} id={`vacancy-${v.id}`} style={{borderRadius:18,border:`1px solid ${highlightVacancyId===v.id?C.teal:C.border}`,background:highlightVacancyId===v.id?`${C.teal}0f`:C.s1,padding:26,transition:"background .4s,border-color .4s"}}>
                <h3 style={{fontFamily:FH,fontWeight:800,fontSize:17,color:C.text,marginBottom:8}}>{t(v.title)}</h3>
                <div style={{display:"flex",gap:16,flexWrap:"wrap",marginBottom:16}}>
                  {v.salaryFrom && (
                    <span style={{display:"flex",alignItems:"center",gap:6,fontSize:13,color:C.text}}>
                      <Wallet size={14} style={{color:C.teal}}/> {v.salaryFrom}–{v.salaryTo} {t("common.perMonth")}
                    </span>
                  )}
                  <span style={{display:"flex",alignItems:"center",gap:6,fontSize:13,color:C.text}}>
                    <Clock size={14} style={{color:C.teal}}/> {t(v.employment)}
                  </span>
                </div>
                <p style={{fontSize:14,color:C.sub,lineHeight:1.7,marginBottom:16}}>{t(v.description)}</p>
                <div style={{display:"flex",flexDirection:"column",gap:7,marginBottom:18}}>
                  {v.requirements.map((r,i)=>(
                    <div key={i} style={{display:"flex",alignItems:"flex-start",gap:8,fontSize:13.5,color:C.sub}}>
                      <CheckCircle size={14} style={{color:C.ok,flexShrink:0,marginTop:2}}/> {t(r)}
                    </div>
                  ))}
                </div>
                <button
                  onClick={()=>{ addApplication(v.id); router.push(chatHref(inst.id)); }}
                  style={{display:"flex",alignItems:"center",gap:8,padding:"10px 20px",borderRadius:10,background:C.teal,color:C.overlay,fontFamily:FH,fontWeight:700,fontSize:13.5,border:"none",cursor:"pointer"}}>
                  <MessageSquare size={14}/> {hasApplied(v.id) ? t({ru:"Вы откликнулись — открыть чат",tg:"Шумо ҷавоб додаед — чатро кушоед"}) : t({ru:"Откликнуться",tg:"Ҷавоб додан"})}
                </button>
              </div>
            ))}
          </div>
        )}

      </div>

      <AmenityModal amenity={amenityOpen} inst={inst} onClose={()=>setAmenityOpen(null)}/>
      <Toast message={reviewToast} onDone={()=>setReviewToast(null)}/>
    </div>
  );
}
