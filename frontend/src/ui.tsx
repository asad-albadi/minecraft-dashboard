import type React from 'react';
import heroValley from './assets/hero-valley-v1.png';
import sidebarCottage from './assets/sidebar-cottage-v1.png';
import overviewIcon from './assets/dashboard-icons/overview.png';
import playersIcon from './assets/dashboard-icons/players.png';
import worldIcon from './assets/dashboard-icons/world.png';
import consoleIcon from './assets/dashboard-icons/console.png';
import backupsIcon from './assets/dashboard-icons/backups.png';
import settingsIcon from './assets/dashboard-icons/settings.png';
import playersOnlineIcon from './assets/dashboard-icons/players-online.png';
import tickRateIcon from './assets/dashboard-icons/tick-rate.png';
import memoryUsageIcon from './assets/dashboard-icons/memory-usage.png';
import uptimeIcon from './assets/dashboard-icons/uptime.png';
import worldNameIcon from './assets/dashboard-icons/world-name.png';
import gameModeIcon from './assets/dashboard-icons/game-mode.png';
import difficultyIcon from './assets/dashboard-icons/difficulty.png';
import seedIcon from './assets/dashboard-icons/seed.png';
import dimensionIcon from './assets/dashboard-icons/dimension.png';
import worldSizeIcon from './assets/dashboard-icons/world-size.png';
import lastBackupIcon from './assets/dashboard-icons/last-backup.png';
import worldTimeIcon from './assets/dashboard-icons/world-time.png';
import weatherIcon from './assets/dashboard-icons/weather.png';
import startServerIcon from './assets/dashboard-icons/start-server.png';
import stopServerIcon from './assets/dashboard-icons/stop-server.png';
import restartServerIcon from './assets/dashboard-icons/restart-server.png';
import createBackupIcon from './assets/dashboard-icons/create-backup.png';
import viewBackupsIcon from './assets/dashboard-icons/view-backups.png';
import onlinePlayersIcon from './assets/dashboard-icons/online-players.png';
import serverConsoleIcon from './assets/dashboard-icons/server-console.png';
import liveStatusIcon from './assets/dashboard-icons/live-status.png';
export const icons = { overview: overviewIcon, players: playersIcon, world: worldIcon, console: consoleIcon, backups: backupsIcon, settings: settingsIcon, 'players-online': playersOnlineIcon, 'tick-rate': tickRateIcon, 'memory-usage': memoryUsageIcon, uptime: uptimeIcon, 'world-name': worldNameIcon, 'game-mode': gameModeIcon, difficulty: difficultyIcon, seed: seedIcon, dimension: dimensionIcon, 'world-size': worldSizeIcon, 'last-backup': lastBackupIcon, 'world-time': worldTimeIcon, weather: weatherIcon, 'start-server': startServerIcon, 'stop-server': stopServerIcon, 'restart-server': restartServerIcon, 'create-backup': createBackupIcon, 'view-backups': viewBackupsIcon, 'online-players': onlinePlayersIcon, 'server-console': serverConsoleIcon, 'live-status': liveStatusIcon } as const;
export type IconName = keyof typeof icons;
export const nav: Array<[IconName, string, string]> = [['overview', 'Overview', '/'], ['players', 'Players', '/players'], ['world', 'World', '/world'], ['console', 'Console', '/console'], ['backups', 'Backups', '/backups'], ['settings', 'Settings', '/settings']];
export function DashboardIcon({ name, className = '' }: { name: IconName; className?: string }) { return <img aria-hidden className={`dashboard-icon ${className}`} src={icons[name]} alt="" />; }
export function Panel({ title, icon, children, className = '' }: { title: string; icon: IconName; children: React.ReactNode; className?: string }) { return <section className={`stone-panel ${className}`}><h2><DashboardIcon name={icon} />{title}</h2>{children}</section>; }
export function Bar({ value = 0, tone = '' }: { value?: number; tone?: string }) { return <span className="meter"><i className={tone} style={{ width: `${Math.max(0, Math.min(100, value))}%` }} /></span>; }
export function Button({ children, tone = '', disabled, onClick }: { children: React.ReactNode; tone?: string; disabled?: boolean; onClick?: () => void }) { return <button className={`pixel-button ${tone}`} disabled={disabled} onClick={onClick}>{children}</button>; }
export {heroValley,sidebarCottage};
