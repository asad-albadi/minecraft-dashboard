import { expect, test } from 'vitest';
import { playerHeadURL } from './players';

test('builds a safely encoded Minecraft head URL from a player name', () => {
  expect(playerHeadURL('Alex Builder')).toBe('https://mc-heads.net/avatar/Alex%20Builder/64');
});
