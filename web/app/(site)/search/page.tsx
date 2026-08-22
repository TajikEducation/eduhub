"use client";

import { Suspense, useEffect, useState } from "react";
import { useRouter, usePathname, useSearchParams } from "next/navigation";
import Link from "next/link";
import { Search, SlidersHorizontal, ChevronDown, X, Map } from "lucide-react";
import { C, FH, FB, CATEGORY_META, REGION_LABEL, REGION_ORDER, type CategoryKey } from "@/lib/data";
import { InstitCard } from "@/components/InstitCard";
import { useReveal, revealStyle } from "@/lib/useReveal";
import { useVisibleInstitutions } from "@/lib/app-state";
import { useT } from "@/lib/i18n";

type SortKey = "score"|"price_asc"|"price_desc"|"reviews";
const CATEGORY_KEYS: CategoryKey[] = ["cat_kg", "cat_school", "cat_center", "cat_uni"];

function Chip({l,on,click}:{l:string;on:boolean;click:()=>void}) {
  return (
    <button onClick={click} style={{fontSize:12.5,fontWeight:600,padding:"6px 12px",borderRadius:8,border:`1px solid ${on?C.teal:C.border}`,background:on?`${C.teal}22`:"transparent",color:on?C.teal:C.sub,transition:"all .13s",fontFamily:FH,whiteSpace:"nowrap",cursor:"pointer"}}>
      {l}
    </button>
  );
}

function Toggle({l,v,fn,color=C.ok}:{l:string;v:boolean;fn:()=>void;color?:string}) {
  return (
    <div style={{display:"flex",alignItems:"center",justifyContent:"space-between",fontSize:13.5,color:v?C.text:C.sub,padding:"6px 0",fontFamily:FB}}>
      <span>{l}</span>
      <button onClick={fn} style={{width:40,height:22,borderRadius:999,position:"relative",background:v?color:C.s3,transition:"background .18s",flexShrink:0,border:"none",cursor:"pointer"}}>
        <span style={{position:"absolute",top:3,width:16,height:16,borderRadius:"50%",background:"#fff",transition:"left .18s",left:v?21:3,boxShadow:"0 1px 3px rgba(0,0,0,.3)"}}/>
      </button>
    </div>
  );
}

function SectionLabel({l}:{l:string}) {
  return <p style={{fontSize:11,fontWeight:700,textTransform:"uppercase",letterSpacing:".08em",color:C.dim,marginBottom:10,fontFamily:FH}}>{l}</p>;
}

function Divider() {
  return <div style={{borderTop:`1px solid ${C.border}`,margin:"14px 0"}}/>;
}

export default function SearchPage() {
  return (
    <Suspense fallback={null}>
      <SearchPageInner />
    </Suspense>
  );
}

function SearchPageInner() {
  const router = useRouter();
  const pathname = usePathname();
  const urlParams = useSearchParams();
  const t = useT();
  const [q,        setQ]        = useState(urlParams.get("q") ?? "");
  const [typeF,    setTypeF]    = useState<CategoryKey|null>((urlParams.get("type") as CategoryKey) || null);
  const [regionF,  setRegionF]  = useState<string|null>(urlParams.get("region"));
  const [areaF,    setAreaF]    = useState<string[]>(urlParams.get("area")?.split(",").filter(Boolean) ?? []);
  const [minPrice, setMinPrice] = useState(Number(urlParams.get("minPrice") ?? 0));
  const [maxPrice, setMaxPrice] = useState(Number(urlParams.get("maxPrice") ?? 2000));
  const [ratingF,  setRatingF]  = useState(Number(urlParams.get("rating") ?? 0));
  const [transF,   setTransF]   = useState(urlParams.get("transport") === "1");
  const [foodF,    setFoodF]    = useState(urlParams.get("food") === "1");
  const [verF,     setVerF]     = useState(urlParams.get("verified") === "1");
  const [sort,     setSort]     = useState<SortKey>((urlParams.get("sort") as SortKey) || "score");
  const [showFilters, setShowFilters] = useState(true);
  const { ref: resultsRef, visible: resultsVisible } = useReveal<HTMLDivElement>();

  // URL — источник истины для фильтров: отфильтрованный вид можно расшарить/забукмаркить (без перезагрузки страницы)
  useEffect(() => {
    const p = new URLSearchParams();
    if (q) p.set("q", q);
    if (typeF) p.set("type", typeF);
    if (regionF) p.set("region", regionF);
    if (areaF.length > 0) p.set("area", areaF.join(","));
    if (minPrice > 0) p.set("minPrice", String(minPrice));
    if (maxPrice < 2000) p.set("maxPrice", String(maxPrice));
    if (ratingF > 0) p.set("rating", String(ratingF));
    if (transF) p.set("transport", "1");
    if (foodF) p.set("food", "1");
    if (verF) p.set("verified", "1");
    if (sort !== "score") p.set("sort", sort);
    const qs = p.toString();
    router.replace(qs ? `${pathname}?${qs}` : pathname, { scroll: false });
  }, [q, typeF, regionF, areaF, minPrice, maxPrice, ratingF, transF, foodF, verF, sort, router, pathname]);

  const toggleArea = (a:string) =>
    setAreaF(prev => prev.includes(a) ? prev.filter(x=>x!==a) : [...prev, a]);

  const activeCount = [typeF,regionF,areaF.length>0,ratingF>0,transF,foodF,verF,minPrice>0,maxPrice<2000].filter(Boolean).length;

  const clearAll = () => { setTypeF(null); setRegionF(null); setAreaF([]); setRatingF(0); setTransF(false); setFoodF(false); setVerF(false); setMinPrice(0); setMaxPrice(2000); };

  const institutions = useVisibleInstitutions();
  let results = institutions.filter(i => {
    if(q) {
      const needle = q.toLowerCase();
      const hay = [i.name.ru, i.name.tg, i.area, t(CATEGORY_META[i.tk].label)].join(" ").toLowerCase();
      if (!hay.includes(needle)) return false;
    }
    if(typeF   && i.tk!==typeF)         return false;
    if(regionF && i.region!==regionF)   return false;
    if(areaF.length>0 && !areaF.includes(i.area)) return false;
    if(i.score < ratingF)              return false;
    if(i.price < minPrice || i.price > maxPrice) return false;
    if(transF  && !i.transport)        return false;
    if(foodF   && !i.food)             return false;
    if(verF    && !i.ver)              return false;
    return true;
  });

  if(sort==="score")      results = [...results].sort((a,b)=>b.score-a.score);
  if(sort==="price_asc")  results = [...results].sort((a,b)=>a.price-b.price);
  if(sort==="price_desc") results = [...results].sort((a,b)=>b.price-a.price);
  if(sort==="reviews")    results = [...results].sort((a,b)=>b.rev-a.rev);

  return (
    <div style={{maxWidth:1260,margin:"0 auto",padding:"28px 28px 72px"}}>
      {/* top bar */}
      <div style={{display:"flex",alignItems:"center",gap:12,marginBottom:20,flexWrap:"wrap"}}>
        <div style={{flex:1,minWidth:220,display:"flex",alignItems:"center",gap:10,background:C.s1,border:`1px solid ${C.border}`,borderRadius:12,padding:"11px 16px"}}>
          <Search size={16} style={{color:C.teal,flexShrink:0}}/>
          <input value={q} onChange={e=>setQ(e.target.value)} placeholder={t({ru:"Поиск по названию или типу…",tg:"Ҷустуҷӯ аз рӯи ном ё навъ…"})}
            style={{flex:1,background:"transparent",border:"none",outline:"none",fontSize:14.5,color:C.text,fontFamily:FB}}/>
          {q && <button onClick={()=>setQ("")} style={{background:"none",border:"none",cursor:"pointer"}}><X size={15} style={{color:C.muted}}/></button>}
        </div>
        <button onClick={()=>setShowFilters(!showFilters)}
          style={{display:"flex",alignItems:"center",gap:7,padding:"11px 16px",borderRadius:12,fontFamily:FH,fontWeight:700,fontSize:13.5,border:`1px solid ${activeCount>0?C.teal:C.border}`,background:activeCount>0?`${C.teal}18`:"transparent",color:activeCount>0?C.teal:C.sub,cursor:"pointer"}}>
          <SlidersHorizontal size={15}/>
          {t({ru:"Фильтры",tg:"Филтрҳо"})}
          {activeCount>0 && <span style={{background:C.teal,color:C.overlay,borderRadius:"50%",width:18,height:18,display:"flex",alignItems:"center",justifyContent:"center",fontSize:11,fontWeight:800}}>{activeCount}</span>}
        </button>
        <div style={{position:"relative"}}>
          <select value={sort} onChange={e=>setSort(e.target.value as SortKey)}
            style={{appearance:"none",padding:"11px 36px 11px 14px",borderRadius:12,border:`1px solid ${C.border}`,background:C.s1,color:C.text,fontFamily:FH,fontWeight:600,fontSize:13.5,outline:"none",cursor:"pointer"}}>
            <option value="score">{t({ru:"По рейтингу",tg:"Аз рӯи рейтинг"})}</option>
            <option value="price_asc">{t({ru:"Дешевле сначала",tg:"Аввал арзонтар"})}</option>
            <option value="price_desc">{t({ru:"Дороже сначала",tg:"Аввал гаронтар"})}</option>
            <option value="reviews">{t({ru:"По отзывам",tg:"Аз рӯи шарҳҳо"})}</option>
          </select>
          <ChevronDown size={14} style={{position:"absolute",right:12,top:"50%",transform:"translateY(-50%)",color:C.muted,pointerEvents:"none"}}/>
        </div>
        <Link href="/map" style={{display:"flex",alignItems:"center",gap:7,padding:"11px 16px",borderRadius:12,fontFamily:FH,fontWeight:700,fontSize:13.5,border:`1px solid ${C.border}`,color:C.sub,textDecoration:"none",whiteSpace:"nowrap"}}>
          <Map size={15}/> {t({ru:"На карте",tg:"Дар харита"})}
        </Link>
        <span style={{fontFamily:FB,fontSize:14,color:C.muted,whiteSpace:"nowrap"}}>
          <b style={{color:C.text}}>{results.length}</b> {t({ru:"учреждений",tg:"муассиса"})}
        </span>
      </div>

      <div style={{display:"grid",gridTemplateColumns:showFilters?"250px 1fr":"1fr",gap:22,alignItems:"start"}}>
        {showFilters && (
          <aside style={{borderRadius:16,border:`1px solid ${C.border}`,background:C.s1,position:"sticky",top:80,alignSelf:"start",overflow:"hidden"}}>
            <div style={{padding:"16px 18px",borderBottom:`1px solid ${C.border}`,display:"flex",justifyContent:"space-between",alignItems:"center"}}>
              <span style={{fontFamily:FH,fontWeight:800,fontSize:16,color:C.text}}>{t({ru:"Фильтры",tg:"Филтрҳо"})}</span>
              {activeCount>0 && <button onClick={clearAll} style={{fontSize:12.5,fontWeight:600,color:C.red,fontFamily:FH,background:"none",border:"none",cursor:"pointer"}}>{t("common.reset")}</button>}
            </div>

            <div style={{padding:"16px 18px",maxHeight:"calc(100vh - 200px)",overflowY:"auto"}}>
              <SectionLabel l={t({ru:"Тип учреждения",tg:"Навъи муассиса"})}/>
              <div style={{display:"grid",gridTemplateColumns:"1fr 1fr",gap:6}}>
                {CATEGORY_KEYS.map(k=>(
                  <Chip key={k} l={t(CATEGORY_META[k].label)} on={typeF===k} click={()=>setTypeF(typeF===k?null:k)}/>
                ))}
              </div>

              <Divider/>

              <SectionLabel l={t("geo.region")}/>
              <div style={{display:"flex",flexWrap:"wrap",gap:6}}>
                {REGION_ORDER.map(r=>(
                  <Chip key={r} l={t(REGION_LABEL[r])} on={regionF===r} click={()=>setRegionF(regionF===r?null:r)}/>
                ))}
              </div>

              <Divider/>

              <SectionLabel l={t({ru:"Район",tg:"Ноҳия"})}/>
              <div style={{display:"flex",flexWrap:"wrap",gap:6}}>
                {["Сино","Фирдавсӣ","И. Сомонӣ","Шоҳмансур"].map(a=>(
                  <Chip key={a} l={a} on={areaF.includes(a)} click={()=>toggleArea(a)}/>
                ))}
              </div>

              <Divider/>

              <SectionLabel l={t({ru:"Стоимость в месяц (сом)",tg:"Нархи моҳона (сомонӣ)"})}/>
              <div style={{marginBottom:8}}>
                <div style={{display:"flex",justifyContent:"space-between",fontSize:13,color:C.sub,fontFamily:FB,marginBottom:8}}>
                  <span>{minPrice === 0 ? t({ru:"Бесплатно",tg:"Ройгон"}) : `${minPrice} сом`}</span>
                  <span>{maxPrice >= 2000 ? t({ru:"Любая",tg:"Дилхоҳ"}) : `${maxPrice} сом`}</span>
                </div>
                <div style={{position:"relative",height:4,borderRadius:999,background:C.s3}}>
                  <div style={{position:"absolute",left:`${minPrice/20}%`,right:`${(2000-maxPrice)/20}%`,height:"100%",background:C.teal,borderRadius:999}}/>
                </div>
                <div style={{display:"flex",gap:8,marginTop:10}}>
                  <input type="range" min={0} max={1900} step={50} value={minPrice} onChange={e=>setMinPrice(+e.target.value)}
                    style={{flex:1,accentColor:C.teal}}/>
                  <input type="range" min={100} max={2000} step={50} value={maxPrice} onChange={e=>setMaxPrice(+e.target.value)}
                    style={{flex:1,accentColor:C.teal}}/>
                </div>
              </div>

              <Divider/>

              <SectionLabel l={t({ru:"Минимальный рейтинг",tg:"Рейтинги ҳадди ақал"})}/>
              <div style={{display:"flex",gap:6}}>
                {[0,3.5,4.0,4.5].map(r=>(
                  <Chip key={r} l={r===0?t({ru:"Любой",tg:"Дилхоҳ"}):`${r}★`} on={ratingF===r} click={()=>setRatingF(r)}/>
                ))}
              </div>

              <Divider/>

              <SectionLabel l={t({ru:"Дополнительно",tg:"Иловагӣ"})}/>
              <div style={{display:"flex",flexDirection:"column",gap:4}}>
                <Toggle l={t({ru:"Есть развозка",tg:"Расонидан мавҷуд аст"})}      v={transF} fn={()=>setTransF(!transF)} color={C.teal}/>
                <Toggle l={t({ru:"Есть питание",tg:"Ғизо мавҷуд аст"})}        v={foodF}  fn={()=>setFoodF(!foodF)}  color={C.teal}/>
                <Toggle l={t({ru:"Проверено МОН РТ",tg:"Аз ВМО ҶТ санҷида шудааст"})}    v={verF}   fn={()=>setVerF(!verF)}    color={C.ok}/>
              </div>
            </div>
          </aside>
        )}

        <div>
          {activeCount>0 && (
            <div style={{display:"flex",flexWrap:"wrap",gap:6,marginBottom:14}}>
              {typeF && <ActiveChip l={t(CATEGORY_META[typeF].label)} clear={()=>setTypeF(null)}/>}
              {regionF && <ActiveChip l={t(REGION_LABEL[regionF as keyof typeof REGION_LABEL])} clear={()=>setRegionF(null)}/>}
              {areaF.map(a=><ActiveChip key={a} l={a} clear={()=>toggleArea(a)}/>)}
              {ratingF>0 && <ActiveChip l={`${ratingF}★+`} clear={()=>setRatingF(0)}/>}
              {transF && <ActiveChip l={t({ru:"Развозка",tg:"Расонидан"})} clear={()=>setTransF(false)}/>}
              {foodF  && <ActiveChip l={t({ru:"Питание",tg:"Ғизо"})}  clear={()=>setFoodF(false)}/>}
              {verF   && <ActiveChip l={t("common.verified")}      clear={()=>setVerF(false)}/>}
            </div>
          )}

          {results.length===0 ? (
            <div style={{padding:56,borderRadius:16,border:`1px dashed ${C.border}`,textAlign:"center",color:C.muted}}>
              <Search size={32} style={{color:C.dim,margin:"0 auto 12px"}}/>
              <p style={{fontFamily:FH,fontWeight:800,fontSize:18,color:C.text,marginBottom:6}}>{t("empty.notFound")}</p>
              <p style={{fontFamily:FB,fontSize:14}}>{t({ru:"Попробуйте изменить фильтры",tg:"Филтрҳоро тағйир диҳед"})}</p>
              {activeCount>0 && <button onClick={clearAll} style={{marginTop:12,padding:"9px 20px",borderRadius:10,background:C.teal,color:C.overlay,fontFamily:FH,fontWeight:700,fontSize:13.5,border:"none",cursor:"pointer"}}>{t({ru:"Сбросить фильтры",tg:"Филтрҳоро тоза кардан"})}</button>}
            </div>
          ) : (
            <div ref={resultsRef} style={{display:"grid",gridTemplateColumns:showFilters?"1fr 1fr":"repeat(3,1fr)",gap:18}}>
              {results.map((inst,i)=>(
                <div key={inst.id} style={revealStyle(resultsVisible,Math.min(i,8)*45)}>
                  <InstitCard inst={inst} onClick={()=>router.push(`/institutions/${inst.id}`)}/>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function ActiveChip({l,clear}:{l:string;clear:()=>void}) {
  return (
    <span style={{display:"inline-flex",alignItems:"center",gap:5,fontSize:12.5,fontWeight:600,padding:"4px 10px",borderRadius:8,background:`${C.teal}20`,color:C.teal,fontFamily:FH}}>
      {l} <button onClick={clear} style={{display:"flex",alignItems:"center",background:"none",border:"none",cursor:"pointer",color:"inherit"}}><X size={12}/></button>
    </span>
  );
}
