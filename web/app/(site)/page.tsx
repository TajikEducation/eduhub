"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Search, ArrowRight, ChevronRight, ShieldCheck, SlidersHorizontal, Star, MessageCircle, Send, Users, Building2, Briefcase, LocateFixed, UtensilsCrossed, Bus, GraduationCap, Trophy } from "lucide-react";
import { C, FH, FB, INSTITUTIONS, ALL_STAFF, PHOTOS, CATEGORY_META, REGION_LABEL, type CategoryKey } from "@/lib/data";
import { InstitCard } from "@/components/InstitCard";
import { useReveal, revealStyle } from "@/lib/useReveal";
import { useAppState, useVisibleInstitutions } from "@/lib/app-state";
import { useT } from "@/lib/i18n";
import { detectCoords, haversine } from "@/lib/geo";

const CATEGORY_KEYS: CategoryKey[] = ["cat_kg", "cat_school", "cat_center", "cat_uni"];

// Витрина реальных данных вместо вымышленных агрегатов платформы (STATS) —
// каждая карточка вычисляется из реальных сидов, ничего не выдумывается.
const foodInst = INSTITUTIONS.find(i => i.foodMenu && i.foodMenu.length > 0);
const foodDay = foodInst?.foodMenu?.[0];
const foodMeal = foodDay?.meals?.[0];
const transportInst = INSTITUTIONS.find(i => i.transportInfo);
const teacherHighlight = ALL_STAFF.find(p => p.achievements.length > 0);
let alumnusHighlight: { instId: number; name: { ru: string; tg: string }; now: { ru: string; tg: string } } | null = null;
for (const inst of INSTITUTIONS) {
  const a = inst.alumni?.[0];
  if (a) { alumnusHighlight = { instId: inst.id, name: a.name, now: a.now }; break; }
}

const REAL_DATA_CARDS = [
  {
    icon: UtensilsCrossed,
    title: {ru:"Реальное меню, не общие слова", tg:"Менюи воқеӣ, на суханони умумӣ"},
    text: foodMeal
      ? {ru:`${foodDay!.day}: ${foodMeal.name}, ${foodMeal.cal}`, tg:`${foodDay!.day}: ${foodMeal.name}, ${foodMeal.cal}`}
      : {ru:"Меню по дням, калории и аллергены — прямо в профиле учреждения", tg:"Меню аз рӯи рӯзҳо, калорияҳо ва аллергенҳо — дар профили муассиса"},
    cta: {ru:"Учреждения с питанием", tg:"Муассисаҳо бо ғизо"}, href:"/search?food=1",
  },
  {
    icon: Bus,
    title: {ru:"Как доедет ребёнок", tg:"Фарзандатон чӣ гуна меравад"},
    text: transportInst?.transportInfo
      ? {ru:`${transportInst.transportInfo.types[0].ru}, от ${transportInst.transportInfo.cost} сомони/мес`, tg:`${transportInst.transportInfo.types[0].tg}, аз ${transportInst.transportInfo.cost} сомонӣ/моҳ`}
      : {ru:"Тип транспорта, районы и стоимость — без звонков в приёмную", tg:"Навъи нақлиёт, ноҳияҳо ва арзиш — бе занг ба қабулгоҳ"},
    cta: {ru:"Учреждения с развозкой", tg:"Муассисаҳо бо расонидан"}, href:"/search?transport=1",
  },
  {
    icon: GraduationCap,
    title: {ru:"Педагоги — не аноним", tg:"Омӯзгорон — на бином"},
    text: teacherHighlight
      ? {ru:`${teacherHighlight.name.ru} — ${teacherHighlight.achievements[0].title.ru}`, tg:`${teacherHighlight.name.tg} — ${teacherHighlight.achievements[0].title.tg}`}
      : {ru:"Образование, стаж и награды каждого педагога — открыто", tg:"Таҳсилот, таҷриба ва мукофотҳои ҳар омӯзгор — кушода"},
    cta: {ru:"Профиль педагога", tg:"Профили омӯзгор"}, href: teacherHighlight ? `/people/${teacherHighlight.id}` : "/search",
  },
  {
    icon: Trophy,
    title: {ru:"Где выпускники сейчас", tg:"Хатмкунандагон ҳоло дар куҷоянд"},
    text: alumnusHighlight
      ? {ru:`${alumnusHighlight.name.ru} — ${alumnusHighlight.now.ru}`, tg:`${alumnusHighlight.name.tg} — ${alumnusHighlight.now.tg}`}
      : {ru:"Реальные истории выпускников — не абстрактный «рейтинг вуза»", tg:"Ҳикояҳои воқеии хатмкунандагон"},
    cta: {ru:"Истории выпускников", tg:"Ҳикояҳои хатмкунандагон"}, href: alumnusHighlight ? `/institutions/${alumnusHighlight.instId}?tab=alumni` : "/search",
  },
];

export default function HomePage() {
  const router = useRouter();
  const [q, setQ] = useState("");
  const { region } = useAppState();
  const t = useT();
  // точные координаты — только в памяти текущей сессии, не персистятся (минимизация PII)
  const [myCoords, setMyCoords] = useState<{ lat: number; lng: number } | null>(null);
  const [locating, setLocating] = useState(false);

  function handleSearch() { router.push(q ? `/search?q=${encodeURIComponent(q)}` : "/search"); }

  function locateMe() {
    setLocating(true);
    detectCoords()
      .then(setMyCoords)
      .catch(() => {})
      .finally(() => setLocating(false));
  }

  const institutions = useVisibleInstitutions();

  // деградация: точные координаты → выбранный регион → вся страна (по рейтингу)
  const nearby = useMemo(() => {
    if (myCoords) {
      return [...institutions]
        .sort((a, b) => haversine(myCoords, a.geo) - haversine(myCoords, b.geo))
        .slice(0, 4);
    }
    const pool = region ? institutions.filter(i => i.region === region) : institutions;
    return [...pool].sort((a,b)=>b.score-a.score).slice(0,4);
  }, [region, myCoords, institutions]);

  const { ref: statsRef, visible: statsVisible } = useReveal<HTMLDivElement>();
  const { ref: missionRef, visible: missionVisible } = useReveal<HTMLDivElement>();
  const { ref: catRef, visible: catVisible } = useReveal<HTMLDivElement>();
  const { ref: topRef, visible: topVisible } = useReveal<HTMLDivElement>();
  const { ref: howRef, visible: howVisible } = useReveal<HTMLDivElement>();

  return (
    <div style={{fontFamily:FB}}>
      {/* ── HERO ── */}
      <section style={{position:"relative",overflow:"hidden",padding:"88px 28px 96px",textAlign:"center"}}>
        <img src={PHOTOS.heroChild} alt="" style={{position:"absolute",inset:0,width:"100%",height:"100%",objectFit:"cover",opacity:0.9}}/>
        <div style={{position:"absolute",inset:0,background:`linear-gradient(180deg, ${C.overlay}6E 0%, ${C.overlay}6E 100%)`}}/>

        <div style={{position:"relative",maxWidth:640,margin:"0 auto"}}>
          <h1 style={{fontFamily:FH,fontWeight:900,fontSize:"clamp(32px,5vw,54px)",lineHeight:1.13,color:"#fff",marginBottom:18,letterSpacing:"-.02em",textShadow:"0 2px 20px rgba(0,0,0,.25)"}}>
            {t({ru:"Постройте будущее ребёнка", tg:"Ояндаи фарзандро бунёд кунед"})}<br/>
            <span style={{color:C.teal}}>{t({ru:"вместе с нами", tg:"якҷоя бо мо"})}</span>
          </h1>
          <p style={{fontSize:"clamp(14px,1.6vw,16px)",color:"rgba(255,255,255,.92)",marginBottom:32,lineHeight:1.65,maxWidth:490,margin:"0 auto 32px",textShadow:"0 1px 12px rgba(0,0,0,.3)"}}>
            {t({ru:"Единый каталог по всем 5 регионам Таджикистана: честный рейтинг по 8 параметрам и отзывы только от родителей с подтверждённой связью — без купленных пятёрок.",
                tg:"Каталоги ягона аз рӯи ҳамаи 5 минтақаи Тоҷикистон: рейтинги ростқавлона аз рӯи 8 меъёр ва шарҳҳо танҳо аз волидайне, ки алоқаи тасдиқшуда доранд — бе панҷҳои харидашуда."})}
          </p>

          <div style={{display:"flex",maxWidth:460,margin:"0 auto",borderRadius:14,overflow:"hidden",boxShadow:"0 8px 40px rgba(0,0,0,.5)",border:`1px solid ${C.teal}44`}}>
            <div style={{flex:1,display:"flex",alignItems:"center",gap:10,background:C.s1,padding:"14px 18px"}}>
              <Search size={17} style={{color:C.teal,flexShrink:0}}/>
              <input value={q} onChange={e=>setQ(e.target.value)} onKeyDown={e=>e.key==="Enter"&&handleSearch()} placeholder={t({ru:"Школа, детсад, район…", tg:"Мактаб, боғча, ноҳия…"})}
                style={{flex:1,background:"transparent",border:"none",outline:"none",fontSize:15,color:C.text,fontFamily:FB}}/>
            </div>
            <button onClick={handleSearch} style={{padding:"0 26px",background:C.teal,color:C.overlay,fontFamily:FH,fontWeight:800,fontSize:14,whiteSpace:"nowrap",flexShrink:0,border:"none",cursor:"pointer"}}>
              {t({ru:"Найти", tg:"Ёфтан"})}
            </button>
          </div>

          <div style={{display:"flex",gap:8,justifyContent:"center",flexWrap:"wrap",marginTop:16}}>
            {[
              {l:{ru:"Лучшие школы",tg:"Беҳтарин мактабҳо"}, href:"/search?type=cat_school&sort=score"},
              {l:{ru:"С развозкой",tg:"Бо расонидан"}, href:"/search?transport=1"},
              {l:{ru:"Проверенные МОН РТ",tg:"Санҷидашуда аз ВМО ҶТ"}, href:"/search?verified=1"},
              {l:{ru:"Высокий рейтинг",tg:"Рейтинги баланд"}, href:"/search?rating=4.5"},
            ].map((c,i)=>(
              <Link key={i} href={c.href} style={{fontSize:12.5,fontWeight:600,padding:"5px 13px",borderRadius:8,border:"1px solid rgba(255,255,255,.22)",color:"rgba(255,255,255,.8)",background:"rgba(255,255,255,.08)",fontFamily:FH,textDecoration:"none"}}>
                {t(c.l)}
              </Link>
            ))}
          </div>

          <Link href="/company" style={{display:"inline-flex",alignItems:"center",gap:6,marginTop:22,fontSize:12.5,fontWeight:600,color:"rgba(255,255,255,.72)",fontFamily:FH,textDecoration:"none"}}>
            <ShieldCheck size={13} style={{color:C.teal}}/> {t({ru:"Как мы проверяем учреждения", tg:"Мо муассисаҳоро чӣ гуна санҷем"})}
          </Link>
        </div>
      </section>

      {/* ── REAL DATA (замена вымышленных агрегатов реальными данными из профилей) ── */}
      <section ref={statsRef} style={{borderTop:`1px solid ${C.border}`,borderBottom:`1px solid ${C.border}`,padding:"48px 28px"}}>
        <div style={{maxWidth:1260,margin:"0 auto"}}>
          <h2 style={{fontFamily:FH,fontWeight:900,fontSize:"clamp(20px,2.6vw,26px)",color:C.text,textAlign:"center",letterSpacing:"-.02em",marginBottom:6}}>
            {t({ru:"Настоящие данные — не рекламные обещания", tg:"Маълумоти воқеӣ — на ваъдаи рекламавӣ"})}
          </h2>
          <p style={{fontSize:13.5,color:C.sub,textAlign:"center",marginBottom:28}}>{t({ru:"Прямо из профилей учреждений на платформе", tg:"Мустақим аз профилҳои муассисаҳо дар платформа"})}</p>
          <div style={{display:"grid",gridTemplateColumns:"repeat(4,1fr)",gap:14}}>
            {REAL_DATA_CARDS.map((c,i)=>{
              const Icon = c.icon;
              return (
                <Link key={c.title.ru} href={c.href} style={{display:"block",borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:20,textDecoration:"none",...revealStyle(statsVisible,i*60)}}>
                  <div style={{width:38,height:38,borderRadius:10,background:`${C.teal}16`,border:`1px solid ${C.teal}33`,display:"flex",alignItems:"center",justifyContent:"center",marginBottom:12}}>
                    <Icon size={17} style={{color:C.teal}}/>
                  </div>
                  <p style={{fontFamily:FH,fontWeight:800,fontSize:14,color:C.text,marginBottom:6}}>{t(c.title)}</p>
                  <p style={{fontSize:12.5,color:C.sub,lineHeight:1.55,marginBottom:12}}>{t(c.text)}</p>
                  <span style={{display:"inline-flex",alignItems:"center",gap:5,fontFamily:FH,fontWeight:700,fontSize:12,color:C.teal}}>
                    {t(c.cta)} <ArrowRight size={12}/>
                  </span>
                </Link>
              );
            })}
          </div>
        </div>
      </section>

      {/* ── MISSION ── */}
      <section ref={missionRef} style={{maxWidth:1260,margin:"0 auto",padding:"64px 28px 0",...revealStyle(missionVisible)}}>
        <div style={{marginBottom:24}}>
          <h2 style={{fontFamily:FH,fontWeight:900,fontSize:"clamp(22px,3vw,30px)",color:C.text,letterSpacing:"-.02em"}}>
            {t({ru:"Руководства",tg:"Дастурҳо"})}
          </h2>
          <p style={{fontSize:14,color:C.sub,marginTop:4}}>
            {t({ru:"Пошаговые инструкции: как пользоваться EduHub родителям, соискателям и учреждениям",tg:"Дастурҳои қадам ба қадам: чӣ гуна аз EduHub истифода баранд волидайн, ҷуяндагони кор ва муассисаҳо"})}
          </p>
        </div>
        <div style={{display:"grid",gridTemplateColumns:"repeat(3,1fr)",gap:18}}>
          {[
            {
              photo: PHOTOS.heroGuideP, icon: Users,
              title: {ru:"Родителям и студентам", tg:"Ба волидайн ва донишҷӯён"},
              desc: {ru:"Бесплатно ищите, сравнивайте по 8 параметрам, читайте отзывы и пишите учреждениям напрямую в чате.",
                     tg:"Ройгон ҷустуҷӯ кунед, аз рӯи 8 меъёр муқоиса кунед, шарҳҳо хонед ва бевосита дар чат ба муассисаҳо нависед."},
              cta: {ru:"Подробнее для родителей", tg:"Муфассал барои волидайн"}, href:"/guide/parents",
            },
            {
              photo: PHOTOS.heroGuideA, icon: Briefcase,
              title: {ru:"Соискателям", tg:"Ба ҷуяндагони кор"},
              desc: {ru:"Находите открытые вакансии в образовательных учреждениях и откликайтесь напрямую в чате.",
                     tg:"Дар муассисаҳои таълимӣ ҷои холии корро ёбед ва бевосита дар чат ҷавоб диҳед."},
              cta: {ru:"Подробнее для соискателей", tg:"Муфассал барои ҷуяндагони кор"}, href:"/guide/applicants",
            },
            {
              photo: PHOTOS.heroGuideI, icon: Building2,
              title: {ru:"Учреждениям", tg:"Ба муассисаҳо"},
              desc: {ru:"Привлекайте новых учеников, получайте честную обратную связь и отвечайте на отзывы в своём кабинете.",
                     tg:"Хонандагони навро ҷалб кунед, бозхӯрди ростқавлона гиред ва дар кабинети худ ба шарҳҳо ҷавоб диҳед."},
              cta: {ru:"Подробнее для учреждений", tg:"Муфассал барои муассисаҳо"}, href:"/guide/institutions",
            },
          ].map(a=>(
            <div key={a.href} style={{borderRadius:18,overflow:"hidden",border:`1px solid ${C.border}`,background:C.s1}}>
              <div style={{position:"relative",height:120,overflow:"hidden"}}>
                <img src={a.photo} alt="" style={{width:"100%",height:"100%",objectFit:"cover"}}/>
                <div style={{position:"absolute",inset:0,background:`linear-gradient(180deg,${C.overlay}10 0%,${C.overlay}D8 100%)`}}/>
                <div style={{position:"absolute",bottom:12,left:14,width:38,height:38,borderRadius:10,background:"rgba(0,0,0,.45)",backdropFilter:"blur(8px)",border:"1px solid rgba(255,255,255,.2)",display:"flex",alignItems:"center",justifyContent:"center"}}>
                  <a.icon size={18} color="#fff"/>
                </div>
              </div>
              <div style={{padding:"20px 22px 24px"}}>
                <h3 style={{fontFamily:FH,fontWeight:800,fontSize:17,color:C.text,marginBottom:8}}>{t(a.title)}</h3>
                <p style={{fontSize:13.5,color:C.sub,lineHeight:1.65,marginBottom:16}}>{t(a.desc)}</p>
                <Link href={a.href} style={{display:"inline-flex",alignItems:"center",gap:6,fontFamily:FH,fontWeight:700,fontSize:13.5,color:C.teal,textDecoration:"none"}}>
                  {t(a.cta)} <ArrowRight size={14}/>
                </Link>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* ── CATEGORIES ── */}
      <section style={{maxWidth:1260,margin:"0 auto",padding:"64px 28px 0"}}>
        <div style={{display:"flex",justifyContent:"space-between",alignItems:"flex-end",marginBottom:28}}>
          <div>
            <h2 style={{fontFamily:FH,fontWeight:900,fontSize:"clamp(22px,3vw,32px)",color:C.text,letterSpacing:"-.02em"}}>{t("nav.categories")}</h2>
            <p style={{fontSize:14,color:C.sub,marginTop:4}}>{t({ru:"Выберите тип учреждения", tg:"Навъи муассисаро интихоб кунед"})}</p>
          </div>
        </div>
        <div ref={catRef} style={{display:"grid",gridTemplateColumns:"repeat(4,1fr)",gap:16}}>
          {CATEGORY_KEYS.map((k,i)=>{
            const meta = CATEGORY_META[k];
            const Icon = meta.icon;
            const count = institutions.filter(x=>x.tk===k).length;
            return (
              <Link key={k} href={`/search?type=${k}`}
                style={{borderRadius:18,overflow:"hidden",border:`1px solid ${C.border}`,background:C.s1,cursor:"pointer",textDecoration:"none",display:"block",...revealStyle(catVisible,i*60)}}>
                <div style={{position:"relative",height:128,overflow:"hidden"}}>
                  <img src={meta.heroPhoto} alt="" style={{width:"100%",height:"100%",objectFit:"cover"}}/>
                  <div style={{position:"absolute",inset:0,background:`linear-gradient(180deg,${C.overlay}10 0%,${C.overlay}D8 100%)`}}/>
                  <div style={{position:"absolute",top:12,left:12,width:36,height:36,borderRadius:10,background:"rgba(0,0,0,.45)",backdropFilter:"blur(8px)",border:"1px solid rgba(255,255,255,.2)",display:"flex",alignItems:"center",justifyContent:"center"}}>
                    <Icon size={17} color="#fff"/>
                  </div>
                  <p style={{position:"absolute",bottom:12,left:14,right:14,fontFamily:FH,fontWeight:800,fontSize:16.5,color:"#fff"}}>{t(meta.plural)}</p>
                </div>
                <div style={{padding:"14px 16px 16px"}}>
                  <p style={{fontSize:12.5,color:C.sub,marginBottom:12,minHeight:32}}>{t(meta.desc)}</p>
                  <div style={{display:"flex",justifyContent:"space-between",alignItems:"center"}}>
                    <span style={{fontSize:12,fontWeight:700,color:meta.color,fontFamily:FH}}>{count} {t({ru:"учр.", tg:"муасс."})}</span>
                    <ArrowRight size={14} style={{color:meta.color}}/>
                  </div>
                </div>
              </Link>
            );
          })}
        </div>
      </section>

      {/* ── NEARBY / TOP RATED ── */}
      <section style={{maxWidth:1260,margin:"0 auto",padding:"56px 28px 0"}}>
        <div style={{display:"flex",justifyContent:"space-between",alignItems:"flex-end",marginBottom:28}}>
          <div>
            <h2 style={{fontFamily:FH,fontWeight:900,fontSize:"clamp(22px,3vw,32px)",color:C.text,letterSpacing:"-.02em"}}>
              {myCoords ? t({ru:"Учреждения рядом с вами", tg:"Муассисаҳо дар наздикии шумо"})
                : region ? t({ru:"Учреждения в вашем регионе", tg:"Муассисаҳо дар минтақаи шумо"})
                : t({ru:"Лучшие учреждения", tg:"Беҳтарин муассисаҳо"})}
            </h2>
            <p style={{fontSize:14,color:C.sub,marginTop:4}}>
              {myCoords ? t({ru:"По расстоянию от вашего местоположения", tg:"Аз рӯи масофа аз ҷои шумо"})
                : region ? t(REGION_LABEL[region])
                : t({ru:"По оценкам родителей Таджикистана", tg:"Аз рӯи бақои волидайни Тоҷикистон"})}
            </p>
          </div>
          <div style={{display:"flex",gap:8,alignItems:"center"}}>
            {!myCoords && (
              <button onClick={locateMe} disabled={locating} style={{display:"flex",alignItems:"center",gap:6,fontFamily:FH,fontWeight:700,fontSize:13.5,color:C.sub,padding:"8px 14px",borderRadius:9,border:`1px solid ${C.border}`,background:"transparent",cursor:locating?"default":"pointer"}}>
                <LocateFixed size={14}/> {locating ? t("nav.gpsLocating") : t({ru:"Точнее по GPS", tg:"Аниқтар аз рӯи GPS"})}
              </button>
            )}
            <Link href="/search" style={{display:"flex",alignItems:"center",gap:5,fontFamily:FH,fontWeight:700,fontSize:13.5,color:C.teal,padding:"8px 14px",borderRadius:9,border:`1px solid ${C.teal}40`,background:`${C.teal}10`,textDecoration:"none"}}>
              {t({ru:"Все учреждения", tg:"Ҳама муассисаҳо"})} <ChevronRight size={14}/>
            </Link>
          </div>
        </div>
        <div ref={topRef} style={{display:"grid",gridTemplateColumns:"repeat(4,1fr)",gap:18}}>
          {nearby.map((inst,i)=>(
            <div key={inst.id} style={revealStyle(topVisible,i*70)}>
              <InstitCard inst={inst} onClick={()=>router.push(`/institutions/${inst.id}`)}/>
            </div>
          ))}
        </div>
      </section>

      {/* ── HOW IT WORKS ── */}
      <section ref={howRef} style={{maxWidth:1260,margin:"0 auto",padding:"56px 28px 80px"}}>
        <h2 style={{fontFamily:FH,fontWeight:900,fontSize:"clamp(22px,3vw,32px)",color:C.text,textAlign:"center",letterSpacing:"-.02em",marginBottom:8}}>{t({ru:"Как это работает", tg:"Ин чӣ гуна кор мекунад"})}</h2>
        <p style={{fontSize:14,color:C.sub,textAlign:"center",marginBottom:44}}>{t({ru:"Четыре шага до осознанного выбора", tg:"Чор қадам то интихоби огоҳона"})}</p>
        <div style={{display:"grid",gridTemplateColumns:"repeat(4,1fr)",gap:16}}>
          {[
            {photo:PHOTOS.student,icon:SlidersHorizontal,t:{ru:"Ищите",tg:"Ҷустуҷӯ кунед"},d:{ru:"Фильтруйте по району, типу, цене, наличию развозки и питания", tg:"Аз рӯи ноҳия, навъ, нарх, мавҷудияти расонидан ва ғизо филтр кунед"}},
            {photo:PHOTOS.classroom2,icon:Star,t:{ru:"Изучайте рейтинг",tg:"Рейтингро омӯзед"},d:{ru:"Честная оценка по 8 параметрам — от качества обучения до безопасности", tg:"Бақои ростқавлона аз рӯи 8 меъёр — аз сифати таълим то бехатарӣ"}},
            {photo:PHOTOS.library,icon:MessageCircle,t:{ru:"Читайте отзывы",tg:"Шарҳҳо хонед"},d:{ru:"Честные мнения родителей и студентов помогут сделать осознанный выбор", tg:"Ақидаҳои ростқавлонаи волидайн ва донишҷӯён ба интихоби огоҳона мадад мерасонанд"}},
            {photo:PHOTOS.school1,icon:Send,t:{ru:"Свяжитесь",tg:"Тамос гиред"},d:{ru:"Напишите учреждению напрямую в чате и запишитесь на визит", tg:"Бевосита дар чат ба муассиса нависед ва ба ташриф сабти ном шавед"}},
          ].map((s,i)=>{
            const Icon = s.icon;
            return (
            <div key={s.t.ru} style={{position:"relative",borderRadius:18,overflow:"hidden",height:220,border:`1px solid ${C.border}`,...revealStyle(howVisible,i*80)}}>
              <img src={s.photo} alt="" style={{position:"absolute",inset:0,width:"100%",height:"100%",objectFit:"cover"}}/>
              <div style={{position:"absolute",inset:0,background:`linear-gradient(180deg, ${C.overlay}26 0%, ${C.overlay}59 45%, ${C.overlay}EB 100%)`}}/>
              <div style={{position:"absolute",top:14,left:14,width:38,height:38,borderRadius:11,background:"rgba(255,255,255,.16)",backdropFilter:"blur(8px)",border:"1px solid rgba(255,255,255,.28)",display:"flex",alignItems:"center",justifyContent:"center"}}>
                <Icon size={18} color="#fff"/>
              </div>
              <div style={{position:"absolute",bottom:0,left:0,right:0,padding:"16px 18px"}}>
                <p style={{fontFamily:FH,fontWeight:800,fontSize:15.5,color:"#fff",marginBottom:5}}>{t(s.t)}</p>
                <p style={{fontSize:11.5,color:"rgba(255,255,255,.78)",lineHeight:1.5}}>{t(s.d)}</p>
              </div>
            </div>
            );
          })}
        </div>
      </section>
    </div>
  );
}
