import { defineStore } from 'pinia';
import piniaPersistConfig from '@/config/pinia-persist';
import { TerminalState } from '../interface';

export const TerminalStore = defineStore({
    id: 'TerminalState',
    state: (): TerminalState => ({
        lineHeight: 1.2,
        letterSpacing: 1.2,
        fontSize: 12,
        fontFamily: '',
        cursorBlink: 'enable',
        cursorStyle: 'underline',
        scrollback: 1000,
        scrollSensitivity: 10,
    }),
    actions: {
        setLineHeight(lineHeight: number) {
            this.lineHeight = lineHeight;
        },
        setLetterSpacing(letterSpacing: number) {
            this.letterSpacing = letterSpacing;
        },
        setFontSize(fontSize: number) {
            this.fontSize = fontSize;
        },
        setFontFamily(fontFamily: string) {
            this.fontFamily = fontFamily;
        },
        setCursorBlink(cursorBlink: string) {
            this.cursorBlink = cursorBlink;
        },
        setCursorStyle(cursorStyle: string) {
            this.cursorStyle = cursorStyle;
        },
        setScrollback(scrollback: number) {
            this.scrollback = scrollback;
        },
        setScrollSensitivity(scrollSensitivity: number) {
            this.scrollSensitivity = scrollSensitivity;
        },
    },
    persist: piniaPersistConfig('TerminalState'),
});

export default TerminalStore;
