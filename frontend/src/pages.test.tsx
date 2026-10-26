import {expect,test} from 'vitest';
import {renderToStaticMarkup} from 'react-dom/server';
import {WorldPage,BackupsPage,SettingsPage,PlayersPage,ConsoleView} from './pages';
import {defaults} from './lib/dashboard';
test('world page identifies unavailable data instead of empty placeholder',()=>{
 const html=renderToStaticMarkup(<WorldPage world={{}} prefs={defaults}/>);
 expect(html).toContain('Unavailable'); expect(html).toContain('World map');
});
test('backups never imply an unverified file is safe to restore',()=>{
 const html=renderToStaticMarkup(<BackupsPage backups={[]} error="Backup directory is not configured" prefs={defaults}/>);
 expect(html).toContain('not configured'); expect(html).toContain('Restore');
});
test('settings show role and actual preferences',()=>{
 const html=renderToStaticMarkup(<SettingsPage identity={{username:'reader',role:'viewer'}} settings={{serverName:'Minecraft Server',host:'',port:'25565',integrations:{}}} prefs={defaults} update={()=>{}} logout={()=>{}}/>);
 expect(html).toContain('viewer'); expect(html).toContain('Timezone');
});
test('console chat does not render system logs as messages',()=>{
 const html=renderToStaticMarkup(<ConsoleView lines={['[12:00:00] [Server thread/INFO]: Alex joined the game']} connection="live" prefs={defaults} initialTab="chat"/>);
 expect(html).not.toContain('Alex joined'); expect(html).toContain('No matching');
});
test('player directory exposes profile interaction',()=>{
 const html=renderToStaticMarkup(<PlayersPage profiles={[{name:'Alex',operator:false,playTime:'Unknown',deaths:'Unknown',walked:'Unknown',mobKills:'Unknown'}]} players={[]} prefs={defaults}/>);
 expect(html).toContain('View Alex'); expect(html).toContain('Search players');
});
