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
        const reg = /^[a-zA-Z0-9\u4e00-\u9fa5]{1}[a-zA-Z0-9_.\u4e00-\u9fa5-]{0,30}$/;
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
        const reg = /[a-zA-Z0-9_\u4e00-\u9fa5]{3,30}$/;
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
        const reg = /^[a-zA-Z0-9\u4e00-\u9fa5]{1}[a-zA-Z0-9_.\u4e00-\u9fa5-]{0,16}$/;
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
        const reg = /^[a-zA-Z0-9]{1}[a-z:A-Z0-9_/.-]{0,150}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.imageName')));
        } else {
            callback();
        }
    }
};

const checkVolumeName = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.volumeName')));
    } else {
        const reg = /^[a-zA-Z0-9]{1}[a-z:A-Z0-9_.-]{1,30}$/;
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
        const reg = /^[^/\\\"'|<>?*]{1,30}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.linuxName', ['/\\:*?\'"<>|'])));
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
        const reg = /^(?![_-])[a-zA-Z0-9_-]{1,29}[a-zA-Z0-9]$/;
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

const checkParamCommon = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.paramName')));
    } else {
        const reg = /^[a-zA-Z0-9]{1}[a-zA-Z0-9._-]{1,29}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.paramName')));
        } else {
            callback();
        }
    }
};

const checkParamComplexity = (rule: any, value: any, callback: any) => {
    if (value === '' || typeof value === 'undefined' || value == null) {
        callback(new Error(i18n.global.t('commons.rule.paramComplexity', ['.%@$!&~_-'])));
    } else {
        const reg = /^[a-zA-Z0-9]{1}[a-zA-Z0-9.%@$!&~_-]{5,29}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.paramComplexity', ['.%@$!&~_-'])));
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
        const reg = /^[A-Za-z0-9\n.]+$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.nginxDoc')));
        } else {
            callback();
        }
    }
};

export function checkNumberRange(min: number, max: number): FormItemRule {
    return {
        required: true,
        trigger: 'blur',
        min: min,
        max: max,
        type: 'number',
        message: i18n.global.t('commons.rule.numberRange', [min, max]),
    };
}

interface CommonRule {
    requiredInput: FormItemRule;
    requiredSelect: FormItemRule;
    requiredSelectBusiness: FormItemRule;
    name: FormItemRule;
    userName: FormItemRule;
    simpleName: FormItemRule;
    dbName: FormItemRule;
    imageName: FormItemRule;
    volumeName: FormItemRule;
    linuxName: FormItemRule;
    password: FormItemRule;
    email: FormItemRule;
    number: FormItemRule;
    integerNumber: FormItemRule;
    ip: FormItemRule;
    host: FormItemRule;
    port: FormItemRule;
    domain: FormItemRule;
    databaseName: FormItemRule;
    nginxDoc: FormItemRule;
    appName: FormItemRule;

    paramCommon: FormItemRule;
    paramComplexity: FormItemRule;
    paramPort: FormItemRule;
    paramExtUrl: FormItemRule;
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
    ip: {
        validator: checkIp,
        required: true,
        trigger: 'blur',
    },
    host: {
        validator: checkHost,
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
};
