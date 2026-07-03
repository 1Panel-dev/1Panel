import i18n from '@/lang';

interface RunningTime {
    days: number;
    hours: number;
    minutes: number;
    seconds: number;
}

const formatDurationUnit = (value: number, unit: string) => `${value}${i18n.global.t(unit, value)}`;

export function formatUptime(runningTime?: RunningTime) {
    if (!runningTime) {
        return '';
    }
    const { days, hours, minutes, seconds } = runningTime;
    if (days <= 0 && hours <= 0 && minutes <= 0 && seconds <= 0) {
        return '';
    }
    if (days !== 0) {
        return (
            formatDurationUnit(days, 'commons.units.day') +
            ' ' +
            formatDurationUnit(hours, 'commons.units.hour') +
            ' ' +
            formatDurationUnit(minutes, 'commons.units.minute') +
            ' ' +
            formatDurationUnit(seconds, 'commons.units.second')
        );
    }
    if (hours !== 0) {
        return (
            formatDurationUnit(hours, 'commons.units.hour') +
            ' ' +
            formatDurationUnit(minutes, 'commons.units.minute') +
            ' ' +
            formatDurationUnit(seconds, 'commons.units.second')
        );
    }
    if (minutes !== 0) {
        return (
            formatDurationUnit(minutes, 'commons.units.minute') +
            ' ' +
            formatDurationUnit(seconds, 'commons.units.second')
        );
    }
    return formatDurationUnit(seconds, 'commons.units.second');
}

export function getCurrentDateFormatted() {
    const now = new Date();
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, '0');
    const day = String(now.getDate()).padStart(2, '0');
    const hours = String(now.getHours()).padStart(2, '0');
    const minutes = String(now.getMinutes()).padStart(2, '0');
    const seconds = String(now.getSeconds()).padStart(2, '0');
    return `${year}${month}${day}${hours}${minutes}${seconds}`;
}

export function dateFormat(row: any, col: any, dataStr: any) {
    const date = new Date(dataStr);
    const y = date.getFullYear();
    let m: string | number = date.getMonth() + 1;
    m = m < 10 ? `0${String(m)}` : m;
    let d: string | number = date.getDate();
    d = d < 10 ? `0${String(d)}` : d;
    let h: string | number = date.getHours();
    h = h < 10 ? `0${String(h)}` : h;
    let minute: string | number = date.getMinutes();
    minute = minute < 10 ? `0${String(minute)}` : minute;
    let second: string | number = date.getSeconds();
    second = second < 10 ? `0${String(second)}` : second;
    return `${String(y)}-${String(m)}-${String(d)}   ${String(h)}:${String(minute)}:${String(second)}`;
}

export function dateFormatSimple(dataStr: any) {
    const date = new Date(dataStr);
    const y = date.getFullYear();
    let m: string | number = date.getMonth() + 1;
    m = m < 10 ? `0${String(m)}` : m;
    let d: string | number = date.getDate();
    d = d < 10 ? `0${String(d)}` : d;
    return `${String(y)}-${String(m)}-${String(d)}`;
}

export function dateFormatSimpleWithSecond(dataStr: any) {
    const date = new Date(dataStr);
    const y = date.getFullYear();
    let m: string | number = date.getMonth() + 1;
    m = m < 10 ? `0${String(m)}` : m;
    let d: string | number = date.getDate();
    d = d < 10 ? `0${String(d)}` : d;
    let h: string | number = date.getHours();
    h = h < 10 ? `0${String(h)}` : h;
    let minute: string | number = date.getMinutes();
    minute = minute < 10 ? `0${String(minute)}` : minute;
    let second: string | number = date.getSeconds();
    second = second < 10 ? `0${String(second)}` : second;
    return `${String(y)}-${String(m)}-${String(d)} ${String(h)}:${String(minute)}:${String(second)}`;
}

export function dateFormatRFC3339(dataStr: any) {
    const date = new Date(dataStr);
    const y = date.getFullYear();
    const m = String(date.getMonth() + 1).padStart(2, '0');
    const d = String(date.getDate()).padStart(2, '0');
    const h = String(date.getHours()).padStart(2, '0');
    const minute = String(date.getMinutes()).padStart(2, '0');
    const second = String(date.getSeconds()).padStart(2, '0');
    const offset = -date.getTimezoneOffset();
    const sign = offset >= 0 ? '+' : '-';
    const absOffset = Math.abs(offset);
    const offsetHour = String(Math.floor(absOffset / 60)).padStart(2, '0');
    const offsetMinute = String(absOffset % 60).padStart(2, '0');
    return `${y}-${m}-${d}T${h}:${minute}:${second}${sign}${offsetHour}:${offsetMinute}`;
}

export function dateFormatForName(dataStr: any) {
    const date = new Date(dataStr);
    const y = date.getFullYear();
    let m: string | number = date.getMonth() + 1;
    m = m < 10 ? `0${String(m)}` : m;
    let d: string | number = date.getDate();
    d = d < 10 ? `0${String(d)}` : d;
    let h: string | number = date.getHours();
    h = h < 10 ? `0${String(h)}` : h;
    let minute: string | number = date.getMinutes();
    minute = minute < 10 ? `0${String(minute)}` : minute;
    let second: string | number = date.getSeconds();
    second = second < 10 ? `0${String(second)}` : second;
    return `${String(y)}${String(m)}${String(d)}${String(h)}${String(minute)}${String(second)}`;
}

export function dateFormatWithoutYear(dataStr: any) {
    const date = new Date(dataStr);
    let m: string | number = date.getMonth() + 1;
    m = m < 10 ? `0${String(m)}` : m;
    let d: string | number = date.getDate();
    d = d < 10 ? `0${String(d)}` : d;
    let h: string | number = date.getHours();
    h = h < 10 ? `0${String(h)}` : h;
    let minute: string | number = date.getMinutes();
    minute = minute < 10 ? `0${String(minute)}` : minute;
    let s: string | number = date.getSeconds();
    s = s < 10 ? `0${String(s)}` : s;
    return `${String(m)}-${String(d)}\n${String(h)}:${String(minute)}:${String(s)}`;
}

export function dateFormatForSecond(dataStr: any) {
    const date = new Date(dataStr);
    let h: string | number = date.getHours();
    h = h < 10 ? `0${String(h)}` : h;
    let minute: string | number = date.getMinutes();
    minute = minute < 10 ? `0${String(minute)}` : minute;
    let second: string | number = date.getSeconds();
    second = second < 10 ? `0${String(second)}` : second;
    return `${String(h)}:${String(minute)}:${String(second)}`;
}

export function loadZero(i: number) {
    return i < 10 ? '0' + i : '' + i;
}

export function getAge(d1: string): string {
    const dateBegin = new Date(d1);
    const dateEnd = new Date();
    const dateDiff = dateEnd.getTime() - dateBegin.getTime();
    const dayDiff = Math.floor(dateDiff / (24 * 3600 * 1000));
    const leave1 = dateDiff % (24 * 3600 * 1000);
    const hours = Math.floor(leave1 / (3600 * 1000));
    const leave2 = leave1 % (3600 * 1000);
    const minutes = Math.floor(leave2 / (60 * 1000));

    let res = '';
    if (dayDiff > 0) {
        res += String(dayDiff) + ' ' + i18n.global.t('commons.units.day', dayDiff) + ' ';
        if (hours <= 0) {
            return res;
        }
    }
    if (hours > 0) {
        res += String(hours) + ' ' + i18n.global.t('commons.units.hour', hours);
        return res;
    }
    if (minutes > 0) {
        res += String(minutes) + ' ' + i18n.global.t('commons.units.minute', minutes);
        return res;
    }
    return i18n.global.t('app.less1Minute');
}

export function getDateStr() {
    let now: Date = new Date();
    let year: number = now.getFullYear();
    let month: number = now.getMonth() + 1;
    let date: number = now.getDate();
    let hours: number = now.getHours();
    let minutes: number = now.getMinutes();
    let seconds: number = now.getSeconds();
    return `${year}-${month < 10 ? '0' + month : month}-${date < 10 ? '0' + date : date}-${
        hours < 10 ? '0' + hours : hours
    }-${minutes < 10 ? '0' + minutes : minutes}-${seconds < 10 ? '0' + seconds : seconds}`;
}
