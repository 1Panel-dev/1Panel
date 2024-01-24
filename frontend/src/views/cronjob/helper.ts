import { Cronjob } from '@/api/interface/cronjob';
import i18n from '@/lang';
import { loadZero } from '@/utils/util';

export const shortcuts = [
    {
        text: i18n.global.t('monitor.today'),
        value: () => {
            const end = new Date(new Date().setHours(23, 59, 59, 999));
            const start = new Date(new Date().setHours(0, 0, 0, 0));
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.yesterday'),
        value: () => {
            const itemDate = new Date(new Date().getTime() - 3600 * 1000 * 24 * 1);
            const end = new Date(itemDate.setHours(23, 59, 59, 999));
            const start = new Date(itemDate.setHours(0, 0, 0, 0));
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.lastNDay', [3]),
        value: () => {
            const itemDate = new Date(new Date().getTime() - 3600 * 1000 * 24 * 3);
            const end = new Date(new Date().setHours(23, 59, 59, 999));
            const start = new Date(itemDate.setHours(0, 0, 0, 0));
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.lastNDay', [7]),
        value: () => {
            const itemDate = new Date(new Date().getTime() - 3600 * 1000 * 24 * 7);
            const end = new Date(new Date().setHours(23, 59, 59, 999));
            const start = new Date(itemDate.setHours(0, 0, 0, 0));
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.lastNDay', [30]),
        value: () => {
            const itemDate = new Date(new Date().getTime() - 3600 * 1000 * 24 * 30);
            const end = new Date(new Date().setHours(23, 59, 59, 999));
            const start = new Date(itemDate.setHours(0, 0, 0, 0));
            return [start, end];
        },
    },
];
export const specOptions = [
    { label: i18n.global.t('cronjob.perMonth'), value: 'perMonth' },
    { label: i18n.global.t('cronjob.perWeek'), value: 'perWeek' },
    { label: i18n.global.t('cronjob.perDay'), value: 'perDay' },
    { label: i18n.global.t('cronjob.perHour'), value: 'perHour' },
    { label: i18n.global.t('cronjob.perNDay'), value: 'perNDay' },
    { label: i18n.global.t('cronjob.perNHour'), value: 'perNHour' },
    { label: i18n.global.t('cronjob.perNMinute'), value: 'perNMinute' },
    { label: i18n.global.t('cronjob.perNSecond'), value: 'perNSecond' },
];
export const weekOptions = [
    { label: i18n.global.t('cronjob.monday'), value: 1 },
    { label: i18n.global.t('cronjob.tuesday'), value: 2 },
    { label: i18n.global.t('cronjob.wednesday'), value: 3 },
    { label: i18n.global.t('cronjob.thursday'), value: 4 },
    { label: i18n.global.t('cronjob.friday'), value: 5 },
    { label: i18n.global.t('cronjob.saturday'), value: 6 },
    { label: i18n.global.t('cronjob.sunday'), value: 0 },
];
function loadWeek(i: number) {
    for (const week of weekOptions) {
        if (week.value === i) {
            return week.label;
        }
    }
    return '';
}

export function loadDefaultSpec(type: string) {
    let item = {} as Cronjob.SpecObj;
    switch (type) {
        case 'shell':
            item.specType = 'perWeek';
            item.week = 1;
            item.hour = 1;
            item.minute = 30;
            break;
        case 'app':
            item.specType = 'perDay';
            item.hour = 2;
            item.minute = 30;
            break;
        case 'database':
            item.specType = 'perDay';
            item.hour = 2;
            item.minute = 30;
            break;
        case 'clean':
        case 'website':
            item.specType = 'perWeek';
            item.week = 1;
            item.hour = 1;
            item.minute = 30;
            break;
        case 'log':
        case 'snapshot':
            item.specType = 'perWeek';
            item.week = 1;
            item.hour = 1;
            item.minute = 30;
            break;
        case 'directory':
            item.specType = 'perDay';
            item.hour = 1;
            item.minute = 30;
            break;
        case 'curl':
            item.specType = 'perWeek';
            item.week = 1;
            item.hour = 1;
            item.minute = 30;
            break;
    }
    return item;
}

export function checkScript(specType: string, week, day, hour, minute, second) {
    switch (specType) {
        case 'perMonth':
            return day > 0 && day < 32 && hour >= 0 && hour < 24 && minute >= 0 && minute < 60;
        case 'perWeek':
            return week >= 0 && week < 7 && hour >= 0 && hour < 24 && minute >= 0 && minute < 60;
        case 'perDay':
            return hour >= 0 && hour < 24 && minute >= 0 && minute < 60;
        case 'perHour':
            return minute >= 0 && minute < 60;
        case 'perNDay':
            return day > 0 && day < 366 && hour >= 0 && hour < 24 && minute >= 0 && minute < 60;
        case 'perNHour':
            return hour > 0 && hour < 8784 && minute >= 0 && minute < 60;
        case 'perNMinute':
            return minute > 0 && minute < 527040;
        case 'perNSecond':
            return second > 0 && second < 31622400;
    }
}

export function transObjToSpec(specType: string, week, day, hour, minute, second): string {
    switch (specType) {
        case 'perMonth':
            return `${minute} ${hour} ${day} * *`;
        case 'perWeek':
            return `${minute} ${hour} * * ${week}`;
        case 'perNDay':
            return `${minute} ${hour} */${day} * *`;
        case 'perDay':
            return `${minute} ${hour} * * *`;
        case 'perNHour':
            return `${minute} */${hour} * * *`;
        case 'perHour':
            return `${minute} * * * *`;
        case 'perNMinute':
            return `@every ${minute}m`;
        case 'perNSecond':
            return `@every ${second}s`;
        default:
            return '';
    }
}

export function transSpecToObj(spec: string) {
    let specs = spec.split(' ');
    let specItem = {
        specType: 'perNMinute',
        week: 0,
        day: 0,
        hour: 0,
        minute: 0,
        second: 0,
    };
    if (specs.length === 2) {
        if (specs[1].indexOf('m') !== -1) {
            specItem.specType = 'perNMinute';
            specItem.minute = Number(specs[1].replaceAll('m', ''));
            return specItem;
        } else {
            specItem.specType = 'perNSecond';
            specItem.second = Number(specs[1].replaceAll('s', ''));
            return specItem;
        }
    }
    if (specs.length !== 5 || specs[0] === '*') {
        return null;
    }
    specItem.minute = Number(specs[0]);
    if (specs[1] === '*') {
        specItem.specType = 'perHour';
        return specItem;
    }
    if (specs[1].indexOf('*/') !== -1) {
        specItem.specType = 'perNHour';
        specItem.hour = Number(specs[1].replaceAll('*/', ''));
        return specItem;
    }
    specItem.hour = Number(specs[1]);
    if (specs[2].indexOf('*/') !== -1) {
        specItem.specType = 'perNDay';
        specItem.day = Number(specs[2].replaceAll('*/', ''));
        return specItem;
    }
    if (specs[2] !== '*') {
        specItem.specType = 'perMonth';
        specItem.day = Number(specs[2]);
        return specItem;
    }
    if (specs[4] !== '*') {
        specItem.specType = 'perWeek';
        specItem.week = Number(specs[4]);
        return specItem;
    }
    specItem.specType = 'perDay';
    return specItem;
}

export function transSpecToStr(spec: string): string {
    const specObj = transSpecToObj(spec);
    let str = '';
    if (specObj.specType.indexOf('N') === -1 || specObj.specType === 'perWeek') {
        str += i18n.global.t('cronjob.' + specObj.specType) + ' ';
    } else {
        str += i18n.global.t('cronjob.per') + ' ';
    }
    switch (specObj.specType) {
        case 'perMonth':
            str +=
                specObj.day +
                i18n.global.t('cronjob.day') +
                ' ' +
                loadZero(specObj.hour) +
                ':' +
                loadZero(specObj.minute);
            break;
        case 'perWeek':
            str += loadWeek(specObj.week) + ' ' + loadZero(specObj.hour) + ':' + loadZero(specObj.minute);
            break;
        case 'perDay':
            str += loadZero(specObj.hour) + ':' + loadZero(specObj.minute);
            break;
        case 'perNDay':
            str +=
                specObj.day +
                i18n.global.t('commons.units.day') +
                ', ' +
                loadZero(specObj.hour) +
                ':' +
                loadZero(specObj.minute);
            break;
        case 'perNHour':
            str += specObj.hour + i18n.global.t('commons.units.hour') + ', ' + loadZero(specObj.minute);
            break;
        case 'perHour':
            str += loadZero(specObj.minute);
            break;
        case 'perNMinute':
            str += loadZero(specObj.minute) + i18n.global.t('commons.units.minute');
            break;
        case 'perNSecond':
            str += loadZero(specObj.second) + i18n.global.t('commons.units.second');
            break;
    }

    return str + ' ' + i18n.global.t('cronjob.handle');
}
