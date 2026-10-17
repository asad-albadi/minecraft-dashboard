export type ServerState = 'online' | 'offline' | 'starting' | 'stopping' | 'unknown';
export const stateLabel = (state: ServerState) => state.toUpperCase();
export const worldHeadline = (state: ServerState) => state === 'online' ? 'WORLD ONLINE' : state === 'offline' ? 'WORLD OFFLINE' : 'WORLD STATUS UNKNOWN';
export const worldCaption = (state: ServerState) => state === 'online' ? 'Adventure continues...' : state === 'offline' ? 'The server is not reachable.' : 'Waiting for server data...';
