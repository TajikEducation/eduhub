"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { ArrowLeft, MapPin, Phone, Globe, Star, CheckCircle, Award, Users, Utensils, Camera, MessageSquare, Newspaper, BookOpen, ChevronRight, Mail, UserSquare2, Medal, Trophy, PenLine, Briefcase, Wallet, Clock, Bus, type LucideIcon } from "lucide-react";
import { C, FH, FB, CATEGORY_META, REGION_LABEL, type Region } from "@/lib/data";
import { useCreateConversationMutation } from "@/app/(site)/messages/api/chatApi";
import { useReveal, revealStyle } from "@/lib/useReveal";
import { toast } from "sonner";
import { useT } from "@/lib/i18n";
import { SingleMap } from "@/components/InstMap";
import { backendTypeToCategory } from "@/lib/backendTypes";
import { useMeQuery } from "@/app/(site)/login/api/authApi";
import {
  useGetInstitutionQuery,
  useListInstitutionReviewsQuery,
  useCreateInstitutionReviewMutation,
  useListInstitutionVacanciesQuery,
  useListInstitutionPublishedNewsQuery,
} from "./api/institutionApi";
import { getAccessToken } from "@/lib/authToken";

type Tab = "about"|"staff"|"gallery"|"achievements"|"alumni"|"menu"|"reviews"|"news"|"vacancies";

const ACH_TIER: Record<string,{icon:LucideIcon;color:string}> = {
  gold:{icon:Medal,color:C.gold}, silver:{icon:Medal,color:C.sub}, bronze:{icon:Medal,color:C.goldD}, special:{icon:Trophy,color:C.teal},
};
const SOCIAL_META = {
  instagram: { abbr:"IG", color:"#E1306C", url:(h:string)=>`https://instagram.com/${h}` },
  telegram:  { abbr:"TG", color:"#229ED9", url:(h:string)=>`https://t.me/${h}` },
  facebook:  { abbr:"FB", color:"#1877F2", url:(h:string)=>`https://facebook.com/${h}` },
} as const;
function Stars({ s, size=13 }:{s:number;size?:number}) {
  return (
    <span style={{display:"flex",gap:2}}>
      {[1,2,3,4,5].map(i=>(<svg key={i} width={size} height={size} viewBox="0 0 24 24" fill={i<=Math.round(s)?C.gold:C.dim}><path d="M12 2l3 7h7l-5.5 4 2 7L12 16l-6.5 4 2-7L2 9h7z"/></svg>))}
    </span>
  );
}

export default function InstitutionProfilePage() {
  const params = useParams<{ id: string }>();
  return <InstitutionProfileInner key={params.id} />;
}

function InstitutionProfileInner() {
  const params = useParams<{ id: string }>();
  const searchParams = useSearchParams();
  const router = useRouter();
  const instId = params.id;
  const t = useT();
  const { ref: staffRef, visible: staffVisible } = useReveal<HTMLDivElement>();

  const { data: inst, isLoading, isError } = useGetInstitutionQuery(instId);
  const { data: reviewsData } = useListInstitutionReviewsQuery(instId);
  const { data: vacanciesData } = useListInstitutionVacanciesQuery(instId);
  const { data: newsData } = useListInstitutionPublishedNewsQuery(instId);
  const { data: me } = useMeQuery(undefined, { skip: !getAccessToken() });
  const [createReview, { isLoading: submitting }] = useCreateInstitutionReviewMutation();
  const [createConversation] = useCreateConversationMutation();

  const [tab, setTab] = useState<Tab>(() => (searchParams.get("tab") as Tab) ?? "about");
  const [lightbox, setLightbox] = useState<string|null>(null);
  const [showFullBio, setShowFullBio] = useState(false);
  const [showReviewForm, setShowReviewForm] = useState(false);
  const [reviewScore, setReviewScore] = useState(5);
  const [reviewText, setReviewText] = useState("");
  const [transportOpen, setTransportOpen] = useState(false);

  const highlightReviewId = searchParams.get("review");

  useEffect(() => {
    if (!highlightReviewId || tab !== "reviews") return;
    const el = document.getElementById(`review-${highlightReviewId}`);
    el?.scrollIntoView({ behavior: "smooth", block: "center" });
  }, [highlightReviewId, tab]);

  if (isLoading) {
    return <div style={{padding:80,textAlign:"center",color:C.muted,fontFamily:FB}}>{t({ru:"Загрузка…",tg:"Боркунӣ…"})}</div>;
  }
  if (isError || !inst) {
    return <div style={{padding:80,textAlign:"center",color:C.muted,fontFamily:FB}}>{t({ru:"Учреждение не найдено",tg:"Муассиса ёфт нашуд"})}</div>;
  }

  const tk = backendTypeToCategory(inst.types[0]);
  const meta = CATEGORY_META[tk];
  const showAlumni = tk === "cat_school" || tk === "cat_uni";
  const descRu = inst.description?.ru ?? "";
  const reviews = reviewsData?.items.filter(r => r.status === "approved") ?? [];
  const news = newsData?.items ?? [];
  const instVacancies = vacanciesData?.items ?? [];
  const hasTransport = inst.transport_routes.length > 0;
  const hasFood = inst.meal_plans.length > 0;

  const isLoggedIn = !!me;

  async function startChat() {
    if (!isLoggedIn) { router.push("/login"); return; }
    try {
      const conv = await createConversation({ counterpart_type: "institution", counterpart_id: instId }).unwrap();
      router.push(`/messages?conv=${conv.id}`);
    } catch {
      toast.error(t({ ru: "Не удалось открыть чат", tg: "Чатро кушода натавонист" }));
    }
  }

  async function submitReview() {
    if (!reviewText.trim() || !isLoggedIn) return;
    try {
      await createReview({ institutionId: instId, rating: reviewScore, text: reviewText.trim() }).unwrap();
      setReviewText(""); setReviewScore(5);
      setShowReviewForm(false);
      toast.success(t({ ru: "Отзыв отправлен на модерацию", tg: "Шарҳ барои санҷиш фиристода шуд" }));
    } catch {
      toast.error(t({ ru: "Не удалось отправить отзыв", tg: "Шарҳро фиристода натавонист" }));
    }
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
        <img src={inst.cover_photo_s3_key || meta.heroPhoto} alt={inst.name.ru} style={{width:"100%",height:"100%",objectFit:"cover"}}/>
        <div style={{position:"absolute",inset:0,background:`linear-gradient(180deg,${C.overlay}26 0%,${C.overlay}D9 75%,${C.overlay} 100%)`}}/>
        <div style={{position:"absolute",top:20,left:28}}>
          <button onClick={()=>router.back()} style={{display:"inline-flex",alignItems:"center",gap:7,padding:"8px 14px",borderRadius:8,background:"rgba(0,0,0,.5)",border:`1px solid rgba(255,255,255,.15)`,color:"#fff",fontFamily:FH,fontWeight:700,fontSize:13,backdropFilter:"blur(8px)",cursor:"pointer"}}>
            <ArrowLeft size={14}/> {t("common.back")}
          </button>
        </div>
        <div style={{position:"absolute",bottom:24,left:28,right:28,display:"flex",alignItems:"flex-end",justifyContent:"space-between"}}>
          <div>
            <div style={{display:"flex",gap:8,marginBottom:10,flexWrap:"wrap"}}>
              <span style={{fontSize:12,fontWeight:700,padding:"4px 10px",borderRadius:7,background:meta.color,color:C.overlay,fontFamily:FH}}>{t(meta.label)}</span>
              {inst.verified && <Link href="/company" style={{fontSize:12,fontWeight:700,padding:"4px 10px",borderRadius:7,background:`${C.ok}22`,border:`1px solid ${C.ok}`,color:C.ok,fontFamily:FH,display:"flex",alignItems:"center",gap:4,textDecoration:"none"}}><CheckCircle size={11}/> {t("common.verified")}</Link>}
              {inst.tag && <span style={{fontSize:12,fontWeight:700,padding:"4px 10px",borderRadius:7,background:`${C.gold}22`,border:`1px solid ${C.gold}`,color:C.gold,fontFamily:FH}}>{inst.tag.ru}</span>}
            </div>
            <h1 style={{fontFamily:FH,fontWeight:900,fontSize:"clamp(22px,3vw,36px)",color:"#fff",lineHeight:1.1,marginBottom:6}}>{inst.name.ru}</h1>
            <div style={{display:"flex",alignItems:"center",gap:12,flexWrap:"wrap"}}>
              <div style={{display:"flex",alignItems:"center",gap:6}}>
                <Stars s={inst.rating_avg ?? 0} size={15}/>
                <span style={{fontFamily:FH,fontWeight:800,fontSize:15,color:"#fff"}}>{(inst.rating_avg ?? 0).toFixed(1)}</span>
                <span style={{fontSize:13,color:"rgba(255,255,255,.55)"}}>({inst.review_count} {t("common.reviews")})</span>
              </div>
              <span style={{fontSize:13,color:"rgba(255,255,255,.6)",display:"flex",alignItems:"center",gap:4}}><MapPin size={12}/>{inst.district ?? ""}{inst.city ? `, ${inst.city.ru}` : ""}</span>
            </div>
          </div>
          <div style={{textAlign:"right"}}>
            <p style={{fontFamily:FH,fontWeight:900,fontSize:30,color:"#fff"}}>{inst.price ?? "—"} <span style={{fontSize:14,fontWeight:400,color:"rgba(255,255,255,.55)"}}>{t("common.perMonth")}</span></p>
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
          <div className="eh-mobile-1col" style={{display:"grid",gridTemplateColumns:"1fr 320px",gap:24,alignItems:"start"}}>
            <div style={{display:"flex",flexDirection:"column",gap:20}}>
              <div style={{borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:26}}>
                <h2 style={{fontFamily:FH,fontWeight:800,fontSize:19,color:C.text,marginBottom:14}}>{t("tab.about")}</h2>
                <p style={{fontSize:14.5,color:C.sub,lineHeight:1.78,maxHeight:showFullBio?undefined:120,overflow:showFullBio?undefined:"hidden"}}>
                  {descRu || t({ru:"Описание пока не добавлено",tg:"Тавсиф ҳанӯз илова нашудааст"})}
                </p>
                {descRu.length > 200 && (
                  <button onClick={()=>setShowFullBio(!showFullBio)} style={{marginTop:10,fontSize:13,color:C.teal,fontFamily:FH,fontWeight:600,background:"none",border:"none",cursor:"pointer"}}>
                    {showFullBio?t("common.collapse"):t("common.readMore")}
                  </button>
                )}
              </div>

              {/* key facts */}
              <div className="eh-mobile-1col" style={{display:"grid",gridTemplateColumns:inst.staff.length>0?"repeat(4,1fr)":"repeat(3,1fr)",gap:12}}>
                {[
                  {l:t({ru:"Основана",tg:"Таъсисёфта"}),v:inst.founded ?? "—"},
                  {l:t({ru:"Учеников",tg:"Хонандагон"}),v:inst.students_count ?? "—"},
                  {l:t({ru:"Возраст",tg:"Синну сол"}),v:inst.age_range ?? "—"},
                  ...(inst.staff.length>0 ? [{l:t({ru:"Учеников на педагога",tg:"Хонанда ба як омӯзгор"}),v: inst.students_count ? `${Math.round(inst.students_count/inst.staff.length)}:1` : "—"}] : []),
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
                <div className="eh-mobile-1col" style={{display:"grid",gridTemplateColumns:"repeat(4,1fr)",gap:12}}>
                  {[
                    {l:t("geo.region"),v:t(REGION_LABEL[inst.region as Region] ?? {ru:inst.region,tg:inst.region})},
                    {l:t("geo.city"),v:inst.city?.ru ?? "—"},
                    {l:t("geo.district"),v:inst.district ?? "—"},
                    {l:t("geo.street"),v:inst.address?.ru ?? "—"},
                  ].map(g=>(
                    <div key={g.l}>
                      <p style={{fontSize:11,fontWeight:700,color:C.dim,textTransform:"uppercase",letterSpacing:".05em",fontFamily:FH,marginBottom:4}}>{g.l}</p>
                      <p style={{fontSize:13.5,color:C.text}}>{g.v}</p>
                    </div>
                  ))}
                </div>
                {inst.lat !== 0 && inst.lng !== 0 && (
                  <div style={{marginTop:18}}>
                    <SingleMap lat={inst.lat} lng={inst.lng} routeLabel={t({ru:"Построить маршрут",tg:"Роҳро тартиб додан"})} />
                  </div>
                )}
              </div>

              {/* features */}
              <div style={{borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:26}}>
                <h3 style={{fontFamily:FH,fontWeight:800,fontSize:17,color:C.text,marginBottom:16}}>{t({ru:"Услуги и возможности",tg:"Хизматрасонӣ ва имконот"})}</h3>
                <div className="eh-mobile-1col" style={{display:"grid",gridTemplateColumns:"repeat(2,1fr)",gap:10}}>
                  <button onClick={()=>setTransportOpen(v=>!v)}
                    style={{display:"flex",alignItems:"center",gap:8,padding:"10px 12px",borderRadius:12,background:hasTransport?`${C.ok}12`:C.s2,border:`1px solid ${hasTransport?C.ok+"33":C.border}`,cursor:"pointer",textAlign:"left"}}>
                    <Bus size={16} style={{color:hasTransport?C.ok:C.dim,flexShrink:0}}/>
                    <span style={{fontSize:12.5,fontFamily:FH,fontWeight:600,color:hasTransport?C.text:C.dim}}>{t({ru:"Развозка",tg:"Интиқол"})}</span>
                    {hasTransport && <CheckCircle size={11} style={{color:C.ok,marginLeft:"auto",flexShrink:0}}/>}
                  </button>
                  <button onClick={()=>setTab("menu")}
                    style={{display:"flex",alignItems:"center",gap:8,padding:"10px 12px",borderRadius:12,background:hasFood?`${C.ok}12`:C.s2,border:`1px solid ${hasFood?C.ok+"33":C.border}`,cursor:"pointer",textAlign:"left"}}>
                    <Utensils size={16} style={{color:hasFood?C.ok:C.dim,flexShrink:0}}/>
                    <span style={{fontSize:12.5,fontFamily:FH,fontWeight:600,color:hasFood?C.text:C.dim}}>{t({ru:"Питание",tg:"Ғизо"})}</span>
                    {hasFood && <CheckCircle size={11} style={{color:C.ok,marginLeft:"auto",flexShrink:0}}/>}
                  </button>
                </div>
                {transportOpen && hasTransport && (
                  <div style={{marginTop:14,display:"flex",flexDirection:"column",gap:8}}>
                    {inst.transport_routes.map(rt=>(
                      <div key={rt.id} style={{padding:"10px 14px",borderRadius:10,background:C.s2,fontSize:12.5,color:C.sub}}>
                        <span style={{color:C.text,fontFamily:FH,fontWeight:700}}>{rt.label?.ru ?? rt.type}</span>
                        {rt.areas.length>0 && <> — {rt.areas.map(a=>a.ru).join(", ")}</>}
                        {rt.cost != null && <> · {rt.cost} {t("common.perMonth")}</>}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>

            {/* sidebar */}
            <div style={{display:"flex",flexDirection:"column",gap:16}}>
              <div style={{borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:22}}>
                <h3 style={{fontFamily:FH,fontWeight:800,fontSize:16,color:C.text,marginBottom:14}}>{t({ru:"Контакты",tg:"Тамос"})}</h3>
                {[
                  {icon:<Phone size={15}/>,l:inst.phone ?? "—"},
                  {icon:<Mail size={15}/>,l:inst.email ?? "—"},
                  {icon:<MapPin size={15}/>,l:inst.address?.ru ?? "—"},
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

                <button onClick={startChat} style={{marginTop:14,width:"100%",padding:"11px",borderRadius:12,background:C.teal,color:C.overlay,fontFamily:FH,fontWeight:800,fontSize:14,display:"flex",alignItems:"center",justifyContent:"center",gap:7,cursor:"pointer",border:"none"}}>
                  <MessageSquare size={15}/> {t("common.writeInChat")}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* STAFF */}
        {tab==="staff" && (
          <div>
            <h2 style={{fontFamily:FH,fontWeight:800,fontSize:22,color:C.text,marginBottom:24}}>{t({ru:"Педагогический состав",tg:"Ҳайати педагогӣ"})}</h2>
            {inst.staff.length === 0 ? (
              <p style={{color:C.muted,fontSize:14}}>{t("empty.staff")}</p>
            ) : (
              <div ref={staffRef} className="eh-mobile-1col" style={{display:"grid",gridTemplateColumns:"repeat(3,1fr)",gap:18}}>
                {inst.staff.map((p,i)=>(
                  <button key={p.id} onClick={()=>router.push(`/people/${p.id}`)}
                    style={{textAlign:"left",borderRadius:18,overflow:"hidden",border:`1px solid ${C.border}`,background:C.s1,cursor:"pointer",...revealStyle(staffVisible,i*70)}}>
                    <div style={{height:220,overflow:"hidden",position:"relative"}}>
                      <img src={p.photo_url || "/logo.svg"} alt={p.name.ru} style={{width:"100%",height:"100%",objectFit:"cover"}}/>
                      <div style={{position:"absolute",inset:0,background:`linear-gradient(180deg,transparent 55%,${C.overlay}E6 100%)`}}/>
                      <div style={{position:"absolute",bottom:14,left:14,right:14}}>
                        <p style={{fontFamily:FH,fontWeight:800,fontSize:16,color:"#fff"}}>{p.name.ru}</p>
                        <p style={{fontSize:12.5,color:C.teal,fontFamily:FH}}>{p.role_label.ru}</p>
                      </div>
                    </div>
                    <div style={{padding:"14px 16px"}}>
                      {p.subject && (
                        <div style={{display:"flex",gap:6,marginBottom:10}}>
                          <span style={{fontSize:11,fontWeight:600,padding:"3px 8px",borderRadius:6,background:`${C.teal}18`,color:C.teal,fontFamily:FH}}>{p.subject.ru}</span>
                        </div>
                      )}
                      {p.exp && (
                        <div style={{display:"flex",justifyContent:"space-between",alignItems:"center"}}>
                          <span style={{fontSize:13,color:C.sub}}>{p.exp} {t({ru:"опыта",tg:"таҷриба"})}</span>
                        </div>
                      )}
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
            {inst.gallery.length === 0 ? (
              <p style={{color:C.muted,fontSize:14}}>{t({ru:"Фотографии не добавлены",tg:"Расмҳо илова нашудаанд"})}</p>
            ) : (
              <div className="eh-mobile-1col" style={{display:"grid",gridTemplateColumns:"repeat(3,1fr)",gap:10}}>
                {inst.gallery.map((g)=>(
                  <div key={g.id} onClick={()=>setLightbox(g.s3_key)} style={{aspectRatio:"4/3",borderRadius:12,overflow:"hidden",cursor:"pointer",position:"relative"}}>
                    <img src={g.s3_key} alt={g.label?.ru ?? ""} style={{width:"100%",height:"100%",objectFit:"cover",transition:"transform .3s"}}
                      onMouseEnter={e=>(e.currentTarget.style.transform="scale(1.05)")}
                      onMouseLeave={e=>(e.currentTarget.style.transform="scale(1)")}/>
                    {g.label && <div style={{position:"absolute",bottom:0,left:0,right:0,padding:"6px 10px",background:"linear-gradient(transparent,rgba(0,0,0,.6))",fontSize:12.5,color:"rgba(255,255,255,.8)",fontFamily:FH}}>{g.label.ru}</div>}
                  </div>
                ))}
              </div>
            )}
            {lightbox && (
              <div onClick={()=>setLightbox(null)} style={{position:"fixed",inset:0,background:"rgba(0,0,0,.9)",zIndex:999,display:"flex",alignItems:"center",justifyContent:"center",cursor:"pointer"}}>
                <img src={lightbox} alt="" style={{maxWidth:"92vw",maxHeight:"92vh",objectFit:"contain",borderRadius:12}}/>
              </div>
            )}
          </div>
        )}

        {/* ACHIEVEMENTS — публичный API отдаёт только тир (gold/silver/bronze/special), без
            группировки institution/student/teacher (это поле, OwnerType, наружу не отдаётся) —
            поэтому единый список, не разбитый по категориям, как было в mock. */}
        {tab==="achievements" && (
          <div>
            <h2 style={{fontFamily:FH,fontWeight:800,fontSize:22,color:C.text,marginBottom:24}}>{t("tab.achievements")}</h2>
            {inst.achievements.length===0 ? (
              <p style={{color:C.muted,fontSize:14}}>{t("empty.achievements")}</p>
            ) : (
              <div className="eh-mobile-1col" style={{display:"grid",gridTemplateColumns:"repeat(3,1fr)",gap:12}}>
                {inst.achievements.map(ach=>{
                  const tier = ACH_TIER[ach.category] ?? ACH_TIER.gold;
                  const TierIcon = tier.icon;
                  return (
                    <div key={ach.id}
                      style={{display:"block",borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:"18px 20px"}}>
                      <TierIcon size={24} style={{color:tier.color,display:"block",marginBottom:10}}/>
                      <p style={{fontFamily:FH,fontWeight:700,fontSize:15,color:C.text,marginBottom:4}}>{ach.title.ru}</p>
                      <p style={{fontSize:12.5,color:C.sub}}>{ach.year}</p>
                      <p style={{fontSize:12.5,color:C.muted,marginTop:4}}>{ach.description.ru}</p>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        )}

        {/* ALUMNI */}
        {tab==="alumni" && (
          <div>
            <h2 style={{fontFamily:FH,fontWeight:800,fontSize:22,color:C.text,marginBottom:24}}>{t("tab.alumni")}</h2>
            {inst.alumni.length===0 ? (
              <p style={{color:C.muted,fontSize:14}}>{t({ru:"Выпускники не добавлены",tg:"Хатмкунандагон илова нашудаанд"})}</p>
            ) : (
              <div className="eh-mobile-1col" style={{display:"grid",gridTemplateColumns:"repeat(4,1fr)",gap:16}}>
                {inst.alumni.map(a=>(
                  <div key={a.id} style={{borderRadius:16,border:`1px solid ${C.border}`,background:C.s1,padding:18,textAlign:"center"}}>
                    <img src={a.photo_url || "/logo.svg"} alt={a.name.ru} style={{width:64,height:64,borderRadius:"50%",objectFit:"cover",margin:"0 auto 12px"}}/>
                    <p style={{fontFamily:FH,fontWeight:700,fontSize:14,color:C.text,marginBottom:3}}>{a.name.ru}</p>
                    <p style={{fontSize:12,color:C.dim,marginBottom:8}}>{t({ru:"Выпуск",tg:"Хатм"})} {a.grad_year}</p>
                    {a.now_label && <p style={{fontSize:12.5,color:C.sub,lineHeight:1.5}}>{a.now_label.ru}</p>}
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* MENU */}
        {tab==="menu" && (
          <div>
            <h2 style={{fontFamily:FH,fontWeight:800,fontSize:22,color:C.text,marginBottom:6}}>{t({ru:"Питание",tg:"Ғизо"})}</h2>
            <p style={{fontSize:14,color:C.sub,marginBottom:24}}>{t({ru:"Варианты питания в учреждении",tg:"Вариантҳои ғизо дар муассиса"})}</p>
            {!hasFood ? (
              <div style={{padding:48,borderRadius:16,border:`1px dashed ${C.border}`,textAlign:"center",color:C.muted}}>
                <Utensils size={28} style={{margin:"0 auto 12px",color:C.dim}}/>
                <p style={{fontFamily:FH,fontWeight:700,fontSize:16,color:C.text,marginBottom:6}}>{t({ru:"Питание не предусмотрено",tg:"Ғизо пешбинӣ нашудааст"})}</p>
                <p style={{fontSize:13.5}}>{t({ru:"В данном учреждении нет организованного питания",tg:"Дар ин муассиса ғизои муташаккил нест"})}</p>
              </div>
            ) : (
              <div className="eh-mobile-1col" style={{display:"grid",gridTemplateColumns:`repeat(${Math.min(inst.meal_plans.length,4)},1fr)`,gap:14}}>
                {inst.meal_plans.map(mp=>(
                  <div key={mp.id} style={{borderRadius:16,border:`1px solid ${C.border}`,background:C.s1,overflow:"hidden"}}>
                    <div style={{padding:"10px 14px",background:C.s2,borderBottom:`1px solid ${C.border}`}}>
                      <p style={{fontFamily:FH,fontWeight:800,fontSize:13,color:C.teal}}>{mp.label?.ru ?? mp.meal_type}</p>
                    </div>
                    <div style={{padding:"12px 14px",display:"flex",flexDirection:"column",gap:8}}>
                      {mp.cost != null && <p style={{fontSize:12.5,color:C.sub}}>{mp.cost} {t("common.perMonth")}</p>}
                      {mp.halal && <p style={{fontSize:11.5,color:C.ok,fontFamily:FH,fontWeight:700}}>Halal</p>}
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
            <div className="eh-mobile-1col" style={{display:"grid",gridTemplateColumns:"240px 1fr",gap:24,marginBottom:32}}>
              <div style={{borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:24,textAlign:"center"}}>
                <p style={{fontFamily:FH,fontWeight:900,fontSize:52,color:C.text,lineHeight:1}}>{(inst.rating_avg ?? 0).toFixed(1)}</p>
                <Stars s={inst.rating_avg ?? 0} size={18}/>
                <p style={{fontSize:13,color:C.sub,marginTop:8}}>{inst.review_count} {t("common.reviews")}</p>
                <div style={{marginTop:16}}>
                  {[5,4,3,2,1].map(n=>{
                    const countN = reviews.filter(r=>r.rating===n).length;
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
                {isLoggedIn ? (
                  <div style={{display:"flex",justifyContent:"flex-end"}}>
                    <button onClick={()=>setShowReviewForm(v=>!v)} style={{display:"flex",alignItems:"center",gap:7,padding:"9px 16px",borderRadius:10,background:showReviewForm?C.s3:`${C.teal}18`,border:`1px solid ${showReviewForm?C.border:C.teal+"44"}`,color:showReviewForm?C.sub:C.teal,fontFamily:FH,fontWeight:700,fontSize:13,cursor:"pointer"}}>
                      <PenLine size={14}/> {t({ru:"Написать отзыв",tg:"Шарҳ навиштан"})}
                    </button>
                  </div>
                ) : (
                  <div style={{borderRadius:14,border:`1px dashed ${C.border}`,background:C.s1,padding:"14px 18px",display:"flex",alignItems:"center",gap:12}}>
                    <PenLine size={16} style={{color:C.dim,flexShrink:0}}/>
                    <div style={{flex:1}}>
                      <p style={{fontSize:13,color:C.sub,lineHeight:1.5}}>{t({ru:"Войдите, чтобы оставить отзыв",tg:"Барои шарҳ гузоштан ворид шавед"})}</p>
                      <button onClick={()=>router.push("/login")} style={{marginTop:6,background:"none",border:"none",color:C.teal,fontFamily:FH,fontWeight:700,fontSize:12.5,cursor:"pointer",padding:0}}>
                        {t("nav.login")}
                      </button>
                    </div>
                  </div>
                )}

                {showReviewForm && isLoggedIn && (
                  <div style={{borderRadius:16,border:`1px solid ${C.teal}44`,background:C.s1,padding:"18px 20px",display:"flex",flexDirection:"column",gap:12}}>
                    <div style={{display:"flex",alignItems:"center",justifyContent:"space-between",gap:8}}>
                      <span style={{fontSize:12.5,color:C.sub}}>{t("review.overall")}</span>
                      <div style={{display:"flex",gap:3}}>
                        {[1,2,3,4,5].map(n=>(
                          <button key={n} type="button" onClick={()=>setReviewScore(n)} aria-label={`${n}`} style={{background:"none",border:"none",cursor:"pointer",padding:1}}>
                            <Star size={18} fill={n<=reviewScore?C.gold:"none"} stroke={n<=reviewScore?C.gold:C.dim}/>
                          </button>
                        ))}
                      </div>
                    </div>
                    <textarea value={reviewText} onChange={e=>setReviewText(e.target.value)} placeholder={t({ru:"Расскажите о своём опыте",tg:"Дар бораи таҷрибаи худ нақл кунед"})} rows={3}
                      style={{padding:"10px 14px",borderRadius:10,border:`1px solid ${C.border}`,background:C.s2,color:C.text,fontFamily:FB,fontSize:14,outline:"none",resize:"vertical"}}/>
                    <button onClick={submitReview} disabled={!reviewText.trim()||submitting} style={{alignSelf:"flex-start",padding:"9px 18px",borderRadius:10,background:C.teal,color:C.overlay,fontFamily:FH,fontWeight:700,fontSize:13.5,border:"none",cursor:"pointer",opacity:(!reviewText.trim()||submitting)?0.5:1}}>
                      {t({ru:"Опубликовать",tg:"Нашр кардан"})}
                    </button>
                  </div>
                )}

                {reviews.map(r=>(
                  <div key={r.id} id={`review-${r.id}`} style={{borderRadius:16,border:`1px solid ${highlightReviewId===r.id?C.teal:C.border}`,background:highlightReviewId===r.id?`${C.teal}0f`:C.s1,padding:"18px 20px",transition:"background .4s,border-color .4s"}}>
                    <div style={{display:"flex",justifyContent:"space-between",alignItems:"flex-start",marginBottom:10}}>
                      <div style={{display:"flex",gap:10}}>
                        <div style={{width:38,height:38,borderRadius:"50%",background:C.teal,display:"flex",alignItems:"center",justifyContent:"center",fontFamily:FH,fontWeight:800,color:C.overlay,fontSize:15}}>{t({ru:"П",tg:"К"})}</div>
                        <div>
                          <p style={{fontFamily:FH,fontWeight:700,fontSize:14,color:C.text}}>{t({ru:"Пользователь",tg:"Корбар"})}</p>
                          <p style={{fontSize:12.5,color:C.sub}}>{new Date(r.created_at).toLocaleDateString("ru-RU")}</p>
                        </div>
                      </div>
                      <Stars s={r.rating}/>
                    </div>
                    <p style={{fontSize:14,color:C.sub,lineHeight:1.7}}>{r.text}</p>
                    {r.reply && (
                      <div style={{marginTop:10,padding:"10px 14px",borderRadius:10,background:`${C.teal}12`}}>
                        <p style={{fontSize:12,fontWeight:700,color:C.teal,fontFamily:FH,marginBottom:4}}>{t({ru:"Ответ учреждения:",tg:"Ҷавоби муассиса:"})}</p>
                        <p style={{fontSize:13.5,color:C.sub}}>{r.reply}</p>
                      </div>
                    )}
                  </div>
                ))}
                {reviews.length===0 && <p style={{color:C.muted,fontSize:14}}>{t("empty.reviews")}</p>}
              </div>
            </div>
          </div>
        )}

        {/* NEWS — реальный backend, FR-24 */}
        {tab==="news" && (
          <div>
            <h2 style={{fontFamily:FH,fontWeight:800,fontSize:22,color:C.text,marginBottom:24}}>{t("tab.news")}</h2>
            <div className="eh-mobile-1col" style={{display:"grid",gridTemplateColumns:"repeat(3,1fr)",gap:18}}>
              {news.map(n=>(
                <Link key={n.id} href={`/news/${n.id}`} style={{display:"block",borderRadius:18,overflow:"hidden",border:`1px solid ${C.border}`,background:C.s1,textDecoration:"none",cursor:"pointer"}}>
                  {n.cover_s3_key && <div style={{height:160,overflow:"hidden"}}><img src={n.cover_s3_key} alt={n.title.ru} style={{width:"100%",height:"100%",objectFit:"cover"}}/></div>}
                  <div style={{padding:"16px 18px"}}>
                    {n.category && <span style={{fontSize:11.5,fontWeight:700,color:C.teal,fontFamily:FH,textTransform:"uppercase",letterSpacing:".05em"}}>{n.category.ru}</span>}
                    <h3 style={{fontFamily:FH,fontWeight:800,fontSize:15.5,color:C.text,margin:"6px 0 8px",lineHeight:1.3}}>{n.title.ru}</h3>
                    <p style={{fontSize:13,color:C.sub,lineHeight:1.6,marginBottom:10}}>{n.content.ru.slice(0,100)}...</p>
                    <div style={{display:"flex",justifyContent:"space-between",alignItems:"center"}}>
                      <p style={{fontSize:12,color:C.muted}}>{new Date(n.created_at).toLocaleDateString("ru-RU")}</p>
                      <p style={{fontSize:12,color:C.muted}}>{n.views_count} {t({ru:"просм.",tg:"дидан"})}</p>
                    </div>
                  </div>
                </Link>
              ))}
              {!news.length && <p style={{color:C.muted,fontSize:14,gridColumn:"1/-1"}}>{t("empty.news")}</p>}
            </div>
          </div>
        )}

        {/* VACANCIES — реальный backend, FR-36 */}
        {tab==="vacancies" && (
          <div style={{display:"flex",flexDirection:"column",gap:12}}>
            {instVacancies.length===0 ? (
              <div style={{padding:56,borderRadius:16,border:`1px dashed ${C.border}`,textAlign:"center",color:C.muted}}>
                <Briefcase size={28} style={{color:C.dim,margin:"0 auto 12px"}}/>
                <p style={{fontFamily:FH,fontWeight:800,fontSize:17,color:C.text}}>{t("empty.vacancies")}</p>
              </div>
            ) : instVacancies.map(v=>(
              <Link key={v.id} href={`/vacancies/${v.id}`} style={{display:"flex",alignItems:"center",justifyContent:"space-between",gap:16,borderRadius:16,border:`1px solid ${C.border}`,background:C.s1,padding:"18px 20px",textDecoration:"none"}}>
                <div style={{minWidth:0}}>
                  <h3 style={{fontFamily:FH,fontWeight:800,fontSize:15.5,color:C.text,marginBottom:6}}>{v.title.ru}</h3>
                  <div style={{display:"flex",gap:14,flexWrap:"wrap"}}>
                    {v.salary_from != null && (
                      <span style={{display:"flex",alignItems:"center",gap:5,fontSize:12.5,color:C.sub}}>
                        <Wallet size={12} style={{color:C.teal}}/> {v.salary_from}–{v.salary_to} {t("common.perMonth")}
                      </span>
                    )}
                    <span style={{display:"flex",alignItems:"center",gap:5,fontSize:12.5,color:C.sub}}>
                      <Clock size={12} style={{color:C.teal}}/> {v.employment.ru}
                    </span>
                  </div>
                </div>
                <ChevronRight size={18} style={{color:C.dim,flexShrink:0}}/>
              </Link>
            ))}
          </div>
        )}

      </div>
    </div>
  );
}
