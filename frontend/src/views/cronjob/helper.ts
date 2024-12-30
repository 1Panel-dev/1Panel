import { Cronjob } from '@/api/interface/cronjob';
import i18n from '@/lang';
import { loadZero } from '@/utils/util';

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
    item.week = 0;
    item.day = 0;
    item.hour = 0;
    item.minute = 0;
    item.second = 0;
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
        case 'cutWebsiteLog':
        case 'ntp':
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
    switch (specObj.specType) {
        case 'perMonth':
            return i18n.global.t('cronjob.perMonthHelper', [specObj.day, specObj.hour, loadZero(specObj.minute)]);
        case 'perWeek':
            return i18n.global.t('cronjob.perWeekHelper', [
                loadWeek(specObj.week),
                specObj.hour,
                loadZero(specObj.minute),
            ]);
        case 'perDay':
            return i18n.global.t('cronjob.perDayHelper', [specObj.hour, loadZero(specObj.minute)]);
        case 'perHour':
            return i18n.global.t('cronjob.perHourHelper', [loadZero(specObj.minute)]);
        case 'perNDay':
            return i18n.global.t('cronjob.perNDayHelper', [specObj.day, specObj.hour, loadZero(specObj.minute)]);
        case 'perNHour':
            return i18n.global.t('cronjob.perNHourHelper', [specObj.hour, loadZero(specObj.minute)]);
        case 'perNMinute':
            return i18n.global.t('cronjob.perNMinuteHelper', [loadZero(specObj.minute)]);
        case 'perNSecond':
            return i18n.global.t('cronjob.perNSecondHelper', [loadZero(specObj.second)]);
    }
}
