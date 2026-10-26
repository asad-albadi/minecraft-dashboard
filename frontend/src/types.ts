export type Identity={username:string;role:'admin'|'viewer'};
export type Status={state:'online'|'offline'|'unknown';players?:number;maxPlayers?:number;version?:string;description?:string;latencyMs?:number;rcon?:string;detail?:string;updatedAt?:string;lastSuccess?:string};
export type World={name?:string;gameMode?:string;difficulty?:string;seed?:string;dimension?:string;size?:string;lastBackup?:string;day?:string;time?:string;weather?:string;hardcore?:string;generator?:string;collectedAt?:string;dimensions?:Record<string,string>};
export type Player={id:string;name:string};
export type Profile={name:string;historyNote?:string;uuid?:string;operator:boolean;playTime:string;deaths:string;walked:string;mobKills:string;joinEvents?:string;firstJoin?:string;lastJoin?:string;stats?:Record<string,Record<string,number>>;advancements?:Array<{id:string;category:string;done:boolean;date?:string;criteria?:Record<string,string>}>;activity?:Array<{type:string;timestamp:string;message?:string}>};
export type Settings={serverName:string;host:string;port:string;integrations:Record<string,string>};
export type Backup={name:string;size:number;modifiedAt:string;type:string;integrity:string};
