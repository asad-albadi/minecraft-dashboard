import { describe, expect, test } from 'vitest';
import { filterLogs, parseChat, selectPlayers, preferences } from './dashboard';

describe('dashboard data views', () => {
  test('chat excludes lifecycle and system messages without losing message text', () => {
    expect(parseChat('[12:30:12] [Server thread/INFO]: <Alex> hello <world>')).toEqual({time:'12:30:12',name:'Alex',message:'hello <world>'});
    expect(parseChat('[12:30:12] [Server thread/INFO]: Alex joined the game')).toBeNull();
  });
  test('severity matches logger prefix not chat content', () => {
    const lines=['[12:30:12] [Server thread/INFO]: <Alex> WARN', '[12:30:13] [Server thread/WARN]: Slow'];
    expect(filterLogs(lines,'slow','WARN')).toEqual([lines[1]]);
    expect(filterLogs(lines,'','WARN')).toEqual([lines[1]]);
  });
  test('player filters combine search and presence; unknown values sort last', () => {
    const rows=[{name:'Alex',playTime:'Unknown',operator:false},{name:'Steve',playTime:'2.0 h',operator:true}];
    expect(selectPlayers(rows,new Set(['alex']),'a','online','name').map(p=>p.name)).toEqual(['Alex']);
    expect(selectPlayers(rows,new Set(),'','all','playTime').map(p=>p.name)).toEqual(['Steve','Alex']);
  });
  test('preferences reject invalid storage and unsafe refresh intervals', () => {
    expect(preferences('{"refresh":1,"timezone":"invalid","landing":"/evil"}')).toMatchObject({refresh:30,timezone:'UTC',landing:'/'});
    expect(preferences('not json').theme).toBe('stone');
  });
  test('unavailable presence never classifies players offline',()=>{
    expect(selectPlayers([{name:'Alex'}],null,'','offline','name')).toEqual([]);
  });
});
