import { describe, expect, it } from 'vitest';
import { stateLabel, worldCaption, worldHeadline } from './format';
describe('live server labels', () => {
  it('never claims an unknown server is online', () => {
    expect(worldHeadline('unknown')).toBe('WORLD STATUS UNKNOWN');
    expect(worldCaption('unknown')).toContain('Waiting');
    expect(stateLabel('unknown')).toBe('UNKNOWN');
  });
});
