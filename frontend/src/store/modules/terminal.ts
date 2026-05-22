import { defineStore } from 'pinia';
import type { StoreDefinition } from 'pinia';
import piniaPersistConfig from '@/config/pinia-persist';
import { TerminalState } from '../interface';

export const TerminalStore = defineStore('TerminalState', {
    state: (): TerminalState => ({
        lineHeight: 1.2,
        letterSpacing: 1.2,
        fontSize: 12,
        fontFamily: "Monaco, Menlo, Consolas, 'Courier New', monospace",
        backgroundColor: '#000000',
        foregroundColor: '#f5f5f5',
        cursorBlink: 'enable',
        cursorStyle: 'underline',
        scrollback: 1000,
        scrollSensitivity: 6,
    }),
    persist: piniaPersistConfig('TerminalState'),
}) as StoreDefinition<'TerminalState', TerminalState, any, any>;

export default TerminalStore;
