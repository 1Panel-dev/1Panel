import i18n from '@/lang';
import { FormItemRule } from 'element-plus';

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

const complexityPassword = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.complexityPassword')));
    } else {
        const reg = /^(?=.*\d)(?=.*[a-zA-Z])(?=.*[~!@#$%^&*.])[\da-zA-Z~!@#$%^&*.]{8,}$/;
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
        const reg = /^[a-zA-Z0-9\u4e00-\u9fa5]{1}[a-zA-Z0-9_.\u4e00-\u9fa5-]{0,30}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.commonName')));
        } else {
            callback();
        }
    }
};

const checkSimpleName = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.simpleName')));
    } else {
        const reg = /^[a-zA-Z0-9]{1}[a-zA-Z0-9_]{0,30}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.simpleName')));
        } else {
            callback();
        }
    }
};

const checkDBName = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.dbName')));
    } else {
        const reg = /^[a-zA-Z0-9]{1}[a-zA-Z0-9_.-]{0,30}$/;
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
        const reg = /^[a-zA-Z0-9\u4e00-\u9fa5]{1}[a-z:A-Z0-9_.\u4e00-\u9fa5-]{0,30}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.imageName')));
        } else {
            callback();
        }
    }
};

const checkLinuxName = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.linuxName', ['/\\:*?"<>|'])));
    } else {
        const reg = /^((?!\\|\/|:|\*|\?|<|>|\||'|%).){1,30}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.linuxName', ['/\\:*?"<>|'])));
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

interface CommonRule {
    requiredInput: FormItemRule;
    requiredSelect: FormItemRule;
    requiredSelectBusiness: FormItemRule;
    name: FormItemRule;
    simpleName: FormItemRule;
    dbName: FormItemRule;
    imageName: FormItemRule;
    linuxName: FormItemRule;
    password: FormItemRule;
    email: FormItemRule;
    number: FormItemRule;
    ip: FormItemRule;
    port: FormItemRule;
    domain: FormItemRule;
    databaseName: FormItemRule;
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
    simpleName: {
        required: true,
        validator: checkSimpleName,
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
    name: {
        required: true,
        validator: checkName,
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
    ip: {
        validator: checkIp,
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
};
