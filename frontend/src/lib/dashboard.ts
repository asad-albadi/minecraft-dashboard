export type Preferences = { theme:'stone'|'obsidian'; density:'comfortable'|'compact'; timezone:string; hour12:boolean; refresh:number; landing:string; autoScroll:boolean; notifications:boolean };
export const defaults: Preferences = {theme:'stone',density:'comfortable',timezone:'UTC',hour12:false,refresh:30,landing:'/',autoScroll:true,notifications:true};
export function preferences(raw:string|null):Preferences {
  try { const p=JSON.parse(raw||'{}'); let timezone='UTC'; try { new Intl.DateTimeFormat('en',{timeZone:p.timezone||'UTC'}); timezone=p.timezone||'UTC'; } catch {}
    return {theme:p.theme==='obsidian'?'obsidian':'stone',density:p.density==='compact'?'compact':'comfortable',timezone,hour12:p.hour12===true,refresh:[15,30,60,120].includes(p.refresh)?p.refresh:30,landing:['/','/players','/world','/console','/backups','/settings'].includes(p.landing)?p.landing:'/',autoScroll:p.autoScroll!==false,notifications:p.notifications!==false};
  } catch { return {...defaults}; }
}
export function parseChat(line:string) {
  const m=line.match(/^\[([^\]]+)\] \[[^\]]+\/INFO\]: (?:\[Not Secure\] )?<([A-Za-z0-9_]{1,16})> (.*)$/);
  return m?{time:m[1],name:m[2],message:m[3]}:null;
}
export function filterLogs(lines:string[],search:string,severity:string) {
  return lines.filter(line=>line.toLowerCase().includes(search.toLowerCase()) && (severity==='all'||new RegExp('^\\[[^\\]]+\\] \\[[^\\]]+/'+severity+'\\]:').test(line)));
}
type SortablePlayer={name:string;operator?:boolean;playTime?:string;deaths?:string;mobKills?:string;firstJoin?:string;lastJoin?:string};
export function selectPlayers<T extends SortablePlayer>(rows:T[],online:Set<string>|null,search:string,filter:string,sort:string):T[] {
  return rows.filter(p=>p.name.toLowerCase().includes(search.toLowerCase()) && (filter==='all'||filter==='online'&&online?.has(p.name.toLowerCase())||filter==='offline'&&online!==null&&!online.has(p.name.toLowerCase())||filter==='operators'&&p.operator||filter==='recent'&&!!p.lastJoin&&Date.now()-Date.parse(p.lastJoin)<30*86400000)).sort((a,b)=>{
    if(sort==='name') return a.name.localeCompare(b.name);
    const av=a[sort as keyof T], bv=b[sort as keyof T];
    const numeric=['playTime','deaths','mobKills'].includes(sort);
    const x=numeric?parseFloat(String(av)):Date.parse(String(av)), y=numeric?parseFloat(String(bv)):Date.parse(String(bv));
    if(!Number.isFinite(x)) return Number.isFinite(y)?1:0;
    if(!Number.isFinite(y)) return -1;
    return sort==='firstJoin'?x-y:y-x;
  });
}
export function dateLabel(value:string|undefined,p:Preferences) {
  if(!value) return 'Unknown';
  if(/^\d{4}-\d{2}-\d{2}$/.test(value)) return value+' (server date)';
  const date=new Date(value); return Number.isNaN(date.getTime())?value:new Intl.DateTimeFormat(undefined,{timeZone:p.timezone,dateStyle:'medium',timeStyle:'short',hour12:p.hour12}).format(date);
}
