"use client";

import Link from "next/link";
import { useParams, usePathname, useRouter, useSearchParams } from "next/navigation";
import { ArrowLeft, Mail, Phone, Award, GraduationCap, Medal, Trophy, type LucideIcon } from "lucide-react";
import { C, FH, FB, ALL_STAFF } from "@/lib/data";
import { chatHref } from "@/lib/chat-window";
import { AchievementModal } from "@/components/AchievementModal";
import { useT } from "@/lib/i18n";

const ACH_TIER: Record<string,{icon:LucideIcon;color:string}> = {
  gold:{icon:Medal,color:C.gold}, silver:{icon:Medal,color:C.sub}, bronze:{icon:Medal,color:C.goldD}, special:{icon:Trophy,color:C.teal},
};

export default function PersonProfilePage() {
  const params = useParams<{ id: string }>();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const router = useRouter();
  const t = useT();
  const person = ALL_STAFF.find(p=>p.id===params.id);

  if(!person) return <div style={{padding:40,color:C.muted,textAlign:"center"}}>{t({ru:"Сотрудник не найден",tg:"Корманд ёфт нашуд"})}</div>;

  return (
    <div style={{maxWidth:900,margin:"0 auto",padding:"28px 28px 80px",fontFamily:FB}}>
      <button onClick={()=>router.push(`/institutions/${person.instId}`)} style={{display:"inline-flex",alignItems:"center",gap:7,fontFamily:FH,fontWeight:700,fontSize:13.5,color:C.teal,marginBottom:24,padding:"8px 14px",borderRadius:9,border:`1px solid ${C.teal}40`,background:`${C.teal}10`,cursor:"pointer"}}>
        <ArrowLeft size={15}/> {t({ru:"Назад к профилю",tg:"Бозгашт ба профил"})}
      </button>

      <div className="eh-mobile-1col" style={{display:"grid",gridTemplateColumns:"300px 1fr",gap:24,alignItems:"start"}}>
        {/* ── LEFT CARD ── */}
        <div style={{borderRadius:18,overflow:"hidden",border:`1px solid ${C.border}`,background:C.s1}}>
          <div style={{height:280,overflow:"hidden",position:"relative"}}>
            <img src={person.photo} alt={t(person.name)} style={{width:"100%",height:"100%",objectFit:"cover"}}/>
            <div style={{position:"absolute",inset:0,background:`linear-gradient(180deg,transparent 50%,${C.overlay}F2 100%)`}}/>
            <div style={{position:"absolute",bottom:16,left:16,right:16}}>
              <p style={{fontFamily:FH,fontWeight:900,fontSize:20,color:"#fff",lineHeight:1.2}}>{t(person.name)}</p>
              <p style={{fontSize:13.5,color:C.teal,fontFamily:FH,marginTop:3}}>{t(person.roleLabel)}</p>
            </div>
          </div>

          <div style={{padding:18}}>
            {person.subject && (
              <div style={{display:"flex",gap:8,flexWrap:"wrap",marginBottom:16}}>
                <span style={{fontSize:12,fontWeight:600,padding:"4px 10px",borderRadius:7,background:`${C.teal}18`,color:C.teal,fontFamily:FH}}>{t(person.subject)}</span>
              </div>
            )}

            <div style={{display:"flex",flexDirection:"column",gap:8}}>
              <div style={{display:"flex",alignItems:"center",gap:8,fontSize:13.5,color:C.sub}}>
                <GraduationCap size={15} style={{color:C.teal,flexShrink:0}}/>
                {t({ru:"Опыт:",tg:"Таҷриба:"})} <b style={{color:C.text}}>{person.exp}</b>
              </div>
              {person.email && (
                <div style={{display:"flex",alignItems:"center",gap:8,fontSize:13,color:C.sub}}>
                  <Mail size={14} style={{color:C.teal,flexShrink:0}}/>{person.email}
                </div>
              )}
              {person.phone && (
                <div style={{display:"flex",alignItems:"center",gap:8,fontSize:13,color:C.sub}}>
                  <Phone size={14} style={{color:C.teal,flexShrink:0}}/>{person.phone}
                </div>
              )}
              {person.rating && (
                <div style={{display:"flex",alignItems:"center",gap:8,fontSize:13.5,color:C.sub}}>
                  <span style={{color:C.gold}}>★</span>
                  <b style={{color:C.text}}>{person.rating}</b>
                  <span style={{color:C.muted}}>({person.reviews} {t("common.reviews")})</span>
                </div>
              )}
            </div>

            <Link href={chatHref(person.instId)} style={{marginTop:16,width:"100%",padding:"10px",borderRadius:12,background:C.teal,color:C.overlay,fontFamily:FH,fontWeight:700,fontSize:13.5,display:"flex",alignItems:"center",justifyContent:"center",gap:7,cursor:"pointer",border:"none",textDecoration:"none"}}>
              <Mail size={14}/> {t("common.write")}
            </Link>
          </div>
        </div>

        {/* ── RIGHT ── */}
        <div style={{display:"flex",flexDirection:"column",gap:20}}>
          {/* Bio */}
          <div style={{borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:24}}>
            <h2 style={{fontFamily:FH,fontWeight:800,fontSize:18,color:C.text,marginBottom:12}}>{t({ru:"О себе",tg:"Дар бораи худ"})}</h2>
            <p style={{fontSize:14.5,color:C.sub,lineHeight:1.75}}>{t(person.bio)}</p>
          </div>

          {/* Education */}
          {person.education.length > 0 && (
            <div style={{borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:24}}>
              <h2 style={{fontFamily:FH,fontWeight:800,fontSize:18,color:C.text,marginBottom:16,display:"flex",alignItems:"center",gap:8}}>
                <GraduationCap size={18} style={{color:C.teal}}/> {t({ru:"Образование",tg:"Таҳсилот"})}
              </h2>
              <div style={{display:"flex",flexDirection:"column",gap:12}}>
                {person.education.map((ed,i)=>(
                  <div key={i} style={{display:"flex",gap:14,alignItems:"flex-start"}}>
                    <div style={{width:8,height:8,borderRadius:"50%",background:C.teal,marginTop:6,flexShrink:0}}/>
                    <p style={{fontSize:14,color:C.sub,lineHeight:1.6}}>{t(ed)}</p>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Achievements */}
          {person.achievements.length > 0 && (
            <div style={{borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:24}}>
              <h2 style={{fontFamily:FH,fontWeight:800,fontSize:18,color:C.text,marginBottom:16,display:"flex",alignItems:"center",gap:8}}>
                <Award size={18} style={{color:C.gold}}/> {t("tab.achievements")}
              </h2>
              <div style={{display:"flex",flexDirection:"column",gap:10}}>
                {person.achievements.map(ach=>{
                  const p = new URLSearchParams(searchParams.toString());
                  p.set("achievement", ach.id);
                  const tier = ACH_TIER[ach.category] ?? ACH_TIER.special;
                  const TierIcon = tier.icon;
                  return (
                    <Link key={ach.id} href={`${pathname}?${p.toString()}`} scroll={false}
                      style={{display:"flex",gap:12,alignItems:"flex-start",padding:"12px 14px",borderRadius:12,background:C.s2,textDecoration:"none",cursor:"pointer"}}>
                      <TierIcon size={20} style={{color:tier.color,flexShrink:0,marginTop:1}}/>
                      <div>
                        <p style={{fontFamily:FH,fontWeight:700,fontSize:14,color:C.text}}>{t(ach.title)}</p>
                        <p style={{fontSize:12.5,color:C.sub}}>{ach.year} · {t(ach.desc)}</p>
                      </div>
                    </Link>
                  );
                })}
              </div>
            </div>
          )}
        </div>
      </div>
      <AchievementModal achievements={person.achievements}/>
    </div>
  );
}
