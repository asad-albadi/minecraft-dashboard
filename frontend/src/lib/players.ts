export function playerHeadURL(playerName: string): string {
  return `https://mc-heads.net/avatar/${encodeURIComponent(playerName)}/64`;
}
