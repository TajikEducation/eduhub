"use client";

import { useState } from "react";
import { CheckCircle, Heart, Bus, UtensilsCrossed } from "lucide-react";
import { C, FH, CATEGORY_META, CURRICULUM_META, type CategoryKey, type Bi, type CurriculumKey } from "@/lib/data";
import { useAppState } from "@/lib/app-state";
import { useT } from "@/lib/i18n";

// Узкий тип вместо полного mock Institution — InstitCard читает только эти поля, что
// позволяет использовать её и с mock-данными (lib/data.ts), и с адаптированными
// карточками из реального backend (см. lib/backendTypes.ts) без риска сломать существующих
// вызывающих: любой mock Institution уже структурно satisfies этот более узкий тип.
export interface InstitCardData {
  id: number | string;
  tk: CategoryKey;
  color: string;
  coverPhoto: string;
  tag: Bi | null;
  area: string;
  name: Bi;
  score: number;
  rev: number;
  ver: boolean;
  transport: boolean;
  food: boolean;
  curriculum?: CurriculumKey[];
  price: number;
}

function Stars({ s, size=12 }:{s:number;size?:number}) {
  return (
    <span style={{display:"flex",gap:2}}>
      {[1,2,3,4,5].map(i=>(
        <svg key={i} width={size} height={size} viewBox="0 0 24 24" fill={i<=Math.round(s)?C.gold:C.dim}>
          <path d="M12 2l3 7h7l-5.5 4 2 7L12 16l-6.5 4 2-7L2 9h7z"/>
        </svg>
      ))}
    </span>
  );
}

export function InstitCard({ inst, onClick }: {
  inst: InstitCardData;
  onClick: ()=>void;
}) {
  const [hov,setHov] = useState(false);
  const { isSaved, toggleSaved } = useAppState();
  const t = useT();
  const saved = isSaved(inst.id);
  const meta = CATEGORY_META[inst.tk];
  const TypeIcon = meta.icon;

  return (
    <div onClick={onClick}
      onMouseEnter={()=>setHov(true)} onMouseLeave={()=>setHov(false)}
      style={{borderRadius:18,overflow:"hidden",cursor:"pointer",background:C.s1,border:`1px solid ${hov?inst.color+"55":C.border}`,transition:"all .22s",transform:hov?"translateY(-4px)":"none",boxShadow:hov?`0 20px 50px rgba(0,0,0,.5)`:undefined}}>

      {/* cover photo */}
      <div style={{height:170,position:"relative",overflow:"hidden",background:inst.color}}>
        <img src={inst.coverPhoto} alt={t(inst.name)} style={{width:"100%",height:"100%",objectFit:"cover",transition:"transform .3s",transform:hov?"scale(1.05)":"scale(1)"}}/>
        <div style={{position:"absolute",inset:0,background:`linear-gradient(180deg,rgba(0,0,0,.15) 0%,rgba(0,0,0,.55) 100%)`}}/>

        {/* tag */}
        {inst.tag && <span style={{position:"absolute",top:12,left:12,fontSize:11,fontWeight:800,fontFamily:FH,padding:"4px 10px",borderRadius:8,background:C.gold,color:C.overlay}}>{t(inst.tag)}</span>}

        {/* save */}
        <button onClick={e=>{e.stopPropagation();toggleSaved(inst.id)}} aria-label={saved?"Убрать из избранного":"В избранное"} style={{position:"absolute",top:10,right:10,width:32,height:32,borderRadius:"50%",background:"rgba(0,0,0,.4)",border:`1px solid rgba(255,255,255,.2)`,display:"flex",alignItems:"center",justifyContent:"center",transition:"transform .15s",transform:saved?"scale(1.08)":"scale(1)"}}>
          <Heart size={13} fill={saved?C.red:"none"} stroke={saved?C.red:"#fff"}/>
        </button>

        {/* bottom meta */}
        <div style={{position:"absolute",bottom:12,left:12,display:"flex",alignItems:"center",gap:7}}>
          <div style={{width:34,height:34,borderRadius:8,background:"rgba(255,255,255,.15)",backdropFilter:"blur(8px)",display:"flex",alignItems:"center",justifyContent:"center"}}>
            <TypeIcon size={17} color="#fff"/>
          </div>
          <div>
            <p style={{fontSize:10.5,fontWeight:700,color:"rgba(255,255,255,.65)",fontFamily:FH,textTransform:"uppercase",letterSpacing:".05em"}}>{t(meta.label)}</p>
            <p style={{fontSize:12,fontWeight:600,color:"#fff",fontFamily:FH}}>{inst.area}</p>
          </div>
        </div>
      </div>

      <div style={{padding:"16px 16px 18px"}}>
        <h3 style={{fontFamily:FH,fontWeight:800,fontSize:15.5,lineHeight:1.3,color:C.text,marginBottom:8,minHeight:40}}>{t(inst.name)}</h3>

        <div style={{display:"flex",alignItems:"center",gap:6,marginBottom:10}}>
          <Stars s={inst.score}/>
          <span style={{fontFamily:FH,fontWeight:700,fontSize:12.5,color:C.text}}>{inst.score}</span>
          <span style={{fontSize:12,color:C.muted}}>({inst.rev})</span>
          {inst.ver && <span style={{marginLeft:"auto",fontSize:10.5,fontWeight:700,color:C.ok,display:"flex",alignItems:"center",gap:3}}><CheckCircle size={10}/> {t("common.verified")}</span>}
        </div>

        <div style={{display:"flex",flexWrap:"wrap",gap:5,marginBottom:13}}>
          {inst.transport && <span style={{display:"flex",alignItems:"center",gap:4,fontSize:10.5,fontWeight:600,padding:"3px 8px",borderRadius:8,background:`${C.teal}18`,color:C.teal,fontFamily:FH}}><Bus size={11}/> {t({ru:"Развозка",tg:"Расонидан"})}</span>}
          {inst.food      && <span style={{display:"flex",alignItems:"center",gap:4,fontSize:10.5,fontWeight:600,padding:"3px 8px",borderRadius:8,background:`${C.s3}`,color:C.sub,fontFamily:FH}}><UtensilsCrossed size={11}/> {t({ru:"Питание",tg:"Ғизо"})}</span>}
          {inst.curriculum?.map(ck => (
            <span key={ck} style={{fontSize:10.5,fontWeight:600,padding:"3px 8px",borderRadius:8,background:C.s3,color:C.sub,fontFamily:FH}}>{t(CURRICULUM_META[ck].label)}</span>
          ))}
        </div>

        <div style={{display:"flex",alignItems:"center",justifyContent:"space-between",paddingTop:12,borderTop:`1px solid ${C.border}`}}>
          <div>
            <span style={{fontFamily:FH,fontWeight:900,fontSize:20,color:C.text}}>{inst.price}</span>
            <span style={{fontSize:11.5,color:C.muted,marginLeft:3}}>{t("common.perMonth")}</span>
          </div>
        </div>
      </div>
    </div>
  );
}
