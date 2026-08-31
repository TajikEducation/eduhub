"use client";

import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Mail, Phone, GraduationCap } from "lucide-react";
import { C, FH, FB } from "@/lib/data";
import { useT } from "@/lib/i18n";
import { useGetStaffMemberQuery } from "../api/peopleApi";
import { useCreateConversationMutation } from "@/app/(site)/messages/api/chatApi";
import { useMeQuery } from "@/app/(site)/login/api/authApi";
import { getAccessToken } from "@/lib/authToken";
import { toast } from "sonner";

export default function PersonProfilePage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const t = useT();
  const { data: person, isLoading, isError } = useGetStaffMemberQuery(params.id);
  const { data: me } = useMeQuery(undefined, { skip: !getAccessToken() });
  const [createConversation] = useCreateConversationMutation();

  async function startChat() {
    if (!person) return;
    if (!me) { router.push("/login"); return; }
    try {
      const conv = await createConversation({ counterpart_type: "institution", counterpart_id: person.institution_id }).unwrap();
      router.push(`/messages?conv=${conv.id}`);
    } catch {
      toast.error(t({ ru: "Не удалось открыть чат", tg: "Чатро кушода натавонист" }));
    }
  }

  if (isLoading) {
    return <div style={{padding:40,color:C.muted,textAlign:"center"}}>{t({ru:"Загрузка…",tg:"Боркунӣ…"})}</div>;
  }
  if (isError || !person) {
    return <div style={{padding:40,color:C.muted,textAlign:"center"}}>{t({ru:"Сотрудник не найден",tg:"Корманд ёфт нашуд"})}</div>;
  }

  return (
    <div style={{maxWidth:900,margin:"0 auto",padding:"28px 28px 80px",fontFamily:FB}}>
      <button onClick={()=>router.push(`/institutions/${person.institution_id}`)} style={{display:"inline-flex",alignItems:"center",gap:7,fontFamily:FH,fontWeight:700,fontSize:13.5,color:C.teal,marginBottom:24,padding:"8px 14px",borderRadius:9,border:`1px solid ${C.teal}40`,background:`${C.teal}10`,cursor:"pointer"}}>
        <ArrowLeft size={15}/> {t({ru:"Назад к профилю",tg:"Бозгашт ба профил"})}
      </button>

      <div className="eh-mobile-1col" style={{display:"grid",gridTemplateColumns:"300px 1fr",gap:24,alignItems:"start"}}>
        {/* ── LEFT CARD ── */}
        <div style={{borderRadius:18,overflow:"hidden",border:`1px solid ${C.border}`,background:C.s1}}>
          <div style={{height:280,overflow:"hidden",position:"relative"}}>
            <img src={person.photo_url || "/logo.svg"} alt={person.name.ru} style={{width:"100%",height:"100%",objectFit:"cover"}}/>
            <div style={{position:"absolute",inset:0,background:`linear-gradient(180deg,transparent 50%,${C.overlay}F2 100%)`}}/>
            <div style={{position:"absolute",bottom:16,left:16,right:16}}>
              <p style={{fontFamily:FH,fontWeight:900,fontSize:20,color:"#fff",lineHeight:1.2}}>{person.name.ru}</p>
              <p style={{fontSize:13.5,color:C.teal,fontFamily:FH,marginTop:3}}>{person.role_label.ru}</p>
            </div>
          </div>

          <div style={{padding:18}}>
            {person.subject && (
              <div style={{display:"flex",gap:8,flexWrap:"wrap",marginBottom:16}}>
                <span style={{fontSize:12,fontWeight:600,padding:"4px 10px",borderRadius:7,background:`${C.teal}18`,color:C.teal,fontFamily:FH}}>{person.subject.ru}</span>
              </div>
            )}

            <div style={{display:"flex",flexDirection:"column",gap:8}}>
              {person.exp && (
                <div style={{display:"flex",alignItems:"center",gap:8,fontSize:13.5,color:C.sub}}>
                  <GraduationCap size={15} style={{color:C.teal,flexShrink:0}}/>
                  {t({ru:"Опыт:",tg:"Таҷриба:"})} <b style={{color:C.text}}>{person.exp}</b>
                </div>
              )}
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
            </div>

            <button onClick={startChat} style={{marginTop:16,width:"100%",padding:"10px",borderRadius:12,background:C.teal,color:C.overlay,fontFamily:FH,fontWeight:700,fontSize:13.5,display:"flex",alignItems:"center",justifyContent:"center",gap:7,cursor:"pointer",border:"none"}}>
              <Mail size={14}/> {t("common.write")}
            </button>
          </div>
        </div>

        {/* ── RIGHT ── */}
        <div style={{display:"flex",flexDirection:"column",gap:20}}>
          {person.bio && (
            <div style={{borderRadius:18,border:`1px solid ${C.border}`,background:C.s1,padding:24}}>
              <h2 style={{fontFamily:FH,fontWeight:800,fontSize:18,color:C.text,marginBottom:12}}>{t({ru:"О себе",tg:"Дар бораи худ"})}</h2>
              <p style={{fontSize:14.5,color:C.sub,lineHeight:1.75}}>{person.bio.ru}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
