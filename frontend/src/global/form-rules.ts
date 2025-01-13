import i18n from '@/lang';
import { FormItemRule } from 'element-plus';

const checkNoSpace = (rule: any, value: any, callback: any) => {
    if (value.indexOf(' ') !== -1) {
        return callback(new Error(i18n.global.t('setting.noSpace')));
    }
    callback();
};

const checkIp = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
    } else {
        const reg =
            /^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.ip')));
        } else {
            callback();
        }
    }
};

const checkIpv4 = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback();
    } else {
        const reg =
            /^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.ip')));
        } else {
            callback();
        }
    }
};

const checkIpV6 = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
    } else {
        const IPv4SegmentFormat = '(?:[0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])';
        const IPv4AddressFormat = `(${IPv4SegmentFormat}[.]){3}${IPv4SegmentFormat}`;
        const IPv6SegmentFormat = '(?:[0-9a-fA-F]{1,4})';
        const IPv6AddressRegExp = new RegExp(
            '^(' +
                `(?:${IPv6SegmentFormat}:){7}(?:${IPv6SegmentFormat}|:)|` +
                `(?:${IPv6SegmentFormat}:){6}(?:${IPv4AddressFormat}|:${IPv6SegmentFormat}|:)|` +
                `(?:${IPv6SegmentFormat}:){5}(?::${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,2}|:)|` +
                `(?:${IPv6SegmentFormat}:){4}(?:(:${IPv6SegmentFormat}){0,1}:${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,3}|:)|` +
                `(?:${IPv6SegmentFormat}:){3}(?:(:${IPv6SegmentFormat}){0,2}:${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,4}|:)|` +
                `(?:${IPv6SegmentFormat}:){2}(?:(:${IPv6SegmentFormat}){0,3}:${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,5}|:)|` +
                `(?:${IPv6SegmentFormat}:){1}(?:(:${IPv6SegmentFormat}){0,4}:${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,6}|:)|` +
                `(?::((?::${IPv6SegmentFormat}){0,5}:${IPv4AddressFormat}|(?::${IPv6SegmentFormat}){1,7}|:))` +
                ')(%[0-9a-zA-Z-.:]{1,})?$',
        );
        if (!IPv6AddressRegExp.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.ip')));
        } else {
            callback();
        }
    }
};

const checkIpV4V6OrDomain = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
    } else {
        const IPv4SegmentFormat = '(?:[0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-4][0-9]|25[0-5])';
        const IPv4AddressFormat = `(${IPv4SegmentFormat}[.]){3}${IPv4SegmentFormat}`;
        const IPv4AddressRegExp = new RegExp(`^${IPv4AddressFormat}$`);
        const IPv6SegmentFormat = '(?:[0-9a-fA-F]{1,4})';
        const IPv6AddressRegExp = new RegExp(
            '^(' +
                `(?:${IPv6SegmentFormat}:){7}(?:${IPv6SegmentFormat}|:)|` +
                `(?:${IPv6SegmentFormat}:){6}(?:${IPv4AddressFormat}|:${IPv6SegmentFormat}|:)|` +
                `(?:${IPv6SegmentFormat}:){5}(?::${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,2}|:)|` +
                `(?:${IPv6SegmentFormat}:){4}(?:(:${IPv6SegmentFormat}){0,1}:${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,3}|:)|` +
                `(?:${IPv6SegmentFormat}:){3}(?:(:${IPv6SegmentFormat}){0,2}:${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,4}|:)|` +
                `(?:${IPv6SegmentFormat}:){2}(?:(:${IPv6SegmentFormat}){0,3}:${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,5}|:)|` +
                `(?:${IPv6SegmentFormat}:){1}(?:(:${IPv6SegmentFormat}){0,4}:${IPv4AddressFormat}|(:${IPv6SegmentFormat}){1,6}|:)|` +
                `(?::((?::${IPv6SegmentFormat}){0,5}:${IPv4AddressFormat}|(?::${IPv6SegmentFormat}){1,7}|:))` +
                ')(%[0-9a-zA-Z-.:]{1,})?$',
        );
        const regHost = /^(?=^.{3,255}$)[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$/;
        if (!regHost.test(value) && !IPv4AddressRegExp.test(value) && !IPv6AddressRegExp.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.ip')));
        } else {
            callback();
        }
    }
};

const checkHost = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
    } else {
        const regIP =
            /^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])$/;
        const regHost = /^(?=^.{3,255}$)[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$/;
        if (!regIP.test(value) && !regHost.test(value)) {
            callback(new Error(i18n.global.t('commons.rule.host')));
        } else {
            callback();
        }
    }
};

const checkIllegal = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.requiredInput')));
        return;
    }
    if (
        value.indexOf('&') !== -1 ||
        value.indexOf('|') !== -1 ||
        value.indexOf(';') !== -1 ||
        value.indexOf('$') !== -1 ||
        value.indexOf("'") !== -1 ||
        value.indexOf('`') !== -1 ||
        value.indexOf('(') !== -1 ||
        value.indexOf(')') !== -1 ||
        value.indexOf("'") !== -1
    ) {
        callback(new Error(i18n.global.t('commons.rule.illegalInput')));
    } else {
        callback();
    }
};

const complexityPassword = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.complexityPassword')));
    } else {
        const reg = /^(?![\d]+$)(?![a-zA-Z]+$)(?![^\da-zA-Z]+$).{8,30}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.complexityPassword')));
        } else {
            callback();
        }
    }
};

const checkName = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.commonName')));
    } else {
        const reg = /^[a-zA-Z0-9\u4e00-\u9fa5]{1}[a-zA-Z0-9_.\u4e00-\u9fa5-]{0,127}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.commonName')));
        } else {
            callback();
        }
    }
};

const checkUserName = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.userName')));
    } else {
        const reg = /^[a-zA-Z0-9_\u4e00-\u9fa5]{3,30}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.userName')));
        } else {
            callback();
        }
    }
};

const checkSimpleName = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.simpleName')));
    } else {
        const reg = /^[a-zA-Z0-9]{1}[a-zA-Z0-9_]{2,29}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.simpleName')));
        } else {
            callback();
        }
    }
};

const checkSimplePassword = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.simplePassword')));
    } else {
        const reg = /^[a-zA-Z0-9]{1}[a-zA-Z0-9_]{0,29}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.simplePassword')));
        } else {
            callback();
        }
    }
};

const checkDBName = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.dbName')));
    } else {
        const reg = /^[a-zA-Z0-9\u4e00-\u9fa5]{1}[a-zA-Z0-9_.\u4e00-\u9fa5-]{0,63}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.dbName')));
        } else {
            callback();
        }
    }
};

const checkImageName = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.imageName')));
    } else {
        const reg = /^[a-zA-Z0-9]{1}[a-z:@A-Z0-9_/.-]{0,256}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.imageName')));
        } else {
            callback();
        }
    }
};

const checkComposeName = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.composeName')));
    } else {
        const reg = /^[a-z0-9]{1}[a-z0-9_-]{0,256}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.composeName')));
        } else {
            callback();
        }
    }
};

const checkVolumeName = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.volumeName')));
    } else {
        const reg = /^[a-zA-Z0-9]{1}[a-zA-Z0-9_.-]{1,30}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.volumeName')));
        } else {
            callback();
        }
    }
};

const checkLinuxName = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.linuxName', ['/\\:*?\'"<>|'])));
    } else {
        const reg = /^[^/\\\"'|<>?*]{1,128}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.linuxName', ['/\\:*?\'"<>|'])));
        } else {
            callback();
        }
    }
};

const checkSupervisorName = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.supervisorName')));
    } else {
        const reg = /^[a-zA-Z0-9]{1}[a-zA-Z0-9_-]{0,127}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.supervisorName')));
        } else {
            callback();
        }
    }
};

const checkDatabaseName = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.databaseName')));
    } else {
        const reg = /^[a-zA-Z0-9]{1}[a-zA-Z0-9_]{0,30}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.databaseName')));
        } else {
            callback();
        }
    }
};

const checkAppName = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.appName')));
    } else {
        const reg = /^(?![_-])[a-z0-9_-]{1,29}[a-zA-Z0-9]$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.appName')));
        } else {
            callback();
        }
    }
};

const checkDomain = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.domain')));
    } else {
        const reg =
            /^([\w\u4e00-\u9fa5\-\*]{1,100}\.){1,10}([\w\u4e00-\u9fa5\-]{1,24}|[\w\u4e00-\u9fa5\-]{1,24}\.[\w\u4e00-\u9fa5\-]{1,24})$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.domain')));
        } else {
            callback();
        }
    }
};

const checkDomainWithPort = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.domain')));
    } else {
        const reg =
            /^([\w\u4e00-\u9fa5\-\*]{1,100}\.){1,10}([\w\u4e00-\u9fa5\-]{1,24}|[\w\u4e00-\u9fa5\-]{1,24}\.[\w\u4e00-\u9fa5\-]{1,24})(:\d{1,5})?$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.domain')));
        } else {
            callback();
        }
    }
};

const checkIntegerNumber = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.integer')));
    } else {
        const reg = /^[1-9]\d*$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.integer')));
        } else {
            callback();
        }
    }
};

const checkIntegerNumberWith0 = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.integer')));
    } else {
        const reg = /^[0-9]*$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.integer')));
        } else {
            callback();
        }
    }
};

const checkFloatNumber = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.integer')));
    } else {
        const reg = /^\d+(\.\d+)?$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.number')));
        } else {
            callback();
        }
    }
};

const checkParamCommon = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.paramName')));
    } else {
        const reg = /^[a-zA-Z0-9]{1}[a-zA-Z0-9._-]{1,63}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.paramName')));
        } else {
            callback();
        }
    }
};

const checkParamComplexity = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.paramComplexity', ['.%@!~_-'])));
    } else {
        const reg = /^[a-zA-Z0-9]{1}[a-zA-Z0-9.%@!~_-]{4,126}[a-zA-Z0-9]{1}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.paramComplexity', ['.%@!~_-'])));
        } else {
            callback();
        }
    }
};

const checkParamUrlAndPort = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.paramUrlAndPort')));
    } else {
        const reg =
            /^(https?:\/\/)?((localhost)|([a-zA-Z0-9_-]+\.)*[a-zA-Z0-9_-]+)(:[1-9]\d{0,3}|:[1-5]\d{4}|:6[0-4]\d{3}|:65[0-4]\d{2}|:655[0-2]\d|:6553[0-5])?(\/\S*)?$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.paramUrlAndPort')));
        } else {
            callback();
        }
    }
};

const checkPort = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.port')));
    } else {
        const reg = /^([1-9](\d{0,3}))$|^([1-5]\d{4})$|^(6[0-4]\d{3})$|^(65[0-4]\d{2})$|^(655[0-2]\d)$|^(6553[0-5])$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.port')));
        } else {
            callback();
        }
    }
};

const checkDoc = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.nginxDoc')));
    } else {
        const reg = /^[A-Za-z0-9\n./]+$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.nginxDoc')));
        } else {
            callback();
        }
    }
};

export function checkNumberRange(min: number, max: number): FormItemRule {
    return {
        required: false,
        trigger: 'blur',
        min: min,
        max: max,
        type: 'number',
        message: i18n.global.t('commons.rule.numberRange', [min, max]),
    };
}

export function checkFloatNumberRange(min: number, max: number): FormItemRule {
    let validatorFunc = function (rule: any, value: any, callback: any) {
        if (value === '' || typeof value === 'undefined' || value == null) {
            callback(new Error(i18n.global.t('commons.rule.disableFunction')));
        } else {
            if ((Number(value) < min || Number(value) > max) && value !== '') {
                callback(new Error(i18n.global.t('commons.rule.disableFunction')));
            } else {
                callback();
            }
        }
    };
    return {
        required: false,
        trigger: 'blur',
        validator: validatorFunc,
        message: i18n.global.t('commons.rule.numberRange', [min, max]),
    };
}

const checkContainerName = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback();
    } else {
        const reg = /^[a-zA-Z0-9]{1}[a-zA-Z0-9_.-]{1,127}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.containerName')));
        } else {
            callback();
        }
    }
};

const checkDisableFunctions = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.disableFunction')));
    } else {
        const reg = /^[a-zA-Z_,]+$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.disableFunction')));
        } else {
            callback();
        }
    }
};

const checkLeechExts = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.leechExts')));
    } else {
        const reg = /^[a-zA-Z0-9,]+$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.leechExts')));
        } else {
            callback();
        }
    }
};

const checkParamSimple = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback();
    } else {
        const reg = /^[a-z0-9][a-z0-9]{1,128}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.paramSimple')));
        } else {
            callback();
        }
    }
};

const checkFilePermission = (rule, value, callback) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.filePermission')));
    } else {
        const regFilePermission = /^[0-7]{3,4}$/;
        if (!regFilePermission.test(value)) {
            callback(new Error(i18n.global.t('commons.rule.filePermission')));
        } else {
            callback();
        }
    }
};

const checkPHPExtensions = (rule, value, callback) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.phpExtension')));
    } else {
        const reg = /^[a-z0-9,_]+$/;
        if (!reg.test(value)) {
            callback(new Error(i18n.global.t('commons.rule.phpExtension')));
        } else {
            callback();
        }
    }
};

const checkHttpOrHttps = (rule, value, callback) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.paramHttp')));
    } else {
        const regHttpHttps = /^(http|https):\/\//;
        if (!regHttpHttps.test(value)) {
            callback(new Error(i18n.global.t('commons.rule.paramHttp')));
        } else {
            callback();
        }
    }
};

const checkPhone = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback();
    } else {
        const reg = /^(?:(?:\+|00)86)?1[3-9]\d{9}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.phone')));
        } else {
            callback();
        }
    }
};

interface CommonRule {
    requiredInput: FormItemRule;
    requiredSelect: FormItemRule;
    requiredSelectBusiness: FormItemRule;
    noSpace: FormItemRule;
    name: FormItemRule;
    userName: FormItemRule;
    simpleName: FormItemRule;
    simplePassword: FormItemRule;
    dbName: FormItemRule;
    imageName: FormItemRule;
    composeName: FormItemRule;
    volumeName: FormItemRule;
    linuxName: FormItemRule;
    password: FormItemRule;
    email: FormItemRule;
    number: FormItemRule;
    integerNumber: FormItemRule;
    integerNumberWith0: FormItemRule;
    floatNumber: FormItemRule;
    ip: FormItemRule;
    ipV6: FormItemRule;
    ipv4: FormItemRule;
    ipV4V6OrDomain: FormItemRule;
    host: FormItemRule;
    illegal: FormItemRule;
    port: FormItemRule;
    domain: FormItemRule;
    databaseName: FormItemRule;
    nginxDoc: FormItemRule;
    appName: FormItemRule;
    containerName: FormItemRule;
    disabledFunctions: FormItemRule;
    leechExts: FormItemRule;
    domainWithPort: FormItemRule;
    filePermission: FormItemRule;
    phpExtensions: FormItemRule;
    supervisorName: FormItemRule;

    paramCommon: FormItemRule;
    paramComplexity: FormItemRule;
    paramPort: FormItemRule;
    paramExtUrl: FormItemRule;
    paramSimple: FormItemRule;
    paramHttp: FormItemRule;
    phone: FormItemRule;
}

export const Rules: CommonRule = {
    requiredInput: {
        required: true,
        message: i18n.global.t('commons.rule.requiredInput'),
        trigger: 'blur',
    },
    requiredSelect: {
        required: true,
        message: i18n.global.t('commons.rule.requiredSelect'),
        trigger: 'change',
    },
    requiredSelectBusiness: {
        required: true,
        min: 1,
        max: 65535,
        type: 'number',
        message: i18n.global.t('commons.rule.requiredSelect'),
        trigger: 'change',
    },
    noSpace: {
        required: true,
        validator: checkNoSpace,
        trigger: 'blur',
    },
    simpleName: {
        required: true,
        validator: checkSimpleName,
        trigger: 'blur',
    },
    simplePassword: {
        required: true,
        validator: checkSimplePassword,
        trigger: 'blur',
    },
    dbName: {
        required: true,
        validator: checkDBName,
        trigger: 'blur',
    },
    imageName: {
        required: true,
        validator: checkImageName,
        trigger: 'blur',
    },
    composeName: {
        required: true,
        validator: checkComposeName,
        trigger: 'blur',
    },
    volumeName: {
        required: true,
        validator: checkVolumeName,
        trigger: 'blur',
    },
    name: {
        required: true,
        validator: checkName,
        trigger: 'blur',
    },
    userName: {
        required: true,
        validator: checkUserName,
        trigger: 'blur',
    },
    linuxName: {
        required: true,
        validator: checkLinuxName,
        trigger: 'blur',
    },
    databaseName: {
        required: true,
        validator: checkDatabaseName,
        trigger: 'blur',
    },
    password: {
        validator: complexityPassword,
        trigger: 'blur',
    },
    email: {
        required: true,
        type: 'email',
        message: i18n.global.t('commons.rule.email'),
        trigger: 'blur',
    },
    number: {
        required: true,
        trigger: 'blur',
        min: 0,
        type: 'number',
        message: i18n.global.t('commons.rule.number'),
    },
    integerNumber: {
        required: true,
        validator: checkIntegerNumber,
        trigger: 'blur',
    },
    integerNumberWith0: {
        required: true,
        validator: checkIntegerNumberWith0,
        trigger: 'blur',
    },
    floatNumber: {
        required: true,
        validator: checkFloatNumber,
        trigger: 'blur',
        min: 0,
        message: i18n.global.t('commons.rule.number'),
    },
    ip: {
        validator: checkIp,
        required: true,
        trigger: 'blur',
    },
    ipV6: {
        validator: checkIpV6,
        required: true,
        trigger: 'blur',
    },
    ipV4V6OrDomain: {
        validator: checkIpV4V6OrDomain,
        required: true,
        trigger: 'blur',
    },
    host: {
        validator: checkHost,
        required: true,
        trigger: 'blur',
    },
    illegal: {
        validator: checkIllegal,
        required: true,
        trigger: 'blur',
    },
    port: {
        required: true,
        trigger: 'blur',
        min: 1,
        max: 65535,
        type: 'number',
        message: i18n.global.t('commons.rule.port'),
    },
    domain: {
        required: true,
        validator: checkDomain,
        trigger: 'blur',
    },
    paramCommon: {
        required: true,
        validator: checkParamCommon,
        trigger: 'blur',
    },
    paramComplexity: {
        required: true,
        validator: checkParamComplexity,
        trigger: 'blur',
    },
    paramPort: {
        required: true,
        trigger: 'blur',
        validator: checkPort,
    },
    paramExtUrl: {
        required: true,
        validator: checkParamUrlAndPort,
        trigger: 'blur',
    },
    nginxDoc: {
        required: true,
        validator: checkDoc,
        trigger: 'blur',
    },
    appName: {
        required: true,
        trigger: 'blur',
        validator: checkAppName,
    },
    containerName: {
        required: false,
        trigger: 'blur',
        validator: checkContainerName,
    },
    disabledFunctions: {
        required: true,
        trigger: 'blur',
        validator: checkDisableFunctions,
    },
    leechExts: {
        required: true,
        trigger: 'blur',
        validator: checkLeechExts,
    },
    supervisorName: {
        required: true,
        trigger: 'blur',
        validator: checkSupervisorName,
    },
    paramSimple: {
        required: true,
        trigger: 'blur',
        validator: checkParamSimple,
    },
    domainWithPort: {
        required: true,
        validator: checkDomainWithPort,
        trigger: 'blur',
    },
    filePermission: {
        required: true,
        validator: checkFilePermission,
        trigger: 'blur',
    },
    phpExtensions: {
        required: true,
        validator: checkPHPExtensions,
        trigger: 'blur',
    },
    paramHttp: {
        required: true,
        validator: checkHttpOrHttps,
        trigger: 'blur',
    },
    ipv4: {
        validator: checkIpv4,
        trigger: 'blur',
    },
    phone: {
        validator: checkPhone,
        trigger: 'blur',
    },
};
